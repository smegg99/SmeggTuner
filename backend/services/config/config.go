package config

import (
	"smegg.me/smeggtuner/backend/common/emit"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/common/i18n"
)

const (
	EventConfigChanged      = "config:changed"
	EventPreferencesChanged = "config:preferences:changed"
)

type Service struct{}

// GetConfig returns the full resolved application config.
func (s *Service) GetConfig() appconfig.Config {
	return *appconfig.Get()
}

// PatchConfig changes only edited fields and returns the current resolved config.
func (s *Service) PatchConfig(changes map[string]any) (appconfig.Config, error) {
	if err := appconfig.Patch(changes); err != nil {
		return appconfig.Config{}, err
	}
	syncLocale()
	emitConfigEvents()
	return *appconfig.Get(), nil
}

// SetPreferences persists updated preferences to disk and memory.
func (s *Service) SetPreferences(p appconfig.Preferences) error {
	if err := appconfig.SetPreferences(p); err != nil {
		return err
	}
	syncLocale()
	emitConfigEvents()
	return nil
}

// SetLoggerConfig persists updated logger configuration.
func (s *Service) SetLoggerConfig(lc appconfig.LoggerConfig) error {
	if err := appconfig.SetLoggerConfig(lc); err != nil {
		return err
	}
	emitConfigEvents()
	return nil
}

func emitConfigEvents() {
	updated := *appconfig.Get()
	emit.Emit(EventPreferencesChanged, updated.Preferences)
	emit.Emit(EventConfigChanged, updated)
}

func syncLocale() {
	if lang := appconfig.Get().Preferences.Language; lang != "" {
		i18n.SetLocale(lang)
	}
}
