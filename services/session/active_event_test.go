// main.go's record light and tray badge hang off EventActive; a broken listener fails silently,
// so these tests pin the emit.
package session

import (
	"testing"
)

func captureEvents(t *testing.T) *[]string {
	t.Helper()
	var names []string
	prev := emitEvent
	emitEvent = func(name string, _ any) { names = append(names, name) }
	t.Cleanup(func() { emitEvent = prev })
	return &names
}

func TestOpeningASessionAnnouncesIt(t *testing.T) {
	s := service(t)
	dto := create(t, s, "Morino", 2, 440)

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	names := captureEvents(t)
	if _, err := s.Open(dto.ID); err != nil {
		t.Fatal(err)
	}

	if !contains(*names, EventActive) {
		t.Fatalf("Open emitted %v, want %q among them: nothing would tell the record light",
			*names, EventActive)
	}
}

func TestCreatingASessionAnnouncesIt(t *testing.T) {
	s := service(t)

	names := captureEvents(t)
	create(t, s, "Morino", 2, 440)

	if !contains(*names, EventActive) {
		t.Fatalf("Create emitted %v, want %q among them", *names, EventActive)
	}
}

func TestClosingASessionAnnouncesIt(t *testing.T) {
	s := service(t)
	create(t, s, "Morino", 2, 440)

	names := captureEvents(t)
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	if !contains(*names, EventActive) {
		t.Fatalf("Close emitted %v, want %q among them", *names, EventActive)
	}
}

func contains(all []string, want string) bool {
	for _, s := range all {
		if s == want {
			return true
		}
	}
	return false
}
