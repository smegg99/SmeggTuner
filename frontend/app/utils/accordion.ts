import { canonicalName } from '~/utils/feet'
import { BANKS, MAX_BASS_REGISTERS, MAX_REGISTERS } from '~/types/session'
import type { Bank, BassRegister, Register } from '~/types/session'

// The name a session is offered. The date alone still identifies it, so an
// accordion nobody has named yet never produces a leading separator.
export function sessionNameFor(accordion: string, date: string): string {
  const name = accordion.trim()
  return name ? `${name} - ${date}` : date
}

// Mirrors services/session maxNotes: the note printed under the card's table. The field
// caps what can be typed, the service refuses what is too long, and tests/reportNotes.spec.ts
// keeps the two from drifting apart.
export const MAX_NOTES = 500

// Mirrors services/session. A reading is a deviation from the reference pitch it was
// measured against, filed under a card column; once one exists the backend refuses
// to move either (SetA4 and SetInstrument, both ErrHasReadings). The panel disables
// those controls rather than offering ones that silently fail.
export function frozenByReadings(readings: number): boolean {
  return readings > 0
}

// A switch IS what it sounds, so the ranks it sounds are its identity and the order they were
// picked in is not. Mirrors core/session bankKey.
function ranksOf(banks: readonly Bank[]): string {
  return BANKS.filter(b => banks.includes(b)).join(' ')
}

// Mirrors core/session feetKey: largest foot first, so clicking 8 then 16 is the same switch as
// clicking 16 then 8.
function feetOfSwitch(feet: readonly number[]): string {
  return [...feet].sort((a, b) => b - a).join('+')
}

// canAddRegister answers the Add key. Every refusal here is one core/session enforces too, so a
// key that stays enabled would put up a click that silently does nothing.
export function canAddRegister(have: readonly Register[], banks: readonly Bank[]): boolean {
  if (banks.length === 0 || have.length >= MAX_REGISTERS) return false

  const wanted = ranksOf(banks)
  const name = canonicalName(banks)
  return !have.some(r => ranksOf(r.banks) === wanted || r.name === name)
}

// canAddBass is the same answer for the bass machine, where the name is typed rather than derived.
export function canAddBass(
  have: readonly BassRegister[],
  feet: readonly number[],
  name: string,
): boolean {
  const wanted = name.trim()
  if (feet.length === 0 || wanted === '' || have.length >= MAX_BASS_REGISTERS) return false

  const sounds = feetOfSwitch(feet)
  return !have.some(r => feetOfSwitch(r.feet) === sounds || r.name === wanted)
}
