# Measurement fixtures

Drop reference recordings here as `<name>.wav` with a matching
`<name>.expect.json`. The fixture test runs each pair through the
measurement engine (free-run, auto or per-expect settings), reduces the
whole measurement stream with `dsp.Aggregate` (levelFrac 0.5) and
asserts the aggregated result. When no WAV files are present the test
skips, so the suite stays green on a bare checkout.

Why aggregation: real hand-pumped recordings never satisfy the lock
tracker (bellows drift exceeds the lock window) and end in a silent
tail where note detection latches onto room noise. Trusting any single
tick therefore reports garbage. Aggregate keeps only running ticks at
or above half the recording's peak input level, votes for the majority
note, and takes per-reed median frequencies over those ticks.

expect.json fields:

    {
      "comment": "origin/license note",  // free text, ignored by the test
      "reeds": 2,            // reed count to measure (1..3)
      "a4": 440,             // optional, default 440
      "note": "A4",          // optional expected detected note name
      "freqs": [440.0, 442.6],  // expected reed frequencies, ascending
      "freqTol": 0.1,        // optional Hz tolerance, default 0.1
      "beats": [2.6],        // optional expected beats between adjacent reeds
      "beatTol": 0.1         // optional Hz tolerance, default 0.1
    }

Golden workflow (measure, then freeze): run the harness on the new
recording without pinning the note,

    go run ./scripts/analyze -wav tests/fixtures/<name>.wav

check that the reported note and frequencies are sensible, then copy
the measured values into `<name>.expect.json` verbatim. The goldens are
regression anchors for the measured behavior, not independent truth:
the recordings are lossy (AAC-sourced) and hand-pumped, so use a
generous freqTol (0.15 Hz for the technician set) and expect values a
few cents off concert pitch.

The six `a-8` .. `h-16` fixtures are technician-supplied reference
samples (2026-07-11), converted from the `sounds/` m4a set with
`ffmpeg -ac 1 -ar 48000`. Names keep the German source labels: Fis =
F#, H = B natural; '8 is the clarinet register at written pitch, '16
the bassoon register an octave below.

Guidelines for recordings: WAV (any common bit depth/rate), a single
sustained note per file, at least 5 seconds, minimal room noise. Name
files like `mm-register-a4-2reeds.wav` or after the source label. State
the recording's origin and license in the comment field of the expect
file if it did not come from this project's maintainer.
