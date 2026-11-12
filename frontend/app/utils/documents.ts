// frontend/app/utils/documents.ts
import { Marked, Renderer } from 'marked'
import privacyEnglish from '../../../PRIVACY.md?raw'
import privacyPolish from '../../../PRIVACY.pl.md?raw'
import usageEnglish from '../../../USAGE.md?raw'
import usagePolish from '../../../USAGE.pl.md?raw'
import license from '../../../LICENSE?raw'

export type DocumentKind = 'guide' | 'privacy' | 'license'
export type DocumentLocale = 'en' | 'pl'

const sources = {
  guide: { en: usageEnglish, pl: usagePolish },
  privacy: { en: privacyEnglish, pl: privacyPolish },
  license: { en: license, pl: license },
}
const assets = import.meta.glob<string>('../../../assets/screenshots/usage-*.png', {
  eager: true, query: '?url', import: 'default',
})
export const documentImages = Object.fromEntries(Object.entries(assets).map(([file, url]) => [file.replace('../../../', ''), url]))

export function documentSource(kind: DocumentKind, locale: DocumentLocale): string {
  return sources[kind][locale]
}

export function documentURL(kind: DocumentKind, locale: DocumentLocale): string {
  if (kind === 'privacy') return `https://smeggtuner.com/${locale === 'pl' ? 'pl/' : ''}privacy`
  return `https://github.com/smegg99/SmeggTuner/blob/main/${kind === 'license' ? 'LICENSE' : locale === 'pl' ? 'USAGE.pl.md' : 'USAGE.md'}`
}

function escapeHTML(value: string): string {
  return value.replace(/[&<>"']/g, char => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', '\'': '&#39;' })[char]!)
}

function slug(text: string): string {
  return text.toLowerCase().replace(/[^\p{L}\p{N}\s-]/gu, '').trim().replace(/\s/g, '-')
}

// Only bundled Markdown is accepted. Raw HTML stays text and every image must be a bundled asset.
export function renderDocument(source: string, prefix: string): string {
  const markdown = new Marked({
    gfm: true,
    renderer: {
      html({ text }) { return escapeHTML(text) },
      heading({ depth, text, tokens }) {
        const level = Math.min(Math.max(depth + 1, 3), 6)
        return `<h${level} id="${escapeHTML(prefix + slug(text))}" tabindex="-1">${this.parser.parseInline(tokens)}</h${level}>`
      },
      image({ href, text }) {
        const image = documentImages[href]
        if (!image) throw new Error(`Document image is not bundled: ${href}`)
        return `<img src="${escapeHTML(image)}" alt="${escapeHTML(text)}" loading="lazy">`
      },
      link({ href, tokens }) {
        const content = this.parser.parseInline(tokens)
        const image = documentImages[href]
        if (image) return `<a href="${escapeHTML(image)}" data-document-image="${escapeHTML(image)}">${content}</a>`
        if (href === 'USAGE.md' || href === 'USAGE.pl.md') {
          return `<a href="#" data-document-locale="${href === 'USAGE.md' ? 'en' : 'pl'}">${content}</a>`
        }
        if (href.startsWith('#')) return `<a href="#${escapeHTML(prefix + href.slice(1))}">${content}</a>`
        const url = new URL(href, 'https://github.com/smegg99/SmeggTuner/blob/main/')
        if (!['https:', 'mailto:'].includes(url.protocol)) return content
        return `<a href="${escapeHTML(url.href)}">${content}</a>`
      },
      table(token) {
        const label = escapeHTML(token.header.map(cell => cell.text).join(' / '))
        return `<div class="document-table" tabindex="0" role="region" aria-label="${label}">${Renderer.prototype.table.call(this, token)}</div>`
      },
    },
  })
  return markdown.parse(source, { async: false })
}
