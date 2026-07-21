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

export const LIGHT: Tokens = {
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

export const DARK: Tokens = {
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

export const TOKENS: Record<'light' | 'dark', Tokens> = { light: LIGHT, dark: DARK }
