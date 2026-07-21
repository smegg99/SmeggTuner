package config

import (
	"github.com/wailsapp/wails/v3/pkg/application"

	appconfig "smegg.me/smeggtuner/common/config"
	"smegg.me/smeggtuner/common/i18n"
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

// SetConfig replaces the entire config on disk and in memory.
func (s *Service) SetConfig(c appconfig.Config) error {
	if err := appconfig.SetConfig(c); err != nil {
		return err
	}
	syncLocale()
	emitConfigEvents()
	return nil
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
	if app := application.Get(); app != nil {
		updated := *appconfig.Get()
		app.Event.Emit(EventPreferencesChanged, updated.Preferences)
		app.Event.Emit(EventConfigChanged, updated)
	}
}

func syncLocale() {
	if lang := appconfig.Get().Preferences.Language; lang != "" {
		i18n.SetLocale(lang)
	}
}
