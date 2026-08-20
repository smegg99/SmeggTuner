export type KeyTone = 'success' | 'warn' | 'error' | 'accent'

// reed is the app's red, goal its green.
const TONE_COLOR: Record<KeyTone, string> = {
  success: 'goal',
  warn: 'warn',
  error: 'reed',
  accent: 'accent',
}

// One cqw was 11px at the 1100px default window; the width prop is still given in those units so
// callers did not all have to change, and it is converted here.
const CQW_PX = 11

// buildKeyStyle folds a key's width (--key-w) and tone colour (--key-tone) into one style binding.
export function buildKeyStyle(fixed?: number, tone?: KeyTone): Record<string, string> | undefined {
  const s: Record<string, string> = {}
  if (fixed !== undefined) s['--key-w'] = `${fixed * CQW_PX}px`
  if (tone) s['--key-tone'] = `var(--v-theme-${TONE_COLOR[tone]})`
  return Object.keys(s).length ? s : undefined
}

// ghostFaces normalises the faces this key sizes to into a list.
export function ghostFaces(sizeFor?: string | readonly string[]): readonly string[] {
  if (!sizeFor) return []
  return typeof sizeFor === 'string' ? [sizeFor] : sizeFor
}
