import type { BeatError, ReedError, TakeRow } from "~/types/record";

// core/session: reeds sounding per note, 1..8; the engine resolves three today but a hardcoded 3 anywhere is a bug.
export const MIN_REEDS = 1;
export const MAX_REEDS = 8;

// EMPTY is what a cell holds when a reed did not sound; deliberately not zero, which reads as a perfectly tuned reed.
export const EMPTY = "-";

export interface ReedGroup {
  kind: "reed";
  key: string;
  /** 0-based, as ReedError.reed is. */
  reed: number;
}

export interface BeatGroup {
  kind: "beat";
  key: string;
  /** "1-2", "2-3": how BeatError.pair spells it, 1-based. */
  pair: string;
  low: number;
  high: number;
}

export type ColumnGroup = ReedGroup | BeatGroup;

export function reedCountOf(count: number): number {
  if (!Number.isFinite(count)) return MIN_REEDS;
  return Math.min(MAX_REEDS, Math.max(MIN_REEDS, Math.trunc(count)));
}

/** The reeds an instrument of this many reeds has, 0-based. */
export function reedIndexes(count: number): number[] {
  return Array.from({ length: reedCountOf(count) }, (_, index) => index);
}

// columnGroups lists a comparison's columns left to right (reed, its beat with the next reed, reed...). Adjacent pairs only, else an eight-reed instrument is twenty-eight beat columns wide.
export function columnGroups(count: number): ColumnGroup[] {
  const out: ColumnGroup[] = [];
  for (const reed of reedIndexes(count)) {
    if (reed > 0) {
      const low = reed - 1;
      out.push({
        kind: "beat",
        key: `beat-${low}`,
        pair: `${low + 1}-${reed + 1}`,
        low,
        high: reed,
      });
    }
    out.push({ kind: "reed", key: `reed-${reed}`, reed });
  }
  return out;
}

// reedGroups is the recorded table's layout: every reed first, then every beat - not columnGroups' order.
export function reedGroups(count: number): ReedGroup[] {
  return reedIndexes(count).map((reed) => ({
    kind: "reed",
    key: `reed-${reed}`,
    reed,
  }));
}

/** The adjacent pairs of an instrument of this many reeds. None at all below two. */
export function beatGroups(count: number): BeatGroup[] {
  const out: BeatGroup[] = [];
  for (const reed of reedIndexes(count)) {
    if (reed === 0) continue;
    const low = reed - 1;
    out.push({
      kind: "beat",
      key: `beat-${low}`,
      pair: `${low + 1}-${reed + 1}`,
      low,
      high: reed,
    });
  }
  return out;
}

// reedsUsable reports whether a row's per-reed numbers may be read as reeds; a merged note not recovered from the beat has none, its lobes are not a reed each. The one place this rule lives.
export function reedsUsable(row: TakeRow): boolean {
  return !row.reedsMerged || row.reedsFromBeat === true;
}

/** Whether the row's reeds were recovered from the beat rather than seen apart. */
export function reedsDerived(row: TakeRow): boolean {
  return row.reedsMerged && row.reedsFromBeat === true;
}

// recordKeys: light.lit and light.disabled are independent - lit means readings are landing, disabled means no session; never both.
export type RecordHint = "recording" | "warmup" | "noSession";

export interface RecordKeys {
  light: { lit: boolean; disabled: boolean; hint: RecordHint };
  undo: { disabled: boolean };
}

export function recordKeys(state: {
  session: boolean;
  armed: boolean;
  readings: number;
  busy: boolean;
}): RecordKeys {
  return {
    light: {
      lit: state.session && state.armed,
      disabled: !state.session,
      hint: !state.session ? "noSession" : state.armed ? "recording" : "warmup",
    },
    // Not gated on arm: undo discards what was captured, not what's being captured now; services/session.UndoTake refuses only ErrNoSession.
    undo: { disabled: !state.session || state.readings === 0 || state.busy },
  };
}

export function reedOf(row: TakeRow, reed: number): ReedError | undefined {
  return row.reeds.find((candidate) => candidate.reed === reed);
}

export function beatOf(row: TakeRow, pair: string): BeatError | undefined {
  return row.beats.find((candidate) => candidate.pair === pair);
}

// byNote sorts a pass into reading order (note, then register), matching the backend; it does not collapse a note played on two registers.
export function byNote(rows: TakeRow[]): TakeRow[] {
  return [...rows].sort(
    (a, b) =>
      a.note - b.note || (a.register ?? "").localeCompare(b.register ?? ""),
  );
}
