import { describe, expect, it } from 'vitest'
import { paneLayout } from '~/composables/render/spectrumPanes'

describe('the per-tone spectra', () => {
  it('splits the plot into equal panes with a gap between', () => {
    const panes = paneLayout(10, 320, 3, 10)
    expect(panes.map(p => p.width)).toEqual([100, 100, 100])
    expect(panes.map(p => p.left)).toEqual([10, 120, 230])
  })

  it('has nothing to lay out without pictures', () => {
    expect(paneLayout(0, 100, 0, 5)).toEqual([])
  })
})
