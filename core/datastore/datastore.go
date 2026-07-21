// Package datastore owns the single SQLite connection; the schema is built from the GORM models via AutoMigrate, not SQL migration files.
package datastore

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"smegg.me/smeggtuner/core/session"
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
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("datastore: create %s: %w", filepath.Dir(path), err)
	}

	gormDB, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return fmt.Errorf("datastore: open %s: %w", path, err)
	}

	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	} {
		if err := gormDB.Exec(pragma).Error; err != nil {
			return fmt.Errorf("datastore: %s: %w", pragma, err)
		}
	}

	// SQLite supports a single writer; one open connection avoids SQLITE_BUSY.
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("datastore: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := gormDB.AutoMigrate(allModels()...); err != nil {
		return fmt.Errorf("datastore: migrate: %w", err)
	}

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
