import { describe, expect, it } from 'vitest'
import { banksOfRegister, banksOfRegisters } from '~/utils/banks'
import type { Bank } from '~/types/session'

/*
 * THIS MUST AGREE WITH core/session.ParseBanks, EXACTLY.
 *
 * The backend decides which column a reading is actually printed under. This only decides
 * what the EDITOR says it will do. If the two disagreed, a technician would type LMMM, be
 * shown four columns, and be handed a card with three - and the reading that fell off the end
 * would be the one they went to file.
 *
 * Every case below is one that backend/core/session/bank_test.go also asserts. If you change one,
 * change both.
 */
describe('reading a register name', () => {
  it('reads the notation that is written on the switch', () => {
    const cases: Record<string, Bank[]> = {
      'LMMM': ['L', 'M1', 'M2', 'M3'],
      'MMM': ['M1', 'M2', 'M3'],
      'LM': ['L', 'M1'],
      'M': ['M1'],
      'LMH': ['L', 'M1', 'H'],

      // Four eight foot ranks is a real switch: the widest musette there is.
      'LMMMM': ['L', 'M1', 'M2', 'M3', 'M4'],

      // Written out, which is the only way to say "the flat one, on its own".
      'M2': ['M2'],
      'M4': ['M4'],
      'LM2': ['L', 'M2'],
      'M1M3': ['M1', 'M3'],

      // Not shouted, and a technician may leave a space around it.
      ' lmmm ': ['L', 'M1', 'M2', 'M3'],
    }

    for (const [name, want] of Object.entries(cases)) {
      expect(banksOfRegister(name), name).toEqual(want)
    }
  })

  /*
   * A NAME THAT IS NOT THE NOTATION IS NOT GUESSED AT.
   *
   * "Musette" is a fine name for a switch and a technician is entitled to use it, but it does
   * not say which ranks sound. It comes back null, the editor says so out loud, and the card
   * numbers those reeds rather than naming them. It must never come back with a plausible
   * answer: a reading filed under the wrong rank looks exactly like one filed under the right
   * one.
   */
  it('refuses a name that is not notation, rather than guessing', () => {
    // LMMMMM is five eight foot ranks and M5 a fifth of them: there is no such switch, and a
    // fourth (LMMMM, M4) is as wide as it goes.
    for (const name of ['', '   ', 'Musette', 'Bandoneon', 'Master', 'LMMMMM', 'M5', 'L2', 'M1M1']) {
      expect(banksOfRegister(name), name).toBeNull()
    }
  })

  // The same rank twice would put two different reeds in one column of the card, and one of
  // them would simply disappear.
  it('refuses one rank twice', () => {
    expect(banksOfRegister('M1M1')).toBeNull()
    expect(banksOfRegister('LL')).toBeNull()
  })
})

describe('an instrument\'s own ranks', () => {
  // The union of what its switches can reach, in CARD ORDER - low to high, not the order the
  // registers happened to be typed in.
  it('is what its switches add up to, in card order', () => {
    const registers = [
      { name: 'M2', banks: banksOfRegister('M2')! },
      { name: 'LMMM', banks: banksOfRegister('LMMM')! },
      { name: 'MMM', banks: banksOfRegister('MMM')! },
    ]
    expect(banksOfRegisters(registers)).toEqual(['L', 'M1', 'M2', 'M3'])
  })

  // An instrument whose switches nobody can read has no columns, and the card numbers its
  // reeds. That is honest, and better than inventing columns for it.
  it('is empty when no switch says what it sounds', () => {
    expect(banksOfRegisters([{ name: 'Bandoneon', banks: [] }])).toEqual([])
    expect(banksOfRegisters([])).toEqual([])
  })
})
