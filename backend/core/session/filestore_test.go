package session

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestATruncatedSessionFileIsRefused(t *testing.T) {
	s := recorded(t, 3)
	path := stsfPath(t, "morino")
	if err := WriteSessionFile(path, s); err != nil {
		t.Fatalf("write: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data[:len(data)/2], filePerm); err != nil {
		t.Fatal(err)
	}

	if _, err := ReadSessionFile(path); !errors.Is(err, ErrNotSessionFile) {
		t.Fatalf("read = %v, want %v", err, ErrNotSessionFile)
	}
}

func TestExportLeavesNoTempFiles(t *testing.T) {
	s := recorded(t, 3)
	dir := t.TempDir()
	path := filepath.Join(dir, "morino"+SessionFileExt)
	for i := 0; i < 5; i++ {
		if err := WriteSessionFile(path, s); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || !strings.HasSuffix(entries[0].Name(), SessionFileExt) {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("directory holds %v, want one session file", names)
	}
}

func TestWriteIsAtomicUnderConcurrentReads(t *testing.T) {
	s := recorded(t, 5)
	path := stsfPath(t, "morino")
	if err := WriteSessionFile(path, s); err != nil {
		t.Fatalf("write: %v", err)
	}

	var wg sync.WaitGroup
	done := make(chan struct{})
	var readErr error
	var mu sync.Mutex

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
			}
			if _, err := ReadSessionFile(path); err != nil {
				mu.Lock()
				readErr = err
				mu.Unlock()
				return
			}
		}
	}()

	for i := 0; i < 50; i++ {
		s.Notes = strings.Repeat("x", i*64) // vary the size so a partial read would show
		if err := WriteSessionFile(path, s); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
	}
	close(done)
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if readErr != nil {
		t.Fatalf("a reader saw a half-written session: %v", readErr)
	}
}
