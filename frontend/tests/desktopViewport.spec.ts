// tests/desktopViewport.spec.ts
import { describe, expect, it } from 'vitest'
import { desktopScale } from '~/utils/desktopViewport'

describe('desktop viewport fitting', () => {
  it.each([[320, 520], [568, 272], [412, 684], [732, 364], [800, 1280]])(
    'fits both desktop dimensions inside %ix%i without changing their proportions',
    (width, height) => {
      const scale = desktopScale(width, height)
      expect(scale).toBeGreaterThan(0)
      expect(scale).toBeLessThanOrEqual(1)
      expect(960 * scale).toBeLessThanOrEqual(width)
      expect(640 * scale).toBeLessThanOrEqual(height)
      expect(Math.max(960 * scale / width, 640 * scale / height)).toBeCloseTo(1, 5)
    },
  )

  it('does not enlarge a desktop window', () => {
    expect(desktopScale(960, 640)).toBe(1)
    expect(desktopScale(1440, 900)).toBe(1)
  })

  it('ignores a transient zero-sized surface', () => {
    expect(desktopScale(0, 640)).toBe(1)
    expect(desktopScale(960, 0)).toBe(1)
  })
})
