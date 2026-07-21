package tuner

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	appconfig "smegg.me/smeggtuner/common/config"
	audiosvc "smegg.me/smeggtuner/services/audio"
)

// These tests initialize common/config, which is process-global state.

// Wails runs SetA4 and SetFilters concurrently; persist once read/mutated/wrote a copy with no lock, so the last writer committed a copy that never saw the other setter's field and a setting silently reverted on disk.
func TestConcurrentSettersDoNotLoseConfigFields(t *testing.T) {
	initConfig(t)
	s := New(audiosvc.New(), nil, nil)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if err := s.SetA4(442); err != nil {
				t.Error(err)
			}
		}()
		go func() {
			defer wg.Done()
			if err := s.SetFilters(true, true); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()

	c := appconfig.Get()
	if c.Tuner.A4 != 442 {
		t.Fatalf("config A4 = %v, want 442: a concurrent SetFilters wrote back the A4 it had read", c.Tuner.A4)
	}
	if !c.Audio.HumFilter50 || !c.Audio.HumFilter60 {
		t.Fatalf("config hum filters = %v/%v, want both on: a concurrent SetA4 wrote back the filters it had read",
			c.Audio.HumFilter50, c.Audio.HumFilter60)
	}
}

// A failed config write must not come back as a silent revert: Start re-reads the config, so a value the engine has but the file never received would roll back at the next device swap, and the next successful write would commit the stale value.
func TestFailedPersistDoesNotRevert(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root writes read-only files, so the config write cannot be made to fail")
	}
	initConfig(t)
	s := New(audiosvc.New(), nil, nil)

	if err := s.SetA4(445); err != nil {
		t.Fatal(err)
	}
	if got := appconfig.Get().Tuner.A4; got != 445 {
		t.Fatalf("config A4 = %v, want 445", got)
	}

	// Make the config unwritable, both through itself and its directory (the writer may replace it).
	path := appconfig.GetConfigPath()
	dir := filepath.Dir(path)
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o700)
		_ = os.Chmod(path, 0o600)
	})
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}

	// The setter still succeeds: the engine has the value, so reporting failure to the UI would be wrong.
	if err := s.SetA4(448); err != nil {
		t.Fatal(err)
	}
	if got := appconfig.Get().Tuner.A4; got != 445 {
		t.Fatalf("config A4 = %v, want the write to have failed at 445 - this test cannot prove anything otherwise", got)
	}
	if got := s.snapshot().A4; got != 448 {
		t.Fatalf("snapshot A4 = %v, want 448: a failed write must not be re-read as truth at the next Start", got)
	}

	// The next write repairs the file instead of committing the stale value.
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := s.SetFilters(true, false); err != nil {
		t.Fatal(err)
	}
	c := appconfig.Get()
	if c.Tuner.A4 != 448 || !c.Audio.HumFilter50 {
		t.Fatalf("config = A4 %v / hum50 %v, want 448 / true: the write after a failed one must carry the pending value",
			c.Tuner.A4, c.Audio.HumFilter50)
	}
	if got := s.snapshot().A4; got != 448 {
		t.Fatalf("snapshot A4 = %v, want 448 once the config agrees again", got)
	}
}

// snapshot is a read: PlayTone calls it, and a reference tone must not move the service's idea of the engine config.
func TestSnapshotDoesNotWrite(t *testing.T) {
	initConfig(t)
	s := New(audiosvc.New(), nil, nil)
	if err := s.SetA4(447); err != nil {
		t.Fatal(err)
	}
	// A settings-page write the service has not adopted yet.
	c := *appconfig.Get()
	c.Tuner.A4 = 433
	if err := appconfig.SetConfig(c); err != nil {
		t.Fatal(err)
	}

	if got := s.snapshot().A4; got != 433 {
		t.Fatalf("snapshot A4 = %v, want the config's 433", got)
	}
	s.mu.Lock()
	stored := s.cfg.A4
	s.mu.Unlock()
	if stored != 447 {
		t.Fatalf("stored A4 = %v, want 447: snapshot must not write", stored)
	}
	if got := s.refresh().A4; got != 433 {
		t.Fatalf("refresh A4 = %v, want 433", got)
	}
	s.mu.Lock()
	stored = s.cfg.A4
	s.mu.Unlock()
	if stored != 433 {
		t.Fatalf("stored A4 = %v, want 433: refresh is what adopts the config", stored)
	}
}
