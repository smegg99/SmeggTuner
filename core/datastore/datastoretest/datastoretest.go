// Package datastoretest opens a throwaway database for a test; it is a separate package so production code never links testing.
package datastoretest

import (
	"path/filepath"
	"testing"

	"smegg.me/smeggtuner/core/datastore"
)

// Init points the shared datastore at a fresh temporary database and tears it down with the test.
func Init(t testing.TB) {
	t.Helper()
	if err := datastore.Close(); err != nil {
		t.Fatalf("close previous datastore: %v", err)
	}
	if err := datastore.Initialize(filepath.Join(t.TempDir(), "test.db")); err != nil {
		t.Fatalf("open test datastore: %v", err)
	}
	t.Cleanup(func() { _ = datastore.Close() })
}
