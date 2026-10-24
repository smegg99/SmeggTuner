import type { RowDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/record/models.js'

// The footer's session label. The in-tune verdict is the backend's (inTol); this only counts it.

/** Only the fields the label reads, so the DTO and the hand-type both satisfy it. */
export interface ProgressSession {
  name: string
  instrument?: { lo?: number, hi?: number } | null
  curve?: { anchors?: unknown[] | null } | null
}

export type SessionProgress
  /** Nothing open: tuning freely. */
  = | { kind: 'none' }
  /** Finished: takes no more readings. */
    | { kind: 'idle', name: string }
  /** A pass being recorded. total is null when the instrument never gave its keyboard span; inTune is null with no goal curve or nothing recorded. */
    | { kind: 'active', name: string, done: number, total: number | null, inTune: number | null }

export function sessionProgress(
  session: ProgressSession | null,
  rows: RowDTO[],
  recording: boolean,
): SessionProgress {
  if (!session) return { kind: 'none' }
  if (!recording) return { kind: 'idle', name: session.name }

  const done = distinctNotes(rows)
  return {
    kind: 'active',
    name: session.name,
    done,
    total: keyboardKeys(session),
    inTune: hasGoalCurve(session) && done > 0 ? notesInTune(rows) : null,
  }
}

// Distinct notes; the same key on two registers counts once.
function distinctNotes(rows: RowDTO[]): number {
  return new Set(rows.map(row => row.note)).size
}

// Keyboard key count, or null when the instrument never gave its span (lo/hi zero, or hi below lo).
function keyboardKeys(session: ProgressSession): number | null {
  const lo = session.instrument?.lo
  const hi = session.instrument?.hi
  if (!lo || !hi || hi < lo) return null
  return hi - lo + 1
}

function hasGoalCurve(session: ProgressSession): boolean {
  return (session.curve?.anchors?.length ?? 0) > 0
}

// A note is in tune only when every row it holds is (both registers, if taken twice).
function notesInTune(rows: RowDTO[]): number {
  const byNote = new Map<number, RowDTO[]>()
  for (const row of rows) {
    const list = byNote.get(row.note)
    if (list) list.push(row)
    else byNote.set(row.note, [row])
  }

  let count = 0
  for (const list of byNote.values()) {
    if (list.every(rowInTune)) count++
  }
  return count
}

// A row is in tune when every slot that sounded is in tolerance; a row with none heard is not a pass.
function rowInTune(row: RowDTO): boolean {
  return row.reeds.length > 0 && row.reeds.every(reed => reed.inTol)
}
