package meter

import (
	"math"
	"math/bits"
)

// RingCapacity is the native analysis ring length in samples.
const RingCapacity = 3_840_000

// UpdatesPerSecond is the native analysis rate that divides the capture sample rate.
const UpdatesPerSecond = 2

const (
	streamCoarseDuration = 1.3653333333333333
	streamFineDuration   = 2.7306666666666666
	streamCoarseCap      = 262_144
	streamFineCap        = 524_288
	streamMinimumWindow  = 32
)

// SampleRing holds float32 samples in int16 amplitude units; reads are quiescent snapshots.
type SampleRing struct {
	samples  []float32
	next     int
	appended int64
}

// NewSampleRing returns a zeroed ring, the state after an explicit capture reset.
func NewSampleRing(capacity int) *SampleRing {
	if capacity < 1 {
		panic("meter: ring capacity must be positive")
	}
	return &SampleRing{samples: make([]float32, capacity)}
}

func (r *SampleRing) Capacity() int { return len(r.samples) }

// NextWrite is the exclusive read endpoint: the newest sample sits just before it.
func (r *SampleRing) NextWrite() int { return r.next }

// Appended counts samples since the last reset; a diagnostic endpoint, not a native field.
func (r *SampleRing) Appended() int64 { return r.appended }

func (r *SampleRing) Reset() {
	clear(r.samples)
	r.next, r.appended = 0, 0
}

// Append stores samples unscaled; an append longer than the ring keeps its newest Capacity samples.
func (r *SampleRing) Append(samples []int16) {
	size := len(r.samples)
	r.appended += int64(len(samples))
	if extra := len(samples) - size; extra > 0 {
		r.next = (r.next + extra) % size
		samples = samples[extra:]
	}
	tail := min(len(samples), size-r.next)
	for i, s := range samples[:tail] {
		r.samples[r.next+i] = float32(s)
	}
	for i, s := range samples[tail:] {
		r.samples[i] = float32(s)
	}
	r.next = (r.next + len(samples)) % size
}

// Keep zeroes all but the n newest samples, as if the capture had begun n samples ago.
func (r *SampleRing) Keep(n int) {
	size := len(r.samples)
	if n >= size {
		return
	}
	start := (r.next - n + size) % size // the oldest kept sample
	if start <= r.next {
		clear(r.samples[r.next:])
		clear(r.samples[:start])
	} else {
		clear(r.samples[r.next:start])
	}
}

// Newest copies the n newest samples, oldest first, into dst and returns n, or 0 without writing when n > len(dst) or n > Capacity.
func (r *SampleRing) Newest(dst []float32, n int) int {
	if n < 0 {
		panic("meter: negative ring request")
	}
	size := len(r.samples)
	if n > len(dst) || n > size {
		return 0
	}
	start := (r.next - n + size) % size
	tail := min(n, size-start)
	copy(dst, r.samples[start:start+tail])
	copy(dst[tail:n], r.samples[:n-tail])
	return n
}

// WindowPlan is one transform request: Samples read from the ring, zero padded to FFTSize.
type WindowPlan struct {
	SampleRate int
	Fine       bool
	Samples    int
	FFTSize    int
}

// NewWindowPlan sizes the coarse or fine window for a nominal rate; the admitted domain is 1..192000 Hz.
func NewWindowPlan(sampleRate int, fine bool) WindowPlan {
	duration, limit := streamCoarseDuration, streamCoarseCap
	if fine {
		duration, limit = streamFineDuration, streamFineCap
	}
	rounded := math.Trunc(float64(float64(sampleRate)*duration) + 0.5)
	samples := int(min(float64(limit), max(streamMinimumWindow, rounded)))
	return WindowPlan{sampleRate, fine, samples, 4 << bits.Len(uint(samples-1))}
}

// MagnitudeCount is the logical spectrum length and the magnitude divisor.
func (p WindowPlan) MagnitudeCount() int { return p.FFTSize / 2 }

// Read fills dst[:FFTSize] with the newest Samples values and trailing zeros; ok is false, with dst zeroed, when the ring is shorter.
func (p WindowPlan) Read(r *SampleRing, dst []float32) (endpoint int64, ok bool) {
	dst = dst[:p.FFTSize]
	got := r.Newest(dst, p.Samples)
	clear(dst[got:])
	return r.Appended(), got == p.Samples
}

// CaptureCadence is the capture-driven analysis trigger: a frame accumulator and a pending-message latch.
type CaptureCadence struct {
	Threshold   int
	Accumulated int // native int32: sums past 2^31-1 are outside the model
	Pending     bool
}

// NewCaptureCadence sets Threshold = sampleRate / updatesPerSecond, which must be positive.
func NewCaptureCadence(sampleRate, updatesPerSecond int) CaptureCadence {
	return CaptureCadence{Threshold: sampleRate / updatesPerSecond}
}

// EventPass consumes one signaled drain pass, subtracting at most one threshold, and reports whether it posts; a timeout is not a pass.
func (c *CaptureCadence) EventPass(appended int, targetPresent bool) bool {
	c.Accumulated += appended
	if c.Accumulated < c.Threshold || !targetPresent {
		return false
	}
	c.Accumulated -= c.Threshold
	if c.Pending {
		return false
	}
	c.Pending = true
	return true
}

// HandlerCompleted is the analysis handler's normal end; it clears the latch without posting.
func (c *CaptureCadence) HandlerCompleted() { c.Pending = false }

// ResetAccumulator is the capture ring reset; it leaves the pending latch alone.
func (c *CaptureCadence) ResetAccumulator() { c.Accumulated = 0 }
