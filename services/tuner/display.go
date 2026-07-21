package tuner

import "smegg.me/smeggtuner/core/dsp"

// rules decides whether a measurement repaints the screen. None of these reach the engine or recorder, so the recorded pass is the same either way.
type rules struct {
	// stopAfterLock holds the reading when the tone locks, releasing when the note does.
	stopAfterLock bool
	// continuousManual keeps a pinned note's numbers moving between locks.
	continuousManual bool
	// manual is whether the note is pinned at all (the detector's own tracking never stops).
	manual bool
}

// holder applies the display rules to the measurement stream; it lives on the emitting goroutine alone, so it needs no lock.
type holder struct{ held bool }

// filter returns the measurement to show. A reading that must hold comes back stripped to its meters (the heartbeat shape) so the strip does not read as a dead engine.
func (h *holder) filter(m dsp.Measurement, r rules) dsp.Measurement {
	if m.ScalePitch <= 0 {
		return m // a heartbeat carries no reading to hold
	}
	if !m.Locked {
		h.held = false
		if r.manual && !r.continuousManual {
			return meters(m)
		}
		return m
	}
	if r.stopAfterLock {
		if h.held {
			return meters(m)
		}
		h.held = true // this one lands; the ones behind it hold still
	}
	return m
}

// meters is a measurement with its reading removed: the heartbeat shape core/dsp emits.
func meters(m dsp.Measurement) dsp.Measurement {
	return dsp.Measurement{
		State:      m.State,
		InputLevel: m.InputLevel,
		Equalizer:  m.Equalizer,
		Waveform:   m.Waveform,
	}
}
