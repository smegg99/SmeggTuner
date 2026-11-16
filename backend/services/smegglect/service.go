//go:build !production

// Package smegglect serves the Smegglect overlay that debug builds carry: it saves exported bundles and forwards
// the overlay's calls to the local agent bridge. Production builds compile none of it.
package smegglect

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// MaxBytes bounds one bundle exported from the overlay, and one bridge request or answer.
const MaxBytes = 64 << 20

var bundleName = regexp.MustCompile(`^smegglect-[A-Za-z0-9_-]{1,64}\.zip$`)

// ErrName means the window sent a name that is not a Smegglect bundle name.
var ErrName = errors.New("smegglect: unexpected file name")

// Service saves bundles exported by the overlay and forwards its calls to the local agent bridge.
type Service struct {
	// Dir, when set, receives bundles without a dialog (automated runs).
	Dir string
	// Root holds the running broker's .smegglect/bridge.json; the working directory when empty.
	Root string
}

// Save writes one bundle and returns its path, or "" when the dialog is cancelled.
func (s *Service) Save(ctx context.Context, name string, data string) (string, error) {
	if !bundleName.MatchString(name) {
		return "", ErrName
	}
	if base64.StdEncoding.DecodedLen(len(data)) > MaxBytes {
		return "", fmt.Errorf("smegglect: bundle over %d bytes", MaxBytes)
	}
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("smegglect: decode: %w", err)
	}
	path := filepath.Join(s.Dir, name)
	if s.Dir == "" {
		dialog := application.Get().Dialog.SaveFile()
		options := &application.SaveFileDialogOptions{Filename: name, CanCreateDirectories: true}
		if win, ok := ctx.Value(application.WindowKey).(*application.WebviewWindow); ok {
			options.Window = win
		}
		dialog.SetOptions(options)
		if path, err = dialog.PromptForSingleSelection(); err != nil || path == "" {
			return "", err
		}
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return "", fmt.Errorf("smegglect: write: %w", err)
	}
	return path, nil
}
