import { describe, expect, it } from 'vitest'
import type { Panel } from '~~bindings/smegg.me/smeggtuner/backend/core/dsp/models.js'
import type { BeatError, ReedError } from '~~bindings/smegg.me/smeggtuner/backend/core/target/models.js'
import { buildBoxes } from '~/utils/boxes'
import type { BoxInput } from '~/utils/boxes'
import { MISSING, toCents } from '~/utils/deviation'

// A reference readout as the engine sends it: "-" is the reference's own print for no value.
const panel = (value: number, target = MISSING, opts: Partial<Panel> = {}): Panel => ({
  value,
  target,
  signed: true,
  decimals: 1,
  dimmed: false,
  shown: value <= MISSING ? '-' : String(value),
  ...opts,
})
const missing = () => panel(MISSING)

const reedError = (reed: number, error: number, inTol: boolean): ReedError =>
  ({ reed, curr: error, goal: 0, error, currHz: 0, goalHz: 0, errorHz: 0, inTol })

const beatError = (low: number, error: number, inTol: boolean): BeatError =>
  ({ pair: `${low + 1}-${low + 2}`, low, high: low + 1, curr: error, goal: 0, error, currHz: 0, goalHz: 0, errorHz: 0, inTol })

const base: BoxInput = {
  panels: [],
  reedErrors: [],
  beatErrors: [],
  tolerance: 3,
  beatTolerance: 0.3,
  reedCount: 3,
  unit: 'reference',
  beatsInHz: true,
  relative: false,
}

// Three slots, each reed reading its value and every pair a beat.
const three = [panel(4.2), panel(2.47, MISSING, { signed: false, decimals: 2 }), panel(1.4), panel(2.51, MISSING, { signed: false, decimals: 2 }), panel(3.7)]

describe('buildBoxes', () => {
  it('lays a beat between every neighbouring pair of reeds', () => {
    const boxes = buildBoxes({ ...base, panels: three })
    expect(boxes.map(b => b.kind)).toEqual(['reed', 'beat', 'reed', 'beat', 'reed'])
  })

  // 1 to 8, and never a special case for three. A bass button sounds five.
  it('handles one reed, and eight', () => {
    expect(buildBoxes({ ...base, reedCount: 1, panels: [panel(0.2)] })).toHaveLength(1)
    expect(buildBoxes({ ...base, reedCount: 8 })).toHaveLength(15)
  })

  /*
   * THE ROW DOES NOT RESHAPE ITSELF. A three voice instrument gets three reed boxes even when
   * only one voice is speaking, and the slots that did not sound say so: the engine measured
   * them and found nothing, which is not the same as nothing sounding at all.
   */
  it('keeps one box per CONFIGURED reed, even when only one is sounding', () => {
    const boxes = buildBoxes({
      ...base,
      panels: [panel(-2.8), missing(), missing(), missing(), missing()],
      reedErrors: [reedError(0, -2.8, false)],
    })

    expect(boxes.map(b => b.kind)).toEqual(['reed', 'beat', 'reed', 'beat', 'reed'])
    expect(boxes[0]?.value).toBeCloseTo(-2.8)
    expect(boxes[0]?.out).toBe(true)

    for (const box of boxes.slice(1)) {
      expect(box.value).toBeNull()
      expect(box.out).toBe(false)
      expect(box.frac).toBeNull()
    }
    expect(boxes.filter(b => b.kind === 'reed').slice(1).every(b => b.blank === 'notHeard')).toBe(true)
    expect(boxes.filter(b => b.kind === 'beat').every(b => b.blank === 'idle')).toBe(true)
  })

  // inTol is the backend's verdict, full stop. A second threshold kept here would be a second
  // opinion, and the day the two disagreed the dial and the report would disagree.
  it('takes the verdict from the DTO and never from the number', () => {
    const boxes = buildBoxes({ ...base, reedCount: 1, panels: [panel(9)], reedErrors: [reedError(0, 9, true)] })
    expect(boxes[0]?.out).toBe(false)
  })

  // The bar's track is four tolerances wide, so one tolerance is always a quarter of it, for a
  // reed and for a beat alike: both verdicts are in reference units.
  it('puts the goal at the centre and one tolerance at a quarter of the track', () => {
    const centred = buildBoxes({ ...base, reedCount: 1, panels: [panel(0)], reedErrors: [reedError(0, 0, true)] })
    expect(centred[0]?.frac).toBeCloseTo(0.5)

    const atTolerance = buildBoxes({ ...base, reedCount: 1, panels: [panel(3)], reedErrors: [reedError(0, 3, false)] })
    expect(atTolerance[0]?.frac).toBeCloseTo(0.625)

    const beat = buildBoxes({
      ...base,
      reedCount: 2,
      panels: [panel(0), panel(1.9, MISSING, { signed: false, decimals: 2 }), panel(0)],
      beatErrors: [beatError(0, 0.3, false)],
    })
    expect(beat.find(b => b.kind === 'beat')?.frac).toBeCloseTo(0.625)
  })

  it('clamps a wild reading into the track instead of drawing off the end', () => {
    const boxes = buildBoxes({
      ...base,
      reedCount: 2,
      panels: [panel(40), panel(1), panel(-40)],
      reedErrors: [reedError(0, 500, false), reedError(1, -500, false)],
    })
    expect(boxes[0]?.frac).toBe(1)
    expect(boxes[2]?.frac).toBe(0)
  })

  // With nothing sounding the row keeps its SHAPE and holds not one number.
  it('keeps the row, and every number in it null, with nothing sounding', () => {
    const boxes = buildBoxes(base)
    expect(boxes.map(b => b.kind)).toEqual(['reed', 'beat', 'reed', 'beat', 'reed'])
    for (const box of boxes) {
      expect(box.value).toBeNull()
      expect(box.goal).toBeNull()
      expect(box.frac).toBeNull()
      expect(box.out).toBe(false)
      expect(box.blank).toBe('idle')
    }
  })

  it('follows the reed count when idle, and never assumes three', () => {
    const five = buildBoxes({ ...base, reedCount: 5 })
    expect(five.filter(b => b.kind === 'reed')).toHaveLength(5)
    expect(five.filter(b => b.kind === 'beat')).toHaveLength(4)
  })

  // A tolerance of zero would be a division by zero, and NaN painted as a bar width vanishes.
  it('survives a tolerance of zero', () => {
    const boxes = buildBoxes({ ...base, reedCount: 1, tolerance: 0, panels: [panel(4)], reedErrors: [reedError(0, 4, false)] })
    expect(boxes[0]?.frac).toBe(0.5)
  })
})

