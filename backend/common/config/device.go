package config

// NewDevice is a capture device the config does not list yet; it matches #Device's defaults.
func NewDevice() Device {
	return Device{Calibration: "uncalibrated", CalibrationPercent: -999999, AutoVolume: true, Volume: 0.5}
}

// Device is what is kept for the capture device id, or NewDevice when nothing is.
func (a Audio) Device(id string) Device {
	if d, ok := a.Devices[id]; ok {
		return d
	}
	return NewDevice()
}
