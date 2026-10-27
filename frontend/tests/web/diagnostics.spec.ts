// tests/web/diagnostics.spec.ts
import { describe, expect, it } from 'vitest'
import { diagnosticsText } from '~/web/diagnostics'

describe('diagnosticsText', () => {
  it('includes every listed item', () => {
    const text = diagnosticsText({
      version: 'v1.4.0',
      userAgent: 'UA',
      platform: 'android',
      track: { echoCancellation: false, noiseSuppression: true, autoGainControl: false, sampleRate: 48000, channelCount: 1 },
      contextRate: 48000,
      baseLatency: 0.01,
      inputLevel: -20,
      perf: { load: 0.25, dropped: 3, behind: false },
    })
    for (const key of ['version', 'browser', 'platform', 'echoCancellation', 'noiseSuppression', 'autoGainControl',
      'micSampleRate', 'channelCount', 'contextSampleRate', 'baseLatency', 'inputLevel', 'engineLoad', 'droppedBlocks', 'behind']) {
      expect(text).toMatch(new RegExp(`^${key}: `, 'm'))
    }
    expect(text).toContain('noiseSuppression: true')
    expect(text).toContain('engineLoad: 0.25')
  })

  it('marks missing values instead of inventing them', () => {
    const text = diagnosticsText({ version: 'v', userAgent: 'UA', platform: 'linux', track: null, contextRate: null, baseLatency: null, inputLevel: null, perf: null })
    expect(text).toContain('echoCancellation: unknown')
    expect(text).toContain('engineLoad: unknown')
  })
})
