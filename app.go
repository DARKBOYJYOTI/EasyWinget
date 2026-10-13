package main

import (
	"context"
	"easywinget/backend/db"
	"easywinget/backend/downloads"
	"easywinget/backend/icons"
	"easywinget/backend/installed"
	"easywinget/backend/jobs"
	"easywinget/backend/models"
	"easywinget/backend/system"
	"easywinget/backend/winget"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// App struct represents the core Wails application
type App struct {
	ctx               context.Context
	dbStorage         *db.Storage
	jobManager        *jobs.Manager
	iconManager       *icons.Manager
	downManager       *downloads.Manager
	dataDir           string
	version           string
	initialTab        string
	updateTriggerChan chan struct{}
}

// NewApp creates a new App application struct
func NewApp() *App {
	dataDir := system.GetEasyWinGetDataDir()
	jobsLogDir := system.GetEasyWinGetJobsDir()
	downloadDir := system.GetEasyWinGetDownloadDir()
	iconsDir := filepath.Join(dataDir, "icons")

	storage, err := db.NewStorage(dataDir)
	if err != nil {
		fmt.Printf("[DB Init Warning] %v\n", err)
	}

	jm, _ := jobs.NewManager(jobsLogDir)
	im, _ := icons.NewManager(iconsDir)
	dm, _ := downloads.NewManager(downloadDir)

	return &App{
		dbStorage:         storage,
		jobManager:        jm,
		iconManager:       im,
		downManager:       dm,
		dataDir:           dataDir,
		version:           "4.0.0",
		initialTab:        "",
		updateTriggerChan: make(chan struct{}, 1),
	}
}

// SetInitialTab sets the tab to open on startup (e.g. from protocol activation)
func (a *App) SetInitialTab(tab string) {
	a.initialTab = tab
}

// GetInitialTab returns the initial tab if set, e.g. "updates"
func (a *App) GetInitialTab() string {
	return a.initialTab
}

// NavigateToTab emits an event to change active tab in frontend
func (a *App) NavigateToTab(tab string) {
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "navigate:tab", tab)
	}
}

// ApplyAutoUpdateSettings signals the background update checker to restart its timer with new interval/state
func (a *App) ApplyAutoUpdateSettings() {
	select {
	case a.updateTriggerChan <- struct{}{}:
	default:
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if a.jobManager != nil {
		a.jobManager.SetContext(ctx)
		a.jobManager.SetOnJobDone(func(e models.JobDoneEvent) {
			if e.Success {
				go a.refreshAllInBackground()
			}
		})
	}

	// Register easywinget:// URL protocol handler in HKCU
	go system.RegisterProtocolHandler()

	// Default startup settings: autostart, minimize to tray, and auto update check default to enabled
	if a.GetSetting("easywinget_autostart_init", "") == "" {
		_ = system.SetAutoStart(true)
		_ = a.SetSetting("easywinget_autostart_init", "done")
		_ = a.SetSetting("easywinget_minimize_to_tray", "true")
		_ = a.SetSetting("easywinget_auto_update_check", "true")
		_ = a.SetSetting("easywinget_update_check_interval", "1440") // 24 hours daily default
	}

	// Realtime listener: when any winget job finishes, sync SQLite and notify frontend
	wailsRuntime.EventsOn(ctx, "job:done", func(optionalData ...interface{}) {
		go a.refreshAllInBackground()
	})

	// If initial tab was requested (e.g. from protocol link), navigate after frontend mounts
	if a.initialTab != "" {
		go func() {
			time.Sleep(800 * time.Millisecond)
			a.NavigateToTab(a.initialTab)
		}()
	}

	// Startup Internet & Update Checker (runs once on startup)
	go a.startNetworkAndStartupUpdateChecker()

	// Periodic Background Update Checker (runs according to user settings)
	go a.startPeriodicUpdateChecker()
}

// RestoreWindow brings the existing EasyWinGet window to foreground and unminimizes it
func (a *App) RestoreWindow() {
	if a.ctx == nil {
		return
	}
	wailsRuntime.WindowShow(a.ctx)
	if wailsRuntime.WindowIsMinimised(a.ctx) {
		wailsRuntime.WindowUnminimise(a.ctx)
	}
	wailsRuntime.WindowSetAlwaysOnTop(a.ctx, true)
	wailsRuntime.WindowSetAlwaysOnTop(a.ctx, false)
	wailsRuntime.EventsEmit(a.ctx, "app:restored")
}

// startNetworkAndStartupUpdateChecker waits for internet connection and checks for updates on startup
func (a *App) startNetworkAndStartupUpdateChecker() {
	// Wait a moment for UI to mount
	time.Sleep(2 * time.Second)

	// Poll until internet connection is confirmed (max 30 tries, 5s interval)
	hasNet := false
	for i := 0; i < 30; i++ {
		if system.CheckInternetConnection() {
			hasNet = true
			break
		}
		time.Sleep(5 * time.Second)
	}

	if !hasNet {
		return
	}

	// Internet is active: scan for updates
	updates, err := winget.ListUpdates()
	if err != nil {
		return
	}

	if a.dbStorage != nil {
		_ = a.dbStorage.SaveAvailableUpdates(updates)
	}

	// Emit updates to UI
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "updates:updated", updates)
	}

	// If updates are available, send Windows system notification & in-app toast
	if len(updates) > 0 {
		title := "EasyWinGet Updates"
		msg := fmt.Sprintf("%d software update(s) are available for your system.", len(updates))
		system.ShowWindowsNotificationWithAction(title, msg, "easywinget://updates", "View Updates")

		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "toast:notify", map[string]string{
				"type":      "info",
				"title":     title,
				"message":   msg,
				"actionTab": "updates",
			})
		}
	}
}

