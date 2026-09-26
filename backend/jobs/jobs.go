package jobs

import (
	"bufio"
	"context"
	"easywinget/backend/models"
	"easywinget/backend/winget"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Job represents a single background execution instance
type Job struct {
	ID          string
	Command     string
	Args        []string
	Cmd         *exec.Cmd
	Stdin       io.WriteCloser
	Done        bool
	ExitCode    int
	StartTime   time.Time
	CompletedAt int64
	Output      strings.Builder
	mu          sync.RWMutex
}

// Manager manages running jobs and streams terminal output to the frontend
type Manager struct {
	mu         sync.RWMutex
	jobs       map[string]*Job
	wailsCtx   context.Context
	jobsLogDir string
}

// NewManager creates a new Job Manager
func NewManager(jobsLogDir string) (*Manager, error) {
	if err := os.MkdirAll(jobsLogDir, 0755); err != nil {
		return nil, err
	}
	m := &Manager{
		jobs:       make(map[string]*Job),
		jobsLogDir: jobsLogDir,
	}
	go m.cleanupRoutine()
	return m, nil
}

// SetContext sets the Wails runtime context for event emission
func (m *Manager) SetContext(ctx context.Context) {
	m.wailsCtx = ctx
}

func (m *Manager) cleanupRoutine() {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now().UnixMilli()
		// Retain completed jobs for 15 minutes
		const retentionMs = 15 * 60 * 1000
		for id, j := range m.jobs {
			j.mu.RLock()
			done := j.Done
			completedAt := j.CompletedAt
			j.mu.RUnlock()

			if done && completedAt > 0 && (now-completedAt) > retentionMs {
				delete(m.jobs, id)
			}
		}
		m.mu.Unlock()
	}
}

// StartJob spawns a process and begins streaming its output in real-time
func (m *Manager) StartJob(command string, args []string) (string, error) {
	jobID := fmt.Sprintf("job-%d", time.Now().UnixNano())

	job := &Job{
		ID:        jobID,
		Command:   command,
		Args:      args,
		StartTime: time.Now(),
	}

	var cmd *exec.Cmd
	isPS := strings.EqualFold(filepath.Base(command), "powershell.exe") || strings.EqualFold(command, "powershell")
	isWinget := strings.EqualFold(filepath.Base(command), "winget.exe") || strings.EqualFold(command, "winget")

	if isPS {
		cmd = exec.Command("powershell.exe", args...)
	} else if isWinget {
		cmd = exec.Command("winget", args...)
	} else {
		cmd = exec.Command(command, args...)
	}

	// CREATE_NO_WINDOW (0x08000000) prevents console popup while allowing GUI installers to display
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000,
	}

	stdinPipe, _ := cmd.StdinPipe()
	job.Stdin = stdinPipe

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to open stdout: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to open stderr: %w", err)
	}

	job.Cmd = cmd

	if err := cmd.Start(); err != nil {
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "elevation") || strings.Contains(lowerErr, "740") {
			// Automatically elevate via PowerShell Start-Process -Verb RunAs
			return m.StartElevatedJob(command, args)
		}
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	m.mu.Lock()
	m.jobs[jobID] = job
	m.mu.Unlock()

	// Handle combined streaming reader
	go m.streamReader(job, stdoutPipe)
	go m.streamReader(job, stderrPipe)

	// Wait for process in background
	go func() {
		err := cmd.Wait()
		exitCode := 0
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			} else {
				exitCode = 1
			}
		}

		job.mu.Lock()
		job.Done = true
		job.ExitCode = exitCode
		job.CompletedAt = time.Now().UnixMilli()
		job.mu.Unlock()

		if m.wailsCtx != nil {
			wailsRuntime.EventsEmit(m.wailsCtx, "job:done", models.JobDoneEvent{
				JobID:    jobID,
				ExitCode: exitCode,
				Success:  exitCode == 0,
			})
		}
	}()

	return jobID, nil
}

