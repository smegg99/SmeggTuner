import type { RowDTO } from '~~bindings/smegg.me/smeggtuner/services/record/models.js'

export interface Step {
  note: number
  /** A bass step asks for a button - a pitch class - not a key: which octave the ladder answers at
   * is the machine's own business, so the note is detected rather than pinned. */
  pc?: number
}

export function buildSteps(lo: number, hi: number): Step[] {
  const out: Step[] = []
  for (let n = lo; n <= hi; n++) out.push({ note: n })
  return out
}

// The bass keyboard is twelve buttons, whatever it spans.
export function buildBassSteps(): Step[] {
  return Array.from({ length: 12 }, (_, pc) => ({ note: -1, pc }))
}

// Idempotent: a replayed note updates its own row, so the sweep never double-advances.
export function isRecorded(rows: RowDTO[], note: number, register: string): boolean {
  return rows.some(r => !r.bass && Number(r.note) === note && String(r.register ?? '') === register)
}

// A bass button is recorded when any bass row of its pitch class stands under the pulled switch.
export function isRecordedBass(rows: RowDTO[], pc: number, register: string): boolean {
  return rows.some(r => Boolean(r.bass) && Number(r.note) % 12 === pc && String(r.register ?? '') === register)
}

export function nextUndone(steps: Step[], from: number, recorded: (s: Step) => boolean): number {
  let i = from
  while (i < steps.length && recorded(steps[i]!)) i++
  return i
}
