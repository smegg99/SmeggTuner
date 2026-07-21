/** Append an alpha channel to a #rrggbb theme color. Formatting, not math. */
export function alpha(color: string, a: number) {
  if (!/^#[0-9a-f]{6}$/i.test(color)) return color
  const byte = Math.round(Math.min(1, Math.max(0, a)) * 255)
  return color + byte.toString(16).padStart(2, '0')
}

// Vuetify types a theme color as string | packed-int | HSV; fillStyle silently
// keeps the previous color for a non-string, so narrow it to a string here.
export function cssColor(value: unknown, fallback = '#000000') {
  return typeof value === 'string' ? value : fallback
}

// Eight-digit hex for a gradient stop, which carries its own alpha and ignores the
// context's globalAlpha. Not `rgb(var(--v-theme-x) / a)`: Vuetify's vars are comma
// triplets, so the slash form is invalid and dropped silently (an invisible color).
export function alphaColor(value: unknown, alpha: number, fallback = '#000000') {
  const hex = cssColor(value, fallback)
  const a = Math.round(Math.max(0, Math.min(1, alpha)) * 255)
  return `${hex}${a.toString(16).padStart(2, '0')}`
}
