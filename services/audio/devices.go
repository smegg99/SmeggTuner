package audio

import (
	"smegg.me/smeggtuner/common/logger"
	coreaudio "smegg.me/smeggtuner/core/audio"
)

// ListDevices enumerates capture devices. A backend failure and an empty list are both ErrNoDevices to the UI.
func (s *Service) ListDevices() ([]DeviceDTO, error) {
	devices, err := enumerate()
	if err != nil {
		return nil, err
	}
	out := make([]DeviceDTO, 0, len(devices))
	for _, d := range devices {
		out = append(out, DeviceDTO{ID: d.ID, Name: d.Name, Default: d.Default})
	}
	logger.Debug(logger.MsgAudioDevicesListed, logger.Int("devices", len(out)))
	return out, nil
}

func enumerate() ([]coreaudio.DeviceInfo, error) {
	devices, err := coreaudio.Devices()
	if err != nil {
		logger.Warn(logger.MsgAudioDeviceEnumFailed, logger.Err(err))
		return nil, ErrNoDevices
	}
	if len(devices) == 0 {
		logger.Warn(logger.MsgAudioNoDevices)
		return nil, ErrNoDevices
	}
	return devices, nil
}

// micName resolves a device's label and checks it is still present. "" follows the system default.
func micName(deviceID string) (string, error) {
	devices, err := enumerate()
	if err != nil {
		return "", err
	}
	for _, d := range devices {
		if deviceID == "" && d.Default {
			return d.Name, nil
		}
		if deviceID != "" && d.ID == deviceID {
			return d.Name, nil
		}
	}
	if deviceID == "" {
		// No device claims the default role; the backend will still pick one.
		return devices[0].Name, nil
	}
	logger.Debug(logger.MsgAudioDeviceGone, logger.Str("device_id", deviceID))
	return "", ErrDeviceGone
}
