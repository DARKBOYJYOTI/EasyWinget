package winget

import (
	"bytes"
	"context"
	"easywinget/backend/models"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"
)

var (
	ansiRegex     = regexp.MustCompile(`[\x1b\x9b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[a-zA-Z]|[\x1b\x9b]][^\x07\x1b]*[\x07\x1b\\]`)
	spinnerRegex1 = regexp.MustCompile(`(\r[-\\|\/]+)|([-\\|\/]+\r)`)
	spinnerRegex2 = regexp.MustCompile(`(?m)^[-\\|\/]\r`)
	spinnerRegex3 = regexp.MustCompile(`(?m)^\s*[-\\|\/]\s*$`)
)

// CleanOutput removes ANSI escape codes, terminal spinner artifacts, and splits into clean trimmed lines
func CleanOutput(raw string) []string {
	clean := ansiRegex.ReplaceAllString(raw, "")
	clean = spinnerRegex1.ReplaceAllString(clean, "")
	clean = spinnerRegex2.ReplaceAllString(clean, "")
	clean = spinnerRegex3.ReplaceAllString(clean, "")

	// Handle Backspace (\x08)
	for strings.Contains(clean, "\x08") {
		clean = regexp.MustCompile(`[^\x08]\x08`).ReplaceAllString(clean, "")
		clean = regexp.MustCompile(`^\x08+`).ReplaceAllString(clean, "")
	}

	rawLines := strings.Split(clean, "\n")
	var lines []string
	for _, l := range rawLines {
		trimmed := strings.TrimRight(strings.TrimLeft(l, " \t\r"), " \t\r")
		if len(trimmed) > 0 {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

// ParseApps parses 'winget search' or 'winget list' output using column positions
func ParseApps(output string) []models.AppItem {
	lines := CleanOutput(output)
	var items []models.AppItem

	colID := -1
	colVersion := -1

	for _, line := range lines {
		// Detect Header
		if colID == -1 && strings.Contains(line, "Name") && strings.Contains(line, "Id") {
			nameIdx := strings.Index(line, "Name")
			colID = strings.Index(line, "Id") - nameIdx
			colVersion = strings.Index(line, "Version") - nameIdx
			continue
		}

		if colID == -1 || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "Name") {
			continue
		}

		if len(line) < colID {
			continue
		}

		name := strings.TrimSpace(line[:colID])
		id := ""
		version := ""

		if colVersion > -1 && len(line) > colVersion {
			id = strings.TrimSpace(line[colID:colVersion])
			rest := strings.TrimSpace(line[colVersion:])
			fields := strings.Fields(rest)
			if len(fields) > 0 {
				version = fields[0]
			}
		} else {
			id = strings.TrimSpace(line[colID:])
		}

		if name != "" && id != "" {
			items = append(items, models.AppItem{
				Name:    name,
				ID:      id,
				Version: version,
			})
		}
	}
	return items
}

// ParseUpdates parses 'winget upgrade' output using column positions
func ParseUpdates(output string) []models.AppItem {
	lines := CleanOutput(output)
	var items []models.AppItem

	colID := -1
	colVersion := -1
	colAvailable := -1

	for _, line := range lines {
		// Detect Header
		if colID == -1 && strings.Contains(line, "Name") && strings.Contains(line, "Id") {
			nameIdx := strings.Index(line, "Name")
			colID = strings.Index(line, "Id") - nameIdx
			colVersion = strings.Index(line, "Version") - nameIdx
			colAvailable = strings.Index(line, "Available") - nameIdx
			continue
		}

		if colID == -1 || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "Name") {
			continue
		}

		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "upgrade") && strings.Contains(lowerLine, "available") {
			continue
		}
		if strings.Contains(lowerLine, "no applicable") || strings.Contains(lowerLine, "no package") {
			continue
		}

		if len(line) < colID {
			continue
		}

		name := strings.TrimSpace(line[:colID])
		id := ""
		current := ""
		available := ""

		if colVersion > -1 && len(line) > colVersion {
			id = strings.TrimSpace(line[colID:colVersion])

			if colAvailable > -1 && len(line) > colAvailable {
				current = strings.TrimSpace(line[colVersion:colAvailable])
				rest := strings.TrimSpace(line[colAvailable:])
				fields := strings.Fields(rest)
				if len(fields) > 0 {
					available = fields[0]
				}
			} else {
				current = strings.TrimSpace(line[colVersion:])
			}
		} else {
			id = strings.TrimSpace(line[colID:])
		}

		if id == "" || len(id) < 3 || strings.Contains(id, " ") {
			continue
		}

		if name != "" && id != "" {
			items = append(items, models.AppItem{
				Name:      name,
				ID:        id,
				Version:   current,
				Available: available,
			})
		}
	}
	return items
}

