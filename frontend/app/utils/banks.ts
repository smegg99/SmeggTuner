import { BANKS } from '~/types/session'
import type { Bank, Register } from '~/types/session'

// banksOfRegister reads a register's bank list from its engraved notation (LMMM, MMM, M2). Must agree EXACTLY with core/session.ParseBanks, which decides the card's columns. An unreadable name returns null, never a guess.
export function banksOfRegister(name: string): Bank[] | null {
  const s = name.trim().toUpperCase()
  if (!s) return null

  const out: Bank[] = []
  let ms = 0

  for (let i = 0; i < s.length; i++) {
    const c = s[i]

    if (c === 'L') {
      out.push('L')
      continue
    }
    if (c === 'H') {
      out.push('H')
      continue
    }
    if (c !== 'M') return null

    // M1..M4 written out, or a run of bare Ms (a musette) counting ranks in order.
    const next = s[i + 1]
    if (next && next >= '1' && next <= '4') {
      out.push(`M${next}` as Bank)
      i++
      continue
    }

    ms++
    if (ms > 4) return null // an instrument has four eight foot ranks, not five
    out.push(`M${ms}` as Bank)
  }

  // A repeated rank would put two reeds in one card column and lose one.
  if (new Set(out).size !== out.length) return null

  return out
}

// banksOfRegisters is the instrument's own ranks: every column any switch reaches, in card order low to high.
export function banksOfRegisters(registers: Register[]): Bank[] {
  const seen = new Set<Bank>()
  for (const r of registers) {
    for (const b of r.banks) seen.add(b)
  }
  return BANKS.filter(b => seen.has(b))
}
