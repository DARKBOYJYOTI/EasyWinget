package system

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows/registry"
)

// GetWindowsAccentColor reads the current Windows 11/10 system accent color from the registry
func GetWindowsAccentColor() string {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\DWM`, registry.QUERY_VALUE)
	if err != nil {
		return "#0078d4" // Default Fluent Blue
	}
	defer k.Close()

	val, _, err := k.GetIntegerValue("AccentColor")
	if err != nil {
		// Try ColorizationColor
		val, _, err = k.GetIntegerValue("ColorizationColor")
		if err != nil {
			return "#0078d4"
		}
	}

	// AccentColor in registry is in ABGR or AARRGGBB format (DWORD)
	// We extract R, G, B
	r := val & 0xFF
	g := (val >> 8) & 0xFF
	b := (val >> 16) & 0xFF

	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// IsWindowsTransparencyEnabled checks if Windows transparency effects are enabled in user Personalize settings
func IsWindowsTransparencyEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	val, _, err := k.GetIntegerValue("EnableTransparency")
	return err == nil && val != 0
}

// GetAutoStart checks if EasyWinGet is set to launch on Windows startup
func GetAutoStart() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	val, _, err := k.GetStringValue("EasyWinGet")
	return err == nil && val != ""
}

// SetAutoStart enables or disables silent auto-start on Windows startup
func SetAutoStart(enable bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	if enable {
		exePath, err := os.Executable()
		if err != nil {
			return err
		}
		cmdVal := fmt.Sprintf(`"%s" -silent`, exePath)
		return k.SetStringValue("EasyWinGet", cmdVal)
	}

	_ = k.DeleteValue("EasyWinGet")
	return nil
}

// GetEasyWinGetDownloadDir returns %USERPROFILE%\Downloads\EasyWinGet
func GetEasyWinGetDownloadDir() string {
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			userProfile = home
		} else {
			userProfile = "."
		}
	}
	dir := filepath.Join(userProfile, "Downloads", "EasyWinGet")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// CheckInternetConnection tests if there is an active internet connection
func CheckInternetConnection() bool {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	// Microsoft network connectivity check URL or Cloudflare DNS
	resp, err := client.Head("http://www.msftconnecttest.com/connecttest.txt")
	if err == nil && resp.StatusCode == 200 {
		return true
	}

	resp2, err2 := client.Head("https://1.1.1.1")
	return err2 == nil && (resp2.StatusCode >= 200 && resp2.StatusCode < 400)
}

// SanitizeFolderName cleans strings for folder creation
func SanitizeFolderName(name string) string {
	replacer := strings.NewReplacer("<", "", ">", "", ":", "", "\"", "", "/", "", "\\", "", "|", "", "?", "", "*", "")
	clean := strings.TrimSpace(replacer.Replace(name))
	if clean == "" {
		return "Package"
	}
	return clean
}

var cachedIconPath string

// InitAppIcon saves the application icon to %LOCALAPPDATA%\EasyWinGet and registers the Start Menu shortcut
func InitAppIcon(iconData []byte) {
	appData := os.Getenv("LOCALAPPDATA")
	if appData == "" {
		appData = os.Getenv("APPDATA")
	}
	if appData == "" {
		return
	}
	appDir := filepath.Join(appData, "EasyWinGet")
	_ = os.MkdirAll(appDir, 0755)

	iconFile := filepath.Join(appDir, "appicon.png")
	if len(iconData) > 0 {
		_ = os.WriteFile(iconFile, iconData, 0644)
	}
	cachedIconPath = iconFile

	// Ensure Start Menu shortcut exists so Windows Action Center associates notifications with EasyWinGet icon
	go func() {
		exePath, err := os.Executable()
		if err != nil {
			return
		}
		programsDir := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs")
		shortcutPath := filepath.Join(programsDir, "EasyWinGet.lnk")
		if _, err := os.Stat(shortcutPath); os.IsNotExist(err) {
			psScript := fmt.Sprintf(`
$wsh = New-Object -ComObject WScript.Shell
$sc = $wsh.CreateShortcut('%s')
$sc.TargetPath = '%s'
$sc.IconLocation = '%s,0'
$sc.Description = 'EasyWinGet - Windows Package Manager Hub'
$sc.Save()
`, strings.ReplaceAll(shortcutPath, `'`, `''`), strings.ReplaceAll(exePath, `'`, `''`), strings.ReplaceAll(exePath, `'`, `''`))
			cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", psScript)
			cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
			_ = cmd.Run()
		}
	}()
}

