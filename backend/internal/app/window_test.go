// backend/internal/app/window_test.go
package app

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/smegg99/s99wails/v2/shell/windowstate"
	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/backend/core/datastore"
)

func TestApplicationOptionsKeepProductionDiagnostics(t *testing.T) {
	if applicationOptions(nil).Logger == nil {
		t.Fatal("Wails defaults to discarding production diagnostics without an explicit logger")
	}
}

func TestWindowOptionsRestorePersistedGeometryBeforeCreation(t *testing.T) {
	if err := datastore.Initialize(filepath.Join(t.TempDir(), "windows.db")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := datastore.Close(); err != nil {
			t.Error(err)
		}
	})
	store := datastore.WindowStore{}
	initial := windowOptions(store)
	if initial.Width != 1100 || initial.Height != 720 {
		t.Fatalf("missing state changed default size: %+v", initial)
	}
	for _, state := range []windowstate.State{
		{X: 125, Y: 80, Width: 1000, Height: 680},
		{X: -1100, Y: 40, Width: 1200, Height: 800, Maximised: true},
	} {
		if err := store.Save(state); err != nil {
			t.Fatal(err)
		}
		opts := windowOptions(store)
		if opts.Width != state.Width || opts.Height != state.Height || opts.X != state.X || opts.Y != state.Y {
			t.Fatalf("saved geometry not adopted before creation: got %dx%d at %d,%d; want %+v", opts.Width, opts.Height, opts.X, opts.Y, state)
		}
		wantStart := application.WindowStateNormal
		if state.Maximised {
			wantStart = application.WindowStateMaximised
		}
		if opts.InitialPosition != application.WindowXY || opts.StartState != wantStart {
			t.Fatalf("saved placement/start state not adopted: position=%v state=%v", opts.InitialPosition, opts.StartState)
		}
		if opts.MinWidth != 960 || opts.MinHeight != 640 {
			t.Fatalf("restoration lost minimum limits: %dx%d", opts.MinWidth, opts.MinHeight)
		}
		saved, ok, err := store.Load()
		if err != nil || !ok || saved != state {
			t.Fatalf("configuration overwrote stored geometry: %+v, %v", saved, err)
		}
	}
}

type unreadableWindowStore struct{}

func (unreadableWindowStore) Load() (windowstate.State, bool, error) {
	return windowstate.State{}, false, errors.New("injected window-state load failure")
}

func (unreadableWindowStore) Save(windowstate.State) error {
	panic("configuring window options must not save state")
}

func TestWindowOptionsKeepDefaultsWhenStoreUnavailable(t *testing.T) {
	for _, store := range []windowstate.Store{nil, unreadableWindowStore{}} {
		opts := windowOptions(store)
		if opts.Width != 1100 || opts.Height != 720 || opts.MinWidth != 960 || opts.MinHeight != 640 {
			t.Fatalf("unavailable store changed default size or minimum limits: %+v", opts)
		}
	}
}
