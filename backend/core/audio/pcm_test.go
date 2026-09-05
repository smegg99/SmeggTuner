//go:build !js

// backend/core/audio/pcm_test.go
package audio

import (
	"encoding/binary"
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/gen2brain/malgo"

	"smegg.me/smeggtuner/backend/core/dsp/meter"
)

func TestCapturePCM(t *testing.T) {
	want := []float32{-1, 32767.0 / 32768, -0.5, 0.25, 0, -1.0 / 32768}
	for _, format := range []malgo.FormatType{malgo.FormatS16, malgo.FormatF32} {
		t.Run(fmt.Sprint(format), func(t *testing.T) {
			var input []byte
			for _, sample := range want {
				if format == malgo.FormatS16 {
					input = binary.LittleEndian.AppendUint16(input, uint16(int16(sample*32768)))
				} else {
					input = binary.LittleEndian.AppendUint32(input, math.Float32bits(sample))
				}
			}
			got := capturePCM(input, len(want), format)
			if !slices.Equal(got, want) {
				t.Fatalf("normalized PCM = %v, want %v", got, want)
			}
			mono, _ := meter.SelectStereo(nil, got)
			if !slices.Equal(mono, []float32{-1, -0.5, 0}) {
				t.Fatalf("selected channel = %v", mono)
			}
		})
	}
}
