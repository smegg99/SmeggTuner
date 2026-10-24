import { palette, type WritableTheme } from '@smegg99/s99ab/theme'

// Vuetify and the canvases both read these names (via useTheme), so DOM and canvas colours can't drift. See docs/architecture.md.

/** The three semantic hues; nothing else may borrow them. */
export const SEMANTIC = ['reed', 'goal', 'warn'] as const

export type SemanticName = typeof SEMANTIC[number]

export type TokenName
  = | 'bg' | 'chrome' | 'chrome2' | 'raised' | 'sunk'
    | 'well' | 'wellLine' | 'line' | 'lineSoft' | 'row'
    | 'ink' | 'ink2' | 'ink3'
    | 'neutral'
    | 'accent'
    | SemanticName

export type Tokens = Record<TokenName, string>

const instrumentLight: Tokens = {
  bg: '#c4c9d2',
  chrome: '#e5e8ee',
  chrome2: '#dadee6',
  raised: '#f7f8fb',
  sunk: '#bcc2cc',

  well: '#ffffff',
  wellLine: '#dce0e8',
  line: '#a9b0bc',
  lineSoft: '#c5cbd5',
  row: '#e8ebf1',

  ink: '#0f1218',
  ink2: '#3f4653',
  ink3: '#737b88',

  neutral: '#68707d',

  reed: '#e5352f',
  goal: '#12a24d',
  warn: '#e58a00',

  accent: '#2563eb',
}

const instrumentDark: Tokens = {
  bg: '#111317',
  chrome: '#1c1f24',
  chrome2: '#23272c',
  raised: '#2d3138',
  sunk: '#15171b',

  well: '#0b0c0f',
  wellLine: '#1e2127',
  line: '#30353c',
  lineSoft: '#24282e',
  row: '#16191d',

  ink: '#e9ebee',
  ink2: '#a3aab3',
  ink3: '#6e757f',

  neutral: '#8b939e',

  reed: '#e85d51',
  goal: '#42c07d',
  warn: '#e0ad4f',

  accent: '#5b8cf5',
}

// The same graphite surfaces as Goptivum; instrument traces retain their meaning.
function graphite(dark: boolean, instrument: Tokens): Tokens {
  const shared = palette({ name: 'graphite', dark, chroma: 0 })
  return Object.fromEntries(Object.entries(instrument).map(([key, value]) => [
    key, SEMANTIC.includes(key as SemanticName) ? value : shared[key] ?? value,
  ])) as Tokens
}

export const LIGHT = graphite(false, instrumentLight)
export const DARK = graphite(true, instrumentDark)
export const TOKENS: Record<'light' | 'dark', Tokens> = { light: LIGHT, dark: DARK }

/** Map the instrument palette onto the shared controls without changing its surfaces. */
export function instrumentColors(name: 'light' | 'dark', accent = TOKENS[name].accent) {
  const t = TOKENS[name]
  return {
    ...t, accent,
    'background': t.bg, 'surface': t.chrome,
    'surface-bright': t.raised, 'surface-light': t.chrome2, 'surface-variant': t.sunk,
    'on-background': t.ink, 'on-surface': t.ink,
    'primary': t.ink2, 'secondary': t.ink3, 'edge': t.line,
    'error': t.reed, 'danger': t.reed, 'danger-lit': t.reed,
    'success': t.goal, 'ok': t.goal, 'ok-lit': t.goal,
    'warning': t.warn, 'warn-lit': t.warn, 'info': t.ink2, 'info-lit': t.ink2,
  }
}

/** Supply raw colours so Vuetify derives every on-colour from the instrument palette. */
export function applyInstrumentTheme(theme: WritableTheme, name: 'light' | 'dark', accent = TOKENS[name].accent) {
  theme.colors = instrumentColors(name, accent)
  Object.assign(theme.variables, {
    'border-color': TOKENS[name].line,
    'border-opacity': 1,
    'high-emphasis-opacity': 1,
    'medium-emphasis-opacity': 1,
  })
}
