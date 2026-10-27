import { describe, expect, it } from 'vitest'
import { banksOfSymbol, canonicalName, feetOf, symbolOf } from '~/utils/feet'
import { banksOfRegister } from '~/utils/banks'
import { BANKS } from '~/types/session'
import type { Bank } from '~/types/session'

/*
 * FEET AND SYMBOLS ARE A VIEW OVER THE BANKS, and they must agree with the model exactly.
 *
 * A technician thinks in feet (16+8+4) and reads the register symbol engraved on the switch; the
 * app stores banks. If the feet layer and the bank layer ever disagreed, the picker would show
 * one register and the card would print another, and a reading would land under the wrong rank.
 */
describe('feet', () => {
  it('says the register in feet, low to high', () => {
    expect(feetOf(['L', 'M1', 'M2', 'M3'])).toBe('16\'+8\'+8\'+8\'')
    expect(feetOf(['M1', 'M2', 'M3'])).toBe('8\'+8\'+8\'')
    expect(feetOf(['L', 'M1', 'H'])).toBe('16\'+8\'+4\'')
    expect(feetOf(['H'])).toBe('4\'')
    expect(feetOf([])).toBe('')
  })

  // Order in, order out: the feet always read low to high whatever order the banks arrive in.
  it('is in card order regardless of input order', () => {
    expect(feetOf(['H', 'L', 'M1'])).toBe('16\'+8\'+4\'')
  })
})

describe('the register symbol', () => {
  it('lights the dots for the ranks a switch sounds', () => {
    expect(symbolOf(['L', 'M1', 'M2', 'M3'])).toEqual({ low: true, middle: 3, high: false })
    expect(symbolOf(['M1', 'M2'])).toEqual({ low: false, middle: 2, high: false })
    expect(symbolOf(['H'])).toEqual({ low: false, middle: 0, high: true })
  })

  /*
   * THE DOTS ROUND-TRIP TO THE BANKS, and the middle fills IN ORDER. Two eight-foot dots are
   * always M1 and M2, never M1 and M3 - that is what "two dots" means, and it is what keeps the
   * picker and the card agreeing which rank is which.
   */
  it('turns the clicked dots back into banks, filling the middle in order', () => {
    expect(banksOfSymbol({ low: true, middle: 3, high: true })).toEqual(['L', 'M1', 'M2', 'M3', 'H'])
    expect(banksOfSymbol({ low: false, middle: 2, high: false })).toEqual(['M1', 'M2'])
    expect(banksOfSymbol({ low: true, middle: 0, high: false })).toEqual(['L'])
    expect(banksOfSymbol({ low: false, middle: 0, high: false })).toEqual([])
  })
})

/*
 * THE PICKER AND THE TYPED NOTATION PRODUCE THE SAME REGISTER.
 *
 * An accordion built by clicking dots and one typed by hand as "LMMM" must come out identical,
 * or two technicians describing the same instrument two ways would get two different cards. This
 * pins the symbol builder against banksOfRegister, which is itself pinned against the Go parser.
 */
describe('a dot-built register equals a typed one', () => {
  const cases: Bank[][] = [
    ['L', 'M1', 'M2', 'M3'],
    ['M1', 'M2', 'M3'],
    ['L', 'M1'],
    ['L', 'M1', 'M2', 'H'],
    ['H'],
  ]

  it('has a canonical name the notation parser reads back to the same banks', () => {
    for (const banks of cases) {
      const name = canonicalName(banks)
      expect(banksOfRegister(name), name).toEqual(banks)
    }
  })

  /*
   * IT WRITES THE NOTATION, not the internal identifiers. A technician writes the three
   * eight-foot ranks as bare Ms - "LMM", never "LM1M2" - and the number is only used to name a
   * rank that does not start the run. These are the same cases backend/core/session/bank_test.go pins on
   * the Go side, read the other way round; if the two ever disagreed, an instrument built by
   * clicking dots and one typed by hand would be two different instruments.
   */
  it('writes the technician\'s notation, collapsing the M-run', () => {
    const cases: [Bank[], string][] = [
      [['L', 'M1', 'M2', 'M3'], 'LMMM'],
      [['M1', 'M2', 'M3'], 'MMM'],
      [['L', 'M1', 'M2'], 'LMM'],
      [['L', 'M1'], 'LM'],
      [['M1'], 'M'],
      [['L', 'M1', 'H'], 'LMH'],
      [['L', 'M1', 'M2', 'H'], 'LMMH'],
      [['H'], 'H'],
      [['L'], 'L'],

      // A rank that does not start the run keeps its number: there is no other way to say
      // "the flat one, on its own".
      [['M2'], 'M2'],
      [['L', 'M2'], 'LM2'],
      [['M1', 'M3'], 'M1M3'],
      [['M2', 'M3'], 'M2M3'],
      [['M3'], 'M3'],
    ]

    for (const [banks, want] of cases) {
      expect(canonicalName(banks), want).toBe(want)
    }
  })

  // It is in card order whatever order the dots arrive in, the same as feetOf.
  it('names in card order regardless of input order', () => {
    expect(canonicalName(['H', 'M1', 'L'])).toBe('LMH')
  })

  /*
   * THE PROPERTY, over every switch an accordion can have: what the writer says, the reader
   * reads back unchanged. This is the invariant the two halves of the app share - break it and a
   * reading is filed under a rank nobody pulled.
   */
  it('round-trips every non-empty set of banks', () => {
    for (let mask = 1; mask < 1 << BANKS.length; mask++) {
      const banks = BANKS.filter((_, i) => mask & (1 << i))
      const name = canonicalName(banks)
      expect(banksOfRegister(name), `${banks.join(',')} -> ${name}`).toEqual(banks)
    }
  })
})
