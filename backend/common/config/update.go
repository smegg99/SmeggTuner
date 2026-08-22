// backend/common/config/update.go
package config

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Patch persists only edited fields, preserving unrelated values and raw references.
func Patch(changes map[string]any) error {
	return patch(changes)
}

// Update serializes a read-modify-write with every other configuration writer.
func Update(mut func(*Config)) error {
	writeMu.Lock()
	defer writeMu.Unlock()
	before := Get()
	after := Get()
	mut(after)
	a, err := configObject(before)
	if err != nil {
		return err
	}
	b, err := configObject(after)
	if err != nil {
		return err
	}
	return patchLocked(changedFields(a, b))
}

func configObject(c *Config) (map[string]any, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("encode config: %w", err)
	}
	var result map[string]any
	err = json.Unmarshal(data, &result)
	return result, err
}

func changedFields(before, after map[string]any) map[string]any {
	changes := make(map[string]any)
	for key, value := range after {
		if reflect.DeepEqual(before[key], value) {
			continue
		}
		old, oldMap := before[key].(map[string]any)
		next, nextMap := value.(map[string]any)
		if oldMap && nextMap {
			changes[key] = changedFields(old, next)
		} else {
			changes[key] = value
		}
	}
	for key := range before {
		if _, exists := after[key]; !exists {
			changes[key] = nil
		}
	}
	return changes
}

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

// SetDevice stores one capture device's entry and leaves every other device as it is.
func SetDevice(id string, d Device) error {
	type audio struct {
		Devices map[string]Device `json:"devices"`
	}
	return patch(struct {
		Audio audio `json:"audio"`
	}{Audio: audio{Devices: map[string]Device{id: d}}})
}

// patch deep-merges v into the on-disk config, validates it, and refreshes Global.
func patch(v any) error {
	writeMu.Lock()
	defer writeMu.Unlock()
	return patchLocked(v)
}
