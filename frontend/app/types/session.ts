// app/types/session.ts
import { Bank } from '~~bindings/smegg.me/smeggtuner/backend/core/session/models.js'

export { Bank } from '~~bindings/smegg.me/smeggtuner/backend/core/session/models.js'
export type { BassRegister, Instrument, Register, Template as InstrumentTemplate } from '~~bindings/smegg.me/smeggtuner/backend/core/session/models.js'

export const BANKS: Bank[] = [Bank.BankL, Bank.BankM1, Bank.BankM2, Bank.BankM3, Bank.BankM4, Bank.BankH]

// Mirrors core/session. The widest treble panel in production is thirteen switches and the widest
// bass machine seven; these sit above both, and the backend refuses anything past them.
export const MAX_REGISTERS = 16
export const MAX_BASS_REGISTERS = 8
export const MAX_SWITCH_NAME = 24
export const MAX_INSTRUMENT_NAME = 60

// A new session as the dialog hands it over: the whole instrument (ranks, switches, reference pitch),
// not just the shown fields.
export interface SessionDraft {
  name: string
  notes: string
  instrument: Instrument
  /** remembered accordion id, so the session can find its photograph */
  instrumentId: string
}
