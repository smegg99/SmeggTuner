// Pure hex <-> HSV for the accent picker's sliders, which move in HSV while the rest of the app stores #rrggbb.

export interface Hsv {
  /** hue, 0-360 */
  h: number
  /** saturation, 0-100 */
  s: number
  /** value, 0-100 */
  v: number
}

// normalizeHex accepts #rgb/#rrggbb, hash optional, any case; null when not yet a colour (mid-type) so the caller leaves sliders put.
export function normalizeHex(input: string): string | null {
  const body = input.trim().replace(/^#/, '')
  if (/^[0-9a-fA-F]{3}$/.test(body)) return `#${body.split('').map(c => c + c).join('').toLowerCase()}`
  if (/^[0-9a-fA-F]{6}$/.test(body)) return `#${body.toLowerCase()}`
  return null
}

export function hexToHsv(hex: string): Hsv {
  const n = normalizeHex(hex) ?? '#000000'
  const r = parseInt(n.slice(1, 3), 16) / 255
  const g = parseInt(n.slice(3, 5), 16) / 255
  const b = parseInt(n.slice(5, 7), 16) / 255
  const max = Math.max(r, g, b)
  const min = Math.min(r, g, b)
  const d = max - min

  let h = 0
  if (d !== 0) {
    if (max === r) h = ((g - b) / d) % 6
    else if (max === g) h = (b - r) / d + 2
    else h = (r - g) / d + 4
    h *= 60
    if (h < 0) h += 360
  }
  const s = max === 0 ? 0 : d / max
  return { h, s: s * 100, v: max * 100 }
}

export function hsvToHex({ h, s, v }: Hsv): string {
  const sat = s / 100
  const val = v / 100
  const c = val * sat
  const x = c * (1 - Math.abs(((h / 60) % 2) - 1))
  const m = val - c

  let r = 0
  let g = 0
  let b = 0
  if (h < 60) [r, g, b] = [c, x, 0]
  else if (h < 120) [r, g, b] = [x, c, 0]
  else if (h < 180) [r, g, b] = [0, c, x]
  else if (h < 240) [r, g, b] = [0, x, c]
  else if (h < 300) [r, g, b] = [x, 0, c]
  else [r, g, b] = [c, 0, x]

  const byte = (n: number) => Math.round((n + m) * 255).toString(16).padStart(2, '0')
  return `#${byte(r)}${byte(g)}${byte(b)}`
}
