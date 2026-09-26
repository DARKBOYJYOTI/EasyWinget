package installed

import (
	"easywinget/backend/models"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

// RegistryApp represents an installed app metadata from Windows Registry (BCUninstaller style)
type RegistryApp struct {
	DisplayName     string
	DisplayVersion  string
	Publisher       string
	InstallLocation string
	DisplayIcon          string
	UninstallString      string
	QuietUninstallString string
}

// Scanner scans Windows registry natively for installed programs and icons
type Scanner struct {
	mu    sync.RWMutex
	apps  map[string]RegistryApp // normalized name/id -> info
	icons map[string]string      // normalized name -> base64 icon
}

var (
	globalScanner *Scanner
	scannerOnce   sync.Once
)

// GetScanner returns the singleton native registry scanner
func GetScanner() *Scanner {
	scannerOnce.Do(func() {
		globalScanner = &Scanner{
			apps:  make(map[string]RegistryApp),
			icons: make(map[string]string),
		}
		globalScanner.Refresh()
	})
	return globalScanner
}

// Refresh re-scans the Windows Registry (HKLM, WOW64, HKCU) in pure Go without PowerShell
func (s *Scanner) Refresh() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.apps = make(map[string]RegistryApp)
	s.icons = make(map[string]string)

	registryRoots := []struct {
		root registry.Key
		path string
	}{
		{registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`},
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`},
		{registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`},
	}

	for _, regTarget := range registryRoots {
		k, err := registry.OpenKey(regTarget.root, regTarget.path, registry.READ)
		if err != nil {
			continue
		}

		subkeys, err := k.ReadSubKeyNames(-1)
		k.Close()
		if err != nil {
			continue
		}

		for _, sub := range subkeys {
			subKey, err := registry.OpenKey(regTarget.root, regTarget.path+`\`+sub, registry.QUERY_VALUE)
			if err != nil {
				continue
			}

			name, _, _ := subKey.GetStringValue("DisplayName")
			name = strings.TrimSpace(name)
			if name == "" {
				subKey.Close()
				continue
			}

			version, _, _ := subKey.GetStringValue("DisplayVersion")
			publisher, _, _ := subKey.GetStringValue("Publisher")
			location, _, _ := subKey.GetStringValue("InstallLocation")
			icon, _, _ := subKey.GetStringValue("DisplayIcon")
			uninstall, _, _ := subKey.GetStringValue("UninstallString")
			quietUninstall, _, _ := subKey.GetStringValue("QuietUninstallString")

			subKey.Close()

			app := RegistryApp{
				DisplayName:          name,
				DisplayVersion:       version,
				Publisher:            publisher,
				InstallLocation:      location,
				DisplayIcon:          icon,
				UninstallString:      uninstall,
				QuietUninstallString: quietUninstall,
			}

			normName := normalizeKey(name)
			normSub := normalizeKey(sub)

			s.apps[normName] = app
			if normSub != "" {
				s.apps[normSub] = app
			}
		}
	}

	// Scan Microsoft Store / AppX packages for native icons
	appModelPath := `Software\Classes\Local Settings\Software\Microsoft\Windows\CurrentVersion\AppModel\Repository\Packages`
	if k, err := registry.OpenKey(registry.CURRENT_USER, appModelPath, registry.READ); err == nil {
		subkeys, _ := k.ReadSubKeyNames(-1)
		k.Close()
		for _, sub := range subkeys {
			subKey, err := registry.OpenKey(registry.CURRENT_USER, appModelPath+`\`+sub, registry.QUERY_VALUE)
			if err != nil {
				continue
			}
			pkgRoot, _, _ := subKey.GetStringValue("PackageRootFolder")
			displayName, _, _ := subKey.GetStringValue("DisplayName")
			pkgID, _, _ := subKey.GetStringValue("PackageID")
			subKey.Close()

			if pkgRoot == "" {
				continue
			}

			b64 := findStoreAppLogo(pkgRoot)
			if b64 != "" {
				if displayName != "" && !strings.HasPrefix(displayName, "@") {
					s.icons[normalizeKey(displayName)] = b64
				}
				if pkgID != "" {
					s.icons[normalizeKey(pkgID)] = b64
				}
				s.icons[normalizeKey(sub)] = b64
			}
		}
	}
}

// findStoreAppLogo searches a package folder for logo images (.png)
func findStoreAppLogo(pkgRoot string) string {
	dirs := []string{
		filepath.Join(pkgRoot, "Assets"),
		filepath.Join(pkgRoot, "Images"),
		pkgRoot,
	}

	var candidates []string
	for _, d := range dirs {
		entries, err := os.ReadDir(d)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			lower := strings.ToLower(e.Name())
			if strings.HasSuffix(lower, ".png") {
				candidates = append(candidates, filepath.Join(d, e.Name()))
			}
		}
	}

	if len(candidates) == 0 {
		return ""
	}

	// Priority heuristics for highest quality icon
	selected := candidates[0]
	for _, c := range candidates {
		lower := strings.ToLower(c)
		if strings.Contains(lower, "square44x44logo") || strings.Contains(lower, "applist") || strings.Contains(lower, "square150x150logo") || strings.Contains(lower, "storelogo") || strings.Contains(lower, "logo") {
			// Skip contrast-white or contrast-black tiles if normal scale available
			if !strings.Contains(lower, "contrast") {
				selected = c
				break
			}
		}
	}

	if data, err := os.ReadFile(selected); err == nil && len(data) > 0 {
		return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
	}
	return ""
}

// GetUninstallString resolves the native uninstaller command string for an app
func (s *Scanner) GetUninstallString(name, id string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	normName := normalizeKey(name)
	normID := normalizeKey(id)

	var app RegistryApp
	var found bool

	if normID != "" {
		app, found = s.apps[normID]
	}
	if !found && normName != "" {
		app, found = s.apps[normName]
	}
	if !found {
		// Fuzzy match
		for k, v := range s.apps {
			if normName != "" && (strings.Contains(k, normName) || strings.Contains(normName, k)) {
				app = v
				found = true
				break
			}
		}
	}

	if !found {
		return ""
	}

	cmdStr := app.QuietUninstallString
	if cmdStr == "" {
		cmdStr = app.UninstallString
	}
	cmdStr = strings.TrimSpace(cmdStr)
	if cmdStr == "" {
		return ""
	}

	// If it's an MSI installer with /I (install/repair), switch to /X (uninstall)
	lower := strings.ToLower(cmdStr)
	if strings.Contains(lower, "msiexec") {
		cmdStr = strings.ReplaceAll(cmdStr, "/I ", "/X ")
		cmdStr = strings.ReplaceAll(cmdStr, "/i ", "/X ")
		cmdStr = strings.ReplaceAll(cmdStr, "/I{", "/X{")
		cmdStr = strings.ReplaceAll(cmdStr, "/i{", "/X{")
		cmdStr = strings.ReplaceAll(cmdStr, "-I ", "-X ")
		cmdStr = strings.ReplaceAll(cmdStr, "-i ", "-X ")
		cmdStr = strings.ReplaceAll(cmdStr, "-I{", "-X{")
		cmdStr = strings.ReplaceAll(cmdStr, "-i{", "-X{")
	}

	return cmdStr
}

// SplitCommandLine parses a Windows command line into executable path and arguments
func SplitCommandLine(cmdStr string) (string, []string) {
	cmdStr = strings.TrimSpace(cmdStr)
	if cmdStr == "" {
		return "", nil
	}

	var exe string
	var rest string

	if strings.HasPrefix(cmdStr, "\"") {
		// Quoted path: "C:\Program Files\App\uninstall.exe" /arg1 /arg2
		end := strings.Index(cmdStr[1:], "\"")
		if end != -1 {
			exe = cmdStr[1 : end+1]
			rest = strings.TrimSpace(cmdStr[end+2:])
		} else {
			exe = strings.Trim(cmdStr, "\"")
		}
	} else {
		// Check for .exe in unquoted path with spaces (e.g. C:\Program Files\App\unins.exe /S)
		lower := strings.ToLower(cmdStr)
		if idx := strings.Index(lower, ".exe"); idx != -1 {
			exe = strings.TrimSpace(cmdStr[:idx+4])
			rest = strings.TrimSpace(cmdStr[idx+4:])
		} else {
			fields := strings.Fields(cmdStr)
			if len(fields) > 0 {
				exe = fields[0]
				if len(fields) > 1 {
					rest = strings.Join(fields[1:], " ")
				}
			}
		}
	}

	exe = strings.Trim(exe, "\"")

	var args []string
	if rest != "" {
		var current strings.Builder
		inQuote := false
		for _, r := range rest {
			if r == '"' {
				inQuote = !inQuote
			} else if r == ' ' && !inQuote {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(r)
			}
		}
		if current.Len() > 0 {
			args = append(args, current.String())
		}
	}

	return exe, args
}

func normalizeKey(s string) string {
	lower := strings.ToLower(strings.TrimSpace(s))
	replacer := strings.NewReplacer(" ", "", "-", "", "_", "", ".", "", "(", "", ")", "")
	return replacer.Replace(lower)
}

var (
	exeExtractSem = make(chan struct{}, 2) // Limit concurrent PowerShell icon extractions to 2
)

// FindIcon looks for an icon in the registry entry DisplayIcon or InstallLocation
func (s *Scanner) FindIcon(name, id string) string {
	normName := normalizeKey(name)
	normID := normalizeKey(id)

	s.mu.RLock()
	if icon, ok := s.icons[normName]; ok && icon != "" {
		s.mu.RUnlock()
		return icon
	}
	if icon, ok := s.icons[normID]; ok && icon != "" {
		s.mu.RUnlock()
		return icon
	}
	s.mu.RUnlock()

	// 1. Direct MSIX package resolution
	if strings.HasPrefix(id, "MSIX\\") {
		pkgName := strings.TrimPrefix(id, "MSIX\\")
		pkgPath := filepath.Join(`C:\Program Files\WindowsApps`, pkgName)
		if b64 := findStoreAppLogo(pkgPath); b64 != "" {
			s.mu.Lock()
			s.icons[normName] = b64
			s.icons[normID] = b64
			s.mu.Unlock()
			return b64
		}
	}

	s.mu.RLock()
	var app RegistryApp
	var found bool

	if app, found = s.apps[normID]; !found {
		if app, found = s.apps[normName]; !found {
			// Try prefix / fuzzy match
			for k, v := range s.apps {
				if normName != "" && (strings.Contains(k, normName) || strings.Contains(normName, k)) {
					app = v
					found = true
					break
				}
			}
		}
	}
	s.mu.RUnlock()

	if !found {
		return ""
	}

	// 2. Inspect DisplayIcon
	if app.DisplayIcon != "" {
		cleanPath := strings.Trim(app.DisplayIcon, `"'`)
		if idx := strings.Index(cleanPath, ","); idx > 0 {
			cleanPath = strings.TrimSpace(cleanPath[:idx])
		}

		if info, err := os.Stat(cleanPath); err == nil && !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(cleanPath))
			if ext == ".ico" {
				if data, err := os.ReadFile(cleanPath); err == nil && len(data) > 0 {
					b64 := "data:image/x-icon;base64," + base64.StdEncoding.EncodeToString(data)
					s.mu.Lock()
					s.icons[normName] = b64
					s.mu.Unlock()
					return b64
				}
			} else if ext == ".png" || ext == ".jpg" || ext == ".webp" {
				if data, err := os.ReadFile(cleanPath); err == nil && len(data) > 0 {
					b64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
					s.mu.Lock()
					s.icons[normName] = b64
					s.mu.Unlock()
					return b64
				}
			} else if ext == ".exe" || ext == ".dll" {
				if b64 := ExtractExeIcon(cleanPath); b64 != "" {
					s.mu.Lock()
					s.icons[normName] = b64
					s.mu.Unlock()
					return b64
				}
			}
		}
	}

	// 3. Inspect InstallLocation for .ico or .exe files
	if app.InstallLocation != "" {
		cleanDir := strings.Trim(app.InstallLocation, `"'`)
		if info, err := os.Stat(cleanDir); err == nil && info.IsDir() {
			entries, _ := os.ReadDir(cleanDir)
			for _, entry := range entries {
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if ext == ".ico" {
					icoPath := filepath.Join(cleanDir, entry.Name())
					if data, err := os.ReadFile(icoPath); err == nil && len(data) > 0 {
						b64 := "data:image/x-icon;base64," + base64.StdEncoding.EncodeToString(data)
						s.mu.Lock()
						s.icons[normName] = b64
						s.mu.Unlock()
						return b64
					}
				} else if ext == ".exe" && (strings.Contains(strings.ToLower(entry.Name()), normName) || strings.EqualFold(entry.Name(), normName+".exe")) {
					exePath := filepath.Join(cleanDir, entry.Name())
					if b64 := ExtractExeIcon(exePath); b64 != "" {
						s.mu.Lock()
						s.icons[normName] = b64
						s.mu.Unlock()
						return b64
					}
				}
			}
		}
	}

	return ""
}

