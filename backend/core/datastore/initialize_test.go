// backend/core/datastore/initialize_test.go
package datastore

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/smegg99/s99wails/v2/shell/windowstate"

	"smegg.me/smeggtuner/backend/core/session"
)

func TestInitializeFailureClosesConnections(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("connection ownership check uses /proc/self/fd")
	}
	for _, kind := range []string{"corrupt", "pragma", "migration", "directory", "invalid DSN"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "data.db")
			dsn := path
			switch kind {
			case "corrupt":
				if err := os.WriteFile(path, []byte("not a SQLite database"), 0600); err != nil {
					t.Fatal(err)
				}
			case "pragma", "migration":
				conn, err := sql.Open("sqlite3", path)
				if err != nil {
					t.Fatal(err)
				}
				query := "CREATE TABLE keep (id INTEGER)"
				if kind == "migration" {
					query = "CREATE VIEW sessions AS SELECT 1 AS id"
				}
				if _, err = conn.Exec(query); err != nil {
					t.Fatal(err)
				}
				if err = conn.Close(); err != nil {
					t.Fatal(err)
				}
				if kind == "pragma" {
					dsn = "file:" + path + "?mode=ro"
				}
			case "directory":
				dsn = dir
			case "invalid DSN":
				dsn = path + "?bad=%GG"
			}
			before := databaseHandles(t, dir)
			for range 4 {
				if err := Initialize(dsn); err == nil {
					_ = Close()
					t.Fatal("invalid database was accepted")
				}
				if db != nil || Path() != "" {
					t.Fatal("failed initialization published datastore")
				}
				if got := databaseHandles(t, dir); got != before {
					t.Fatalf("connection leak: descriptors before=%d after=%d", before, got)
				}
			}
			if err := Initialize(filepath.Join(dir, "retry.db")); err != nil {
				t.Fatalf("retry: %v", err)
			}
			if err := Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func databaseHandles(t *testing.T, dir string) int {
	t.Helper()
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range entries {
		target, _ := os.Readlink(filepath.Join("/proc/self/fd", entry.Name()))
		if strings.HasPrefix(target, dir+"/") {
			count++
		}
	}
	return count
}

func TestInitializeKeepsPersistedModelsAndDriverErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")
	if err := Initialize(path); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close() })
	at := time.Date(2026, 10, 26, 12, 13, 14, 123456789, time.FixedZone("test", 7200))
	sn := session.New("persisted", session.Instrument{ReedCount: 1}, 442)
	sn.Created, sn.Updated, sn.Notes = at, at, "kept"
	if err := Get().Create(sn).Error; err != nil {
		t.Fatal(err)
	}
	photo := session.Template{ID: session.NewID(), Name: "accordion", Instrument: sn.Instrument, Image: []byte{0, 1, 255}, HasImage: true, ImageRev: 42}
	if err := Get().Create(&photo).Error; err != nil {
		t.Fatal(err)
	}
	geometry := windowstate.State{X: 10, Y: 20, Width: 900, Height: 700, Maximised: true}
	if err := (WindowStore{}).Save(geometry); err != nil {
		t.Fatal(err)
	}
	err := Get().Create(sn).Error
	var native sqlite3.Error
	if !errors.As(err, &native) || native.ExtendedCode != sqlite3.ErrConstraintPrimaryKey {
		t.Fatalf("duplicate error: %T %v", err, err)
	}
	if err := Close(); err != nil {
		t.Fatal(err)
	}
	if err := Initialize(path); err != nil {
		t.Fatal(err)
	}
	var back session.Session
	if err := Get().First(&back, "id = ?", sn.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !back.Created.Equal(at) || back.Notes != "kept" || back.Instrument.ReedCount != 1 || back.A4 != 442 {
		t.Fatalf("session changed: %+v", back)
	}
	var restored session.Template
	if err := Get().First(&restored, "id = ?", photo.ID).Error; err != nil {
		t.Fatal(err)
	}
	if string(restored.Image) != string(photo.Image) || !restored.HasImage || restored.ImageRev != 42 {
		t.Fatalf("image changed: %+v", restored)
	}
	g, ok, err := (WindowStore{}).Load()
	if err != nil || !ok || g != geometry {
		t.Fatalf("geometry: %+v %v %v", g, ok, err)
	}
	var enabled int
	if err := Get().Raw("PRAGMA foreign_keys").Scan(&enabled).Error; err != nil || enabled != 1 {
		t.Fatalf("foreign keys: %d %v", enabled, err)
	}
}

func TestReplacementConnectionRetainsPragmas(t *testing.T) {
	if err := Initialize(filepath.Join(t.TempDir(), "data.db")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close() })
	conn, err := Get().DB()
	if err != nil {
		t.Fatal(err)
	}
	conn.SetMaxIdleConns(0)
	for _, pragma := range []struct {
		name string
		want int
	}{{"foreign_keys", 1}, {"busy_timeout", 5000}} {
		var got int
		if err := conn.QueryRow("PRAGMA " + pragma.name).Scan(&got); err != nil || got != pragma.want {
			t.Fatalf("replacement %s = %d, %v", pragma.name, got, err)
		}
	}
}
