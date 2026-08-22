// backend/common/serviceerror/serviceerror.go

// Package serviceerror holds the one error shape every service hands the frontend.
package serviceerror

// ServiceError is the only error shape handed to the frontend; Key is an i18n key, the Go error is logged only.
type ServiceError struct {
	Key string `json:"key"`
}

func (e *ServiceError) Error() string { return e.Key }
