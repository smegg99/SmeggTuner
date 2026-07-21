package config

import "fmt"

// SetConfig replaces the full config on disk and in memory.
func SetConfig(c Config) error {
	return patch(c)
}

// SetPreferences updates only the preferences section.
func SetPreferences(p Preferences) error {
	return patch(struct {
		Preferences Preferences `json:"preferences"`
	}{Preferences: p})
}

// SetLoggerConfig updates only the logger section.
func SetLoggerConfig(lc LoggerConfig) error {
	return patch(struct {
		Logger LoggerConfig `json:"logger"`
	}{Logger: lc})
}

// patch deep-merges v into the on-disk config, validates it, and refreshes Global.
func patch(v any) error {
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
