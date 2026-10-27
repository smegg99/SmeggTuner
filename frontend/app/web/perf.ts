// app/web/perf.ts
import type { PerfDTO } from './protocol'

const WINDOW_MS = 5000
const BEHIND_LOAD = 0.9
const BEHIND_FOR_MS = 3000

// PerfMeter tracks engine time against audio time; "behind" means it has not kept up for a while, not one slow block.
export class PerfMeter {
  private samples: { at: number, work: number, audio: number }[] = []
  private overSince: number | null = null
  private dropped = 0

  observe(workMs: number, audioMs: number, now: number) {
    this.samples.push({ at: now, work: workMs, audio: audioMs })
    while (this.samples.length && this.samples[0]!.at < now - WINDOW_MS) this.samples.shift()
    if (this.load() > BEHIND_LOAD) this.overSince ??= now
    else this.overSince = null
  }

  drop() {
    this.dropped++
  }

  snapshot(now: number): PerfDTO {
    return {
      load: this.load(),
      dropped: this.dropped,
      behind: this.overSince !== null && now - this.overSince >= BEHIND_FOR_MS,
    }
  }

  private load(): number {
    let work = 0
    let audio = 0
    for (const s of this.samples) {
      work += s.work
      audio += s.audio
    }
    return audio > 0 ? work / audio : 0
  }
}
