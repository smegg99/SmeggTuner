// backend/common/config/memory_test.go
package config

import "testing"

func TestSeedEmptyGivesDefaults(t *testing.T) {
	c, err := decodeSeed(nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.Tuner.A4 != 440 || c.Preferences.Language != "en" {
		t.Fatalf("defaults A4 %v language %q, want 440 and en", c.Tuner.A4, c.Preferences.Language)
	}
}

func TestSeedInvalidGivesDefaults(t *testing.T) {
	for name, raw := range map[string]string{"not json": "{", "wrong type": `{"tuner":{"a4":"loud"}}`} {
		t.Run(name, func(t *testing.T) {
			c, err := decodeSeed([]byte(raw))
			if err == nil {
				t.Fatal("a rejected seed must report its error")
			}
			if c.Tuner.A4 != 440 {
				t.Fatalf("A4 = %v after a rejected seed, want the default 440", c.Tuner.A4)
			}
		})
	}
}

func TestSeedValidIsKeptOverDefaults(t *testing.T) {
	c, err := decodeSeed([]byte(`{"tuner":{"a4":442}}`))
	if err != nil {
		t.Fatal(err)
	}
	if c.Tuner.A4 != 442 || c.Preferences.Language != "en" {
		t.Fatalf("A4 %v language %q, want the stored 442 over the default en", c.Tuner.A4, c.Preferences.Language)
	}
}

func TestPatchConfigKeepsUnrelatedFields(t *testing.T) {
	c, _ := decodeSeed([]byte(`{"tuner":{"a4":442}}`))
	next, err := patchConfig(c, map[string]any{"preferences": map[string]any{"language": "pl"}})
	if err != nil {
		t.Fatal(err)
	}
	if next.Tuner.A4 != 442 || next.Preferences.Language != "pl" {
		t.Fatalf("patched A4 %v language %q, want 442 kept and pl set", next.Tuner.A4, next.Preferences.Language)
	}
}