// ParseDetails parses 'winget show --id <id>' output
func ParseDetails(id string, output string) models.AppDetails {
	lines := strings.Split(output, "\n")
	details := models.AppDetails{
		ID:   id,
		Tags: []string{},
	}

	currentSection := ""
	var descLines []string
	collectingDesc := false

	foundRegex := regexp.MustCompile(`^Found (.+?) \[(.+?)\]`)
	kvRegex := regexp.MustCompile(`^([A-Za-z\s]+):\s*(.*)$`)

	for _, rawLine := range lines {
		line := strings.TrimRight(rawLine, "\r\n")

		if match := foundRegex.FindStringSubmatch(line); len(match) > 1 {
			details.Name = strings.TrimSpace(match[1])
			continue
		}

		if kvMatch := kvRegex.FindStringSubmatch(line); len(kvMatch) > 2 {
			key := strings.ToLower(strings.TrimSpace(kvMatch[1]))
			val := strings.TrimSpace(kvMatch[2])

			if collectingDesc && key != "description" {
				details.Description = strings.Join(descLines, " ")
				collectingDesc = false
			}

			switch key {
			case "version":
				details.Version = val
			case "publisher":
				details.Publisher = val
			case "author":
				details.Author = val
			case "moniker":
				details.Moniker = val
			case "description":
				collectingDesc = true
				if val != "" {
					descLines = append(descLines, val)
				}
			case "license":
				details.License = val
			case "copyright":
				details.Copyright = val
			case "homepage":
				details.Homepage = val
			case "release notes url":
				details.ReleaseNotesURL = val
			case "installer type":
				details.InstallerType = val
			case "installer url":
				details.InstallerURL = val
			case "installer sha256":
				details.InstallerSha256 = val
			case "installer":
				currentSection = "installer"
			case "tags":
				currentSection = "tags"
			}
		} else if strings.HasPrefix(line, "  ") && currentSection == "tags" {
			tag := strings.TrimSpace(line)
			if tag != "" && !strings.Contains(tag, ":") {
				details.Tags = append(details.Tags, tag)
			}
		} else if collectingDesc && strings.TrimSpace(line) != "" {
			descLines = append(descLines, strings.TrimSpace(line))
		}
	}

	if collectingDesc {
		details.Description = strings.Join(descLines, " ")
	}

	return details
}

// Invoke executes a winget command with UTF-8 encoding support
func Invoke(ctx context.Context, args ...string) (string, error) {
	// cmd.exe /c chcp 65001 > nul && winget <args...>
	cmdArgs := append([]string{"/c", "chcp", "65001", ">", "nul", "&&", "winget"}, args...)
	cmd := exec.CommandContext(ctx, "cmd.exe", cmdArgs...)

	// Hide window
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	outStr := stdout.String()

	// Winget sometimes returns non-zero exit code even when output is valid
	if err != nil && outStr == "" {
		return "", fmt.Errorf("winget error: %v, stderr: %s", err, stderr.String())
	}

	return outStr, nil
}

// Search searches for packages in the winget repository
func Search(query string) ([]models.AppItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	out, err := Invoke(ctx, "search", query, "-s", "winget", "--accept-source-agreements")
	if err != nil && out == "" {
		return nil, err
	}
	return ParseApps(out), nil
}

// ListInstalled lists all installed applications on the system
func ListInstalled() ([]models.AppItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	out, err := Invoke(ctx, "list", "--accept-source-agreements")
	if err != nil && out == "" {
		return nil, err
	}
	return ParseApps(out), nil
}

// ListUpdates lists all installed applications that have updates available
func ListUpdates() ([]models.AppItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	out, err := Invoke(ctx, "upgrade", "--accept-source-agreements")
	if err != nil && out == "" {
		return nil, err
	}
	return ParseUpdates(out), nil
}

// Show gets manifest details for a given package ID
func Show(id string) (*models.AppDetails, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	out, err := Invoke(ctx, "show", "--id", id, "--accept-source-agreements", "--disable-interactivity")
	if err != nil && out == "" {
		return nil, err
	}
	details := ParseDetails(id, out)
	return &details, nil
}

// Export runs 'winget export' to save installed apps to a file
func Export(targetFilePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err := Invoke(ctx, "export", "-o", targetFilePath, "--accept-source-agreements")
	return err
}

// Import runs 'winget import' to install apps from an exported file
func Import(sourceFilePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	_, err := Invoke(ctx, "import", "-i", sourceFilePath, "--accept-source-agreements", "--accept-package-agreements")
	return err
}