// startPeriodicUpdateChecker checks for updates at user-configured intervals when internet is active
func (a *App) startPeriodicUpdateChecker() {
	// Initial wait to let the startup scan complete cleanly
	time.Sleep(25 * time.Second)

	for {
		enabled := a.GetSetting("easywinget_auto_update_check", "true") != "false"
		intervalStr := a.GetSetting("easywinget_update_check_interval", "1440")
		intervalMins, err := strconv.Atoi(intervalStr)
		if err != nil || intervalMins <= 0 {
			intervalMins = 1440 // default 24 hours
		}

		duration := time.Duration(intervalMins) * time.Minute
		timer := time.NewTimer(duration)

		select {
		case <-timer.C:
			if !enabled {
				continue
			}
			// Only check if internet is active
			if !system.CheckInternetConnection() {
				continue
			}

			updates, err := winget.ListUpdates()
			if err != nil {
				continue
			}

			if a.dbStorage != nil {
				_ = a.dbStorage.SaveAvailableUpdates(updates)
			}
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, "updates:updated", updates)
			}

			if len(updates) > 0 {
				title := "EasyWinGet Updates"
				msg := fmt.Sprintf("%d software update(s) are available for your system.", len(updates))
				system.ShowWindowsNotificationWithAction(title, msg, "easywinget://updates", "View Updates")

				if a.ctx != nil {
					wailsRuntime.EventsEmit(a.ctx, "toast:notify", map[string]string{
						"type":      "info",
						"title":     title,
						"message":   msg,
						"actionTab": "updates",
					})
				}
			}

		case <-a.updateTriggerChan:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			// Settings changed: loop re-runs immediately with updated interval/enablement
		}
	}
}

// RefreshAll triggers a fresh concurrent rescan of registry, winget installed, and winget updates
func (a *App) RefreshAll() {
	go a.refreshAllInBackground()
}

// refreshAllInBackground silently refreshes installed apps and available updates
func (a *App) refreshAllInBackground() {
	// 1. Rescan native Windows registry cache (BCUninstaller style)
	if scanner := installed.GetScanner(); scanner != nil {
		scanner.Refresh()
	}

	// 2. Scan installed apps and available updates concurrently for maximum speed
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if apps, err := winget.ListInstalled(); err == nil && len(apps) > 0 {
			if a.dbStorage != nil {
				_ = a.dbStorage.SaveInstalledApps(apps)
			}
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, "installed:updated", apps)
			}
		}
	}()

	go func() {
		defer wg.Done()
		if updates, err := winget.ListUpdates(); err == nil {
			if a.dbStorage != nil {
				_ = a.dbStorage.SaveAvailableUpdates(updates)
			}
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, "updates:updated", updates)
			}
		}
	}()

	wg.Wait()
}

