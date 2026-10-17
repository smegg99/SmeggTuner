package audio

import (
	appconfig "smegg.me/smeggtuner/backend/common/config"
	configsvc "smegg.me/smeggtuner/backend/services/config"
)

// device is what the config keeps for a capture device key; with no config loaded, the defaults.
func device(key string) appconfig.Device {
	if appconfig.GetConfigPath() == "" {
		return appconfig.NewDevice()
	}
	return appconfig.Get().Audio.Device(key)
}

// storeDevice changes one device's entry in the config file and tells the settings page; with no
// config loaded (tests, headless tools) it does nothing.
func (s *Service) storeDevice(key string, mut func(*appconfig.Device)) error {
	if appconfig.GetConfigPath() == "" {
		return nil
	}
	err := appconfig.Update(func(c *appconfig.Config) {
		d := c.Audio.Device(key)
		mut(&d)
		if c.Audio.Devices == nil {
			c.Audio.Devices = make(map[string]appconfig.Device)
		}
		c.Audio.Devices[key] = d
	})
	if err != nil {
		return err
	}
	// Publish the persisted snapshot so the settings page can refresh untouched fields.
	emitEvent(configsvc.EventConfigChanged, *appconfig.Get())
	return nil
}
