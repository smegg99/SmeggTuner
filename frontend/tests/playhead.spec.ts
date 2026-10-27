import { beforeEach, describe, expect, it, vi } from 'vitest'
import { usePlayhead } from '~/composables/usePlayhead'

/*
 * The needle broke three times. These are those three failures, written down.
 */
describe('the playhead', () => {
  let clock = 0

  beforeEach(() => {
    clock = 0
    vi.spyOn(performance, 'now').mockImplementation(() => clock)
  })

  const tick = (ms: number) => {
    clock += ms
  }

  /*
   * IT NEVER GOES BACKWARDS. Not once, not by a millisecond.
   *
   * Coasting overshoots, so the next landing is usually a hair BEHIND where the needle
   * already is. Anything that moves the needle TO the landing therefore drags it backwards
   * - snapping does it, and easing the position does it too, just by less. Thirty times a
   * second, that is a needle twitching left and right, and it is what shipped.
   *
   * The position is never corrected. The RATE is: a landing behind us means the clock is
   * running fast, so it is slowed, and it drifts back onto the truth while still moving
   * forward the whole way.
   */
  it('never moves backwards, however far behind the truth lands', () => {
    const { land, at } = usePlayhead()

    land(0, true, 10, true)

    let last = 0
    for (let i = 0; i < 200; i++) {
      tick(33)

      // Go lands a value that is always a little BEHIND the coast: the worst case, and the
      // one that actually happens, because the sound card reports in 30 ms steps.
      const truth = i * 0.033 - 0.02
      land(Math.max(0, truth), true, 10)

      const drawn = at(true, 10)
      expect(drawn).toBeGreaterThanOrEqual(last)
      last = drawn
    }
  })

  /*
   * IT DOES NOT SAIL AWAY WHEN THE ENGINE STOPS. Stopping does not pause anything - it
   * kills the event stream - so a needle that coasted whenever it was "not paused" went on
   * coasting from a dead anchor and walked off down the recording on its own.
   */
  it('freezes where the sound stopped when nothing is advancing', () => {
    const { land, at } = usePlayhead()

    land(4, true, 10, true)
    tick(8000)

    expect(at(false, 10)).toBe(4)
  })

  it('coasts at about one second per second while the audio advances', () => {
    const { land, at } = usePlayhead()

    land(4, true, 10, true)
    tick(100)

    expect(at(true, 10)).toBeCloseTo(4.1, 3)
  })

  // A stalled backend must not look exactly like playback.
  it('gives up and freezes rather than running away on a guess', () => {
    const { land, at } = usePlayhead()

    land(4, true, 10, true)
    tick(60_000)

    expect(at(true, 10)).toBeLessThanOrEqual(4.25)
  })

  it('never coasts past the end of what is being played', () => {
    const { land, at } = usePlayhead()

    land(2.95, true, 3, true)
    tick(200)

    expect(at(true, 3)).toBe(3)
  })

  // A seek is not drift. Past a threshold the difference is the truth arriving, and the
  // needle simply goes there - backwards, if that is where it was sent.
  it('goes straight to a seek', () => {
    const { land, at, heard } = usePlayhead()

    land(4, true, 10, true)
    tick(200)

    land(1, true, 10, true)

    expect(at(true, 10)).toBeCloseTo(1, 3)
    expect(heard.value).toBe(1)
  })

  it('snaps rather than easing when the truth is far away, as a loop coming round is', () => {
    const { land, at } = usePlayhead()

    land(1.9, true, 2, true)
    tick(33)

    land(1.0, true, 2) // the loop wrapped: not a drift, a jump
    expect(at(true, 2)).toBeCloseTo(1.0, 2)
  })
})