// GetAppVersion returns the current EasyWinGet version
func (a *App) GetAppVersion() string {
	return a.version
}

// GetInstalled returns installed applications instantly from SQLite if cached
func (a *App) GetInstalled(refresh bool) ([]models.AppItem, error) {
	if a.dbStorage != nil && !refresh {
		cached, err := a.dbStorage.GetInstalledApps()
		if err == nil && len(cached) > 0 {
			return cached, nil
		}
	}

	// If refresh requested or database is empty, scan winget
	apps, err := winget.ListInstalled()
	if err != nil {
		// Fallback to SQLite cache if available
		if a.dbStorage != nil {
			if cached, dbErr := a.dbStorage.GetInstalledApps(); dbErr == nil && len(cached) > 0 {
				return cached, nil
			}
		}
		return nil, err
	}

	if a.dbStorage != nil && len(apps) > 0 {
		_ = a.dbStorage.SaveInstalledApps(apps)
	}
	return apps, nil
}

// GetUpdates returns packages with available updates
func (a *App) GetUpdates(refresh bool) ([]models.AppItem, error) {
	if a.dbStorage != nil && !refresh {
		cached, err := a.dbStorage.GetAvailableUpdates()
		if err == nil && len(cached) > 0 {
			return cached, nil
		}
	}

	updates, err := winget.ListUpdates()
	if err != nil {
		if a.dbStorage != nil {
			if cached, dbErr := a.dbStorage.GetAvailableUpdates(); dbErr == nil {
				return cached, nil
			}
		}
		return nil, err
	}

	if a.dbStorage != nil {
		_ = a.dbStorage.SaveAvailableUpdates(updates)
		return a.dbStorage.GetAvailableUpdates()
	}

	return updates, nil
}

// SearchApps searches for packages via WinGet CLI
func (a *App) SearchApps(query string) ([]models.AppItem, error) {
	trimmed := strings.TrimSpace(query)
	if len(trimmed) < 2 {
		return []models.AppItem{}, nil
	}
	return winget.Search(trimmed)
}

// GetAppDetails retrieves full manifest metadata for an application
func (a *App) GetAppDetails(id string) (*models.AppDetails, error) {
	return winget.Show(id)
}

// GetManifestIcon fetches favicon / app icon from package metadata or domain
func (a *App) GetManifestIcon(id string, homepage string) models.ManifestIcon {
	if a.iconManager == nil {
		return models.ManifestIcon{}
	}
	return a.iconManager.ResolveManifestIcon(id, homepage)
}

// GetLocalAppIcon gets Windows icon for an installed application
func (a *App) GetLocalAppIcon(id string, name string) string {
	if a.iconManager == nil {
		return ""
	}
	return a.iconManager.GetLocalAppIcon(id, name)
}

// InstallPackage starts a background winget install job
func (a *App) InstallPackage(id string, interactive bool, force bool, scope string) (string, error) {
	args := []string{
		"install",
		"--id", id,
		"--accept-source-agreements",
		"--accept-package-agreements",
		"--verbose",
	}

	if interactive {
		args = append(args, "--interactive")
	}

	if force {
		args = append(args, "--force")
	}

	if scope == "user" {
		args = append(args, "--scope", "user")
	} else if scope == "machine" {
		args = append(args, "--scope", "machine")
	}

	return a.jobManager.StartJob("winget", args)
}

// UpgradePackage starts a background winget upgrade job for a specific package
func (a *App) UpgradePackage(id string) (string, error) {
	args := []string{
		"upgrade",
		"--id", id,
		"--accept-source-agreements",
		"--accept-package-agreements",
		"--verbose",
	}

	return a.jobManager.StartJob("winget", args)
}

// UpgradeAll upgrades all outdated packages without --include-unknown
func (a *App) UpgradeAll() (string, error) {
	args := []string{
		"upgrade",
		"--all",
		"--accept-source-agreements",
		"--accept-package-agreements",
		"--verbose",
	}

	return a.jobManager.StartJob("winget", args)
}

