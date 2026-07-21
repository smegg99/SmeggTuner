// The goal curve as core/target writes it to JSON; every field mirrors a Go field, keep the two in step.

/** Unit the anchors were authored in; storage is always cents. */
export type CurveUnit = 'cent' | 'hz'

export const CURVE_UNITS: readonly CurveUnit[] = ['cent', 'hz']

// core/tuning: notes the tuner tracks, as MIDI numbers.
export const NOTE_MIN = 16
export const NOTE_MAX = 120

// core/target: reeds a curve may describe; don't assume 3 (the bass side runs to five).
export const MIN_REEDS = 1
export const MAX_REEDS = 8

/** RefReed when no reed is defined as sounding at pitch. */
export const NO_REF_REED = -1

/** core/target MaxAsymmetry: the percent bounds of Curve.asymmetry, either way. */
export const MAX_ASYMMETRY = 100

/** A beating needs two reeds; core/target refuses fewer. */
export const MIN_BEATING_REEDS = 2

/** One note the curve is pinned at. reeds holds cents from that note's scale pitch, index 0 = reed 1, and may be shorter than reedCount (a missing reed is at zero). */
export interface Anchor {
  note: number
  reeds: number[]
}

/** The goal: sparse anchors, interpolated between and held flat outside. Zero anchors is legal, not degenerate - it makes the tuner a pure indicator; nothing may treat it as an error or empty state. */
export interface Curve {
  name: string
  reedCount: number
  /** 0-based reed defined as sounding at pitch, or NO_REF_REED. A convention, never an assumption. */
  refReed: number
  /** Sparse, sorted by note, at most one per note. */
  anchors: Anchor[]
  unit: CurveUnit

  /** Where the reference reed sits in the tremolo, percent, -MAX_ASYMMETRY..MAX_ASYMMETRY; 0 symmetric. Read by the backend when a beating is entered, never re-applied to what's already entered. */
  asymmetry: number

  /** Interpolation between/beyond anchors; all three default true in core/target. Never build a Curve from a partial object: the generated Curve.createFrom defaults a missing bool to false. Bind the backend's DTO and pass it unchanged. */
  interpolate: boolean
  extrapolateLeft: boolean
  extrapolateRight: boolean
}
