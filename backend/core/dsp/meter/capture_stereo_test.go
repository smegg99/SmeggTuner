package meter

import (
	"math"
	"testing"
)

func math32bits(v float32) uint32 { return math.Float32bits(v) }

func TestSelectStereoFollowsTheReferenceThresholds(t *testing.T) {
	stereo := func(left, right float32) []float32 {
		out := make([]float32, 64)
		for i := 0; i < 32; i++ {
			out[2*i], out[2*i+1] = left, right
		}
		return out
	}
	cases := []struct {
		left, right float32
		want        int
	}{
		{0, 0.5, 1}, // dead left, live right
		{captureLeftQuiet, captureRightActive, 1}, // both boundaries are inclusive
		{0.3, 0.5, 0},                     // both live: left
		{0, captureRightActive * 0.9, 0},  // right not loud enough
		{captureLeftQuiet * 1.01, 0.5, 0}, // left not quiet enough
	}
	for _, c := range cases {
		got, ch := SelectStereo(nil, stereo(c.left, c.right))
		if ch != c.want || len(got) != 32 || got[5] != [2]float32{c.left, c.right}[c.want] {
			t.Fatalf("left %v right %v: channel %d samples %v", c.left, c.right, ch, got[:2])
		}
	}
	// The float choice equals the Float32 capture path's choice on the same bytes.
	in := stereo(0, 0.25)
	raw := make([]byte, 4*len(in))
	for i, v := range in {
		b := math32bits(v)
		raw[4*i], raw[4*i+1], raw[4*i+2], raw[4*i+3] = byte(b), byte(b>>8), byte(b>>16), byte(b>>24)
	}
	chunk, err := ConvertCapture(nil, raw, 32, 2, CaptureFloat32, 0)
	if _, ch := SelectStereo(nil, in); err != nil || chunk.Channel != ch {
		t.Fatalf("capture path channel %d err %v, float path %d", chunk.Channel, err, ch)
	}
}