// UninstallPackage uninstalls an application smartly (routes to Windows native uninstaller or WinGet)
func (a *App) UninstallPackage(id string, name string, forceNative bool) (string, error) {
	isArpOrMsix := strings.HasPrefix(id, "ARP\\") || strings.HasPrefix(id, "MSIX\\") || strings.HasPrefix(id, "User\\") || strings.HasPrefix(id, "Machine\\") || !strings.Contains(id, ".")

	// 1. If it's an ARP/Registry entry or forceNative was requested:
	if isArpOrMsix || forceNative {
		searchName := name
		if searchName == "" {
			parts := strings.Split(id, ".")
			searchName = parts[len(parts)-1]
		}
		uninstCmd := installed.GetScanner().GetUninstallString(searchName, id)
		if uninstCmd != "" {
			exe, args := installed.SplitCommandLine(uninstCmd)
			if exe != "" {
				return a.jobManager.StartJob(exe, args)
			}
		}
		if isArpOrMsix && forceNative {
			return "", fmt.Errorf("native uninstaller not found for: %s", name)
		}
	}

	// 2. Try WinGet CLI if it looks like a valid WinGet package ID
	if strings.Contains(id, ".") || len(id) == 12 {
		args := []string{
			"uninstall",
			"--id", id,
			"--accept-source-agreements",
			"--verbose",
		}
		return a.jobManager.StartJob("winget", args)
	}

	// 3. Fallback: try native uninstaller for the name
	uninstCmd := installed.GetScanner().GetUninstallString(name, id)
	if uninstCmd != "" {
		exe, args := installed.SplitCommandLine(uninstCmd)
		if exe != "" {
			return a.jobManager.StartJob(exe, args)
		}
	}

	// 4. Final attempt with winget by name
	args := []string{
		"uninstall",
		"--name", name,
		"--accept-source-agreements",
		"--verbose",
	}
	return a.jobManager.StartJob("winget", args)
}

// DownloadPackage downloads installer directly into %USERPROFILE%\Downloads\EasyWinGet\<SoftwareName>
func (a *App) DownloadPackage(id string, name string) (string, error) {
	targetDir := system.GetEasyWinGetDownloadDir()
	if name != "" {
		safeName := system.SanitizeFolderName(name)
		if safeName != "" {
			targetDir = filepath.Join(targetDir, safeName)
		}
	}
	_ = os.MkdirAll(targetDir, 0755)

	args := []string{
		"download",
		"--id", id,
		"--download-directory", targetDir,
		"--accept-source-agreements",
		"--verbose",
	}
	return a.jobManager.StartJob("winget", args)
}

// CancelJob cancels an active job
func (a *App) CancelJob(jobID string) bool {
	if a.jobManager == nil {
		return false
	}
	return a.jobManager.CancelJob(jobID)
}

// SendJobInput sends interactive stdin text to a running job
func (a *App) SendJobInput(jobID string, input string) error {
	if a.jobManager == nil {
		return fmt.Errorf("job manager unavailable")
	}
	return a.jobManager.SendInput(jobID, input)
}

// GetJobStatus retrieves the status and log of a job
func (a *App) GetJobStatus(jobID string) (*models.JobStatus, bool) {
	if a.jobManager == nil {
		return nil, false
	}
	return a.jobManager.GetJob(jobID)
}

// GetIgnoredApps returns all ignored applications from SQLite
func (a *App) GetIgnoredApps() ([]models.IgnoredItem, error) {
	if a.dbStorage == nil {
		return []models.IgnoredItem{}, nil
	}
	return a.dbStorage.GetIgnoredUpdates()
}

// IgnoreApp adds an app to the ignored updates list
func (a *App) IgnoreApp(id string, name string) error {
	if a.dbStorage == nil {
		return nil
	}
	return a.dbStorage.AddIgnoredUpdate(id, name)
}

// UnignoreApp removes an app from the ignored updates list
func (a *App) UnignoreApp(id string) error {
	if a.dbStorage == nil {
		return nil
	}
	return a.dbStorage.RemoveIgnoredUpdate(id)
}

// GetDownloadedFiles scans the downloads folder (%USERPROFILE%\Downloads\EasyWinGet)
func (a *App) GetDownloadedFiles() ([]models.FileItem, error) {
	if a.downManager == nil {
		return []models.FileItem{}, nil
	}
	return a.downManager.GetFiles()
}

// DeleteDownloadedFile deletes an installer file
func (a *App) DeleteDownloadedFile(relPath string) error {
	if a.downManager == nil {
		return nil
	}
	return a.downManager.DeleteFile(relPath)
}

