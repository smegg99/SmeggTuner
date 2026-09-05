// backend/core/audio/endpoint/wasapi_windows_test.go
package endpoint

import "testing"

func TestEndpointIDStopsAtTheNUL(t *testing.T) {
	raw := make([]byte, 128)
	for i, r := range "{0.0.1.00000000}.{ab}" {
		raw[2*i] = byte(r)
	}
	if got := endpointID(raw); got != "{0.0.1.00000000}.{ab}" {
		t.Fatalf("id %q", got)
	}
}
