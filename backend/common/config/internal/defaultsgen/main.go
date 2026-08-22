// backend/common/config/internal/defaultsgen/main.go

// Command defaultsgen writes defaults.json, the schema's defaults, for the web edition that carries no CUE validator.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/smegg99/s99config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "defaultsgen:", err)
		os.Exit(1)
	}
}

func run() error {
	schema, err := os.ReadFile("config.cue")
	if err != nil {
		return err
	}
	l, err := s99config.New(schema)
	if err != nil {
		return err
	}
	raw, err := l.DefaultsJSON()
	if err != nil {
		return err
	}
	var out bytes.Buffer
	if err := json.Indent(&out, raw, "", "  "); err != nil {
		return err
	}
	out.WriteByte('\n')
	return os.WriteFile("defaults.json", out.Bytes(), 0o644)
}
