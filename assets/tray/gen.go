// assets/tray/gen.go

// Package trayicons holds the tray artwork and the generated accessors for it.
//
// Everything here except the two source SVGs and trayicons.cue is generated:
// run `go generate ./assets/tray` (or `go generate ./...`) after touching
// either. traygen renders through resvg in wasm, so this needs no cairo, no
// venv, and no system libraries - just the Go toolchain.
package trayicons

//go:generate go tool traygen -config trayicons.cue
