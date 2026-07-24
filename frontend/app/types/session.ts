// Mirrors core/session. Hand-written so presentational components can be built and tested before the services that expose them are bound.

/** A bank is a rank of reeds and the card's column. L is 16', M1..M4 the 8' musette ranks, H the 4'. Frequency doesn't reveal the rank, so a register names its banks and a take names its register. */
export type Bank = 'L' | 'M1' | 'M2' | 'M3' | 'M4' | 'H'

// Card order, low to high: the order columns print in.
export const BANKS: Bank[] = ['L', 'M1', 'M2', 'M3', 'M4', 'H']

export interface Register {
  name: string
  banks: Bank[]
}

/** One of the bass machine's switches: which ranks it sounds, by foot (32, 16, 8...). */
export interface BassRegister {
  name: string
  feet: number[]
}

export interface Instrument {
  /** What the technician calls this accordion. Travels with the session, since a session opens on machines whose shelf never heard of this instrument. */
  name?: string
  serial: string
  // Reeds per note on the bench, 1..8; not assumed to be 3 (a bass sounds five voices on one button).
  reedCount: number
  // Every rank in card order. Absent on an undescribed instrument, whose card numbers its columns instead.
  banks?: Bank[]
  registers?: Register[]
  // Lowest and highest key. Zero for an instrument nobody described.
  lo?: number
  hi?: number
  // The bass machine: how many octave-stacked ranks it sounds (2..6, usually 4 or 5); absent means
  // no bass section described. Its switches, when it has any - an older fixed machine has none.
  bassReeds?: number
  bassRegisters?: BassRegister[]
  // This accordion's own reference pitch (e.g. 442 vs 440); absent falls back to the app default. See core/session.Instrument.
  a4?: number
  // How tight this accordion is judged, in cents; absent falls back to the app default. See core/session.Instrument.Tolerances.
  tolerance?: number
  beatTolerance?: number
}

// An already-described instrument, reused. Builtin ones ship with the app and cannot be changed.
export interface InstrumentTemplate {
  id: string
  name: string
  instrument: Instrument
  /** There is a photograph. The bytes are never here - a shelf of fifty over JSON-RPC would be megabytes of base64 - it's fetched from /instruments/<id>/image. See useInstruments.imageOf. */
  hasImage: boolean
  /** Changes when the photograph does and goes in the image URL, so a replaced image isn't served stale from the webview's cache. See useInstruments.imageOf. */
  imageRev?: number
}

// The session itself (readings, summary, bench) is not mirrored here: it crosses the bindings as
// generated types (SessionDTO, Summary, Bench), and a hand-written copy only drifts.

export const MIN_REEDS = 1
export const MAX_REEDS = 8

// A4 range the schema allows (common/config: 430..450).
export const A4_MIN = 430
export const A4_MAX = 450
export const A4_DEFAULT = 440

export function emptyInstrument(): Instrument {
  return { make: '', model: '', serial: '', reedCount: 1 }
}
