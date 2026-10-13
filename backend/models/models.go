package models

import "time"

// AppItem represents a WinGet package (installed, available update, or search result)
type AppItem struct {
	Name      string `json:"name"`
	ID        string `json:"id"`
	Version   string `json:"version"`
	Available string `json:"available,omitempty"` // For updates
	Source    string `json:"source,omitempty"`
}

// AppDetails represents the detailed manifest of a package
type AppDetails struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Version         string   `json:"version"`
	Publisher       string   `json:"publisher"`
	Author          string   `json:"author"`
	Moniker         string   `json:"moniker"`
	Description     string   `json:"description"`
	License         string   `json:"license"`
	Copyright       string   `json:"copyright"`
	Homepage        string   `json:"homepage"`
	ReleaseNotesURL string   `json:"releaseNotesUrl"`
	Tags            []string `json:"tags"`
	InstallerType   string   `json:"installerType"`
	InstallerURL    string   `json:"installerUrl"`
	InstallerSha256 string   `json:"installerSha256"`
}

// JobStatus represents a background execution task (install, update, uninstall, etc.)
type JobStatus struct {
	ID          string    `json:"id"`
	Command     string    `json:"command"`
	Args        []string  `json:"args"`
	Done        bool      `json:"done"`
	ExitCode    int       `json:"exitCode"`
	Output      string    `json:"output"`
	StartTime   time.Time `json:"startTime"`
	CompletedAt int64     `json:"completedAt,omitempty"`
}

// JobLogEvent is emitted in real-time when new output arrives
type JobLogEvent struct {
	JobID string `json:"jobId"`
	Chunk string `json:"chunk"`
}

// JobDoneEvent is emitted when a job completes
type JobDoneEvent struct {
	JobID    string `json:"jobId"`
	ExitCode int    `json:"exitCode"`
	Success  bool   `json:"success"`
}

// FileItem represents a downloaded installer file
type FileItem struct {
	Name          string `json:"name"`
	Length        int64  `json:"length"`
	LastWriteTime string `json:"lastWriteTime"`
}

// IgnoredItem represents an app whose updates are ignored
type IgnoredItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ManifestIcon represents icon scraping result
type ManifestIcon struct {
	Success bool   `json:"success"`
	Domain  string `json:"domain"`
	URL     string `json:"url"`
	IconURL string `json:"iconUrl"`
}

// AppUpdateInfo represents EasyWinGet application release update info
type AppUpdateInfo struct {
	HasUpdate      bool   `json:"hasUpdate"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseName    string `json:"releaseName"`
	ReleaseNotes   string `json:"releaseNotes"`
	ReleaseURL     string `json:"releaseUrl"`
	DownloadURL    string `json:"downloadUrl"`
	PublishedAt    string `json:"publishedAt"`
}