// OpenDownloadedFolder opens Windows Explorer at file
func (a *App) OpenDownloadedFolder(relPath string) error {
	if a.downManager == nil {
		return nil
	}
	return a.downManager.OpenFolder(relPath)
}

// OpenDownloadsDirectory opens %USERPROFILE%\Downloads\EasyWinGet directly in Windows Explorer
func (a *App) OpenDownloadsDirectory() error {
	if a.downManager == nil {
		return nil
	}
	return a.downManager.OpenFolder("")
}

// RunCommand executes a custom shell or winget command in the terminal
func (a *App) RunCommand(commandStr string) (string, error) {
	commandStr = strings.TrimSpace(commandStr)
	if commandStr == "" {
		return "", fmt.Errorf("command is empty")
	}

	parts := strings.Fields(commandStr)
	if strings.EqualFold(parts[0], "winget") {
		return a.jobManager.StartJob("winget", parts[1:])
	}

	args := []string{"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", commandStr}
	return a.jobManager.StartJob("powershell.exe", args)
}

// RunDownloadedInstaller executes downloaded installer with elevated prompt
func (a *App) RunDownloadedInstaller(relPath string) (string, error) {
	fullPath := filepath.Join(a.downManager.GetDownloadDir(), relPath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("installer file not found: %s", relPath)
	}

	args := []string{
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-Command",
		fmt.Sprintf("Start-Process -FilePath '%s' -PassThru", strings.ReplaceAll(fullPath, "'", "''")),
	}
	return a.jobManager.StartJob("powershell.exe", args)
}

// ClearCache clears all cached lists from SQLite
func (a *App) ClearCache() error {
	if a.dbStorage == nil {
		return nil
	}
	return a.dbStorage.ClearAllCache()
}

// GetSystemAccentColor returns Windows 11/10 system accent color hex
func (a *App) GetSystemAccentColor() string {
	return system.GetWindowsAccentColor()
}

// GetWindowsTransparencyEnabled returns whether Windows transparency effects are enabled
func (a *App) GetWindowsTransparencyEnabled() bool {
	return system.IsWindowsTransparencyEnabled()
}

// GetAutoStart checks if application launches on Windows startup
func (a *App) GetAutoStart() bool {
	return system.GetAutoStart()
}

// SetAutoStart sets application auto-start on Windows startup (silently)
func (a *App) SetAutoStart(enable bool) error {
	return system.SetAutoStart(enable)
}

// CheckInternetConnection tests if active network connection exists
func (a *App) CheckInternetConnection() bool {
	return system.CheckInternetConnection()
}

// CheckForUpdatesStartup triggers manual check
func (a *App) CheckForUpdatesStartup() {
	go a.startNetworkAndStartupUpdateChecker()
}

// GetSetting gets a stored setting from SQLite
func (a *App) GetSetting(key string, defVal string) string {
	if a.dbStorage == nil {
		return defVal
	}
	return a.dbStorage.GetSetting(key, defVal)
}

// SetSetting saves a setting into SQLite
func (a *App) SetSetting(key string, val string) error {
	if a.dbStorage == nil {
		return nil
	}
	return a.dbStorage.SetSetting(key, val)
}

// SelectExportPath opens native Windows save dialog
func (a *App) SelectExportPath() (string, error) {
	return wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "Export Installed Packages",
		DefaultFilename: "winget-packages.json",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
		},
	})
}

// SelectImportPath opens native Windows open file dialog
func (a *App) SelectImportPath() (string, error) {
	return wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Import Packages File",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
}

// ExportApps exports installed packages to selected file
func (a *App) ExportApps(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("no file path provided")
	}
	return winget.Export(filePath)
}

// ImportApps installs packages from selected export file
func (a *App) ImportApps(filePath string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("no file path provided")
	}
	args := []string{
		"import",
		"-i", filePath,
		"--accept-source-agreements",
		"--accept-package-agreements",
		"--verbose",
	}
	return a.jobManager.StartJob("winget", args)
}

