// The panel without a session: mode and reed boxes as services/tuner keeps them, laid out for the screen.

export type PanelMode = 'tones' | 'octaves' | 'chords'

export const PANEL_MODES: readonly PanelMode[] = ['tones', 'octaves', 'chords']

/** One of the reference's three reed boxes. */
export interface PanelBox {
  /** 1..3 */
  box: number
  on: boolean
  /** octaves mode measures its base always */
  fixed: boolean
  /** the only box on; services/tuner refuses to clear it (ErrLastBox) */
  last: boolean
}

/** Octaves mode: box 2 is the octave above the base, box 3 two octaves above. */
export const octaveOfBox = (box: number) => (box - 1) * 12

/** panelBoxes lists the boxes a mode toggles; chords has none. */
export function panelBoxes(mode: PanelMode, reedBoxes: readonly number[], octaveOffsets: readonly number[]): PanelBox[] {
  if (mode === 'chords') return []
  if (mode === 'octaves') {
    return [1, 2, 3].map((box) => {
      if (box === 1) return { box, on: true, fixed: true, last: false }
      const on = octaveOffsets.includes(octaveOfBox(box))
      return { box, on, fixed: false, last: on && octaveOffsets.length === 1 }
    })
  }
  return [1, 2, 3].map((box) => {
    const on = reedBoxes.includes(box)
    return { box, on, fixed: false, last: on && reedBoxes.length === 1 }
  })
}

/** The measured slots' reed boxes, in Measurement.Reeds order, as dsp.ReedMeasure.Reed numbers them. */
export function panelSlotIds(mode: PanelMode, reedBoxes: readonly number[], octaveOffsets: readonly number[]): number[] {
  if (mode === 'chords') return [1, 2, 3]
  if (mode === 'octaves') return [1, ...[...octaveOffsets].sort((a, b) => a - b).map(o => o / 12 + 1)]
  return [...reedBoxes].sort((a, b) => a - b)
}

/** Each measured slot's semitones above the key. */
export function panelSlotOctaves(mode: PanelMode, reedBoxes: readonly number[], octaveOffsets: readonly number[]): number[] {
  if (mode === 'octaves') return [0, ...[...octaveOffsets].sort((a, b) => a - b)]
  return panelSlotIds(mode, reedBoxes, octaveOffsets).map(() => 0)
}
