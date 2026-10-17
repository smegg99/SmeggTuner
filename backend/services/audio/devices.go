package audio

import (
	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/common/logger"
	coreaudio "smegg.me/smeggtuner/backend/core/audio"
)

// ListDevices enumerates capture devices. A backend failure and an empty list are both ErrNoDevices to the UI.
func (s *Service) ListDevices() ([]DeviceDTO, error) {
	devices, err := s.enumerate()
	if err != nil {
		return nil, err
	}
	out := make([]DeviceDTO, 0, len(devices))
	for _, d := range devices {
		out = append(out, DeviceDTO{ID: d.ID, Name: d.Name, Default: d.Default})
	}
	logger.Debug(i18n.KeyLogAudioDevicesListed, logger.Int("devices", len(out)))
	return out, nil
}

func (s *Service) enumerate() ([]coreaudio.DeviceInfo, error) {
	devices, err := s.listDevices()
	if err != nil {
		logger.Warn(i18n.KeyLogAudioDeviceEnumFailed, logger.Err(err))
		return nil, ErrNoDevices
	}
	if len(devices) == 0 {
		logger.Warn(i18n.KeyLogAudioNoDevices)
		return nil, ErrNoDevices
	}
	return devices, nil
}

// micName resolves a device's label and stored key and checks it is still present. "" follows the
// system default, whose key is the device that claims the role ("" when none does).
func (s *Service) micName(deviceID string) (name, key string, err error) {
	devices, err := s.enumerate()
	if err != nil {
		return "", "", err
	}
	for _, d := range devices {
		if deviceID == "" && d.Default {
			return d.Name, d.ID, nil
		}
		if deviceID != "" && d.ID == deviceID {
			return d.Name, d.ID, nil
		}
	}
	if deviceID == "" {
		// No device claims the default role; the backend will still pick one.
		return devices[0].Name, "", nil
	}
	logger.Debug(i18n.KeyLogAudioDeviceGone, logger.Str("device_id", deviceID))
	return "", "", ErrDeviceGone
}

// keyOf is the stored key of a capture device: its id, or for "" the current system default.
func (s *Service) keyOf(deviceID string) string {
	if deviceID != "" {
		return deviceID
	}
	_, key, _ := s.micName("")
	return key
}

// micKey is the stored key of the selected mic, resolved once per selection.
func (s *Service) micKey() string {
	s.mu.RLock()
	key, keyed, id := s.key, s.keyed, s.current.DeviceID
	s.mu.RUnlock()
	if keyed {
		return key
	}
	key = s.keyOf(id)
	s.mu.Lock()
	if s.current.Kind == SourceMic && s.current.DeviceID == id {
		s.key, s.keyed = key, true
	}
	s.mu.Unlock()
	return key
}
