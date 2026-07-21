// core/audio/filepeaks.go
package audio

import (
	"time"

	gaudio "github.com/go-audio/audio"
)

// downmixMono averages the buffer's channels into a mono float32 signal. The format guard
// matters: go-audio validates bit depth but copies the fmt chunk's channel count through
// unchecked, so a zero channel count would divide by zero and a zero bit depth shift by -1,
// both panics on arbitrary user files.
func downmixMono(buf *gaudio.IntBuffer) ([]float32, error) {
	ch := buf.Format.NumChannels
	if ch <= 0 || buf.SourceBitDepth <= 0 {
		return nil, errUnsupportedFormat
	}
	scale := float32(int64(1) << (buf.SourceBitDepth - 1))
	frames := len(buf.Data) / ch
	mono := make([]float32, frames)
	for i := 0; i < frames; i++ {
		var sum float32
		for c := 0; c < ch; c++ {
			sum += float32(buf.Data[i*ch+c]) / scale
		}
		mono[i] = sum / float32(ch)
	}
	return mono, nil
}

// Peaks reduces [from, to) to one min/max pair per bucket: the waveform at the view's zoom.
// Min AND max (not an average of magnitudes) because the envelope is the information - where
// the note starts, where the bellows turned, whether the take is clipped.
func (s *FileSource) Peaks(from, to time.Duration, buckets int) []Peak {
	if buckets <= 0 {
		return nil
	}
	lo := clampInt(s.sampleAt(from), 0, len(s.samples))
	hi := clampInt(s.sampleAt(to), lo, len(s.samples))
	if hi <= lo {
		return nil
	}

	out := make([]Peak, buckets)
	span := hi - lo
	for b := range out {
		// Each boundary scaled from the exact fraction, not stepped by a truncated bucket
		// width: with fewer samples than buckets a truncated width of zero leaves the span
		// undrawn.
		start := lo + b*span/buckets
		end := lo + (b+1)*span/buckets
		if end <= start {
			end = start + 1
		}
		if end > hi {
			end = hi
		}

		mn, mx := s.samples[start], s.samples[start]
		for _, v := range s.samples[start:end] {
			if v < mn {
				mn = v
			}
			if v > mx {
				mx = v
			}
		}
		out[b] = Peak{Min: mn, Max: mx}
	}
	return out
}

func (s *FileSource) timeAt(sample int) time.Duration {
	if s.sampleRate <= 0 {
		return 0
	}
	return time.Duration(float64(sample) / float64(s.sampleRate) * float64(time.Second))
}

func (s *FileSource) sampleAt(at time.Duration) int {
	return int(at.Seconds() * float64(s.sampleRate))
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
