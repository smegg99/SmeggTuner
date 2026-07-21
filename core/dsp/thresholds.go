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
