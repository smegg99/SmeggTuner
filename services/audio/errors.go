package audio

// ServiceError is the only error shape handed to the frontend; Key is an i18n key, the Go error is logged only.
type ServiceError struct {
	Key string `json:"key"`
}

func (e *ServiceError) Error() string { return e.Key }

var (
	ErrNoDevices      = &ServiceError{Key: "tuner.error.noDevices"}      // no capture device enumerated
	ErrFileUnreadable = &ServiceError{Key: "tuner.error.fileUnreadable"} // recording will not open or decode
	ErrDeviceGone     = &ServiceError{Key: "tuner.error.deviceGone"}     // capture device no longer present
)
