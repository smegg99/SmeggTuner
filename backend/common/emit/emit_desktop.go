//go:build !js

// backend/common/emit/emit_desktop.go
package emit

import "github.com/wailsapp/wails/v3/pkg/application"

// Emit sends a Wails event; a no-op when no application is running.
func Emit(name string, data any) {
	if app := application.Get(); app != nil {
		app.Event.Emit(name, data)
	}
}
