package session

import (
	"archive/zip"
	"bytes"
	"errors"
	"strings"
	"testing"
)

func archiveEntry(t *testing.T, body string, corrupt bool) *zip.Reader {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	entry, err := w.CreateHeader(&zip.FileHeader{Name: "data.json", Method: zip.Store})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if corrupt {
		offset, err := z.File[0].DataOffset()
		if err != nil {
			t.Fatal(err)
		}
		data[int(offset)+len(body)-1] ^= 1
	}
	return z
}

func TestReadEntryRejectsOversizeInsteadOfAcceptingJSONPrefix(t *testing.T) {
	const limit = 32
	prefix := `{}` + strings.Repeat(" ", limit-2)
	for _, tail := range []string{"", " ", "untrusted tail"} {
		body, err := readEntry(archiveEntry(t, prefix+tail, false), "data.json", limit)
		if tail == "" {
			if err != nil || string(body) != prefix {
				t.Fatalf("at limit: %q, %v", body, err)
			}
		} else if err == nil || body != nil {
			t.Fatalf("oversize entry accepted: %q, %v", body, err)
		}
	}
}

func TestReadEntryChecksCRCAtExactLimit(t *testing.T) {
	const body = `{"name":"instrument"}`
	got, err := readEntry(archiveEntry(t, body, true), "data.json", int64(len(body)))
	if !errors.Is(err, zip.ErrChecksum) || got != nil {
		t.Fatalf("corrupt exact-limit entry: %q, %v", got, err)
	}
}
