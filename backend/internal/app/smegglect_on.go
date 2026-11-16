//go:build !production && !android

// backend/internal/app/smegglect_on.go
package app

import (
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/backend/services/smegglect"
)

// smegglectRoot is where the bridge `just dev` starts keeps its address: frontend/, the folder .mcp.json names.
const smegglectRoot = "frontend"

// smegglectServices registers the native save and the bridge proxy of the Smegglect overlay that debug builds
// carry; see smegglect_off.go.
func smegglectServices() []application.Service {
	root := os.Getenv("SMEGGLECT_ROOT")
	if root == "" {
		root = smegglectRoot
	}
	return []application.Service{application.NewService(&smegglect.Service{
		Dir:  os.Getenv("SMEGGLECT_EXPORT_DIR"),
		Root: root,
	})}
}
