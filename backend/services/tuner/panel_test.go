package tuner

import (
	"errors"
	"slices"
	"testing"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/core/dsp"
	audiosvc "smegg.me/smeggtuner/backend/services/audio"
)

func freshService(t *testing.T) *Service {
	t.Helper()
	initConfig(t)
	return New(audiosvc.NewTunerInput(audiosvc.New()), nil, nil)
}

func mustOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func wantErr(t *testing.T, err, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
}

// Tones mode reads any non-empty set of reed boxes, and the last box on cannot be cleared.
func TestReedBoxes(t *testing.T) {
	s := freshService(t)
	if c := s.snapshot(); !slices.Equal(c.Reeds, []int{1}) || c.ReedCount != 1 || c.Chords || c.Octaves != nil {
		t.Fatalf("default = reeds %v / count %d / chords %v / octaves %v, want tones on box 1", c.Reeds, c.ReedCount, c.Chords, c.Octaves)
	}
	mustOK(t, s.SetReedBox(3, true))
	mustOK(t, s.SetReedBox(2, true))
	mustOK(t, s.SetReedBox(1, false))
	if c := s.snapshot(); !slices.Equal(c.Reeds, []int{2, 3}) || c.ReedCount != 2 {
		t.Fatalf("reeds %v / count %d, want boxes 2 and 3", c.Reeds, c.ReedCount)
	}
	mustOK(t, s.SetReedBox(2, false))
	wantErr(t, s.SetReedBox(3, false), ErrLastBox)
	wantErr(t, s.SetReedBox(4, true), ErrInvalidBox)
	wantErr(t, s.SetReedBox(0, true), ErrInvalidBox)
	if c := s.snapshot(); !slices.Equal(c.Reeds, []int{3}) {
		t.Fatalf("reeds %v, want box 3 alone", c.Reeds)
	}
	if got := appconfig.Get().Tuner.ReedBoxes; !slices.Equal(got, []int{3}) {
		t.Fatalf("config reed boxes = %v, want [3]", got)
	}
}

// Octaves mode is the base and one or both octaves above it, untransposed, keeping the manual note.
func TestOctavesMode(t *testing.T) {
	s := freshService(t)
	mustOK(t, s.SetTranspose(5))
	mustOK(t, s.SetManualNote(60))
	mustOK(t, s.SetMode("octaves"))

	c := s.snapshot()
	if want := []dsp.OctaveRequest{{Offset: 0, Reeds: 1}, {Offset: 12, Reeds: 1}}; !slices.Equal(c.Octaves, want) {
		t.Fatalf("layout = %+v, want %+v", c.Octaves, want)
	}
	if c.Transpose != 0 || c.ManualNote != 60 || c.Reeds != nil {
		t.Fatalf("transpose %d / note %d / reeds %v, want 0 / 60 / none", c.Transpose, c.ManualNote, c.Reeds)
	}
	mustOK(t, s.SetOctave(24, true))
	mustOK(t, s.SetOctave(12, false))
	if want := []dsp.OctaveRequest{{Offset: 0, Reeds: 1}, {Offset: 24, Reeds: 1}}; !slices.Equal(s.snapshot().Octaves, want) {
		t.Fatalf("layout = %+v, want %+v", s.snapshot().Octaves, want)
	}
	wantErr(t, s.SetOctave(24, false), ErrLastBox)
	wantErr(t, s.SetOctave(36, true), ErrInvalidBox)
	wantErr(t, s.SetTranspose(3), ErrTransposeMode)
	if st := s.Settings(); st.Mode != "octaves" || !slices.Equal(st.OctaveOffsets, []int{24}) || st.ReedCount != 2 || st.Transpose != 0 {
		t.Fatalf("settings = %+v", st)
	}
	if got := appconfig.Get().Tuner; got.Mode != "octaves" || !slices.Equal(got.OctaveOffsets, []int{24}) || got.Transpose != 5 {
		t.Fatalf("config = mode %q / octaves %v / transpose %d, want octaves / [24] / the saved 5", got.Mode, got.OctaveOffsets, got.Transpose)
	}

	// Back in tones the saved transpose returns.
	mustOK(t, s.SetMode("tones"))
	if c := s.snapshot(); c.Transpose != 5 || c.Octaves != nil || c.ManualNote != 60 {
		t.Fatalf("tones = transpose %d / octaves %v / note %d, want 5 / none / 60", c.Transpose, c.Octaves, c.ManualNote)
	}
}

// Chords pick their own notes: the manual note is cleared and stays refused, and nothing is transposed.
func TestChordsMode(t *testing.T) {
	s := freshService(t)
	mustOK(t, s.SetTranspose(2))
	mustOK(t, s.SetManualNote(64))
	mustOK(t, s.SetMode("chords"))

	c := s.snapshot()
	if !c.Chords || c.ManualNote != 0 || c.Transpose != 0 || c.ReedCount != 3 {
		t.Fatalf("chords = %v / note %d / transpose %d / count %d, want on / 0 / 0 / 3", c.Chords, c.ManualNote, c.Transpose, c.ReedCount)
	}
	wantErr(t, s.SetManualNote(64), ErrChordsAuto)
	_, err := s.StepNote(1)
	wantErr(t, err, ErrChordsAuto)
	_, err = s.SetAuto(false)
	wantErr(t, err, ErrChordsAuto)
	wantErr(t, s.SetTranspose(1), ErrTransposeMode)
	wantErr(t, s.SetMode("bass"), ErrInvalidMode)

	// The note stays cleared when tones return; the transpose comes back.
	mustOK(t, s.SetMode("tones"))
	if c := s.snapshot(); c.Chords || c.ManualNote != 0 || c.Transpose != 2 {
		t.Fatalf("tones = chords %v / note %d / transpose %d, want off / 0 / 2", c.Chords, c.ManualNote, c.Transpose)
	}
}

// The panel survives a restart of the app: a new service reads it back from the config file.
func TestPanelPersists(t *testing.T) {
	s := freshService(t)
	mustOK(t, s.SetReedBox(2, true))
	mustOK(t, s.SetOctave(24, true))
	mustOK(t, s.SetMode("octaves"))

	again := New(audiosvc.NewTunerInput(audiosvc.New()), nil, nil)
	st := again.Settings()
	if st.Mode != "octaves" || !slices.Equal(st.ReedBoxes, []int{1, 2}) || !slices.Equal(st.OctaveOffsets, []int{12, 24}) {
		t.Fatalf("settings after a restart = %+v", st)
	}
	if st.ReedCount != 3 {
		t.Fatalf("reed count = %d, want the base and both octaves", st.ReedCount)
	}
}

// Every setting change tells the UI what the engine now measures.
func TestSettersAnnounce(t *testing.T) {
	s := freshService(t)
	var got []SettingsDTO
	prev := emitEvent
	emitEvent = func(name string, data any) {
		if dto, ok := data.(SettingsDTO); ok && name == EventSettings {
			got = append(got, dto)
		}
	}
	t.Cleanup(func() { emitEvent = prev })

	mustOK(t, s.SetReedBox(2, true))
	if len(got) != 1 || got[0].ReedCount != 2 {
		t.Fatalf("settings events = %+v, want one with two reeds", got)
	}
	wantErr(t, s.SetReedBox(9, true), ErrInvalidBox)
	if len(got) != 1 {
		t.Fatal("a refused change must not announce")
	}
}
