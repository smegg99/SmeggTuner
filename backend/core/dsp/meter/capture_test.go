package meter

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math"
	"slices"
	"testing"
)

func captureFloats(values ...float32) []byte {
	b := make([]byte, 0, 4*len(values))
	for _, v := range values {
		b = binary.LittleEndian.AppendUint32(b, math.Float32bits(v))
	}
	return b
}

func capturePCMBytes(width int, values ...int32) []byte {
	b := make([]byte, 0, width*len(values))
	for _, v := range values {
		for i := range width {
			b = append(b, byte(v>>(8*i)))
		}
	}
	return b
}

func captureRun(t *testing.T, data []byte, frames, channels int, encoding CaptureEncoding, flags CaptureFlags) CaptureChunk {
	t.Helper()
	chunk, err := ConvertCapture(nil, data, frames, channels, encoding, flags)
	if err != nil {
		t.Fatalf("ConvertCapture: %v", err)
	}
	return chunk
}

func captureExpect(t *testing.T, chunk CaptureChunk, channel int, samples ...int16) {
	t.Helper()
	if chunk.Channel != channel || !slices.Equal(chunk.Samples, samples) {
		t.Fatalf("got channel %d samples %v, want %d %v", chunk.Channel, chunk.Samples, channel, samples)
	}
}

func TestCapturePCM16AllCodes(t *testing.T) {
	var codes []int32
	var want []int16
	for v := math.MinInt16; v <= math.MaxInt16; v++ {
		codes, want = append(codes, int32(v)), append(want, int16(v))
	}
	captureExpect(t, captureRun(t, capturePCMBytes(2, codes...), len(codes), 1, CapturePCM16, 0), 0, want...)
}

func TestCapturePCMStereoInclusiveThresholds(t *testing.T) {
	for _, c := range []struct {
		left, right, want int32
		channel           int
	}{{128, 512, 512, 1}, {129, 512, 129, 0}, {128, 511, 128, 0}} {
		for _, sign := range []int32{-1, 1} {
			data := capturePCMBytes(2, sign*c.left, sign*c.right)
			captureExpect(t, captureRun(t, data, 1, 2, CapturePCM16, 0), c.channel, int16(sign*c.want))
		}
	}
	captureExpect(t, captureRun(t, capturePCMBytes(2, 0, -32768), 1, 2, CapturePCM16, 0), 1, -32768)
}

func TestCapturePCM16In32ReadsLowWord(t *testing.T) {
	data, _ := hex.DecodeString("800034120002cdab7fff00000080ffff")
	captureExpect(t, captureRun(t, data, 4, 1, CapturePCM16In32, 0), 0, 128, 512, -129, -32768)
	captureExpect(t, captureRun(t, data[:8], 1, 2, CapturePCM16In32, 0), 1, 512)
}

func TestCapturePCM24ShiftFloorsNegativeCodes(t *testing.T) {
	data := capturePCMBytes(3, -8388608, -257, -256, -1, 0, 255, 256, 8388607)
	captureExpect(t, captureRun(t, data, 8, 1, CapturePCM24, 0), 0, -32768, -2, -1, -1, 0, 0, 1, 32767)
	captureExpect(t, captureRun(t, capturePCMBytes(3, 128*256+255, 512*256), 1, 2, CapturePCM24, 0), 1, 512)
}

func TestCapturePCM32MonoUsesHighWord(t *testing.T) {
	data := capturePCMBytes(4, math.MinInt32, -65537, -1, 0, 65535, 65536, math.MaxInt32)
	captureExpect(t, captureRun(t, data, 7, 1, CapturePCM32, 0), 0, -32768, -2, -1, 0, 0, 1, 32767)
}

