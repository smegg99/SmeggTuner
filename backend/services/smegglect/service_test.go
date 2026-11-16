//go:build !production

package smegglect_test

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"smegg.me/smeggtuner/backend/services/smegglect"
)

func TestSaveWritesTheBundleToDir(t *testing.T) {
	dir := t.TempDir()
	svc := &smegglect.Service{Dir: dir}
	path, err := svc.Save(context.Background(), "smegglect-bat_1.zip", base64.StdEncoding.EncodeToString([]byte("zip")))
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(dir, "smegglect-bat_1.zip") {
		t.Fatalf("path = %q", path)
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != "zip" {
		t.Fatalf("read %q, %v", raw, err)
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %v", info.Mode().Perm())
	}
}

func TestSaveRefusesOtherNamesAndBadData(t *testing.T) {
	svc := &smegglect.Service{Dir: t.TempDir()}
	for _, name := range []string{"../smegglect-x.zip", "notes.zip", "smegglect-x.zip.exe", "smegglect-.zip"} {
		if _, err := svc.Save(context.Background(), name, "AA=="); !errors.Is(err, smegglect.ErrName) {
			t.Fatalf("%q: err = %v", name, err)
		}
	}
	if _, err := svc.Save(context.Background(), "smegglect-x.zip", "not base64!"); err == nil {
		t.Fatal("bad base64 was accepted")
	}
	big := strings.Repeat("A", base64.StdEncoding.EncodedLen(smegglect.MaxBytes+3))
	if _, err := svc.Save(context.Background(), "smegglect-x.zip", big); err == nil {
		t.Fatal("an oversized bundle was accepted")
	}
}
