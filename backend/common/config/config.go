//go:generate go run cuelang.org/go/cmd/cue@v0.16.1 exp gengotypes .
//go:generate go run ./internal/defaultsgen

package config

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

const (
	configFileName = "config.json"
	keyringService = "Smeggtuner"
)

var (
	// Global holds the loaded configuration.
	Global Config

	mu          sync.RWMutex
	writeMu     sync.Mutex
	resolvedCfg string
)

// Get returns an owned snapshot; use Update or Patch to persist changes.
func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()
	snapshot := Global
	snapshot.Audio.Devices = maps.Clone(Global.Audio.Devices)
	snapshot.Tuner.ReedBoxes = slices.Clone(Global.Tuner.ReedBoxes)
	snapshot.Tuner.OctaveOffsets = slices.Clone(Global.Tuner.OctaveOffsets)
	return &snapshot
}

// GetConfigPath returns the resolved path to the config file.
func GetConfigPath() string {
	mu.RLock()
	defer mu.RUnlock()
	return resolvedCfg
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
