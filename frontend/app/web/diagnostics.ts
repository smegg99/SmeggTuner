// app/web/diagnostics.ts
import type { PerfDTO } from './protocol'

export interface DiagnosticsInput {
  version: string
  userAgent: string
  platform: string
  track: MediaTrackSettings | null
  contextRate: number | null
  baseLatency: number | null
  inputLevel: number | null
  perf: PerfDTO | null
}

// diagnosticsText is one "key: value" line per item, for pasting into a bug report; unknown stays unknown.
export function diagnosticsText(d: DiagnosticsInput): string {
  const v = (x: unknown) => (x === null || x === undefined ? 'unknown' : String(x))
  const rows: [string, unknown][] = [
    ['version', d.version],
    ['browser', d.userAgent],
    ['platform', d.platform],
    ['echoCancellation', d.track?.echoCancellation],
    ['noiseSuppression', d.track?.noiseSuppression],
    ['autoGainControl', d.track?.autoGainControl],
    ['micSampleRate', d.track?.sampleRate],
    ['channelCount', d.track?.channelCount],
    ['contextSampleRate', d.contextRate],
    ['baseLatency', d.baseLatency],
    ['inputLevel', d.inputLevel],
    ['engineLoad', d.perf ? Math.round(d.perf.load * 100) / 100 : null],
    ['droppedBlocks', d.perf?.dropped],
    ['behind', d.perf?.behind],
  ]
  return rows.map(([k, x]) => `${k}: ${v(x)}`).join('\n')
}
