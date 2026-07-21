import { EMPTY } from '~/utils/record'

// Central number formatting so every component rounds cents and beats identically.

// Cents to one decimal, beats to two: the engine resolves a beat to ~0.1 Hz.
const CENT_DIGITS = 1
const HZ_DIGITS = 2

export function useRecordFormat() {
  const { n } = useI18n()

  function fixed(value: number, digits: number) {
    if (!Number.isFinite(value)) return EMPTY
    return n(value, { minimumFractionDigits: digits, maximumFractionDigits: digits })
  }

  /** A signed deviation. n() writes the locale's minus, so only the plus is added. */
  function signed(value: number, digits: number) {
    if (!Number.isFinite(value)) return EMPTY
    const text = fixed(value, digits)
    return value > 0 ? `+${text}` : text
  }

  const cents = (value: number) => fixed(value, CENT_DIGITS)
  const signedCents = (value: number) => signed(value, CENT_DIGITS)
  const hertz = (value: number) => fixed(value, HZ_DIGITS)
  const signedHertz = (value: number) => signed(value, HZ_DIGITS)

  /** Parse a typed number, accepting the locale's comma decimal separator. */
  function parse(text: string): number | null {
    const trimmed = text.trim()
    // Number("") is 0; an emptied cell must not read as a perfectly tuned reed.
    if (!trimmed) return null
    const value = Number(trimmed.replace(',', '.'))
    return Number.isFinite(value) ? value : null
  }

  return { cents, signedCents, hertz, signedHertz, parse, EMPTY }
}
