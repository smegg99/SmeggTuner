package main

import (
	"embed"
	"os"

	"smegg.me/smeggtuner/internal/app"
)

// Regenerate the s99wails composables: go run github.com/smegg99/s99wails/frontendgen -out frontend/app/composables/s99wails

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

func main() {
	os.Exit(app.Main(assets, appIcon))
}
