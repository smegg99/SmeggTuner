package datastore

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/smegg99/s99wails/windowstate"
	"gorm.io/gorm"
)

// windowState is the saved main-window geometry (one row); it lives in the database rather than app config because geometry changes on every move/resize.
type windowState struct {
	ID        int `gorm:"primaryKey"`
	X         int
	Y         int
	Width     int
	Height    int
	Maximised bool
}

func (windowState) TableName() string { return "window_state" }

// The table holds exactly one row; this is its key.
const windowStateID = 1

// WindowStore is a windowstate.Store backed by the datastore.
type WindowStore struct{}

// Load reads the saved geometry. A missing row means no state was saved yet.
func (WindowStore) Load() (windowstate.State, bool, error) {
	var row windowState
	err := Get().First(&row, "id = ?", windowStateID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return windowstate.State{}, false, nil
	}
	if err != nil {
		return windowstate.State{}, false, err
	}
	return windowstate.State{
		X:         row.X,
		Y:         row.Y,
		Width:     row.Width,
		Height:    row.Height,
		Maximised: row.Maximised,
	}, true, nil
}

// ImportLegacyWindowState seeds geometry from the pre-database windowstate.json once; an existing row wins over the file, and a missing or unparseable file means no saved state.
func ImportLegacyWindowState(path string) error {
	if _, ok, err := (WindowStore{}).Load(); err != nil || ok {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var state windowstate.State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil
	}
	return (WindowStore{}).Save(state)
}

// Save writes the geometry, replacing any previous one.
func (WindowStore) Save(state windowstate.State) error {
	return Get().Save(&windowState{
		ID:        windowStateID,
		X:         state.X,
		Y:         state.Y,
		Width:     state.Width,
		Height:    state.Height,
		Maximised: state.Maximised,
	}).Error
}
