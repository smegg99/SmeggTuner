//go:build !js

// backend/core/audio/pcm.go
package audio

import (
	"encoding/binary"
	"math"

	"github.com/gen2brain/malgo"
)

// capturePCM normalizes the configured device format without changing the sample clock.
func capturePCM(input []byte, count int, format malgo.FormatType) []float32 {
	samples := make([]float32, count)
	if format == malgo.FormatS16 {
		for i := range samples {
			samples[i] = float32(int16(binary.LittleEndian.Uint16(input[i*2:]))) / 32768
		}
		return samples
	}
	for i := range samples {
		samples[i] = math.Float32frombits(binary.LittleEndian.Uint32(input[i*4:]))
	}
	return samples
}
