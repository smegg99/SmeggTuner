// backend/generate.go

package main

// TypeScript bindings and S99Wails composables for frontend/, which has no generator of its own.

//go:generate go tool wails3 generate bindings -silent -ts -clean=true -d ../frontend/bindings ./...
//go:generate go tool frontendgen -out ../frontend/app/composables/s99wails -s99ab
