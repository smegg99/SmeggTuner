package meter

import (
	"encoding/binary"
	"math"
	"slices"
)

// CaptureEncoding is a resolved sample layout: format tag, effective valid bits and storage bytes.
type CaptureEncoding int

const (
	CaptureFloat32     CaptureEncoding = iota // IEEE float, 32 bits in 4 bytes
	CapturePCM16                              // 16 bits in 2 bytes
	CapturePCM16In32                          // 16 bits in the low word of 4 bytes
	CapturePCM24                              // 24 bits packed in 3 bytes
	CapturePCM32                              // 32 bits in 4 bytes, mono only
	CaptureUnsupported                        // any layout without a conversion branch; only silent chunks convert
)

// CaptureFlags are the capture buffer status bits; other bits are ignored.
type CaptureFlags uint32

const (
	CaptureDiscontinuity  CaptureFlags = 1
	CaptureSilent         CaptureFlags = 2
	CaptureTimestampError CaptureFlags = 4
)

const (
	MaxCaptureFrames = 3_840_000
	CaptureNoChannel = -1
)

const (
	captureLeftQuiet   float32 = 128.0 / 32767
	captureRightActive float32 = 512.0 / 32767
	capturePCMQuiet            = 128
	capturePCMActive           = 512
)

// CaptureError reports a chunk outside the admitted capture contract.
type CaptureError string

func (e CaptureError) Error() string { return string(e) }

const (
	ErrCaptureFormat  CaptureError = "meter: unsupported capture encoding or channel count"
	ErrCaptureFrames  CaptureError = "meter: capture frame count outside one chunk"
	ErrCapturePayload CaptureError = "meter: capture payload size does not match frame count"
)

// CaptureChunk is one converted chunk; Channel is CaptureNoChannel for silent or empty chunks.
type CaptureChunk struct {
	Samples []int16
	Channel int
	Silent  bool
	// TimestampEligible means the chunk may feed clock accounting, not that its timestamp is valid.
	TimestampEligible bool
}

// ConvertCapture converts one chunk to int16 units in dst's storage; stereo keeps one whole-chunk channel, so callers must not merge packets.
func ConvertCapture(dst []int16, data []byte, frames, channels int, encoding CaptureEncoding, flags CaptureFlags) (CaptureChunk, error) {
	if frames < 0 || frames > MaxCaptureFrames {
		return CaptureChunk{}, ErrCaptureFrames
	}
	chunk := CaptureChunk{
		Channel:           CaptureNoChannel,
		TimestampEligible: flags&(CaptureDiscontinuity|CaptureTimestampError) == 0,
	}
	if flags&CaptureSilent != 0 {
		chunk.Samples, chunk.Silent = slices.Grow(dst[:0], frames)[:frames], true
		clear(chunk.Samples)
		return chunk, nil
	}
	width, ok := captureWidth(encoding, channels)
	if !ok {
		return CaptureChunk{}, ErrCaptureFormat
	}
	if len(data) != frames*channels*width {
		return CaptureChunk{}, ErrCapturePayload
	}
	samples := slices.Grow(dst[:0], frames)[:frames]
	chunk.Samples = samples
	if frames == 0 {
		return chunk, nil
	}
	chunk.Channel = 0
	if encoding == CaptureFloat32 {
		at := func(i int) float32 { return math.Float32frombits(binary.LittleEndian.Uint32(data[4*i:])) }
		if channels == 2 {
			chunk.Channel = captureChannel(at, frames, captureLeftQuiet, captureRightActive)
		}
		for i := range samples {
			samples[i] = captureQuantize(at(i*channels + chunk.Channel))
		}
		return chunk, nil
	}
	at := func(i int) int32 { return capturePCM(data, i, encoding) }
	if channels == 2 {
		chunk.Channel = captureChannel(at, frames, capturePCMQuiet, capturePCMActive)
	}
	for i := range samples {
		samples[i] = int16(at(i*channels + chunk.Channel))
	}
	return chunk, nil
}

func captureWidth(encoding CaptureEncoding, channels int) (int, bool) {
	if channels != 1 && channels != 2 {
		return 0, false
	}
	switch encoding {
	case CapturePCM16:
		return 2, true
	case CapturePCM24:
		return 3, true
	case CaptureFloat32, CapturePCM16In32:
		return 4, true
	case CapturePCM32:
		return 4, channels == 1
	}
	return 0, false
}

// capturePCM returns sample i in int16 units: arithmetic shifts floor negative codes.
func capturePCM(data []byte, i int, encoding CaptureEncoding) int32 {
	switch encoding {
	case CapturePCM16:
		return int32(int16(binary.LittleEndian.Uint16(data[2*i:])))
	case CapturePCM16In32:
		return int32(int16(binary.LittleEndian.Uint16(data[4*i:])))
	case CapturePCM24:
		b := data[3*i:]
		return int32(uint32(b[0])<<8|uint32(b[1])<<16|uint32(b[2])<<24) >> 16
	}
	return int32(binary.LittleEndian.Uint32(data[4*i:])) >> 16
}

type captureSample interface{ ~float32 | ~int32 }

func captureAbs[T captureSample](v T) T {
	if v < 0 {
		return -v
	}
	return v
}

// captureChannel picks right only when the left peak is quiet and the right peak active, inclusive.
func captureChannel[T captureSample](at func(int) T, frames int, quiet, active T) int {
	var left, right T // peaks start at zero and maxss keeps them when a sample is NaN
	for i := range frames {
		if v := captureAbs(at(2 * i)); v > left {
			left = v
		}
		if v := captureAbs(at(2*i + 1)); v > right {
			right = v
		}
	}
	if left <= quiet && right >= active {
		return 1
	}
	return 0
}

// captureQuantize clamps in native minss/maxss operand order, then rounds the binary32 product half away from zero.
func captureQuantize(x float32) int16 {
	m := x
	if 1 < x {
		m = 1
	}
	c := m
	if -1 > m {
		c = -1
	}
	if c != c {
		return 0 // cvttss2si gives 0x80000000, whose stored low word is 0
	}
	y := float32(c * 32767)
	if y < 0 {
		return int16(float32(y - 0.5))
	}
	return int16(float32(y + 0.5))
}

// QuantizeFloat32 converts mono full-scale samples to int16 units as the Float32 capture path does.
func QuantizeFloat32(dst []int16, src []float32) []int16 {
	dst = slices.Grow(dst[:0], len(src))[:len(src)]
	for i, x := range src {
		dst[i] = captureQuantize(x)
	}
	return dst
}

// SelectStereo is the reference Float32 stereo choice for one capture chunk: left unless left is quiet
// and right active. It writes the chosen channel, unconverted, into dst and returns it with the channel.
func SelectStereo(dst, interleaved []float32) ([]float32, int) {
	frames := len(interleaved) / 2
	channel := captureChannel(func(i int) float32 { return interleaved[i] }, frames, captureLeftQuiet, captureRightActive)
	dst = slices.Grow(dst[:0], frames)[:frames]
	for i := range dst {
		dst[i] = interleaved[2*i+channel]
	}
	return dst, channel
}
