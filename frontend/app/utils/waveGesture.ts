// How a press on the file waveform reads. The timeline lane and the playhead itself scrub; a press
// anywhere else is a click (move the playhead there) until it travels, then it drags a range.

/** Pixels either side of the playhead that grab it. */
export const NEEDLE_GRAB_PX = 6

/** Pixels a press may travel and still count as a click. */
export const DRAG_SLOP_PX = 3

export function pressKind(x: number, needleX: number | null, onLane: boolean): 'scrub' | 'press' {
  if (onLane) return 'scrub'
  if (needleX !== null && Math.abs(x - needleX) <= NEEDLE_GRAB_PX) return 'scrub'
  return 'press'
}

/** A seek outside the selected range: Go would clamp it into the range, so the range has to go first. */
export function outsideRange(at: number, from: number, to: number): boolean {
  return at < from || at > to
}
