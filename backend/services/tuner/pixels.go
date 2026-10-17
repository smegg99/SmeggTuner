package tuner

import (
	"encoding/base64"
	"math"
)

// Pictures go over as quantized bytes, not floats: the float arrays were 91% of a 6.1 kB Wails event compiled as JS ~12/s.

// pixels is a base64 picture on the wire; a string not []byte so the binding generator infers the frontend type.
type pixels string

func encode(b []byte) pixels {
	return pixels(base64.StdEncoding.EncodeToString(b))
}

// pack maps [lo, hi] onto a byte; 0 stays 0 so a silent band stays silent.
func pack(src []float32, lo, hi float32) []byte {
	if len(src) == 0 {
		return nil
	}
	span := hi - lo
	out := make([]byte, len(src))
	for i, v := range src {
		out[i] = byte(math.Round(float64(clamp((v-lo)/span, 0, 1) * 255)))
	}
	return out
}

// packSigned maps [-1, 1] onto a byte, zero at 128 exactly so silence gets no DC offset.
func packSigned(src []float32) []byte {
	if len(src) == 0 {
		return nil
	}
	out := make([]byte, len(src))
	for i, v := range src {
		out[i] = byte(128 + int(math.Round(float64(clamp(v, -1, 1)*127))))
	}
	return out
}

func clamp(v, lo, hi float32) float32 {
	return min(hi, max(lo, v))
}