// RegisterProtocolHandler registers the easywinget:// URL protocol in Windows Registry (HKCU)
func RegisterProtocolHandler() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}

	k, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Classes\easywinget`, registry.ALL_ACCESS)
	if err != nil {
		return
	}
	defer k.Close()

	_ = k.SetStringValue("", "URL:EasyWinGet Protocol")
	_ = k.SetStringValue("URL Protocol", "")

	cmdKey, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Classes\easywinget\shell\open\command`, registry.ALL_ACCESS)
	if err != nil {
		return
	}
	defer cmdKey.Close()

	_ = cmdKey.SetStringValue("", fmt.Sprintf(`"%s" "%%1"`, exePath))
}

// ShowWindowsNotification displays a native Windows system notification with app icon and protocol activation
func ShowWindowsNotification(title, message string) {
	ShowWindowsNotificationWithAction(title, message, "easywinget://updates", "View Updates")
}

// ShowWindowsNotificationWithAction displays a Windows notification with click-to-activate protocol & action button
func ShowWindowsNotificationWithAction(title, message, launchUri, actionTitle string) {
	iconSrc := cachedIconPath
	if iconSrc == "" {
		appData := os.Getenv("LOCALAPPDATA")
		if appData != "" {
			iconSrc = filepath.Join(appData, "EasyWinGet", "appicon.png")
		}
	}

	if launchUri == "" {
		launchUri = "easywinget://updates"
	}
	if actionTitle == "" {
		actionTitle = "View Updates"
	}

	var script string
	if iconSrc != "" {
		script = fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null

$xml = @"
<toast activationType="protocol" launch="%s">
  <visual>
    <binding template="ToastGeneric">
      <text>%s</text>
      <text>%s</text>
      <image placement="appLogoOverride" hint-crop="circle" src="%s"/>
    </binding>
  </visual>
  <actions>
    <action content="%s" activationType="protocol" arguments="%s"/>
  </actions>
</toast>
"@

$doc = [Windows.Data.Xml.Dom.XmlDocument]::new()
$doc.LoadXml($xml)
$toast = [Windows.UI.Notifications.ToastNotification]::new($doc)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("EasyWinGet").Show($toast)
`, strings.ReplaceAll(launchUri, `"`, `\"`), strings.ReplaceAll(title, `"`, `\"`), strings.ReplaceAll(message, `"`, `\"`), strings.ReplaceAll(iconSrc, `"`, `\"`), strings.ReplaceAll(actionTitle, `"`, `\"`), strings.ReplaceAll(launchUri, `"`, `\"`))
	} else {
		script = fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null

$xml = @"
<toast activationType="protocol" launch="%s">
  <visual>
    <binding template="ToastGeneric">
      <text>%s</text>
      <text>%s</text>
    </binding>
  </visual>
  <actions>
    <action content="%s" activationType="protocol" arguments="%s"/>
  </actions>
</toast>
"@

$doc = [Windows.Data.Xml.Dom.XmlDocument]::new()
$doc.LoadXml($xml)
$toast = [Windows.UI.Notifications.ToastNotification]::new($doc)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("EasyWinGet").Show($toast)
`, strings.ReplaceAll(launchUri, `"`, `\"`), strings.ReplaceAll(title, `"`, `\"`), strings.ReplaceAll(message, `"`, `\"`), strings.ReplaceAll(actionTitle, `"`, `\"`), strings.ReplaceAll(launchUri, `"`, `\"`))
	}

	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000} // CREATE_NO_WINDOW
	_ = cmd.Run()
}


