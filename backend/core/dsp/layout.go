package dsp

import "smegg.me/smeggtuner/backend/core/dsp/meter"

// spans reports whether the layout reaches outside the key's own octave.
func (c *EngineConfig) spans() bool {
	for _, o := range c.Octaves {
		if o.Offset != 0 {
			return true
		}
	}
	return false
}

// bands is the layout relative to its lowest band, and the semitones from that band up to the key.
func (c *EngineConfig) bands() (bands []meter.Band, keyAbove int) {
	if !c.spans() {
		return nil, 0
	}
	low := c.Octaves[0].Offset
	for _, o := range c.Octaves {
		if n := len(bands); n > 0 && bands[n-1].Offset == o.Offset-low {
			bands[n-1].Reeds += o.Reeds
			continue
		}
		bands = append(bands, meter.Band{Offset: o.Offset - low, Reeds: o.Reeds})
	}
	if c.Bass {
		return bands, 0
	}
	return bands, -low
}

// dirkBass reports whether the layout runs the reference Bass mode.
func (c *EngineConfig) dirkBass() bool {
	bands, _ := c.bands()
	_, ok := meter.DirkBass(bands)
	return ok && !c.Chords
}

// reedIDs is the enabled reed boxes of one band.
func (c *EngineConfig) reedIDs() []int {
	if len(c.Reeds) > 0 {
		return c.Reeds
	}
	ids := make([]int, c.ReedCount)
	for i := range ids {
		ids[i] = i + 1
	}
	return ids
}

// meterSlots is the meter readout slot each measured slot reads: the enabled reed boxes of one band,
// the base and enabled upper boxes of the reference Bass mode, otherwise every slot in order.
func (c *EngineConfig) meterSlots() []int {
	var ids []int
	switch {
	case c.Chords:
		return []int{0, 1, 2}
	case !c.spans():
		for _, id := range c.reedIDs() {
			ids = append(ids, id-1)
		}
		return ids
	case c.dirkBass():
		bands, _ := c.bands()
		for _, b := range bands {
			ids = append(ids, b.Offset/12)
		}
		return ids
	}
	bands, _ := c.bands()
	for _, b := range bands {
		for range b.Reeds {
			ids = append(ids, len(ids))
		}
	}
	return ids
}

// slots is how many reed readouts the configuration measures.
func (c *EngineConfig) slots() int { return len(c.meterSlots()) }

// pairs says which neighbouring measured slots beat: reference neighbours only, never chord tones.
func (c *EngineConfig) pairs() []bool {
	ids := c.meterSlots()
	out := make([]bool, len(ids)-1)
	for k := range out {
		out[k] = ids[k+1] == ids[k]+1 && !c.Chords
	}
	return out
}

// pairPanels reports whether the pair readouts show: the reference Bass mode has octave beat boxes only in Hz.
func (c *EngineConfig) pairPanels() bool { return c.BeatsInHertz || !c.dirkBass() }

// slotOctaves is each slot's semitones from the key in the layout; one band and chords are all zero.
func (c *EngineConfig) slotOctaves() []int {
	out := make([]int, c.slots())
	if c.Chords {
		return out
	}
	bands, keyAbove := c.bands()
	i := 0
	for _, b := range bands {
		for range b.Reeds {
			out[i] = b.Offset - keyAbove
			i++
		}
	}
	return out
}
