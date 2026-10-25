import { EMPTY } from '~/utils/record'

// Central number formatting so every component rounds deviations identically.

// Deviations to one decimal, as the reference prints a reed.
const DEVIATION_DIGITS = 1

// Beats in hertz to two decimals, as the reference prints a beat box.
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

  const deviation = (value: number) => fixed(value, DEVIATION_DIGITS)
  const signedDeviation = (value: number) => signed(value, DEVIATION_DIGITS)
  const hz = (value: number) => fixed(value, HZ_DIGITS)

  /** Parse a typed number, accepting the locale's comma decimal separator. */
  function parse(text: string): number | null {
    const trimmed = text.trim()
    // Number("") is 0; an emptied cell must not read as a perfectly tuned reed.
    if (!trimmed) return null
    const value = Number(trimmed.replace(',', '.'))
    return Number.isFinite(value) ? value : null
  }

  return { deviation, signedDeviation, hz, parse, EMPTY }
}
