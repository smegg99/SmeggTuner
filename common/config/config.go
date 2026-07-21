//go:generate go run cuelang.org/go/cmd/cue@v0.16.1 exp gengotypes .

package config

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
	"github.com/smegg99/s99config"
)

//go:embed config.cue
var schema []byte

const (
	configFileName = "config.json"
	keyringService = "Smeggtuner"
)

var (
	// Global holds the loaded configuration.
	Global Config

	mu          sync.RWMutex
	loader      *s99config.Loader
	resolvedCfg string
)

// Initialize loads and validates the config into Global, writing a default file when none exists, and returns the resolved path.
func Initialize() (string, error) {
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

// Get returns a snapshot copy of the configuration; mutating it does nothing, use SetConfig to write.
func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()
	snapshot := Global
	return &snapshot
}

// GetConfigPath returns the resolved path to the config file.
func GetConfigPath() string {
	mu.RLock()
	defer mu.RUnlock()
	return resolvedCfg
}

// decodeLocked refreshes Global from the loader; callers must hold at least a read lock (it takes the write lock itself).
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

// resolveConfigPath: CONFIG_PATH wins, installed builds use the platform config dir, else the portable default in the working dir.
func resolveConfigPath() string {
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return filepath.Join(p, configFileName)
		}
		return p
	}
	if DetectFormat().IsInstalled() {
		if dir, err := userConfigDir(); err == nil {
			return filepath.Join(dir, configFileName)
		}
	}
	return configFileName
}

// resolveDataDir is where @{datadir:...} references land: the platform data dir when installed, the config's own dir when portable.
func resolveDataDir(configDir string) string {
	if DetectFormat().IsInstalled() {
		if dir, err := userDataDir(); err == nil {
			return dir
		}
	}
	return configDir
}
