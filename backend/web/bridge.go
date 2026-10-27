// backend/web/bridge.go
package web

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"math"

	coreaudio "smegg.me/smeggtuner/backend/core/audio"
)

// reply is one binding call's answer to the worker.
type reply struct {
	OK      bool            `json:"ok"`
	Data    json.RawMessage `json:"data,omitempty"`
	Kind    string          `json:"kind,omitempty"`
	Message string          `json:"message,omitempty"`
	Cause   json.RawMessage `json:"cause,omitempty"`
}

// callJSON runs one call whose arguments arrive as a JSON array and answers in JSON.
func callJSON(d *Dispatcher, id uint32, argsJSON string) string {
	var args []json.RawMessage
	var out reply
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		out = reply{Kind: "TypeError", Message: "arguments are not a JSON array: " + err.Error()}
	} else if data, err := d.Call(id, args); err != nil {
		var ce *CallError
		if !errors.As(err, &ce) {
			ce = &CallError{Kind: "RuntimeError", Message: err.Error()}
		}
		out = reply{Kind: ce.Kind, Message: ce.Message, Cause: ce.Cause}
	} else {
		out = reply{OK: true, Data: data}
	}
	b, _ := json.Marshal(out)
	return string(b)
}

// floatsFromBytes reads little-endian float32 samples, the layout of a Float32Array in wasm memory.
func floatsFromBytes(b []byte) []float32 {
	out := make([]float32, len(b)/4)
	for i := range out {
		out[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[4*i:]))
	}
	return out
}

// Input is the browser microphone as services/tuner reads it.
type Input struct{ src *BrowserSource }

// NewInput wraps the browser source for the tuner.
func NewInput(src *BrowserSource) Input { return Input{src: src} }

func (in Input) Build() (coreaudio.Source, bool, error) { return in.src, true, nil }
func (Input) SourceName() string                        { return "browser" }
func (Input) CorrectionPercent() float64                { return 0 }
