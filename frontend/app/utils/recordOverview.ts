import type { RowDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/record/models.js'
import { restate } from '~/utils/deviation'
import type { DeviationUnit } from '~/utils/deviation'
import { byNote } from '~/utils/record'

// The record table's graph: one column per note, each reed's error against its goal as a mark.

export interface OverviewMark {
  /** the reed's error in the display unit */
  error: number
  inTol: boolean
}

export interface OverviewColumn {
  note: number
  noteName: string
  /** the note's rows (one per register), in table order; a click cycles through them */
  keys: string[]
  marks: OverviewMark[]
}

/** The row key recordedRows gives a voice. */
export const rowKey = (row: Pick<RowDTO, 'note' | 'register'>) => `${row.note}:${row.register ?? ''}`

export function overviewColumns(rows: readonly RowDTO[], unit: DeviationUnit): OverviewColumn[] {
  const out: OverviewColumn[] = []
  for (const row of byNote([...rows])) {
    let col = out[out.length - 1]
    if (col?.note !== row.note) {
      col = { note: row.note, noteName: row.noteName, keys: [], marks: [] }
      out.push(col)
    }
    col.keys.push(rowKey(row))
    for (const r of row.reeds) col.marks.push({ error: restate(r, unit).error, inTol: r.inTol })
  }
  return out
}

/** nextKey: a click selects the column's first row, or the one after the row already selected there. */
export function nextKey(col: OverviewColumn, selected: string | null | undefined): string {
  const at = selected ? col.keys.indexOf(selected) : -1
  return col.keys[(at + 1) % col.keys.length]!
}

/** overviewSpan is the graph's half height: the largest error, at least twice the tolerance, in whole steps of 5. */
export function overviewSpan(columns: readonly OverviewColumn[], tolerance: number): number {
  const worst = Math.max(2 * tolerance, ...columns.flatMap(c => c.marks.map(m => Math.abs(m.error))))
  return Math.max(5, Math.ceil(worst / 5) * 5)
}
