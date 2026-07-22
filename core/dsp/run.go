package dsp

import (
	"context"
	"math"
	"time"

	"smegg.me/smeggtuner/common/logger"
	"smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/tuning"
)

// Run blocks until ctx is cancelled or the source channel closes. Freeze state does not survive a
// run: a source swap starts thawed with an empty latch.
func (e *Engine) Run(ctx context.Context, src audio.Source) error {
	blocks, err := src.Start(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = src.Stop() }()

	e.frozen = false
	e.latched = Measurement{}

	sr := src.Info().SampleRate
	ring := NewRing(sr * 10)
	coarse := NewCoarse(sr, e.cfg.A4)
	noise := NewNoiseFloor(e.cfg.CalibSecs)
	det := NewDetector()
	zoom := NewZoom(sr)
	lock := NewLockTracker(e.cfg.LockHold.Seconds(), e.cfg.LockEpsilonHz)

	var hum50, hum60, hp *Biquad
	rebuildFilters := func() {
		hum50, hum60, hp = nil, nil, nil
		if e.cfg.Hum50 {
			hum50 = NewNotch(float64(sr), 50, 30)
		}
		if e.cfg.Hum60 {
			hum60 = NewNotch(float64(sr), 60, 30)
		}
		if e.cfg.Highpass {
			hp = NewHighpass(float64(sr), 15)
		}
	}
	rebuildFilters()

	// Both stages run on every delivered block; these are floors, not timers. With a 4096-sample block
	// (85 ms at 48 kHz) that is ~12 measurements a second. The analysis window is unchanged at three
	// seconds and slides rather than steps; only how often it is read changed.
	coarseEvery := sr / 20
	fineEvery := coarseEvery
	sinceCoarse, sinceFine := 0, 0
	var lastCoarse CoarseResult
	// The newest audio looked at, as a place in the SOURCE. Zero for a microphone (no timeline).
	_, hasTimeline := src.(audio.Transport)
	var srcAt float64
	var floor [tuning.NumNotes]float64
	// floor with the leak guard applied: what detection and the equalizer read.
	var refFloor [tuning.NumNotes]float64
	// The zooms the phase refinement reaches back into.
	var zoomHist []pastZoom
	// When the last fine result landed, so the lock is told how long the reading held.
	var lastFineAt float64
	// The sample the tracked note was adopted at: the spectrum is drawn over this note's own audio.
	var noteAt int
	var consumed int
	var curNote tuning.Note
	// A note waiting to be confirmed, and for how many hops it has held up.
	var pendingNote tuning.Note
	var pendingHops int
	// The shape of the last fine result, so a one-off is not reported.
	var lastShape reedShape
	// level and clipped describe the last completed coarse hop.
	var hopSum, hopPeak float64
	var level float64
	// The loudest hop lately, and whether this one is loud enough to be a note.
	var peakLevel float64
	var sounding bool
	// Hop levels across the fine window, keyed by samples consumed (a hop lands once per block, not per coarseEvery samples).
	type hopLevel struct {
		at    int // samples consumed when this hop closed
		level float64
	}
	windowSamples := int(e.cfg.FineWindow.Seconds() * float64(sr))
	var levelHist []hopLevel
	var steady bool
	var clipped bool

	// A seek shows up as a block that does not start where the last one ended. seamGap sits far above
	// At's rounding and far below any real move, so playback never trips it and a seek always does.
	seamGap := sr / 10
	lastEndSample := 0
	haveLast := false

	for {
		select {
		case <-ctx.Done():
			return nil
		case on := <-e.frzCh:
			e.frozen = on
		case mut := <-e.mutCh:
			oldA4 := e.cfg.A4
			mut(&e.cfg)
			e.cfg.fill()
			if e.cfg.A4 != oldA4 {
				coarse.SetA4(e.cfg.A4)
			}
			rebuildFilters()
			// Reconfigure and restart the settle clock so nothing carries across a live config change.
			lock.Configure(e.cfg.LockHold.Seconds(), e.cfg.LockEpsilonHz)
		case <-e.recalibCh:
			// Measure the room afresh: re-enter the quiet warm-up and drop the note in progress so a
			// reading does not survive the reset. A no-op window (a file, CalibSecs 0) settles at once.
			noise = NewNoiseFloor(e.cfg.CalibSecs)
			lock.Reset()
			zoomHist = zoomHist[:0]
			curNote, pendingNote, pendingHops = 0, 0, 0
			lastShape = reedShape{}
		case b, ok := <-blocks:
			if !ok {
				// A source that can fail at runtime (the mic) reports why it closed; freeze must not
				// mask this. Reuse the state shape so the fields serialize as arrays/numbers, not null.
				if es, isErrSrc := src.(audio.ErrorSource); isErrSrc {
					if err := es.Err(); err != nil {
						e.emit(e.stateMeasurement(StateDeviceLost, lastCoarse, floor, 0, srcAt))
						return err
					}
				}
				return nil
			}
			if len(b.Samples) == 0 {
				continue
			}

			// A seek, caught: a block starting somewhere other than where the last ended means the
			// playhead moved, and the window would then span two unrelated stretches. So on a jump the
			// engine returns to how it was at the run's first block. A microphone has no timeline and
			// never trips this.
			if hasTimeline {
				startSample := int(math.Round(b.At.Seconds() * float64(sr)))
				gap := startSample - lastEndSample
				if gap < 0 {
					gap = -gap
				}
				if haveLast && gap > seamGap {
					ring.Reset()
					lock.Reset()
					zoomHist = zoomHist[:0]
					levelHist = levelHist[:0]
					curNote, pendingNote, pendingHops = 0, 0, 0
					lastShape = reedShape{}
					peakLevel, sounding = 0, false
					steady, e.steady = false, false
					hopSum, hopPeak = 0, 0
					sinceCoarse, sinceFine = 0, 0
					noteAt = consumed
					lastFineAt = float64(consumed) / float64(sr)
				}
				lastEndSample = startSample + len(b.Samples)
				haveLast = true
			}

			// The newest audio looked at, as a place in the recording: the END of this block. Only a
			// source with a timeline has an answer; b.At > 0 cannot be the test (the first block is at
			// zero too, and a mic's blocks are at zero forever).
			if hasTimeline && b.SampleRate > 0 {
				srcAt = (b.At + time.Duration(float64(len(b.Samples))/float64(b.SampleRate)*float64(time.Second))).Seconds()
			}

			if hp != nil {
				hp.Process(b.Samples)
			}
			if hum50 != nil {
				hum50.Process(b.Samples)
			}
			if hum60 != nil {
				hum60.Process(b.Samples)
			}
			// After the filters (the trace shows what the engine hears) and before anything is measured (every emit carries it).
			e.decimateWave(b.Samples)
			for _, s := range b.Samples {
				x := float64(s)
				if a := math.Abs(x); a > hopPeak {
					hopPeak = a
				}
				hopSum += x * x
			}
			ring.Write(b.Samples)
			consumed += len(b.Samples)
			sinceCoarse += len(b.Samples)
			sinceFine += len(b.Samples)

			hopped := false
			if sinceCoarse >= coarseEvery {
				hopped = true
				dt := float64(sinceCoarse) / float64(sr)
				level = math.Sqrt(hopSum / float64(sinceCoarse))
				clipped = hopPeak > 0.98
				hopSum, hopPeak = 0, 0
				sinceCoarse = 0

				if level > peakLevel {
					peakLevel = level
				} else {
					peakLevel *= math.Exp(-dt / peakLevelTau)
				}
				sounding = level >= quietLevel && level >= peakLevel*soundingFrac

				// Keep the levels across the window so the fine stage can tell a note that is holding
				// from one arriving or leaving.
				levelHist = append(levelHist, hopLevel{at: consumed, level: level})
				for len(levelHist) > 1 && consumed-levelHist[0].at > windowSamples {
					levelHist = levelHist[1:]
				}

				// Not "the level held still": two reeds beating swing the level. What disqualifies a
				// window is a trend through it, so compare its two halves.
				steady = false
				if len(levelHist) >= 4 && consumed-levelHist[0].at >= windowSamples*9/10 {
					half := len(levelHist) / 2
					var first, second float64
					for i, h := range levelHist {
						if i < half {
							first += h.level
						} else {
							second += h.level
						}
					}
					first /= float64(half)
					second /= float64(len(levelHist) - half)
					if first > quietLevel && second > quietLevel {
						steady = second <= steadyRatio*first && first <= steadyRatio*second
					}
				}
				e.steady = steady
				lastCoarse = coarse.Analyze(ring)
				noise.Update(lastCoarse, dt)
				floor = noise.Floor()

				if noise.Calibrating() {
					e.emit(e.stateMeasurement(StateInitializing, lastCoarse, floor, level, srcAt))
					continue
				}

				refFloor = refFloorFor(lastCoarse, floor)

				note := e.cfg.ManualNote
				if note == 0 && sounding {
					if n, ok := det.Detect(lastCoarse, refFloor); ok {
						note = n
					} else if curNote != 0 && snrAt(lastCoarse, refFloor, curNote) > 2.5 {
						note = curNote // hold through detector dropouts
					}
				}
				// A note has to be heard more than once before the tuner commits: as a reed starts and
				// dies the window straddles silence and tone, and a single hop in there can name
				// anything. Manual selection is taken at once.
				if note != curNote {
					if e.cfg.ManualNote != 0 || note == pendingNote {
						pendingHops++
					} else {
						pendingNote, pendingHops = note, 1
					}
					if e.cfg.ManualNote != 0 || pendingHops >= confirmHops(curNote, note) {
						curNote = note
						pendingNote, pendingHops = 0, 0
						lock.Reset()
						// A new note is a new band, and has no audio of its own yet: drop the old zooms
						// and grow the window from here.
						zoomHist = zoomHist[:0]
						lastFineAt = float64(consumed) / float64(sr)
						noteAt = consumed
						logger.Debug(msgTrackedNoteChanged, logger.Int("note", int(curNote)))
					}
				} else {
					pendingNote, pendingHops = 0, 0
				}
			}

			measured := false

			// The fine stage runs on the note that is sounding. Whether that is a READING or only a
			// PICTURE is decided below: reeds, beats and the lock wait for `steady`; the spectrum does
			// not, because a short window draws it with fat lobes rather than wrong ones.
			if curNote != 0 && sounding && sinceFine >= fineEvery && !noise.Calibrating() {
				sinceFine = 0
				target := curNote.Transpose(e.cfg.Transpose)
				fc := target.Freq(e.cfg.A4)
				span := math.Max(16, fc*0.035)

				// `steady` is the whole test of a reading; nothing else may be added to it.
				full := steady

				// Only when it is NOT a reading does the window shrink, and then only to this note's own audio.
				window := e.cfg.FineWindow
				if !full {
					if held := consumed - noteAt; held < windowSamples {
						window = time.Duration(float64(held) / float64(sr) * float64(time.Second))
					}
				}

				compound := e.compound()

				zr := ZoomResult{}
				if window >= minSpectrumWindow && !(full && compound) {
					zr = zoom.Analyze(ring, fc, span, window)
				}

				if zr.Valid && !full {
					// The band and the note it is centred on, and nothing else: no number is read off a window too short to give one.
					m := e.stateMeasurement(hopState(level, clipped), lastCoarse, refFloor, level, srcAt)
					m.Note = curNote
					m.NoteName = curNote.Name(e.cfg.Scale)
					m.ScalePitch = fc
					m.Spectrum = e.spectrumFor(zr, fc)
					e.publish(m)
					measured = true
				}

				// The compound stage: one band per octave the register sounds, in place of the single
				// band and its pair machinery. The phase-refinement chain is not run here - the bands
				// re-centre as the key resolves, and the number a compound register is tuned by is the
				// beat within one window, not a tenth of a cent on one line.
				if full && compound {
					now := float64(consumed) / float64(sr)
					if cf, ok := e.compoundAnalyze(zoom, ring, curNote, window, now); ok {
						locked := lock.Observe(cf.freqs, now-lastFineAt)
						lastFineAt = now
						m := e.buildCompound(cf, lastCoarse, refFloor, level, locked, srcAt)
						m.LockProgress = lock.Progress()
						e.attachProfile(&m, zoom, ring, window)
						if shape := compoundShape(cf, len(m.Beats)); shape == lastShape {
							m.State = hopState(level, clipped)
							e.publish(m)
						} else {
							// Not trusted yet, but the meters still have to move.
							lastShape = shape
							e.publish(e.stateMeasurement(hopState(level, clipped), lastCoarse, refFloor, level, srcAt))
						}
						measured = true
					}
				}

				if zr.Valid && full && !compound {
					peaks := FindPeaks(zr, e.cfg.ReedCount, e.lobeWidth())

					// seconds of samples consumed, never wall clock: RefinePhase's error sensitivity is f*dt/hop.
					now := float64(consumed) / float64(sr)

					// The phase reference is not the previous hop: the stage runs every block now, and a
					// reference 85 ms back would triple the frequency error. The history reaches back a
					// quarter of a second, as far as it did when that was as often as the stage ran.
					ref, hop := refZoom(zoomHist, now)
					freqs := make([]float64, len(peaks))
					for i, p := range peaks {
						f := p.Freq
						if ref != nil {
							f = RefinePhase(*ref, zr, f, hop)
						}
						freqs[i] = f * (1 - e.cfg.ClockPPM*1e-6)
					}
					zoomHist = pushZoom(zoomHist, zr, now)

					locked := lock.Observe(freqs, now-lastFineAt)
					lastFineAt = now
					m := e.buildMeasurement(curNote, fc, freqs, peaks, zr, lastCoarse, refFloor, level, locked, srcAt)
					m.LockProgress = lock.Progress()
					e.attachProfile(&m, zoom, ring, window)

					// Wait for the shape of the result to repeat before reporting it: a stray sideband
					// that dressed itself up as a second reed does not survive that.
					shape := reedShape{reeds: len(m.Reeds), beats: len(m.Beats), fromBeat: m.ReedsFromBeat}
					if shape == lastShape {
						m.State = hopState(level, clipped)
						e.publish(m)
					} else {
						// Not trusted yet, but the meters still have to move.
						lastShape = shape
						e.publish(e.stateMeasurement(hopState(level, clipped), lastCoarse, refFloor, level, srcAt))
					}
					measured = true
				}
			}

			// A hop that produced no fine result still reports state, level and equalizer, or the engine
			// goes silent in a quiet room and the UI cannot tell that from a dead engine.
			if hopped && !measured {
				e.publish(e.stateMeasurement(hopState(level, clipped), lastCoarse, refFloor, level, srcAt))
			}
		}
	}
}
