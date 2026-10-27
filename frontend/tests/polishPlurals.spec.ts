// tests/polishPlurals.spec.ts
import { describe, expect, it } from 'vitest'
import { createI18n } from 'vue-i18n'
import { withS99ab } from '@smegg99/s99ab/i18n'
import { pluralRules } from '../i18n/generated/plurals.gen'
import curvePl from '../i18n/generated/pl/curve.json'
import recordPl from '../i18n/generated/pl/record.json'
import sessionPl from '../i18n/generated/pl/session.json'
import tunerPl from '../i18n/generated/pl/tuner.json'
import instrumentPl from '../i18n/generated/pl/instrument.json'
import recordEn from '../i18n/generated/en/record.json'
import sessionEn from '../i18n/generated/en/session.json'
import tunerEn from '../i18n/generated/en/tuner.json'

function translator(locale: 'en' | 'pl') {
  return createI18n({
    legacy: false,
    locale,
    fallbackLocale: 'en',
    pluralRules,
    messages: withS99ab({
      en: { ...recordEn, ...sessionEn, ...tunerEn },
      pl: { ...curvePl, ...recordPl, ...sessionPl, ...tunerPl, ...instrumentPl },
    }),
  }).global.t
}

describe('generated locale messages through vue-i18n', () => {
  it.each([
    [0, '0 odczytów'], [1, '1 odczyt'], [2, '2 odczyty'], [4, '4 odczyty'],
    [5, '5 odczytów'], [12, '12 odczytów'], [14, '14 odczytów'], [21, '21 odczytów'],
    [22, '22 odczyty'], [24, '24 odczyty'], [25, '25 odczytów'], [112, '112 odczytów'],
  ])('renders Polish count %i as %s', (count, expected) => {
    expect(translator('pl')('session.card.readings', { count }, count)).toBe(expected)
  })

  it('uses the plural category for the counted argument, independently of other numbers', () => {
    expect(translator('pl')('tuner.panel.boxOctave', { box: 3, octaves: 2 }, 2))
      .toBe('Pole 3: 2 oktawy nad podstawą')
    expect(translator('en')('tuner.panel.boxOctave', { box: 2, octaves: 1 }, 1))
      .toBe('Box 2: 1 octave above the base')
  })

  it('preserves the instrument namespace and named placeholders', () => {
    expect(translator('pl')('instrument.says', { banks: 'L M M', reeds: 3 }, 3))
      .toBe('Chóry głosowe L M M. Najszerszy rejestr uruchamia 3 głosy.')
  })

  it('inflects recorded readings in each language', () => {
    expect(translator('pl')('session.card.readings', { count: 22 }, 22)).toBe('22 odczyty')
    const en = translator('en')
    expect(en('session.card.readings', { count: 0 }, 0)).toBe('0 readings')
    expect(en('session.card.readings', { count: 1 }, 1)).toBe('1 reading')
    expect(en('session.card.readings', { count: 22 }, 22)).toBe('22 readings')
  })
})
