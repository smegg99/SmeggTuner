package session

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// A .stsf (SmeggTuner Session File) is a zip, on the same terms as the .stif; import also accepts the bare legacy .session.json.

// SessionFileVersion is stamped into every .stsf; an unknown version is refused.
const SessionFileVersion = 1

// SessionFileExt is the SmeggTuner Session File extension.
const SessionFileExt = ".stsf"

// sessionFileKind is written into the manifest so a renamed zip of something else is refused.
const sessionFileKind = "smeggtuner.session"

const entrySession = "session.json"

var (
	ErrNotSessionFile = fmt.Errorf("session: that is not a session file")
	ErrSessionFileVer = fmt.Errorf("session: unsupported session file version")
)

// WriteSessionFile writes a session out as a .stsf, atomically via temp-and-rename.
func WriteSessionFile(path string, s *Session) error {
	if err := s.Validate(); err != nil {
		return err
	}

	var buf bytes.Buffer
	z := zip.NewWriter(&buf)

	if err := writeEntry(z, entryManifest, mustJSON(manifest{V: SessionFileVersion, Kind: sessionFileKind})); err != nil {
		return err
	}

	body, err := json.MarshalIndent(sessionFile{V: Version, Session: *s}, "", "  ")
	if err != nil {
		return err
	}
	if err := writeEntry(z, entrySession, append(body, '\n')); err != nil {
		return err
	}

	if err := z.Close(); err != nil {
		return err
	}
	return writeFileAtomic(path, buf.Bytes())
}

// ReadSessionFile reads a .stsf, checking the manifest before decoding the session.
func ReadSessionFile(path string) (*Session, error) {
	z, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotSessionFile, err)
	}
	defer z.Close()

	man, err := readEntry(&z.Reader, entryManifest, 4<<10)
	if err != nil {
		return nil, fmt.Errorf("%w: no manifest", ErrNotSessionFile)
	}
	var m manifest
	if err := json.Unmarshal(man, &m); err != nil || m.Kind != sessionFileKind {
		return nil, fmt.Errorf("%w: %q", ErrNotSessionFile, path)
	}
	if m.V != SessionFileVersion {
		return nil, fmt.Errorf("%w: %d", ErrSessionFileVer, m.V)
	}

	// Generous next to a real session, tight next to a zip bomb.
	body, err := readEntry(&z.Reader, entrySession, 16<<20)
	if err != nil {
		return nil, fmt.Errorf("%w: no session in it", ErrNotSessionFile)
	}
	s, err := decodeSession(body)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return s, nil
}

// ReadSessionAny reads a session from either format: the .stsf or the bare legacy JSON.
func ReadSessionAny(path string) (*Session, error) {
	if strings.HasSuffix(strings.ToLower(path), SessionFileExt) {
		return ReadSessionFile(path)
	}
	return Read(path)
}

// SuggestSessionFileName is the suggested filename for an exported session.
func SuggestSessionFileName(name string) string {
	return suggestBase(name, "session") + SessionFileExt
}