/*
 * THE BOXES SHOW THE REFERENCE SCREEN. The number is the engine's readout - the deviation from
 * the scale, or the distance from the goal when the technician drives to zero - and the goal
 * under it is the readout's own target.
 */
describe('the screen readout', () => {
  it('shows the deviation and its goal when driving to the goal', () => {
    const boxes = buildBoxes({ ...base, reedCount: 1, panels: [panel(12, 10)], reedErrors: [reedError(0, 2, true)] })
    expect(boxes[0]?.value).toBeCloseTo(12)
    expect(boxes[0]?.goal).toBeCloseTo(10)
  })

  it('shows the readout as the engine computed it when driving to zero', () => {
    const boxes = buildBoxes({ ...base, relative: true, reedCount: 1, panels: [panel(2, 10)] })
    expect(boxes[0]?.value).toBeCloseTo(2)
    expect(boxes[0]?.goal).toBeCloseTo(10)
  })

  // cents = 1200*log2(1 + 0.0006*d); a residual converts both ends, as core/target restates an error.
  it('restates deviations in cents when the setting says cents', () => {
    const scale = buildBoxes({ ...base, unit: 'cents', reedCount: 1, panels: [panel(10, 5)] })
    expect(scale[0]?.unit).toBe('cents')
    expect(scale[0]?.value).toBeCloseTo(toCents(10))
    expect(scale[0]?.goal).toBeCloseTo(toCents(5))

    const residual = buildBoxes({ ...base, unit: 'cents', relative: true, reedCount: 1, panels: [panel(2, 10)] })
    expect(residual[0]?.value).toBeCloseTo(toCents(12) - toCents(10))
  })

  // The beat box must not call reference units Hz, nor Hz units.
  it('labels a beat in hertz only when the engine reads beats in hertz', () => {
    const pair = [panel(0), panel(2.34, MISSING, { signed: false, decimals: 2 }), panel(0)]
    const hz = buildBoxes({ ...base, reedCount: 2, panels: pair })
    expect(hz[1]?.unit).toBe('hz')
    expect(hz[1]?.value).toBeCloseTo(2.34)
    expect(hz[1]?.decimals).toBe(2)
    expect(hz[1]?.signed).toBe(false)

    const units = buildBoxes({ ...base, reedCount: 2, beatsInHz: false, unit: 'cents', panels: [panel(0), panel(9), panel(0)] })
    expect(units[1]?.unit).toBe('cents')
    expect(units[1]?.value).toBeCloseTo(toCents(9))
    expect(units[0]?.unit).toBe('cents')
  })

  // The reference prints an unusable beat as "-", whatever number sits behind it.
  it('prints no beat where the reference prints none', () => {
    const boxes = buildBoxes({ ...base, reedCount: 2, panels: [panel(0), panel(-1, MISSING, { signed: false, shown: '-' }), panel(0)] })
    expect(boxes[1]?.value).toBeNull()
    expect(boxes[1]?.blank).toBe('idle')
  })

  // No note: the reference keeps the last numbers and dims them; a stale number carries no verdict.
  it('keeps a dimmed reading, dimmed and without a verdict', () => {
    const boxes = buildBoxes({
      ...base,
      reedCount: 1,
      panels: [panel(6, MISSING, { dimmed: true })],
      reedErrors: [reedError(0, 6, false)],
    })
    expect(boxes[0]?.value).toBeCloseTo(6)
    expect(boxes[0]?.dimmed).toBe(true)
    expect(boxes[0]?.out).toBe(false)
  })
})

