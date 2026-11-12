// backend/core/datastore/update.go
package datastore

import (
	"errors"
	"fmt"

	"github.com/smegg99/s99wails/v2/shell/update"
	"gorm.io/gorm"
)

type updateMemory struct {
	ID      int `gorm:"primaryKey"`
	Ignored string
	Notes   string
}

func (updateMemory) TableName() string { return "update_memory" }

// UpdateStore persists update preferences in the datastore.
type UpdateStore struct{}

// Load reads remembered versions, reporting false when none have been saved.
func (UpdateStore) Load() (update.Memory, bool, error) {
	var row updateMemory
	err := Get().First(&row, "id = ?", 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return update.Memory{}, false, nil
	}
	if err != nil {
		return update.Memory{}, false, fmt.Errorf("datastore: load update memory: %w", err)
	}
	return update.Memory{Ignored: row.Ignored, Notes: row.Notes}, true, nil
}

// Save replaces both remembered versions in the singleton row.
func (UpdateStore) Save(memory update.Memory) error {
	err := Get().Save(&updateMemory{ID: 1, Ignored: memory.Ignored, Notes: memory.Notes}).Error
	if err != nil {
		return fmt.Errorf("datastore: save update memory: %w", err)
	}
	return nil
}

// HasHistory reports whether saved window geometry or a session exists.
func HasHistory() bool {
	var exists bool
	err := Get().Raw("SELECT EXISTS(SELECT 1 FROM window_state UNION ALL SELECT 1 FROM sessions)").Scan(&exists).Error
	return err == nil && exists
}
