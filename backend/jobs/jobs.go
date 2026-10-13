package jobs

import (
	"bufio"
	"context"
	"easywinget/backend/models"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	onJobDone  func(models.JobDoneEvent)
}

// SetOnJobDone registers a callback invoked whenever any job completes
func (m *Manager) SetOnJobDone(cb func(models.JobDoneEvent)) {
	m.mu.Lock()
	m.onJobDone = cb
	m.mu.Unlock()
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

	// Environment: Ensure UTF-8 and modern terminal support so winget & tools output full animations and colors
	cmd.Env = append(os.Environ(),
		"PYTHONIOENCODING=utf-8",
		"WT_SESSION=1",
		"TERM=xterm-256color",
	)

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

	if isWinget {
		go m.monitorWingetProgress(job)
	}

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

		event := models.JobDoneEvent{
			JobID:    jobID,
			ExitCode: exitCode,
			Success:  exitCode == 0,
		}

		if m.wailsCtx != nil {
			wailsRuntime.EventsEmit(m.wailsCtx, "job:done", event)
		}

		m.mu.RLock()
		cb := m.onJobDone
		m.mu.RUnlock()
		if cb != nil {
			go cb(event)
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
	cmd.Env = append(os.Environ(),
		"PYTHONIOENCODING=utf-8",
		"WT_SESSION=1",
		"TERM=xterm-256color",
	)
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

		event := models.JobDoneEvent{
			JobID:    jobID,
			ExitCode: exitCode,
			Success:  exitCode == 0,
		}

		if m.wailsCtx != nil {
			wailsRuntime.EventsEmit(m.wailsCtx, "job:done", event)
		}

		m.mu.RLock()
		cb := m.onJobDone
		m.mu.RUnlock()
		if cb != nil {
			go cb(event)
		}
	}()

	return jobID, nil
}

func (m *Manager) streamReader(j *Job, r io.Reader) {
	buf := make([]byte, 2048)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			rawChunk := string(buf[:n])
			j.mu.Lock()
			j.Output.WriteString(rawChunk)
			j.mu.Unlock()

			if m.wailsCtx != nil {
				wailsRuntime.EventsEmit(m.wailsCtx, "job:log", models.JobLogEvent{
					JobID: j.ID,
					Chunk: rawChunk,
				})
			}
		}
		if err != nil {
			break
		}
	}
}

// monitorWingetProgress watches the winget diagnostic output log for real-time Delivery Optimization progress
func (m *Manager) monitorWingetProgress(j *Job) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return
	}
	diagDir := filepath.Join(localAppData, "Packages", "Microsoft.DesktopAppInstaller_8wekyb3d8bbwe", "LocalState", "DiagOutputDir")
	if _, err := os.Stat(diagDir); os.IsNotExist(err) {
		return
	}

	searchThreshold := j.StartTime.Add(-2 * time.Second)
	var logFile string

	// Poll up to 4 seconds (40 x 100ms) to locate the WinGet log file created for this job
	for i := 0; i < 40; i++ {
		j.mu.RLock()
		done := j.Done
		j.mu.RUnlock()
		if done {
			return
		}

		entries, err := os.ReadDir(diagDir)
		if err == nil {
			var newestPath string
			var newestTime time.Time
			for _, e := range entries {
				name := e.Name()
				if strings.HasSuffix(name, ".log") && strings.HasPrefix(name, "WinGet-") {
					info, err := e.Info()
					if err == nil && info.ModTime().After(searchThreshold) {
						if info.ModTime().After(newestTime) {
							newestTime = info.ModTime()
							newestPath = filepath.Join(diagDir, name)
						}
					}
				}
			}
			if newestPath != "" {
				logFile = newestPath
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	if logFile == "" {
		return
	}

	file, err := os.Open(logFile)
	if err != nil {
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	reDO := regexp.MustCompile(`DO State \d+,\s*(\d+)\s*/\s*(\d+)`)

	var lastPct float64 = -1
	var lastCurBytes int64 = 0
	completedDownload := false

	for {
		j.mu.RLock()
		done := j.Done
		j.mu.RUnlock()
		if done {
			break
		}

		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			match := reDO.FindStringSubmatch(line)
			if len(match) == 3 {
				cur, _ := strconv.ParseInt(match[1], 10, 64)
				tot, _ := strconv.ParseInt(match[2], 10, 64)
				if tot > 0 {
					// Handle new download in multi-package jobs
					if cur < lastCurBytes {
						completedDownload = false
						lastPct = -1
					}
					lastCurBytes = cur

					pct := float64(cur) * 100.0 / float64(tot)
					if pct > 100.0 {
						pct = 100.0
					}

					// Update when percentage increments or reaches 100%
					if pct != lastPct && (pct-lastPct >= 0.2 || pct == 100.0 || lastPct < 0) {
						lastPct = pct
						curMB := float64(cur) / (1024.0 * 1024.0)
						totMB := float64(tot) / (1024.0 * 1024.0)

						barWidth := 25
						filled := int((pct / 100.0) * float64(barWidth))
						if filled > barWidth {
							filled = barWidth
						}
						empty := barWidth - filled
						bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

						var chunk string
						if pct >= 100.0 && !completedDownload {
							completedDownload = true
							chunk = fmt.Sprintf("\r  [Download] [%s] 100.0%% (%.1f MB / %.1f MB) \x1b[32m✔ Verified\x1b[0m\n", bar, totMB, totMB)
						} else if !completedDownload {
							chunk = fmt.Sprintf("\r  [Download] [%s] %5.1f%% (%.1f MB / %.1f MB)", bar, pct, curMB, totMB)
						}

						if chunk != "" {
							j.mu.Lock()
							j.Output.WriteString(chunk)
							j.mu.Unlock()

							if m.wailsCtx != nil {
								wailsRuntime.EventsEmit(m.wailsCtx, "job:log", models.JobLogEvent{
									JobID: j.ID,
									Chunk: chunk,
								})
							}
						}
					}
				}
			}
		}

		if err != nil {
			time.Sleep(80 * time.Millisecond)
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

	event := models.JobDoneEvent{
		JobID:    jobID,
		ExitCode: -1,
		Success:  false,
	}

	if m.wailsCtx != nil {
		wailsRuntime.EventsEmit(m.wailsCtx, "job:done", event)
	}

	m.mu.RLock()
	cb := m.onJobDone
	m.mu.RUnlock()
	if cb != nil {
		go cb(event)
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

