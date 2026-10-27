// backend/web/bridge_test.go
package web

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"testing"
)

func TestCallJSONWrapsResult(t *testing.T) {
	d := NewDispatcher(&fakeService{})
	var ok struct {
		OK   bool `json:"ok"`
		Data int  `json:"data"`
	}
	if err := json.Unmarshal([]byte(callJSON(d, fakeID("Double"), "[4]")), &ok); err != nil || !ok.OK || ok.Data != 8 {
		t.Fatalf("Double(4) reply %+v, %v; want ok with 8", ok, err)
	}
	var bad struct {
		OK      bool            `json:"ok"`
		Kind    string          `json:"kind"`
		Message string          `json:"message"`
		Cause   json.RawMessage `json:"cause"`
	}
	if err := json.Unmarshal([]byte(callJSON(d, fakeID("Fail"), "[]")), &bad); err != nil {
		t.Fatal(err)
	}
	if bad.OK || bad.Kind != "RuntimeError" || bad.Message != "fake.error" || string(bad.Cause) != `{"key":"fake.error"}` {
		t.Fatalf("Fail() reply %+v", bad)
	}
}

func TestCallJSONRejectsMalformedArgs(t *testing.T) {
	var reply struct {
		OK   bool   `json:"ok"`
		Kind string `json:"kind"`
	}
	_ = json.Unmarshal([]byte(callJSON(NewDispatcher(&fakeService{}), fakeID("Double"), "{")), &reply)
	if reply.OK || reply.Kind != "TypeError" {
		t.Fatalf("reply %+v, want a TypeError", reply)
	}
}

func TestFloatsFromBytes(t *testing.T) {
	want := []float32{0.5, -1, 0.25}
	b := make([]byte, 4*len(want))
	for i, f := range want {
		binary.LittleEndian.PutUint32(b[4*i:], math.Float32bits(f))
	}
	got := floatsFromBytes(b)
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("floatsFromBytes = %v, want %v", got, want)
		}
	}
}

func TestInputBuildsTheBrowserSource(t *testing.T) {
	src := NewBrowserSource(44100)
	in := NewInput(src)
	got, isMic, err := in.Build()
	if err != nil || !isMic || got != src {
		t.Fatalf("Build() = %v, %v, %v; want the browser source as a mic", got, isMic, err)
	}
	if in.CorrectionPercent() != 0 {
		t.Fatal("a browser input has no sound-card correction")
	}
}
