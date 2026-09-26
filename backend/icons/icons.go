package icons

import (
	"context"
	"easywinget/backend/installed"
	"easywinget/backend/models"
	"easywinget/backend/winget"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	urlRegex  = regexp.MustCompile(`(?i):\s+(https?://[^\s]+)`)
	linkRegex = regexp.MustCompile(`(?i)<link[^>]*rel=["'](?:shortcut )?icon["'][^>]*href=["']([^"']+)["']`)
	appleIcon = regexp.MustCompile(`(?i)<link[^>]*rel=["']apple-touch-icon["'][^>]*href=["']([^"']+)["']`)
)

// Manager handles application and manifest icon resolution and caching
type Manager struct {
	iconsDir      string
	manifestCache map[string]models.ManifestIcon
	mu            sync.RWMutex
	client        *http.Client
}

// NewManager creates an icon manager
func NewManager(iconsDir string) (*Manager, error) {
	if err := os.MkdirAll(iconsDir, 0755); err != nil {
		return nil, err
	}
	return &Manager{
		iconsDir:      iconsDir,
		manifestCache: make(map[string]models.ManifestIcon),
		client: &http.Client{
			Timeout: 4 * time.Second,
		},
	}, nil
}

func sanitizeKey(s string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9.-]`)
	return reg.ReplaceAllString(s, "_")
}

// GetLocalAppIcon extracts the icon natively from Windows registry (BCUninstaller style) or returns cached base64
func (m *Manager) GetLocalAppIcon(id string, name string) string {
	rawKey := id
	if rawKey == "" {
		rawKey = name
	}
	if rawKey == "" {
		return ""
	}

	cacheKey := sanitizeKey(rawKey)
	iconFile := filepath.Join(m.iconsDir, cacheKey+".png")

	if _, err := os.Stat(iconFile); err == nil {
		data, err := os.ReadFile(iconFile)
		if err == nil && len(data) > 0 {
			return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
		}
	}

	// Native Windows Registry Lookup (BCUninstaller style + exe extraction)
	if nativeB64 := installed.GetScanner().FindIcon(name, id); nativeB64 != "" {
		if idx := strings.Index(nativeB64, ","); idx > 0 {
			if dec, err := base64.StdEncoding.DecodeString(nativeB64[idx+1:]); err == nil && len(dec) > 0 {
				_ = os.WriteFile(iconFile, dec, 0644)
			}
		}
		return nativeB64
	}

	return ""
}

// ScrapeIcon tries to find the website icon for a given domain/URL
func (m *Manager) ScrapeIcon(targetURL string) string {
	if targetURL == "" {
		return ""
	}
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "https://" + targetURL
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := m.client.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		return ""
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	html := string(bodyBytes)

	var iconHref string
	if match := appleIcon.FindStringSubmatch(html); len(match) > 1 {
		iconHref = match[1]
	} else if match := linkRegex.FindStringSubmatch(html); len(match) > 1 {
		iconHref = match[1]
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		return ""
	}

	if iconHref != "" {
		if strings.HasPrefix(iconHref, "//") {
			return u.Scheme + ":" + iconHref
		}
		if strings.HasPrefix(iconHref, "/") {
			return fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, iconHref)
		}
		if strings.HasPrefix(iconHref, "http") {
			return iconHref
		}
		return fmt.Sprintf("%s://%s/%s", u.Scheme, u.Host, iconHref)
	}

	// Fallback to Google Favicon API
	return fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s&sz=128", u.Host)
}

// ResolveManifestIcon extracts icon information from winget show or package homepage
func (m *Manager) ResolveManifestIcon(id string, homepage string) models.ManifestIcon {
	m.mu.RLock()
	cached, ok := m.manifestCache[id]
	m.mu.RUnlock()
	if ok {
		return cached
	}

	targetURL := homepage
	if targetURL == "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		rawShow, _ := winget.Invoke(ctx, "show", "--id", id, "--accept-source-agreements", "--disable-interactivity")
		matches := urlRegex.FindAllStringSubmatch(rawShow, -1)
		for _, m := range matches {
			if len(m) > 1 && !strings.Contains(m[1], "github") && !strings.Contains(m[1], "cdn") {
				targetURL = m[1]
				break
			}
		}
		if targetURL == "" && len(matches) > 0 {
			targetURL = matches[0][1]
		}
	}

	result := models.ManifestIcon{
		Success: false,
	}

	if targetURL != "" {
		if parsed, err := url.Parse(targetURL); err == nil {
			result.Domain = parsed.Host
			result.URL = targetURL
			result.IconURL = m.ScrapeIcon(targetURL)
			if result.IconURL != "" {
				result.Success = true
			}
		}
	}

	m.mu.Lock()
	m.manifestCache[id] = result
	m.mu.Unlock()

	return result
}
