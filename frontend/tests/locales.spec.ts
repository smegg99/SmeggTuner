// frontend/tests/locales.spec.ts
import schema from '../../backend/common/config/config.cue?raw'
import { describe, expect, it } from 'vitest'
import { isLocale, savedLocale, localeOptions } from '../i18n/locales'
import { locales } from '../i18n/generated/keys.gen'
import { documentSource, documentURL, renderDocument } from '../app/utils/documents'

const configured = [...schema.match(/^#Language:.*$/m)![0].matchAll(/"([^"]+)"/g)].map(match => match[1])

it('keeps persisted configuration and generated catalogs in sync', () => {
  expect(new Set(configured)).toEqual(new Set(locales))
})

it('rejects unsupported persisted values', () => {
  for (const value of [undefined, null, '', 'xx', {}, 'constructor']) {
    expect(isLocale(value)).toBe(false)
    expect(savedLocale(value)).toBe('en')
  }
})

describe.each(localeOptions)('$code language', ({ code, name }) => {
  it('survives restoring saved preferences with a readable picker label', () => {
    expect(savedLocale(code)).toBe(code)
    expect(name).not.toBe(code)
  })
  it('bundles its translated privacy policy and links to the matching page', () => {
    const source = documentSource('privacy', code)
    if (code !== 'en') expect(source).not.toBe(documentSource('privacy', 'en'))
    const article = document.createElement('article')
    article.innerHTML = renderDocument(source, 'privacy-')
    expect(article.querySelectorAll('h3')).toHaveLength(8)
    expect(article.textContent).toContain('me.smegg.smeggtuner')
    for (const href of [
      'mailto:support@smegg.me',
      'https://www.cloudflare.com/privacypolicy/',
      'https://www.cloudflare.com/turnstile-privacy-policy/',
    ]) {
      expect(article.querySelector(`a[href="${href}"]`)).not.toBeNull()
    }
    expect(documentURL('privacy', code)).toBe(`https://smeggtuner.com/${code === 'en' ? '' : `${code}/`}privacy`)
  })
})
