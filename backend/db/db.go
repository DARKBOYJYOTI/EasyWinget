package db

import (
	"database/sql"
	"easywinget/backend/models"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Storage handles SQLite persistence for packages, updates, settings, and cache
type Storage struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewStorage initializes SQLite database and creates tables if they don't exist
func NewStorage(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "easywinget.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Optimize connection pool for desktop app
	db.SetMaxOpenConns(1)

	s := &Storage{db: db}
	if err := s.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return s, nil
}

func (s *Storage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS installed_apps (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		version TEXT,
		source TEXT,
		updated_at INTEGER
	);

	CREATE TABLE IF NOT EXISTS available_updates (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		version TEXT,
		available TEXT,
		source TEXT,
		updated_at INTEGER
	);

	CREATE TABLE IF NOT EXISTS ignored_updates (
		id TEXT PRIMARY KEY,
		name TEXT
	);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);
	`
	_, err := s.db.Exec(schema)
	return err
}

// Close closes database handle
func (s *Storage) Close() error {
	return s.db.Close()
}

// --- Installed Apps ---

// GetInstalledApps retrieves all cached installed apps from SQLite
func (s *Storage) GetInstalledApps() ([]models.AppItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT id, name, version, source FROM installed_apps ORDER BY name COLLATE NOCASE ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []models.AppItem
	for rows.Next() {
		var app models.AppItem
		var source sql.NullString
		if err := rows.Scan(&app.ID, &app.Name, &app.Version, &source); err != nil {
			continue
		}
		if source.Valid {
			app.Source = source.String
		}
		apps = append(apps, app)
	}
	return apps, nil
}

// SaveInstalledApps replaces the installed apps cache in a single transaction
func (s *Storage) SaveInstalledApps(apps []models.AppItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM installed_apps"); err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO installed_apps (id, name, version, source, updated_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().Unix()
	for _, app := range apps {
		if _, err := stmt.Exec(app.ID, app.Name, app.Version, app.Source, now); err != nil {
			continue
		}
	}

	return tx.Commit()
}

// HasInstalledApps checks if there are cached installed apps
func (s *Storage) HasInstalledApps() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int
	_ = s.db.QueryRow("SELECT COUNT(*) FROM installed_apps").Scan(&count)
	return count > 0
}

// --- Available Updates ---

// GetAvailableUpdates retrieves cached updates
func (s *Storage) GetAvailableUpdates() ([]models.AppItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT u.id, u.name, u.version, u.available, u.source 
		FROM available_updates u
		LEFT JOIN ignored_updates i ON u.id = i.id
		WHERE i.id IS NULL
		ORDER BY u.name COLLATE NOCASE ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var updates []models.AppItem
	for rows.Next() {
		var u models.AppItem
		var source sql.NullString
		if err := rows.Scan(&u.ID, &u.Name, &u.Version, &u.Available, &source); err != nil {
			continue
		}
		if source.Valid {
			u.Source = source.String
		}
		updates = append(updates, u)
	}
	return updates, nil
}

// SaveAvailableUpdates saves updates to SQLite
func (s *Storage) SaveAvailableUpdates(updates []models.AppItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM available_updates"); err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO available_updates (id, name, version, available, source, updated_at) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().Unix()
	for _, u := range updates {
		if _, err := stmt.Exec(u.ID, u.Name, u.Version, u.Available, u.Source, now); err != nil {
			continue
		}
	}

	return tx.Commit()
}

// HasAvailableUpdates checks if updates are cached
func (s *Storage) HasAvailableUpdates() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int
	_ = s.db.QueryRow("SELECT COUNT(*) FROM available_updates").Scan(&count)
	return count > 0
}

// --- Ignored Updates ---

// GetIgnoredUpdates retrieves ignored apps
func (s *Storage) GetIgnoredUpdates() ([]models.IgnoredItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT id, name FROM ignored_updates ORDER BY name COLLATE NOCASE ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.IgnoredItem
	for rows.Next() {
		var item models.IgnoredItem
		if err := rows.Scan(&item.ID, &item.Name); err == nil {
			list = append(list, item)
		}
	}
	return list, nil
}

// AddIgnoredUpdate adds app to ignored list
func (s *Storage) AddIgnoredUpdate(id string, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("INSERT OR REPLACE INTO ignored_updates (id, name) VALUES (?, ?)", id, name)
	return err
}

// RemoveIgnoredUpdate removes app from ignored list
func (s *Storage) RemoveIgnoredUpdate(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM ignored_updates WHERE id = ?", id)
	return err
}

// --- Settings Storage ---

// GetSetting gets a setting value from SQLite
func (s *Storage) GetSetting(key string, defaultVal string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var val string
	err := s.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&val)
	if err != nil {
		return defaultVal
	}
	return val
}

// SetSetting saves a setting value to SQLite
func (s *Storage) SetSetting(key string, val string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", key, val)
	return err
}

// ClearAllCache clears cached installed apps and updates tables
func (s *Storage) ClearAllCache() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err1 := s.db.Exec("DELETE FROM installed_apps")
	_, err2 := s.db.Exec("DELETE FROM available_updates")
	if err1 != nil {
		return err1
	}
	return err2
}
