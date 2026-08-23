// backend/core/datastore/datastore.go
// Package datastore owns the single SQLite connection; the schema is built from the GORM models via AutoMigrate, not SQL migration files.
package datastore

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"smegg.me/smeggtuner/backend/core/session"
)

var (
	mu     sync.Mutex
	db     *gorm.DB
	dbPath string
)

func allModels() []any {
	return []any{
		&session.Session{},
		&session.Template{},
		&windowState{},
	}
}

// Initialize opens the database at path, sets the SQLite pragmas, and builds the schema; it errors if a database is already open.
func Initialize(path string) error {
	mu.Lock()
	defer mu.Unlock()

	if db != nil {
		return fmt.Errorf("datastore: already initialised at %s", dbPath)
	}
	dsn, err := connectionDSN(path)
	if err != nil {
		return fmt.Errorf("datastore: open %s: %w", path, err)
	}
	diskPath, _, _ := strings.Cut(path, "?")
	if strings.HasPrefix(diskPath, "file:") {
		uri, err := url.Parse(diskPath)
		if err != nil {
			return fmt.Errorf("datastore: open %s: %w", path, err)
		}
		diskPath = uri.Path
		if diskPath == "" {
			diskPath = uri.Opaque
		}
	}
	if err := os.MkdirAll(filepath.Dir(diskPath), 0o755); err != nil {
		return fmt.Errorf("datastore: create %s: %w", filepath.Dir(diskPath), err)
	}

	sqlDB, err := sql.Open(sqlite.DriverName, dsn)
	if err != nil {
		return fmt.Errorf("datastore: open %s: %w", path, err)
	}
	ready := false
	defer func() {
		if !ready {
			_ = sqlDB.Close()
		}
	}()
	sqlDB.SetMaxOpenConns(1)

	gormDB, err := gorm.Open(sqlite.New(sqlite.Config{Conn: sqlDB}), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return fmt.Errorf("datastore: open %s: %w", path, err)
	}

	if err := gormDB.Exec("PRAGMA journal_mode=WAL").Error; err != nil {
		return fmt.Errorf("datastore: PRAGMA journal_mode=WAL: %w", err)
	}

	if err := gormDB.AutoMigrate(allModels()...); err != nil {
		return fmt.Errorf("datastore: migrate: %w", err)
	}

	ready = true
	db = gormDB
	dbPath = path
	return nil
}

// Get returns the shared connection; it panics if Initialize has not run.
func Get() *gorm.DB {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		panic("datastore: not initialised, call Initialize first")
	}
	return db
}

// Path returns the database file location, or "" before Initialize.
func Path() string {
	mu.Lock()
	defer mu.Unlock()
	return dbPath
}

// Close releases the connection. Safe to call when nothing is open.
func Close() error {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	db = nil
	dbPath = ""
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping reports whether the database answers.
func Ping() error {
	sqlDB, err := Get().DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// connectionDSN reapplies connection-local pragmas when database/sql replaces a connection.
func connectionDSN(path string) (string, error) {
	base, query, _ := strings.Cut(path, "?")
	values, err := url.ParseQuery(query)
	if err != nil {
		return "", err
	}
	values.Del("_fk")
	values.Del("_timeout")
	values.Set("_foreign_keys", "on")
	values.Set("_busy_timeout", "5000")
	return base + "?" + values.Encode(), nil
}
