import type { RowDTO } from '~~bindings/smegg.me/smeggtuner/services/record/models.js'

export interface Step {
  note: number
}

export function buildSteps(lo: number, hi: number): Step[] {
  const out: Step[] = []
  for (let n = lo; n <= hi; n++) out.push({ note: n })
  return out
}

// Idempotent: a replayed note updates its own row, so the sweep never double-advances.
export function isRecorded(rows: RowDTO[], note: number, register: string): boolean {
  return rows.some(r => Number(r.note) === note && String(r.register ?? '') === register)
}

export function nextUndone(steps: Step[], from: number, recorded: (s: Step) => boolean): number {
  let i = from
  while (i < steps.length && recorded(steps[i]!)) i++
  return i
}
