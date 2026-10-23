# Measurement fixtures

Drop reference recordings here as `<name>.wav` with a matching
`<name>.expect.json`. `TestFixtures` plays each pair through the engine
the way the app plays a file (no background learning) and asserts:

- the note: the one most analyses name, holding at least two thirds of
  them;
- each slot's frequency: the median over the analyses on that note made
  while the reed sounds, meaning an input level of at least half the
  recording's peak. That drops the analyses whose window straddles a
  bellows reversal;
- a take: the recording locks at least once.

These committed fixtures are required. Missing WAV files or invalid audio fail
the check. The separate tests using the optional `sounds/` folder skip only
when their recordings are absent; decode and engine failures fail the check.

Why a median and not the take: the take is the analysis where the
reference lock first rises, one instant of a hand-pumped note that
drifts a few tenths of a hertz across the recording. The median over the
sounding analyses is where the needle sits while the note sustains.

expect.json fields:

    {
      "comment": "origin/license note",  // free text, ignored by the test
      "reeds": 2,            // reed slots to measure (1..6)
      "a4": 440,             // optional, default 440
      "note": "A4",          // optional expected note name
      "freqs": [440.0, 442.6],  // expected slot frequencies, in slot order
      "freqTol": 0.1,        // optional Hz tolerance, default 0.1
      "beats": [2.6],        // optional expected beats between adjacent slots, Hz
      "beatTol": 0.1         // optional Hz tolerance, default 0.1
    }

Golden workflow (measure, then freeze): run the harness on the new
recording without pinning the note,

    go run ./scripts/analyze -wav tests/fixtures/<name>.wav -trace

check that the reported note and frequencies are sensible, then copy the
median of the sounding analyses into `<name>.expect.json`. The goldens
are regression anchors, not independent truth: the recordings are lossy
(AAC-sourced) and hand-pumped, so use a generous freqTol (0.15 Hz for
the technician set) and expect values a few cents off concert pitch. The
current values were frozen from the previous estimator and still hold on
the reference-compatible engine; the largest difference measured on
2026-10-24 was 0.094 Hz (fis-8).

The six `a-8` .. `h-16` fixtures are technician-supplied reference
samples (2026-07-11), converted from the `sounds/` m4a set with
`ffmpeg -ac 1 -ar 48000`. Names keep the German source labels: Fis =
F#, H = B natural; '8 is the clarinet register at written pitch, '16
the bassoon register an octave below.

Guidelines for recordings: WAV (any common bit depth/rate), a single
sustained note per file, at least 6 seconds (the first take needs about
five), minimal room noise. Name files like `mm-register-a4-2reeds.wav`
or after the source label. State the recording's origin and license in
the comment field of the expect file if it did not come from this
project's maintainer.
