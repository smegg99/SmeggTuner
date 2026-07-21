// i18n/plurals.ts

/*
 * POLISH COUNTS IN FOUR, AND vue-i18n COUNTS IN TWO.
 *
 * Without this rule every Polish count in the app is wrong, and wrong in a way nothing catches:
 * the default is the English one/other, so a three-part "brak | jeden | {count} wpisów" prints
 * "2 wpisów" - the genitive plural where the nominative belongs, which reads to a Pole exactly
 * as "2 entrys" reads to an English speaker. Four-part strings fared no better: the fourth form
 * was unreachable, so five reeds came out "5 głosy".
 *
 * The bands the language actually has:
 *
 *   0          brak wpisów      a "brak" form, so a count can read as a word
 *   1          1 wpis           nominative singular
 *   2-4        2 wpisy          nominative plural, but NOT 12-14
 *   otherwise  5 wpisów         genitive plural: 5, 12, 25
 *
 * 22 is few and 12 is many, which is why the tens are tested and not just the ones. That pair is
 * the whole reason this cannot be `n < 5`.
 *
 * The index is clamped to what the message offers, so a two-part string still behaves as
 * singular/plural rather than reaching past its own end. Every counted string should still carry
 * all four: a missing part silently collapses a band. tests/polishPlurals.spec.ts holds both ends.
 */
export function polishPlural(choice: number, choicesLength: number): number {
  const n = Math.abs(Math.floor(choice))
  if (choicesLength < 3) return n === 1 ? 0 : 1
  if (n === 0) return 0
  if (n === 1) return 1

  const ones = n % 10
  const tens = n % 100
  const few = ones >= 2 && ones <= 4 && (tens < 10 || tens >= 20)
  return Math.min(few ? 2 : 3, choicesLength - 1)
}
