package dsp

import "time"

// The engine's load-bearing, empirically-fixed detection thresholds. They are deliberately package
// constants, not config knobs: a stray value here quietly degrades detection.

// leakGuardRatio floors the detection SNR reference at this fraction of the loudest note, so coarse
// spectral leakage (measured up to 1/8 of a sounding note's energy) stays below the detector's gates
// while a real note keeps snr 1/0.04 = 25. On the headless path (CalibSecs 0) the floor starts at a
// degenerate 1e-6, so this is the only reference there is. Corridor: the golden low-E1 case sits ~8x
// inside it.
const leakGuardRatio = 0.04

// quietLevel is the RMS below which a hop counts as silence. A reed a phone's length away lands ~0.03.
const quietLevel = 1e-3

// soundingFrac: detection's floor is relative, so in silence it collapses onto the noise. A hop is
// only worth identifying within this fraction of the loudest hop lately.
const soundingFrac = 0.02

// peakLevelTau is how fast the loudest-hop memory fades, so a quiet passage after a loud one is heard on its own terms.
const peakLevelTau = 5.0

// steadyRatio is how far the two halves of the beat window may differ in mean level and still count as one note holding.
const steadyRatio = 1.8

// minBeatHz is the slowest beat worth believing. Below it a single reed's own amplitude moves as much
// (bellows, attack, decay); a real E4 reed was reported as a pair beating at 0.41 Hz for that reason.
const minBeatHz = 0.6

// reedBeatDepth is how deeply the amplitude must swing for a modulation to count, on its own, as two
// reeds beating rather than one reed breathing. It is a floor on the EVIDENCE, not on what a pair may
// be: a pair with one reed 6 dB down lands under this bar and is proven a pair by splitPair instead.
const reedBeatDepth = 0.5

// pairDepthSlack is how far the band's beat may fall SHORT of the beat a fitted pair owes and still be
// that pair. One-sided: a bellows working both reeds can push it high, but a pitch modulation putting
// no swing in the amplitude makes it fall short (measured 0.12-0.15). The slack is thin because the
// two populations nearly touch, and where they touch nothing can separate them.
const pairDepthSlack = 0.05

// pairExplained is the least of the band's energy the two fitted lines must account for. Over 432
// synthetic pairs the two lines never held less than 0.94; what a pair is confused with leaves more
// behind (three merged reeds leave 0.67, a worked single reed 0.84-0.87). The bar sits between.
const pairExplained = 0.90

// minSpectrumWindow is the shortest window worth drawing a spectrum over. Too coarse to measure a
// reed, ample to draw one and its sidebands until the real window fills in behind it.
const minSpectrumWindow = 500 * time.Millisecond

// waveFloor is the quietest block that still fills the input strip: below it the scale is held so a
// quiet room's noise is not magnified to full deflection. 0.05 of full scale, ~-26 dBFS.
const waveFloor = 0.05

// The compound-register constants below are calibrated on the recordings in sounds/: single-rank
// takes ('16 and '8 alone), 16+8 pairs and 16+8+8+8 registers, measured by scripts once and fixed here.

// crossBandFloor is how loud a line in any octave band must be, against the loudest line across all
// of the register's bands, to be a reed at all. FindPeaks floors within one band, so an octave whose
// rank is silent would otherwise report its noise. Genuine quiet ranks measured down to 0.13 of the
// loudest; an empty band's noise never above 0.007.
const crossBandFloor = 0.04

// ghostFloorHz and ghostRelTol give the half-window around an exact multiple of a lower rank's line
// within which a line is that rank's partial, not a reed: max(ghostFloorHz, f*ghostRelTol). A reed's
// partials are exact multiples (a periodic oscillation has no choice), measured within 0.06 Hz;
// the nearest genuine reed measured 1.7 Hz away, and a merged reed-plus-harmonic composite is pulled
// at least 0.18 Hz off the multiple.
const ghostFloorHz = 0.12
const ghostRelTol = 3e-4

// partialImageTol is how close an upper band's line spacing must sit to k times the lower band's
// envelope beat to be that cluster's partial image (see markPartialImages). The spacings are the
// same physics measured twice, so they agree to the envelope's own error, well under this.
const partialImageTol = 0.25

// plvLocked is the phase-locking value above which an octave band is only the lower rank's harmonic.
// A harmonic rides its fundamental's every wobble, so against the fundamental's doubled phase it
// measures near 1; an independent reed drifts on its own. On a clean three-second window blocked
// ranks measure 0.954..0.997 and independent reeds 0.006..0.865, but live sliding windows overlap
// (a reed in tune to under a third of a hertz holds 0.92+ too), so this alone rules only until the
// residual's rotation has been watched long enough - see rotIndependent.
const plvLocked = 0.92

// rotIndependent is how fast the residual's angle must walk, in rad/s, for the band to hold an
// independent reed rather than the lower rank's partial: 2*pi times the slowest octave beat worth
// calling a second voice (~0.16 Hz). A partial's angle holds still only in principle - pressure
// changes ramp and step it (a reed's onset is not its sustain), and the measurement's own bias
// sweeps as drift walks the lines across the analysis grid - so blocked-rank recordings show
// apparent rates up to 0.85 rad/s, while the slowest genuine second voice measured 1.19. A rank in
// tune beyond this floor reads as harmonic-only, which a real pair leaves within seconds of pumping.
const rotIndependent = 1.0

// profileExcess is how far a band's line must stand over the calibrated partial (RankProfile) to
// prove an independent reed by amplitude alone - the judgement that works at zero beat. A rank's
// ratio swings up to 1.7x across a take as the bellows works (measured over the solo-rank
// recordings in sounds/); a genuine second voice adds 3..20x the profiled partial.
const profileExcess = 2.5

// profileMinRatio is the calibrated ratio below which a profile entry is treated as absent: some
// reeds' partials vanish at low pressure (measured to 0.00), and an expectation of nearly nothing
// would call anything at all an independent voice.
const profileMinRatio = 0.02

// plvWarmIndependent rules the solo case before the rotation tracker has settled: only a band this
// decisively unlocked (a beat fast enough to decorrelate within one window) is called independent
// on sight. A blocked rank's onset dips to 0.80; a fast-beating real pair sits well under this.
const plvWarmIndependent = 0.7
