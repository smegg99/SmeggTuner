import { BANKS } from '~/types/session'
import type { Bank } from '~/types/session'

// Feet notation over banks: L is 16', M1..M4 the 8' musette ranks, H is 4'. What's engraved on the instrument.
const FEET: Record<Bank, string> = {
  L: '16\'',
  M1: '8\'',
  M2: '8\'',
  M3: '8\'',
  M4: '8\'',
  H: '4\'',
}

// octaveOf is where a bank sounds, in semitones from the played key; mirrors core/session Bank.Octave.
export function octaveOf(bank: Bank): number {
  if (bank === 'L') return -12
  if (bank === 'H') return 12
  return 0
}

// bassFeetOf is the bass machine's ranks by foot for a voice count, largest (lowest) first and
// always ending at 2': 5 voices -> 32,16,8,4,2. Mirrors core/session.BassFeet.
export function bassFeetOf(count: number): number[] {
  if (count < 2 || count > 6) return []
  return Array.from({ length: count }, (_, i) => 1 << (count - i))
}

// bassSlotsOf lays sounding bass ranks out as boxes slots: the largest foot is the base, each
// halving climbs an octave. Mirrors core/session.OctavesOfFeet.
export function bassSlotsOf(feet: readonly number[]): { name: string, octave: number }[] {
  const sorted = [...feet].sort((a, b) => b - a)
  const top = sorted[0]
  if (!top) return []
  return sorted.map(f => ({ name: `${f}'`, octave: 12 * Math.round(Math.log2(top / f)) }))
}

// feetOf renders a register's banks in feet, e.g. [L,M1,M2,M3] -> "16'+8'+8'+8'", in card order low to high.
export function feetOf(banks: readonly Bank[]): string {
  const ordered = BANKS.filter(b => banks.includes(b))
  if (ordered.length === 0) return ''
  return ordered.map(b => FEET[b]).join('+')
}

// RegisterDots is the engraved register symbol as booleans and a count, so a component can draw it without knowing banks.
export interface RegisterDots {
  /** 4' at the top */
  high: boolean
  /** how many 8' dots in the middle: 0 to 4 */
  middle: number
  /** 16' bass at the bottom */
  low: boolean
}

export function symbolOf(banks: readonly Bank[]): RegisterDots {
  return {
    high: banks.includes('H'),
    middle: (['M1', 'M2', 'M3', 'M4'] as Bank[]).filter(m => banks.includes(m)).length,
    low: banks.includes('L'),
  }
}

// banksOfSymbol maps clicked dots back to banks; the middle count fills M1,M2,M3,M4 in order so picker and card agree on which rank is which.
export function banksOfSymbol(sym: RegisterDots): Bank[] {
  const out: Bank[] = []
  if (sym.low) out.push('L')
  for (let i = 0; i < Math.min(4, Math.max(0, sym.middle)); i++) {
    out.push(`M${i + 1}` as Bank)
  }
  if (sym.high) out.push('H')
  return out
}

// canonicalName is a register's own notation, e.g. "LMMM" (bare M for a run counted from M1); explicit M-numbers only where a run doesn't start at M1, e.g. "M2", "M1M3".
export function canonicalName(banks: readonly Bank[]): string {
  const ordered = BANKS.filter(b => banks.includes(b))
  const ms = ordered.filter(b => b !== 'L' && b !== 'H')

  // A gapless run from M1 collapses to bare M; a gap (M1,M3) names both ranks.
  const counted = ms.every((b, i) => b === `M${i + 1}`)

  return ordered.map(b => (b !== 'L' && b !== 'H' && counted ? 'M' : b)).join('')
}