func TestCaptureSilentPrecedesFormatDispatch(t *testing.T) {
	for _, c := range []struct {
		channels int
		encoding CaptureEncoding
	}{{1, CaptureUnsupported}, {2, CapturePCM32}, {3, CaptureFloat32}, {0, CapturePCM16}} {
		chunk := captureRun(t, nil, 2, c.channels, c.encoding, CaptureSilent)
		captureExpect(t, chunk, CaptureNoChannel, 0, 0)
		if !chunk.Silent {
			t.Fatalf("%+v: silent chunk %+v", c, chunk)
		}
		if _, err := ConvertCapture(nil, nil, 2, c.channels, c.encoding, 0); !errors.Is(err, ErrCaptureFormat) {
			t.Fatalf("%+v: loud chunk err %v", c, err)
		}
	}
	if _, err := ConvertCapture(nil, nil, MaxCaptureFrames+1, 1, CaptureUnsupported, CaptureSilent); !errors.Is(err, ErrCaptureFrames) {
		t.Fatalf("the frame cap precedes the silent path: %v", err)
	}
}

func TestCaptureSilentIgnoresPayload(t *testing.T) {
	for _, payload := range [][]byte{nil, []byte("invalid"), captureFloats(float32(math.NaN()))} {
		for _, flags := range []CaptureFlags{2, 3, 6, 7} {
			chunk := captureRun(t, payload, 3, 2, CaptureFloat32, flags)
			captureExpect(t, chunk, CaptureNoChannel, 0, 0, 0)
			if !chunk.Silent || chunk.TimestampEligible != (flags == CaptureSilent) {
				t.Fatalf("flags %d: %+v", flags, chunk)
			}
		}
	}
}

func TestCaptureDiscontinuityKeepsSamples(t *testing.T) {
	for _, flags := range []CaptureFlags{0, CaptureDiscontinuity, CaptureTimestampError, 5} {
		chunk := captureRun(t, captureFloats(0.5, -0.5), 1, 2, CaptureFloat32, flags)
		captureExpect(t, chunk, 0, 16384)
		if chunk.Silent || chunk.TimestampEligible != (flags == 0) {
			t.Fatalf("flags %d: %+v", flags, chunk)
		}
	}
}

func TestCaptureReusesDestination(t *testing.T) {
	dst := make([]int16, 7, 16)
	chunk := captureRun(t, capturePCMBytes(2, 5, -6), 2, 1, CapturePCM16, 0)
	reused, _ := ConvertCapture(dst, capturePCMBytes(2, 5, -6), 2, 1, CapturePCM16, 0)
	if &reused.Samples[0] != &dst[0] || !slices.Equal(reused.Samples, chunk.Samples) {
		t.Fatalf("dst not reused: %v", reused.Samples)
	}
}

func TestCaptureContract(t *testing.T) {
	if chunk := captureRun(t, nil, 0, 2, CaptureFloat32, 0); len(chunk.Samples) != 0 || chunk.Channel != CaptureNoChannel {
		t.Fatalf("empty chunk: %+v", chunk)
	}
	for _, c := range []struct {
		data             []byte
		frames, channels int
		encoding         CaptureEncoding
		want             error
	}{
		{nil, -1, 2, CaptureFloat32, ErrCaptureFrames},
		{nil, MaxCaptureFrames + 1, 2, CaptureFloat32, ErrCaptureFrames},
		{nil, 0, 3, CaptureFloat32, ErrCaptureFormat},
		{nil, 0, 2, CapturePCM32, ErrCaptureFormat},
		{nil, 0, 1, CaptureUnsupported, ErrCaptureFormat},
		{nil, 1, 2, CaptureFloat32, ErrCapturePayload},
		{captureFloats(1), 1, 2, CaptureFloat32, ErrCapturePayload},
		{captureFloats(1, 2), 1, 1, CaptureFloat32, ErrCapturePayload},
	} {
		if _, err := ConvertCapture(nil, c.data, c.frames, c.channels, c.encoding, 0); !errors.Is(err, c.want) {
			t.Fatalf("%+v: err %v, want %v", c, err, c.want)
		}
	}
}
