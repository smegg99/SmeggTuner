package session

import (
	"archive/zip"
	"bytes"
	"errors"
	"image"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func stifPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(t.TempDir(), name+InstrumentFileExt)
}

func TestAnInstrumentSurvivesTheRoundTrip(t *testing.T) {
	want := &Template{
		Name: "Castagnari Tommy",
		Instrument: Instrument{
			Banks:     []Bank{BankM1, BankM2},
			Registers: []Register{{Name: "MM", Banks: []Bank{BankM1, BankM2}}},
			ReedCount: 2,
		},
	}
	jpg, err := PrepareImage(bytes.NewReader(photo(t, 900, 600, "jpeg")))
	if err != nil {
		t.Fatal(err)
	}

	path := stifPath(t, "tommy")
	if err := WriteInstrumentFile(path, want, jpg); err != nil {
		t.Fatal(err)
	}

	got, back, err := ReadInstrumentFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != want.Name {
		t.Fatalf("came back as %+v", got)
	}
	if !slices.Equal(got.Instrument.Banks, want.Instrument.Banks) {
		t.Fatalf("banks came back as %v", got.Instrument.Banks)
	}
	if len(got.Instrument.Registers) != 1 || !slices.Equal(got.Instrument.Registers[0].Banks, want.Instrument.Registers[0].Banks) {
		t.Fatalf("registers came back as %+v", got.Instrument.Registers)
	}
	if !got.HasImage || len(back) == 0 {
		t.Fatal("the photograph did not survive")
	}
	if _, _, err := image.Decode(bytes.NewReader(back)); err != nil {
		t.Fatalf("what came back is not an image: %v", err)
	}
}

func TestAnInstrumentNeedNoPhotograph(t *testing.T) {
	path := stifPath(t, "plain")
	if err := WriteInstrumentFile(path, &Template{Name: "Plain", Instrument: Instrument{ReedCount: 1}}, nil); err != nil {
		t.Fatal(err)
	}

	got, jpg, err := ReadInstrumentFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.HasImage || jpg != nil {
		t.Fatal("an instrument nobody photographed came back with a photograph")
	}
}

func TestTheFileIsAZipAnybodyCanOpen(t *testing.T) {
	jpg, _ := PrepareImage(bytes.NewReader(photo(t, 500, 500, "jpeg")))
	path := stifPath(t, "morino")
	if err := WriteInstrumentFile(path, &Template{Name: "Morino", Instrument: Instrument{ReedCount: 3}}, jpg); err != nil {
		t.Fatal(err)
	}

	z, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("it is not a zip: %v", err)
	}
	defer z.Close()

	var names []string
	for _, f := range z.File {
		names = append(names, f.Name)
	}
	slices.Sort(names)
	if !slices.Equal(names, []string{entryImage, entryInstrument, entryManifest}) {
		t.Fatalf("the file holds %v", names)
	}
}

func TestAZipThatIsNotAnInstrumentIsRefused(t *testing.T) {
	path := stifPath(t, "holiday")

	var buf bytes.Buffer
	z := zip.NewWriter(&buf)
	w, _ := z.Create("beach.jpg")
	_, _ = w.Write([]byte("not an accordion"))
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), filePerm); err != nil {
		t.Fatal(err)
	}

	if _, _, err := ReadInstrumentFile(path); !errors.Is(err, ErrNotInstrumentFile) {
		t.Fatalf("read = %v, want %v", err, ErrNotInstrumentFile)
	}
}

func TestSomethingThatIsNotAFileFormatIsRefused(t *testing.T) {
	path := stifPath(t, "junk")
	if err := os.WriteFile(path, []byte("hello"), filePerm); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadInstrumentFile(path); !errors.Is(err, ErrNotInstrumentFile) {
		t.Fatalf("read = %v, want %v", err, ErrNotInstrumentFile)
	}
}

func TestAnInstrumentFileFromTheFutureIsRefused(t *testing.T) {
	path := stifPath(t, "future")

	var buf bytes.Buffer
	z := zip.NewWriter(&buf)
	w, _ := z.Create(entryManifest)
	_, _ = w.Write([]byte(`{"v":99,"kind":"smeggtuner.instrument"}`))
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), filePerm); err != nil {
		t.Fatal(err)
	}

	if _, _, err := ReadInstrumentFile(path); !errors.Is(err, ErrInstrumentFileVer) {
		t.Fatalf("read = %v, want %v", err, ErrInstrumentFileVer)
	}
}

func TestAPhotographInsideAnInstrumentFileIsNotTrusted(t *testing.T) {
	path := stifPath(t, "liar")

	var buf bytes.Buffer
	z := zip.NewWriter(&buf)

	m, _ := z.Create(entryManifest)
	_, _ = m.Write(mustJSON(manifest{V: InstrumentFileVersion, Kind: instrumentFileKind}))

	i, _ := z.Create(entryInstrument)
	_, _ = i.Write([]byte(`{"v":1,"id":"","name":"Morino","instrument":{"reedCount":3}}`))

	// named image.jpg, but not a JPEG
	img, _ := z.Create(entryImage)
	_, _ = img.Write([]byte("gotcha"))

	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), filePerm); err != nil {
		t.Fatal(err)
	}

	if _, _, err := ReadInstrumentFile(path); !errors.Is(err, ErrNotAnImage) {
		t.Fatalf("read = %v, want %v", err, ErrNotAnImage)
	}
}

func TestTheSuggestedNameIsTheInstrumentsOwn(t *testing.T) {
	cases := map[string]string{
		"Hohner Morino V N": "hohner-morino-v-n.stif",
		"  Castagnari  ":    "castagnari.stif",
		"////":              "instrument.stif",
		"":                  "instrument.stif",
	}
	for in, want := range cases {
		if got := SuggestFileName(in); got != want {
			t.Fatalf("SuggestFileName(%q) = %q, want %q", in, got, want)
		}
	}
}