// OpenURLInDefaultBrowser opens any URL in the user's default Windows browser (via explorer/start)
func (a *App) OpenURLInDefaultBrowser(targetURL string) error {
	targetURL = strings.TrimSpace(targetURL)
	if targetURL == "" {
		return nil
	}
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") && !strings.HasPrefix(targetURL, "mailto:") {
		targetURL = "https://" + targetURL
	}

	// 1. Primary: Use Windows Shell explorer.exe / start
	cmd := exec.Command("cmd.exe", "/c", "start", "", targetURL)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Start(); err == nil {
		return nil
	}

	// 2. Secondary: rundll32 url.dll,FileProtocolHandler
	cmd2 := exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", targetURL)
	cmd2.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd2.Start(); err == nil {
		return nil
	}

	// 3. Fallback: Wails BrowserOpenURL runtime
	if a.ctx != nil {
		wailsRuntime.BrowserOpenURL(a.ctx, targetURL)
	}
	return nil
}

type gitHubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type gitHubReleaseResponse struct {
	TagName     string               `json:"tag_name"`
	Name        string               `json:"name"`
	Body        string               `json:"body"`
	HTMLURL     string               `json:"html_url"`
	PublishedAt string               `json:"published_at"`
	Assets      []gitHubReleaseAsset `json:"assets"`
}

// CheckAppSelfUpdate queries GitHub releases to see if a newer version of EasyWinGet is available
func (a *App) CheckAppSelfUpdate() (*models.AppUpdateInfo, error) {
	currentVersion := a.GetAppVersion()
	result := &models.AppUpdateInfo{
		HasUpdate:      false,
		CurrentVersion: currentVersion,
	}

	client := &http.Client{
		Timeout: 7 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://api.github.com/repos/DARKBOYJYOTI/EasyWinget/releases/latest", nil)
	if err != nil {
		return result, nil
	}
	req.Header.Set("User-Agent", "EasyWinGet-Desktop")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return result, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, nil
	}

	var rel gitHubReleaseResponse
	if err := json.Unmarshal(body, &rel); err != nil {
		return result, nil
	}

	latestVersion := strings.TrimSpace(rel.TagName)
	latestVersion = strings.TrimPrefix(latestVersion, "v")
	latestVersion = strings.TrimPrefix(latestVersion, "V")

	result.LatestVersion = latestVersion
	result.ReleaseName = rel.Name
	result.ReleaseNotes = rel.Body
	result.ReleaseURL = rel.HTMLURL
	result.PublishedAt = rel.PublishedAt

	// Find suitable download URL from assets (installer preferred, then exe, then release URL)
	downloadURL := rel.HTMLURL
	for _, asset := range rel.Assets {
		lowerName := strings.ToLower(asset.Name)
		if strings.Contains(lowerName, "installer") && strings.HasSuffix(lowerName, ".exe") {
			downloadURL = asset.BrowserDownloadURL
			break
		} else if strings.HasSuffix(lowerName, ".exe") && downloadURL == rel.HTMLURL {
			downloadURL = asset.BrowserDownloadURL
		}
	}
	result.DownloadURL = downloadURL

	if isVersionNewer(latestVersion, currentVersion) {
		result.HasUpdate = true
	}

	return result, nil
}

// isVersionNewer compares semver strings (e.g., "4.1.0" > "4.0.0")
func isVersionNewer(latest, current string) bool {
	cleanVer := func(v string) []int {
		v = strings.TrimSpace(v)
		v = strings.TrimPrefix(v, "v")
		v = strings.TrimPrefix(v, "V")
		if idx := strings.IndexAny(v, "-+"); idx != -1 {
			v = v[:idx]
		}
		parts := strings.Split(v, ".")
		nums := make([]int, len(parts))
		for i, p := range parts {
			n, _ := strconv.Atoi(p)
			nums[i] = n
		}
		return nums
	}

	latestParts := cleanVer(latest)
	currentParts := cleanVer(current)

	maxLen := len(latestParts)
	if len(currentParts) > maxLen {
		maxLen = len(currentParts)
	}

	for i := 0; i < maxLen; i++ {
		l := 0
		c := 0
		if i < len(latestParts) {
			l = latestParts[i]
		}
		if i < len(currentParts) {
			c = currentParts[i]
		}
		if l > c {
			return true
		}
		if l < c {
			return false
		}
	}
	return false
}

// IsPortableMode returns true if the application is running in portable mode
func (a *App) IsPortableMode() bool {
	return system.IsPortableMode()
}

