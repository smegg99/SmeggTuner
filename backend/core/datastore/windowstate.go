package datastore

import (
	"errors"

	"github.com/smegg99/s99wails/v2/shell/windowstate"
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
