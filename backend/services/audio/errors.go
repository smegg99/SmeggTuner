package audio

import "smegg.me/smeggtuner/backend/common/serviceerror"

// ServiceError is the shared frontend error shape.
type ServiceError = serviceerror.ServiceError

var (
	ErrCalibrationSave = &ServiceError{Key: "settings.calibration.saveFailed"}
	ErrAutoVolumeSave  = &ServiceError{Key: "settings.autoVolume.saveFailed"}
	ErrNoDevices       = &ServiceError{Key: "tuner.error.noDevices"}      // no capture device enumerated
	ErrFileUnreadable  = &ServiceError{Key: "tuner.error.fileUnreadable"} // recording will not open or decode
	ErrDeviceGone      = &ServiceError{Key: "tuner.error.deviceGone"}     // capture device no longer present
)
