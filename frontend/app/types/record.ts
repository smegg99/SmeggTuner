import type { Bank } from '~/types/session'

// DTOs mirroring backend JSON (core/target, services/record); nothing here is computed frontend-side.

/** The unit a value is read in. The same two strings core/target's Unit uses. */
export type RecordUnit = 'cent' | 'hz'

/** One reed of one note, measured against the goal curve. core/target.ReedError. */
export interface ReedError {
  /** 0-based, as the curve's RefReed is: 0 is reed 1. */
  reed: number
  /** Cents from scale pitch, measured. */
  curr: number
  /** Cents from scale pitch, from the curve. 0 when the session has no curve. */
  goal: number
  /** Curr - Goal: what has to come off the reed. */
  error: number

  /** The same three in hertz; the backend carries them because cents-to-hertz depends on scale pitch, which the frontend never computes. */
  currHz: number
  goalHz: number
  errorHz: number

  /** The backend's verdict; never re-derived here. */
  inTol: boolean
}

/** One pair of reeds of one note: the beat between them. core/target.BeatError. */
export interface BeatError {
  /** "1-2", "1-3", "2-3": 1-based, as dsp.BeatMeasure spells it. */
  pair: string
  /** 0-based reed indices, lower pitch first. */
  low: number
  high: number

  curr: number
  goal: number
  error: number

  /** The same three as beat rates at this note. */
  currHz: number
  goalHz: number
  errorHz: number

  inTol: boolean
  /** Beat read off the amplitude envelope, not the spectrum; on a merged note it's the only reliable reading. */
  fromEnvelope: boolean
}

/** One row of the tuning table. services/record.RowDTO. One reading per voice - replaying replaces it - so `take` is its index into the session's readings. */
export interface TakeRow {
  /** MIDI note number. */
  note: number
  noteName: string
  /** The switch this note was played on; distinguishes readings of different reeds at the same note. Absent when the session describes no register, and the table then holds one row per note. */
  register?: string
  /** Which column each reed belongs in: reeds[i] is bank banks[i]. Absent when the mapping isn't certain; the table then numbers its reeds rather than naming them, never guessing. */
  banks?: Bank[]
  /** Index into the session's readings of the one this row shows: what an edit or removal aims at. */
  take: number
  at: string
  reeds: ReedError[]
  beats: BeatError[]
  /** Spectrum couldn't separate this note's reeds, so per-reed numbers are lobes of one peak - unless reedsFromBeat is set. See reedsUsable() in app/utils/record.ts. */
  reedsMerged: boolean
  /** The reeds were recovered from the measured beat. Optional: older tables omit it. */
  reedsFromBeat?: boolean
  /** A Curr was hand-edited: typed, not heard. */
  manual: boolean
}

/** The session's readings, against the goal. services/record.TableDTO. */
export interface PassView {
  id: string
  label: string
  at: string
  /** The reference every reading was measured against. */
  a4: number
  /** The instrument's ranks in card order: the columns the table prints. Absent on an instrument nobody described, whose table numbers its reeds instead. */
  banks?: Bank[]
  rows: TakeRow[]
}

/** What the tuning table emits when a Curr cell is committed. */
export interface CurrEdit {
  /** The take the row stands for: services/record.EditReed takes this index. */
  take: number
  /** 0-based reed. */
  reed: number
  value: number
  /** The unit `value` was typed in, i.e. the unit the table is read in; services/record.EditReed converts. */
  unit: RecordUnit
}

/** How the card is exported. `pdf` isn't a third renderer: it's the HTML sheet rendered to A4 by a headless browser on the backend. */
export type ReportFormat = 'html' | 'pdf' | 'csv'

/** What the report dialog emits, the genuinely per-export fields. The letterhead is not here: it's read from Settings when the sheet is written, so it can't go stale. */
export interface ReportOptions {
  format: ReportFormat
  /** ISO yyyy-mm-dd. */
  date: string
}
