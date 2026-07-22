import { EQUALIZER_FULL_SCALE_DB } from "~/composables/useTuner";

// Drawing geometry only. One of the two places the frontend may do arithmetic (architecture.md rule 1) - every input came from a DTO.

// RULER_CENTS: fifty cents either side - one semitone, past which the engine tracks the next note.
export const RULER_CENTS = 50;

const clamp = (v: number, lo: number, hi: number) =>
  Math.max(lo, Math.min(hi, v));

// centToFrac maps a cent deviation onto 0..1 across the ruler; clamps, else a badly-out reed draws off-canvas and invisible.
export function centToFrac(cent: number): number {
  return clamp((cent + RULER_CENTS) / (RULER_CENTS * 2), 0, 1);
}

// bandToFrac maps a note-strip band (decibels over the engine's reference, bounded by EQUALIZER_FULL_SCALE_DB, not the wire's ceiling) onto 0..1. The clamp stays: a byte off the wire is not a promise.
export function bandToFrac(db: number): number {
  return clamp(db / EQUALIZER_FULL_SCALE_DB, 0, 1);
}
