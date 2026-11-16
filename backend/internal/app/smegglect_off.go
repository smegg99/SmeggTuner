//go:build production || android

// backend/internal/app/smegglect_off.go
package app

import "github.com/wailsapp/wails/v3/pkg/application"

// smegglectServices is empty in production and Android builds: they carry no Smegglect overlay; see smegglect_on.go.
func smegglectServices() []application.Service { return nil }
