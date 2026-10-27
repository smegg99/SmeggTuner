import { describe, expect, it } from 'vitest'
import { recordKeys } from '~/utils/record'

/*
 * NO SESSION, NO RECORDING - and no recording until it is asked for.
 *
 * A session opens in WARM-UP: the key is live, because there is a session to record into,
 * and unlit, because nothing is being saved yet. That is the state a technician plays
 * himself in on. Pressing the key arms it.
 *
 * This key has been wrong three times. Disabled unconditionally, so its hint told a
 * technician to open the session he had open. Then disabled AND lit at once, an accent
 * block at a third opacity. Then lit the moment a session opened, filing the notes played
 * to warm up. Each time the rule lived in a template expression where nothing could test
 * it. It lives here.
 */
const off = { session: false, armed: false, readings: 0, busy: false }

describe('the record light', () => {
  it('is dead and dark with no session, because there is nowhere to record', () => {
    const { light } = recordKeys(off)

    expect(light.lit).toBe(false)
    expect(light.disabled).toBe(true)
    expect(light.hint).toBe('noSession')
  })

  /*
   * WARM-UP: the state the whole feature exists for. Live, so it can be pressed; unlit,
   * because nothing is landing. A technician plays a few notes here and files none of them.
   */
  it('is live but unlit in a session nobody has pressed record in', () => {
    const { light } = recordKeys({ ...off, session: true })

    expect(light.lit).toBe(false)
    expect(light.disabled).toBe(false)
    expect(light.hint).toBe('warmup')
  })

  it('is live and lit once armed', () => {
    const { light } = recordKeys({ ...off, session: true, armed: true })

    expect(light.lit).toBe(true)
    expect(light.disabled).toBe(false)
    expect(light.hint).toBe('recording')
  })

  /*
   * WARM-UP IS NOT THE SAME STATE AS NO SESSION, and must not borrow its words. One says
   * "you have nowhere to put a reading", the other says "press this when you are ready".
   */
  it('does not tell a technician in a session to open a session', () => {
    expect(recordKeys({ ...off, session: true }).light.hint).not.toBe('noSession')
  })

  /*
   * LIT AND DISABLED IS THE STATE THAT MUST NOT EXIST. One says "this is happening", the
   * other says "this is not available", and a key wearing both says neither.
   */
  it('is never lit and disabled at once, in any state it can be in', () => {
    for (const session of [false, true]) {
      for (const armed of [false, true]) {
        for (const readings of [0, 3]) {
          const { light } = recordKeys({ session, armed, readings, busy: false })
          const state = JSON.stringify({ session, armed, readings })
          expect(light.lit && light.disabled, state).toBe(false)
        }
      }
    }
  })
})

describe('the undo key', () => {
  it('is dead with no session', () => {
    expect(recordKeys(off).undo.disabled).toBe(true)
  })

  // Stale readings from a session that has been closed must not leave it live: the count is
  // not the question, the session is.
  it('is dead with no session even if a reading count is still on screen', () => {
    expect(recordKeys({ ...off, readings: 5 }).undo.disabled).toBe(true)
  })

  it('is dead in a session with nothing recorded yet', () => {
    expect(recordKeys({ ...off, session: true }).undo.disabled).toBe(true)
  })

  /*
   * IT DOES NOT WAIT FOR THE ARM. A reading recorded, then disarmed, is still a reading the
   * technician may want back: undo throws away what was captured, and what was captured is
   * not a function of what is being captured now.
   */
  it('is live on a session that holds readings, armed or not', () => {
    expect(recordKeys({ ...off, session: true, armed: true, readings: 1 }).undo.disabled).toBe(false)
    expect(recordKeys({ ...off, session: true, readings: 1 }).undo.disabled).toBe(false)
  })

  it('is dead while a call is in flight', () => {
    expect(recordKeys({ ...off, session: true, readings: 2, busy: true }).undo.disabled).toBe(true)
  })
})