// ExtractExeIcon extracts the Windows associated icon from an executable or DLL to a PNG base64 string
// Uses semaphore to prevent spawning multiple concurrent PowerShell processes
func ExtractExeIcon(cleanPath string) string {
	if cleanPath == "" {
		return ""
	}
	clean := filepath.Clean(cleanPath)
	if _, err := os.Stat(clean); err != nil {
		return ""
	}

	// Non-blocking try: if 2 workers are already running, do not block the thread
	select {
	case exeExtractSem <- struct{}{}:
		defer func() { <-exeExtractSem }()
	default:
		// Saturated: return empty to avoid freezing UI
		return ""
	}

	psScript := fmt.Sprintf(`Add-Type -AssemblyName System.Drawing; $i=[System.Drawing.Icon]::ExtractAssociatedIcon('%s'); if($null -ne $i){$ms=New-Object System.IO.MemoryStream; $i.ToBitmap().Save($ms,[System.Drawing.Imaging.ImageFormat]::Png); [Convert]::ToBase64String($ms.ToArray())}`, strings.ReplaceAll(clean, "'", "''"))
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	b64 := strings.TrimSpace(string(out))
	if len(b64) > 50 {
		return "data:image/png;base64," + b64
	}
	return ""
}

// GetRegistryApps returns all installed applications detected natively via registry
func (s *Scanner) GetRegistryApps() []models.AppItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []models.AppItem
	seen := make(map[string]bool)

	for _, app := range s.apps {
		if seen[app.DisplayName] {
			continue
		}
		seen[app.DisplayName] = true

		list = append(list, models.AppItem{
			ID:      app.DisplayName,
			Name:    app.DisplayName,
			Version: app.DisplayVersion,
			Source:  "Registry",
		})
	}

	return list
}
