import { describe, expect, it } from 'vitest'
import { frozenByReadings, sessionNameFor } from '~/utils/accordion'

/*
 * What the new-session dialog works out for itself. The accordion's switches are not among it:
 * they are described on the Accordion tab, one at a time, once the session is open.
 */

describe('the offered session name', () => {
  it('is the accordion and the date', () => {
    expect(sessionNameFor('Bugari 508/ARS', '19 Aug 2026'))
      .toBe('Bugari 508/ARS - 19 Aug 2026')
  })

  it('is the date alone when the accordion has no name yet', () => {
    expect(sessionNameFor('', '19 Aug 2026')).toBe('19 Aug 2026')
    expect(sessionNameFor('   ', '19 Aug 2026')).toBe('19 Aug 2026')
  })
})

describe('what a reading freezes', () => {
  /*
   * Mirrors services/session. The reference pitch is refused outright once there
   * are takes (SetA4, ErrHasReadings), and so are the card's columns
   * (SetInstrument, columnsDiffer). The panel has to say so rather than offering
   * a control that silently fails.
   */
  it('leaves a session that has heard nothing alone', () => {
    expect(frozenByReadings(0)).toBe(false)
  })

  it('freezes the moment there is one reading', () => {
    expect(frozenByReadings(1)).toBe(true)
    expect(frozenByReadings(37)).toBe(true)
  })
})