/*
 * A REGISTER THAT SPANS OCTAVES NAMES ITS BOXES BY ITS RANKS. Slots are positional in card
 * order, so a silent 16' leaves the 16' box empty instead of shifting every reading one place.
 */
describe('a register spanning octaves', () => {
  const lm = { ...base, reedCount: 2, slots: [{ name: 'L', octave: -12 }, { name: 'M1', octave: 0 }] }

  it('maps each slot onto its own rank and the beat between them', () => {
    const boxes = buildBoxes({ ...lm, panels: [panel(-2.5), panel(2.3, MISSING, { signed: false, decimals: 2 }), panel(1.4)] })
    expect(boxes.map(b => b.kind)).toEqual(['reed', 'beat', 'reed'])
    expect(boxes[0]?.rank).toBe('L')
    expect(boxes[0]?.value).toBeCloseTo(-2.5)
    expect(boxes[1]?.rank).toBe('L-M1')
    expect(boxes[1]?.value).toBeCloseTo(2.3)
    expect(boxes[2]?.rank).toBe('M1')
    expect(boxes[2]?.value).toBeCloseTo(1.4)
  })

  it('leaves the 16\' box in place when only the 8\' sounds', () => {
    const boxes = buildBoxes({ ...lm, panels: [missing(), missing(), panel(1.4)] })
    expect(boxes[0]?.rank).toBe('L')
    expect(boxes[0]?.value).toBeNull()
    expect(boxes[0]?.blank).toBe('notHeard')
    expect(boxes[1]?.value).toBeNull()
    expect(boxes[2]?.value).toBeCloseTo(1.4)
  })

  it('keeps the whole row idle-empty with nothing sounding', () => {
    const boxes = buildBoxes(lm)
    expect(boxes.map(b => b.kind)).toEqual(['reed', 'beat', 'reed'])
    expect(boxes.every(b => b.value === null)).toBe(true)
    expect(boxes[0]?.blank).toBe('idle')
  })

  // The verdict belongs to the pair by its slots, not its place in the list of sounding pairs.
  it('picks the beat verdict by its pair, not its position in the list', () => {
    const boxes = buildBoxes({
      ...base,
      reedCount: 4,
      slots: [{ name: 'L', octave: -12 }, { name: 'M1', octave: 0 }, { name: 'M2', octave: 0 }, { name: 'M3', octave: 0 }],
      panels: [panel(0), missing(), missing(), missing(), panel(0), panel(2.8), panel(1)],
      beatErrors: [beatError(2, 0.9, false)],
    })
    const beats = boxes.filter(b => b.kind === 'beat')
    expect(beats.map(b => b.out)).toEqual([false, false, true])
  })
})

