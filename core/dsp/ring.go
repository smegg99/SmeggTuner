// Package dsp implements the measurement pipeline: coarse per-note
// energies, noise floor, note detection, zoom analysis, peaks, beats,
// and the engine that ties them to an audio.Source.
package dsp

type Ring struct {
	buf   []float64
	pos   int
	count int
}

func NewRing(capacity int) *Ring {
	return &Ring{buf: make([]float64, capacity)}
}

func (r *Ring) Write(samples []float32) {
	for _, s := range samples {
		r.buf[r.pos] = float64(s)
		r.pos = (r.pos + 1) % len(r.buf)
	}
	r.count += len(samples)
	if r.count > len(r.buf) {
		r.count = len(r.buf)
	}
}

// Len reports samples written, capped at capacity.
func (r *Ring) Len() int { return r.count }

// Reset empties the ring without freeing it; the engine calls it on a source seek so a window never spans the jump.
func (r *Ring) Reset() {
	r.pos = 0
	r.count = 0
}

// Tail copies the newest n samples into dst, oldest-first, and returns the count copied (< n if not filled that far).
func (r *Ring) Tail(n int, dst []float64) int {
	if n > r.count {
		n = r.count
	}
	start := (r.pos - n + len(r.buf)) % len(r.buf)
	for i := 0; i < n; i++ {
		dst[i] = r.buf[(start+i)%len(r.buf)]
	}
	return n
}
