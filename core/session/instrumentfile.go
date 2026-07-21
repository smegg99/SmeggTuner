package session

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// The .stif (SmeggTuner Instrument File) packs an accordion and its photo into one zip. The on-disk
// library is plain .json + .jpg (see template.go); .stif is only the import/export format.

// InstrumentFileVersion is stamped into every .stif; an unknown version is refused.
const InstrumentFileVersion = 1

// InstrumentFileExt is the instrument export extension (a zip, not JSON).
const InstrumentFileExt = ".stif"

// instrumentFileKind is stored in the manifest so a renamed foreign zip is refused.
const instrumentFileKind = "smeggtuner.instrument"

const (
	entryInstrument = "instrument.json"
	entryImage      = "image.jpg"
)

var (
	ErrNotInstrumentFile = errors.New("session: that is not an instrument file")
	ErrInstrumentFileVer = errors.New("session: unsupported instrument file version")
)

// WriteInstrumentFile writes an accordion as a .stif; a nil photo is fine, and the write is atomic.
func WriteInstrumentFile(path string, t *Template, jpg []byte) error {
	out := *t
	out.HasImage = jpg != nil
	out.ImageRev = 0 // a mod time here means nothing on another machine
	out.ID = ""      // the id stays with the shelf; Templates.Import replaces it anyway

	var buf bytes.Buffer
	z := zip.NewWriter(&buf)

	if err := writeEntry(z, entryManifest, mustJSON(manifest{V: InstrumentFileVersion, Kind: instrumentFileKind})); err != nil {
		return err
	}

	body, err := json.MarshalIndent(templateFile{V: TemplateVersion, Template: out}, "", "  ")
	if err != nil {
		return err
	}
	if err := writeEntry(z, entryInstrument, append(body, '\n')); err != nil {
		return err
	}

	if jpg != nil {
		// Stored, not deflated: a JPEG is already compressed.
		w, err := z.CreateHeader(&zip.FileHeader{Name: entryImage, Method: zip.Store})
		if err != nil {
			return err
		}
		if _, err := w.Write(jpg); err != nil {
			return err
		}
	}

	if err := z.Close(); err != nil {
		return err
	}
	return writeFileAtomic(path, buf.Bytes())
}

// ReadInstrumentFile reads a .stif: the accordion and its photo if present; the photo is decoded, not trusted.
func ReadInstrumentFile(path string) (*Template, []byte, error) {
	z, err := zip.OpenReader(path)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrNotInstrumentFile, err)
	}
	defer z.Close()

	man, err := readEntry(&z.Reader, entryManifest, 4<<10)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: no manifest", ErrNotInstrumentFile)
	}
	var m manifest
	if err := json.Unmarshal(man, &m); err != nil || m.Kind != instrumentFileKind {
		return nil, nil, fmt.Errorf("%w: %q", ErrNotInstrumentFile, path)
	}
	if m.V != InstrumentFileVersion {
		return nil, nil, fmt.Errorf("%w: %d", ErrInstrumentFileVer, m.V)
	}

	body, err := readEntry(&z.Reader, entryInstrument, 1<<20)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: no instrument in it", ErrNotInstrumentFile)
	}
	t, err := decodeTemplate(body)
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", path, err)
	}

	// No photo is an ordinary instrument file, not a broken one.
	f, err := open(&z.Reader, entryImage)
	if err != nil {
		t.HasImage = false
		return t, nil, nil
	}
	defer f.Close()

	jpg, err := PrepareImage(f)
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", path, err)
	}
	t.HasImage = true
	return t, jpg, nil
}

// SuggestFileName suggests an export file name from the instrument's own name.
func SuggestFileName(name string) string {
	return suggestBase(name, "instrument") + InstrumentFileExt
}
