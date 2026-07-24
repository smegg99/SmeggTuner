package session

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

// photo is a real, decodable image of the given size, so "is this an image" is answered by decoding it.
func photo(t *testing.T, w, h int, format string) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := range w {
		for y := range h {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 90, A: 255})
		}
	}

	var buf bytes.Buffer
	var err error
	if format == "png" {
		err = png.Encode(&buf, img)
	} else {
		err = jpeg.Encode(&buf, img, nil)
	}
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestAPhotographIsCappedOnTheWayIn(t *testing.T) {
	jpg, err := PrepareImage(bytes.NewReader(photo(t, 4000, 3000, "jpeg")))
	if err != nil {
		t.Fatal(err)
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(jpg))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Width != ImageMaxEdge {
		t.Fatalf("a 4000px photograph was stored at %dpx, want %d", cfg.Width, ImageMaxEdge)
	}
	// And it kept its shape.
	if want := 3000 * ImageMaxEdge / 4000; cfg.Height != want {
		t.Fatalf("height = %d, want %d: the photograph was distorted", cfg.Height, want)
	}
}

func TestASmallPhotographIsLeftAlone(t *testing.T) {
	jpg, err := PrepareImage(bytes.NewReader(photo(t, 800, 600, "jpeg")))
	if err != nil {
		t.Fatal(err)
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(jpg))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Width != 800 || cfg.Height != 600 {
		t.Fatalf("an 800x600 photograph came back %dx%d", cfg.Width, cfg.Height)
	}
}

// A PNG goes in and a JPEG comes out.
func TestAPhotographIsAlwaysStoredAsOneThing(t *testing.T) {
	jpg, err := PrepareImage(bytes.NewReader(photo(t, 400, 400, "png")))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := jpeg.Decode(bytes.NewReader(jpg)); err != nil {
		t.Fatalf("a PNG was not stored as a JPEG: %v", err)
	}
}

// A file is proved to be an image by decoding it, never by its name.
func TestSomethingThatIsNotAPhotographIsRefused(t *testing.T) {
	_, err := PrepareImage(bytes.NewReader([]byte("this is not an accordion")))
	if !errors.Is(err, ErrNotAnImage) {
		t.Fatalf("PrepareImage = %v, want %v", err, ErrNotAnImage)
	}
}

func TestATemplateNeedsAName(t *testing.T) {
	tpl := &Template{Name: "  ", Instrument: Instrument{ReedCount: 3}}
	if err := tpl.Validate(); !errors.Is(err, ErrTemplateName) {
		t.Fatalf("validate = %v, want %v", err, ErrTemplateName)
	}
}

// FromSession keeps the model, not the one accordion: the serial is left behind, and it is a copy.
func TestSavingTheInstrumentOnTheBenchKeepsTheModelAndNotTheAccordion(t *testing.T) {
	s := New("Jan K.", Instrument{
		Name:      "Hohner Morino",
		Serial:    "12345",
		Banks:     []Bank{BankM1, BankM2, BankM3},
		Registers: []Register{{Name: "MMM", Banks: []Bank{BankM1, BankM2, BankM3}}},
		ReedCount: 3,
	}, 442)

	tpl := FromSession(s, "")
	if tpl.Name != "Hohner Morino" {
		t.Fatalf("name = %q, want it to fall back to the accordion's own name", tpl.Name)
	}
	if tpl.Instrument.Serial != "" {
		t.Fatal("the template kept the serial number of one particular accordion")
	}

	// And it is a copy: editing the template must not reach back into the session.
	tpl.Instrument.Banks[0] = BankL
	tpl.Instrument.Registers[0].Banks[0] = BankL
	if s.Instrument.Banks[0] != BankM1 || s.Instrument.Registers[0].Banks[0] != BankM1 {
		t.Fatal("the template shares its banks with the session it came from")
	}
}

// The saved instrument keeps the model's pitch, not the one accordion's serial.
func TestSavingTheBenchInstrumentKeepsItsPitch(t *testing.T) {
	s := New("Jan K.", Instrument{Serial: "12345", ReedCount: 3, A4: 442}, 442)
	tpl := FromSession(s, "")
	if tpl.Instrument.A4 != 442 {
		t.Fatalf("A4 = %v, want 442", tpl.Instrument.A4)
	}
	if tpl.Instrument.Serial != "" {
		t.Fatal("the template kept a serial")
	}
}

// A negative pitch is refused; zero is "not said" and falls back to the app default.
func TestAReferencePitchIsAFrequencyOrNothing(t *testing.T) {
	if err := (Instrument{ReedCount: 1, A4: -1}).validate(); !errors.Is(err, ErrA4) {
		t.Fatalf("validate = %v, want %v", err, ErrA4)
	}
	if err := (Instrument{ReedCount: 1, A4: 0}).validate(); err != nil {
		t.Fatalf("an instrument with no stated pitch was refused: %v", err)
	}
}

// The reference pitch survives the trip between benches in a .stif.
func TestAnInstrumentFileKeepsItsReferencePitch(t *testing.T) {
	mine := &Template{
		Name:       "Hohner, old",
		Instrument: Instrument{ReedCount: 3, Banks: []Bank{BankM1, BankM2, BankM3}, A4: 442},
	}
	path := t.TempDir() + "/old" + InstrumentFileExt
	if err := WriteInstrumentFile(path, mine, nil); err != nil {
		t.Fatal(err)
	}
	back, _, err := ReadInstrumentFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if back.Instrument.A4 != 442 {
		t.Fatalf("A4 came back as %v", back.Instrument.A4)
	}
}
