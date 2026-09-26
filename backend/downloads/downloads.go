package downloads

import (
	"easywinget/backend/models"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

var installerExts = map[string]bool{
	".exe":  true,
	".msi":  true,
	".zip":  true,
	".7z":   true,
	".rar":  true,
	".iso":  true,
	".msix": true,
	".appx": true,
}

// Manager handles downloaded package installers
type Manager struct {
	downloadDir string
}

// NewManager creates a downloads manager
func NewManager(downloadDir string) (*Manager, error) {
	absDir, err := filepath.Abs(downloadDir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absDir, 0755); err != nil {
		return nil, err
	}
	return &Manager{downloadDir: absDir}, nil
}

// GetDownloadDir returns the absolute download directory
func (m *Manager) GetDownloadDir() string {
	return m.downloadDir
}

func (m *Manager) isPathSafe(relPath string) (string, bool) {
	cleanRel := filepath.Clean(relPath)
	if strings.HasPrefix(cleanRel, "..") || filepath.IsAbs(cleanRel) {
		return "", false
	}
	fullPath := filepath.Join(m.downloadDir, cleanRel)
	if !strings.HasPrefix(fullPath, m.downloadDir) {
		return "", false
	}
	return fullPath, true
}

// GetFiles recursively scans download directory for installer packages
func (m *Manager) GetFiles() ([]models.FileItem, error) {
	var files []models.FileItem

	err := filepath.Walk(m.downloadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if strings.EqualFold(info.Name(), "dependencies") {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if installerExts[ext] {
			rel, relErr := filepath.Rel(m.downloadDir, path)
			if relErr == nil {
				files = append(files, models.FileItem{
					Name:          filepath.ToSlash(rel),
					Length:        info.Size(),
					LastWriteTime: info.ModTime().Format(time.RFC3339),
				})
			}
		}
		return nil
	})

	return files, err
}

// DeleteFile deletes a downloaded installer or its parent directory
func (m *Manager) DeleteFile(relPath string) error {
	fullPath, safe := m.isPathSafe(relPath)
	if !safe {
		return fmt.Errorf("invalid file path")
	}

	parts := strings.Split(filepath.ToSlash(relPath), "/")
	if len(parts) > 1 {
		// Subfolder deletion
		subfolder := filepath.Join(m.downloadDir, parts[0])
		return os.RemoveAll(subfolder)
	}

	return os.Remove(fullPath)
}

// openFolderNative opens a directory path natively using Windows ShellExecute without shell quote issues
func openFolderNative(dirPath string) error {
	dirUTF16 := windows.StringToUTF16Ptr(dirPath)
	verbUTF16 := windows.StringToUTF16Ptr("open")
	err := windows.ShellExecute(0, verbUTF16, dirUTF16, nil, nil, windows.SW_SHOWNORMAL)
	if err != nil {
		// Fallback to explorer.exe with raw CmdLine to prevent space escaping bugs
		cmd := exec.Command("explorer.exe")
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CmdLine:       fmt.Sprintf(`explorer.exe "%s"`, dirPath),
			CreationFlags: 0x08000000,
		}
		return cmd.Start()
	}
	return nil
}

// OpenFolder opens Windows Explorer at the installer file location or the base download directory
func (m *Manager) OpenFolder(relPath string) error {
	relPath = strings.TrimSpace(relPath)
	_ = os.MkdirAll(m.downloadDir, 0755)

	if relPath == "" {
		return openFolderNative(m.downloadDir)
	}

	fullPath, safe := m.isPathSafe(relPath)
	if !safe {
		fullPath = filepath.Join(m.downloadDir, relPath)
	}

	cleanPath := filepath.Clean(fullPath)
	if info, err := os.Stat(cleanPath); err == nil {
		if info.IsDir() {
			return openFolderNative(cleanPath)
		}
		// If it's an installer file, reveal it with /select,"<cleanPath>"
		// Note: Explorer expects /select,"path with spaces" WITHOUT space after comma and without quotes around /select.
		cmd := exec.Command("explorer.exe")
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CmdLine:       fmt.Sprintf(`explorer.exe /select,"%s"`, cleanPath),
			CreationFlags: 0x08000000,
		}
		if startErr := cmd.Start(); startErr == nil {
			return nil
		}
		// Fallback to opening parent directory
		return openFolderNative(filepath.Dir(cleanPath))
	}

	// If cleanPath does not exist directly, check its parent directory
	parentDir := filepath.Dir(cleanPath)
	if info, err := os.Stat(parentDir); err == nil && info.IsDir() {
		return openFolderNative(parentDir)
	}

	return openFolderNative(m.downloadDir)
}
