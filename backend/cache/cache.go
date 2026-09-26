package cache

import (
	"easywinget/backend/models"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Manager handles thread-safe disk caching of JSON data
type Manager struct {
	mu      sync.RWMutex
	dataDir string
}

// NewManager initializes cache directory and returns a Manager
func NewManager(dataDir string) (*Manager, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}
	return &Manager{dataDir: dataDir}, nil
}

func (m *Manager) filePath(key string) string {
	return filepath.Join(m.dataDir, key+".json")
}

// Save marshals and writes data to disk
func (m *Manager) Save(key string, data interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.filePath(key), bytes, 0644)
}

// Load reads and unmarshals data from disk. Returns true if file existed and loaded successfully.
func (m *Manager) Load(key string, target interface{}) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fp := m.filePath(key)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return false, nil
	}

	bytes, err := os.ReadFile(fp)
	if err != nil {
		return false, err
	}

	if err := json.Unmarshal(bytes, target); err != nil {
		return false, err
	}
	return true, nil
}

// Clear removes a specific cache file
func (m *Manager) Clear(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fp := m.filePath(key)
	if err := os.Remove(fp); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ClearAll removes all json files in data directory except ignored.json
func (m *Manager) ClearAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entries, err := os.ReadDir(m.dataDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" && entry.Name() != "ignored.json" {
			_ = os.Remove(filepath.Join(m.dataDir, entry.Name()))
		}
	}
	return nil
}

// GetIgnored returns list of ignored package updates
func (m *Manager) GetIgnored() ([]models.IgnoredItem, error) {
	var list []models.IgnoredItem
	found, err := m.Load("ignored", &list)
	if err != nil {
		return nil, err
	}
	if !found {
		return []models.IgnoredItem{}, nil
	}
	return list, nil
}

// AddIgnored adds a package to the ignored updates list
func (m *Manager) AddIgnored(id string, name string) error {
	items, err := m.GetIgnored()
	if err != nil {
		items = []models.IgnoredItem{}
	}

	for _, item := range items {
		if item.ID == id {
			return nil // already ignored
		}
	}

	items = append(items, models.IgnoredItem{ID: id, Name: name})
	return m.Save("ignored", items)
}

// RemoveIgnored un-ignores a package
func (m *Manager) RemoveIgnored(id string) error {
	items, err := m.GetIgnored()
	if err != nil {
		return err
	}

	var filtered []models.IgnoredItem
	for _, item := range items {
		if item.ID != id {
			filtered = append(filtered, item)
		}
	}

	return m.Save("ignored", filtered)
}
