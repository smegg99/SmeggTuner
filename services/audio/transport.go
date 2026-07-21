package audio

import (
	"time"

	coreaudio "smegg.me/smeggtuner/core/audio"
)

// TransportDTO is the file transport state the view draws; Available is false for a mic, times in seconds not samples.
type TransportDTO struct {
	Available  bool    `json:"available"`
	Name       string  `json:"name"`
	Duration   float64 `json:"duration"`
	Position   float64 `json:"position"`
	From       float64 `json:"from"`
	To         float64 `json:"to"`
	Paused     bool    `json:"paused"`
	Moving     bool    `json:"moving"` // sound is actually coming out now
	Loop       bool    `json:"loop"`
	SampleRate int     `json:"sampleRate"` // the recording's own rate, from the header
}

// PeakDTO is one drawn waveform column: the extremes of the audio, not its average.
type PeakDTO struct {
	Min float32 `json:"min"`
	Max float32 `json:"max"`
}

// transport is the current file, or nil (a typed nil in an interface is not a nil interface).
func (s *Service) transport() coreaudio.Transport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.file == nil || s.current.Kind != SourceFile {
		return nil
	}
	return s.file
}

func secs(d time.Duration) float64 { return d.Seconds() }

func dur(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
}

// Transport reports the playhead, selection and run state; a fresh view calls this on mount.
func (s *Service) Transport() TransportDTO {
	t := s.transport()
	if t == nil {
		return TransportDTO{Available: false}
	}
	from, to := t.Selection()

	s.mu.RLock()
	name, loop := s.current.Name, s.current.Loop
	s.mu.RUnlock()

	s.mu.RLock()
	rate := 0
	if s.file != nil {
		rate = s.file.Info().SampleRate
	}
	s.mu.RUnlock()

	return TransportDTO{
		Available:  true,
		Name:       name,
		Duration:   secs(t.Duration()),
		Position:   secs(t.Position()),
		From:       secs(from),
		To:         secs(to),
		Paused:     t.Paused(),
		Moving:     t.Moving(),
		Loop:       loop,
		SampleRate: rate,
	}
}

// Seek moves the playhead; core/audio clamps it to the selection.
func (s *Service) Seek(seconds float64) TransportDTO {
	if t := s.transport(); t != nil {
		t.Seek(dur(seconds))
	}
	return s.Transport()
}

func (s *Service) SetPaused(paused bool) TransportDTO {
	if t := s.transport(); t != nil {
		t.SetPaused(paused)
	}
	return s.Transport()
}

// SetRange selects a fragment; an empty or backwards range means the whole file.
func (s *Service) SetRange(from, to float64) TransportDTO {
	if t := s.transport(); t != nil {
		t.SetRange(dur(from), dur(to))
	}
	return s.Transport()
}

// SetLoop is stored on the selection too, because the selection is what survives a restart and what the config remembers.
func (s *Service) SetLoop(loop bool) TransportDTO {
	if t := s.transport(); t != nil {
		t.SetLoop(loop)
	}
	s.mu.Lock()
	if s.current.Kind == SourceFile {
		s.current.Loop = loop
	}
	s.mu.Unlock()
	return s.Transport()
}

// Peaks is the waveform between two times, one column per bucket; no cache, the in-memory sweep is cheaper than the IPC call.
func (s *Service) Peaks(from, to float64, buckets int) []PeakDTO {
	t := s.transport()
	if t == nil {
		return nil
	}

	if buckets > maxPeakBuckets {
		buckets = maxPeakBuckets
	}

	peaks := t.Peaks(dur(from), dur(to), buckets)
	out := make([]PeakDTO, len(peaks))
	for i, p := range peaks {
		out[i] = PeakDTO{Min: p.Min, Max: p.Max}
	}
	return out
}

// maxPeakBuckets caps requested columns, far wider than any real display.
const maxPeakBuckets = 8192
