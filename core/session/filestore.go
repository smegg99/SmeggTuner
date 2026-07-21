package session

import (
	"archive/zip"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// entryManifest is checked before a container's contents are trusted.
const entryManifest = "manifest.json"

// manifest is what a container claims to be.
type manifest struct {
	V    int    `json:"v"`
	Kind string `json:"kind"`
}

func writeEntry(z *zip.Writer, name string, body []byte) error {
	w, err := z.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(body)
	return err
}

func open(z *zip.Reader, name string) (io.ReadCloser, error) {
	for _, f := range z.File {
		// Zip entry names are attacker-controlled; exact match avoids path traversal.
		if f.Name == name {
			return f.Open()
		}
	}
	return nil, fs.ErrNotExist
}

func readEntry(z *zip.Reader, name string, limit int64) ([]byte, error) {
	f, err := open(z, name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Limited against a zip bomb; nothing legitimate is near these limits.
	return io.ReadAll(io.LimitReader(f, limit))
}

// suggestBase flattens a name into a filesystem-safe base, or fallback if nothing survives.
func suggestBase(name, fallback string) string {
	clean := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			return r
		case r >= 'A' && r <= 'Z':
			return r + 32
		case r == ' ' || r == '-' || r == '_':
			return '-'
		}
		return -1
	}, strings.TrimSpace(name))

	clean = strings.Trim(clean, "-")
	if clean == "" {
		return fallback
	}
	return clean
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// writeFileAtomic writes to a temp file and renames onto the target, so a failed write leaves the old file intact.
func writeFileAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".stif-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(name, filePerm); err != nil {
		return err
	}
	return os.Rename(name, path)
}
