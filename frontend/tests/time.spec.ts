import { describe, expect, it } from 'vitest'
import { formatTime, timeTicks } from '~/utils/time'

describe('the timeline ruler', () => {
  /*
   * A ruler is only a ruler if its marks land on numbers a person would say out loud.
   * Dividing the window into ten gives marks at 0.37s, 0.74s, 1.11s: arithmetically
   * even, unreadable, and the reason this is a table of steps and not a division.
   */
  it('counts in steps a clock would count in, at every zoom', () => {
    for (const [from, to] of [[0, 90], [0, 10], [3, 3.5], [1.2, 1.23]] as const) {
      const majors = timeTicks(from, to, 1200).filter(t => t.major)
      expect(majors.length).toBeGreaterThan(1)

      // whatever the step turns out to be, the marks are evenly spaced on it and it
      // divides them exactly
      const step = majors[1]!.at - majors[0]!.at
      for (const m of majors) {
        expect(Math.abs(Math.round(m.at / step) * step - m.at)).toBeLessThan(step / 100)
      }
    }
  })

  it('never crowds its labels closer than they can be read', () => {
    const width = 800
    const [from, to] = [0, 37.5]

    const majors = timeTicks(from, to, width).filter(t => t.major)
    const px = (t: number) => ((t - from) / (to - from)) * width

    for (let i = 1; i < majors.length; i++) {
      expect(px(majors[i]!.at) - px(majors[i - 1]!.at)).toBeGreaterThanOrEqual(64)
    }
  })

  it('puts unlabelled minors between the labelled majors, which is what makes it a ruler', () => {
    const ticks = timeTicks(0, 10, 1000)

    expect(ticks.filter(t => !t.major).length).toBeGreaterThan(ticks.filter(t => t.major).length)
    for (const t of ticks) {
      expect(t.label === '').toBe(!t.major)
    }
  })

  it('stays inside the window it was asked for', () => {
    for (const t of timeTicks(2, 4, 600)) {
      expect(t.at).toBeGreaterThanOrEqual(2)
      expect(t.at).toBeLessThanOrEqual(4)
    }
  })

  it('draws nothing for a window with no width', () => {
    expect(timeTicks(3, 3, 600)).toEqual([])
    expect(timeTicks(4, 2, 600)).toEqual([])
    expect(timeTicks(0, 10, 0)).toEqual([])
  })
})

describe('the clock', () => {
  // A ninety second take is read in minutes and seconds; a reed's attack is over
  // inside a tenth of one, and "0:00" four times running is not a ruler.
  it('reads in minutes past a second, and in decimals below one', () => {
    expect(formatTime(75, 5)).toBe('1:15')
    expect(formatTime(42, 1)).toBe('42s')
    expect(formatTime(1.25, 0.05)).toBe('1.25s')
    expect(formatTime(0.004, 0.001)).toBe('0.004s')
  })

  // enough decimals to tell one tick from the next, and not one more
  it('shows no more precision than the step it is counting in', () => {
    expect(formatTime(1.5, 0.5)).toBe('1.5s')
    expect(formatTime(1.23, 0.01)).toBe('1.23s')
  })
})