// The bass side is the same row taller: five feet, each an octave up from the last.
describe('a bass ladder', () => {
  it('labels each box by its foot and keeps a weak lowest voice its place', () => {
    const boxes = buildBoxes({
      ...base,
      reedCount: 5,
      slots: [
        { name: '32\'', octave: 0 }, { name: '16\'', octave: 12 }, { name: '8\'', octave: 24 },
        { name: '4\'', octave: 36 }, { name: '2\'', octave: 48 },
      ],
      panels: [missing(), missing(), panel(1.2), panel(0.9), panel(0.4), panel(0.5), panel(-0.8), panel(0.2), panel(0.1)],
    })

    expect(boxes.filter(b => b.kind === 'reed').map(b => b.rank)).toEqual(['32\'', '16\'', '8\'', '4\'', '2\''])
    expect(boxes[0]?.blank).toBe('notHeard')
    // The beat between the silent 32' and the 16' has no pair to read.
    expect(boxes[1]?.value).toBeNull()
    expect(boxes[2]?.value).toBeCloseTo(1.2)
    expect(boxes[3]?.value).toBeCloseTo(0.9)
  })
})

/*
 * THE ROW DOES NOT CHANGE SHAPE WHILE IT IS BEING READ. How many boxes there are is DECLARED -
 * the reeds the pulled register sounds, or the setting - and never measured off the reading.
 */
describe('the shape of the row', () => {
  it('holds its shape whatever readouts the engine sends this frame', () => {
    const shapes = [three, three.slice(0, 3), []]
      .map(panels => buildBoxes({ ...base, panels }).map(b => b.kind))
    for (const shape of shapes) {
      expect(shape).toEqual(['reed', 'beat', 'reed', 'beat', 'reed'])
    }
  })

  // With NOTHING declared - no session, no setting yet - the engine's readout row is all there is.
  it('falls back to the readouts only when nothing has been declared', () => {
    const boxes = buildBoxes({ ...base, reedCount: 0, panels: [panel(4.2), panel(1), panel(1.4)] })
    expect(boxes.filter(b => b.kind === 'reed')).toHaveLength(2)
  })
})

/*
 * THE PANEL WITHOUT A SESSION. Reed boxes 1 and 3 have no pair; chord tones never beat; and an
 * octave pair of the reference Bass mode reads its beat only in Hz, so in units its box says so.
 */
describe('absent pair readouts', () => {
  const absent = panel(MISSING, MISSING, { absent: true, dimmed: true })

  it('drops the pair of reed boxes 1 and 3 and numbers the reeds by their boxes', () => {
    const boxes = buildBoxes({ ...base, reedCount: 2, ids: [1, 3], octaves: [0, 0], panels: [panel(1), absent, panel(2)] })
    expect(boxes.map(b => b.kind)).toEqual(['reed', 'reed'])
    expect(boxes.map(b => b.number)).toEqual([1, 3])
  })

  it('says an octave pair reads only in Hz while beats are in units', () => {
    const boxes = buildBoxes({ ...base, beatsInHz: false, reedCount: 2, ids: [1, 2], octaves: [0, 12], panels: [panel(1), absent, panel(2)] })
    expect(boxes[1]).toMatchObject({ kind: 'beat', value: null, blank: 'hzOnly', pair: [1, 2] })
  })

  // Two octaves apart (0/24) there is no neighbour pair at all, in either unit.
  it('drops a pair two octaves apart', () => {
    const boxes = buildBoxes({ ...base, beatsInHz: false, reedCount: 2, ids: [1, 3], octaves: [0, 24], panels: [panel(1), absent, panel(2)] })
    expect(boxes.map(b => b.kind)).toEqual(['reed', 'reed'])
  })

  it('names chord tones and carries each tone\'s note', () => {
    const tone = (v: number, text: string) => panel(v, MISSING, { text })
    const boxes = buildBoxes({ ...base, tones: true, ids: [1, 2, 3], octaves: [0, 0, 0], panels: [tone(1, 'C4'), absent, tone(2, 'E4'), absent, tone(3, 'G4')] })
    expect(boxes.map(b => b.kind)).toEqual(['reed', 'reed', 'reed'])
    expect(boxes.every(b => b.tone)).toBe(true)
    expect(boxes.map(b => b.text)).toEqual(['C4', 'E4', 'G4'])
  })
})
