// frontend/tests/documents.spec.ts
import { describe, expect, it } from 'vitest'
import { documentImages, documentSource, documentURL, renderDocument } from '../app/utils/documents'

describe.each(['en', 'pl'] as const)('bundled %s documents', (locale) => {
  it('renders the complete policy and direct contact from the bundled source', () => {
    const article = document.createElement('article')
    article.innerHTML = renderDocument(documentSource('privacy', locale), 'privacy-')
    expect(article.textContent!.length).toBeGreaterThan(7000)
    expect(article.textContent).toContain('smegg99')
    expect(article.querySelector('a[href="mailto:support@smegg.me"]')).not.toBeNull()
    expect(article.querySelectorAll('h3').length).toBe(8)
    expect(article.querySelector('iframe, script, img')).toBeNull()
  })

  it('renders the guide tables, local screenshots and working section links', () => {
    const article = document.createElement('article')
    article.innerHTML = renderDocument(documentSource('guide', locale), 'guide-')
    expect(article.querySelectorAll('table').length).toBeGreaterThanOrEqual(6)
    expect(article.querySelectorAll('img').length).toBe(6)
    for (const img of article.querySelectorAll('img')) {
      expect(Object.values(documentImages)).toContain(img.getAttribute('src'))
      expect(img.alt.length).toBeGreaterThan(10)
    }
    for (const link of article.querySelectorAll<HTMLAnchorElement>('a[href^="#"]:not([data-document-locale])')) {
      expect(article.querySelector(`[id="${decodeURIComponent(link.hash.slice(1))}"]`)).not.toBeNull()
    }
    expect(article.querySelector('a[data-document-locale]')).not.toBeNull()
    expect(article.querySelectorAll('a[data-document-image]').length).toBe(6)
  })

  it('keeps the online version separate from the bundled text', () => {
    expect(documentURL('privacy', locale)).toBe(`https://smeggtuner.com/${locale === 'pl' ? 'pl/' : ''}privacy`)
    expect(documentSource('license', locale)).toContain('GNU AFFERO GENERAL PUBLIC LICENSE')
  })
})

it('renders raw HTML as text and drops executable link schemes', () => {
  const element = document.createElement('div')
  element.innerHTML = renderDocument('<script>alert(1)</script>\n\n[unsafe](javascript:alert%281%29)', 'test-')
  expect(element.querySelector('script, a')).toBeNull()
  expect(element.textContent).toContain('<script>')
})

it('rejects image URLs that would need the network', () => {
  expect(() => renderDocument('![remote](https://example.com/image.png)', 'test-')).toThrow('not bundled')
})
