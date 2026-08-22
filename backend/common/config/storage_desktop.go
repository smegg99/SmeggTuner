//go:build !js

// backend/common/config/storage_desktop.go
package config

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/smegg99/s99config"
)

//go:embed config.cue
var schema []byte

var loader *s99config.Loader

// decodeLocked publishes the loader's current snapshot while the caller holds writeMu.
func decodeLocked() error {
	mu.RLock()
	l := loader
	mu.RUnlock()
	if l == nil {
		return fmt.Errorf("config not initialized")
	}

	var next Config
	if err := l.Decode(&next); err != nil {
		return fmt.Errorf("decode config: %w", err)
	}

	mu.Lock()
	Global = next
	mu.Unlock()
	return nil
}

// Initialize loads and validates the config into Global, writing a default file when none exists, and returns the resolved path.
func Initialize() (string, error) {
	writeMu.Lock()
	defer writeMu.Unlock()
	_ = godotenv.Load()

	path := resolveConfigPath()
	dir := filepath.Dir(path)

	l, err := s99config.New(
		schema,
		s99config.WithReferences(s99config.ReferenceOptions{
			ConfigDir:      dir,
			DataDir:        resolveDataDir(dir),
			KeyringService: keyringService,
		}),
	)
	if err != nil {
		return "", fmt.Errorf("compile config schema: %w", err)
	}

	if _, statErr := os.Stat(path); errors.Is(statErr, fs.ErrNotExist) {
		if err := l.WriteDefaults(path); err != nil {
			return path, fmt.Errorf("write default config %s: %w", path, err)
		}
	}

	if err := l.Load(path); err != nil {
		return path, fmt.Errorf("load config %s: %w", path, err)
	}

	mu.Lock()
	loader = l
	resolvedCfg = path
	mu.Unlock()

	if err := decodeLocked(); err != nil {
		return path, err
	}

	return path, nil
}

func patchLocked(v any) error {
	mu.RLock()
	l := loader
	mu.RUnlock()
	if l == nil {
		return fmt.Errorf("config not initialized")
	}

	if err := l.Patch(v); err != nil {
		return fmt.Errorf("patch config: %w", err)
	}

	return decodeLocked()
}