// StartElevatedJob launches an executable with Windows Administrator privileges (triggers UAC)
func (m *Manager) StartElevatedJob(command string, args []string) (string, error) {
	jobID := fmt.Sprintf("job-elevated-%d", time.Now().UnixNano())

	job := &Job{
		ID:        jobID,
		Command:   command,
		Args:      args,
		StartTime: time.Now(),
	}

	// Prepare PowerShell Start-Process with -Verb RunAs (clean argument handling)
	var psScript string
	escapedCmd := strings.ReplaceAll(command, `'`, `''`)

	if len(args) > 0 {
		var escapedArgs []string
		for _, a := range args {
			a = strings.TrimSpace(a)
			if a != "" {
				escapedArgs = append(escapedArgs, fmt.Sprintf(`'%s'`, strings.ReplaceAll(a, `'`, `''`)))
			}
		}
		if len(escapedArgs) > 0 {
			psScript = fmt.Sprintf(`Start-Process -FilePath '%s' -ArgumentList @(%s) -Verb RunAs -Wait`,
				escapedCmd, strings.Join(escapedArgs, ", "))
		} else {
			psScript = fmt.Sprintf(`Start-Process -FilePath '%s' -Verb RunAs -Wait`, escapedCmd)
		}
	} else {
		psScript = fmt.Sprintf(`Start-Process -FilePath '%s' -Verb RunAs -Wait`, escapedCmd)
	}

	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000,
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to open stdout: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to open stderr: %w", err)
	}

	job.Cmd = cmd

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start elevated command: %w", err)
	}

	m.mu.Lock()
	m.jobs[jobID] = job
	m.mu.Unlock()

	go m.streamReader(job, stdoutPipe)
	go m.streamReader(job, stderrPipe)

	go func() {
		err := cmd.Wait()
		exitCode := 0
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			} else {
				exitCode = 1
			}
		}

		job.mu.Lock()
		job.Done = true
		job.ExitCode = exitCode
		job.CompletedAt = time.Now().UnixMilli()
		job.mu.Unlock()

		if m.wailsCtx != nil {
			wailsRuntime.EventsEmit(m.wailsCtx, "job:done", models.JobDoneEvent{
				JobID:    jobID,
				ExitCode: exitCode,
				Success:  exitCode == 0,
			})
		}
	}()

	return jobID, nil
}

func (m *Manager) streamReader(j *Job, r io.Reader) {
	reader := bufio.NewReader(r)
	buf := make([]byte, 1024)

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			chunk := string(buf[:n])
			cleanChunk := strings.Join(winget.CleanOutput(chunk), "\n")
			if cleanChunk != "" {
				cleanChunk += "\n"
				j.mu.Lock()
				j.Output.WriteString(cleanChunk)
				j.mu.Unlock()

				if m.wailsCtx != nil {
					wailsRuntime.EventsEmit(m.wailsCtx, "job:log", models.JobLogEvent{
						JobID: j.ID,
						Chunk: cleanChunk,
					})
				}
			}
		}
		if err != nil {
			break
		}
	}
}

// CancelJob terminates the process tree for the given job
func (m *Manager) CancelJob(jobID string) bool {
	m.mu.RLock()
	job, exists := m.jobs[jobID]
	m.mu.RUnlock()

	if !exists || job.Cmd == nil || job.Cmd.Process == nil {
		return false
	}

	job.mu.RLock()
	if job.Done {
		job.mu.RUnlock()
		return false
	}
	pid := job.Cmd.Process.Pid
	job.mu.RUnlock()

	// Windows taskkill to terminate entire process tree
	killCmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
	killCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = killCmd.Run()

	job.mu.Lock()
	job.Done = true
	job.ExitCode = -1
	job.CompletedAt = time.Now().UnixMilli()
	job.Output.WriteString("\n[Process cancelled by user]\n")
	job.mu.Unlock()

	if m.wailsCtx != nil {
		wailsRuntime.EventsEmit(m.wailsCtx, "job:done", models.JobDoneEvent{
			JobID:    jobID,
			ExitCode: -1,
			Success:  false,
		})
	}
	return true
}

// GetJob returns the current status and output of a job
func (m *Manager) GetJob(jobID string) (*models.JobStatus, bool) {
	m.mu.RLock()
	job, exists := m.jobs[jobID]
	m.mu.RUnlock()

	if !exists {
		return nil, false
	}

	job.mu.RLock()
	defer job.mu.RUnlock()

	return &models.JobStatus{
		ID:          job.ID,
		Command:     job.Command,
		Args:        job.Args,
		Done:        job.Done,
		ExitCode:    job.ExitCode,
		Output:      job.Output.String(),
		StartTime:   job.StartTime,
		CompletedAt: job.CompletedAt,
	}, true
}

// SendInput writes interactive user input to the running job's stdin
func (m *Manager) SendInput(jobID string, input string) error {
	m.mu.RLock()
	job, exists := m.jobs[jobID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.mu.RLock()
	done := job.Done
	stdin := job.Stdin
	job.mu.RUnlock()

	if done || stdin == nil {
		return fmt.Errorf("job is not active or stdin closed")
	}

	_, err := stdin.Write([]byte(input + "\r\n"))
	return err
}

