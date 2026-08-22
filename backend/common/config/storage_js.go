//go:build js

// backend/common/config/storage_js.go
package config

import (
	"encoding/json"
	"fmt"

	"smegg.me/smeggtuner/backend/common/emit"
)

var seedRaw []byte

// Seed hands Initialize the config the browser stored; call it before Initialize.
func Seed(raw []byte) { seedRaw = raw }

// Initialize loads the seeded config into Global; a rejected seed still leaves the defaults loaded and is returned.
func Initialize() (string, error) {
	writeMu.Lock()
	defer writeMu.Unlock()
	c, err := decodeSeed(seedRaw)
	mu.Lock()
	Global = c
	resolvedCfg = "browser"
	mu.Unlock()
	return "browser", err
}

// patchLocked merges v into Global and posts the result to the main thread, which stores it.
func patchLocked(v any) error {
	next, err := patchConfig(*Get(), v)
	if err != nil {
		return fmt.Errorf("patch config: %w", err)
	}
	data, err := json.Marshal(next)
	if err != nil {
		return err
	}
	mu.Lock()
	Global = next
	mu.Unlock()
	emit.Post(map[string]any{"type": "config", "json": string(data)})
	return nil
}
