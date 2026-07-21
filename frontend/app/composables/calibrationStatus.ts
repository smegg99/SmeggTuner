import { computed } from 'vue'
import type { ComputedRef, Ref } from 'vue'
import { AUTO_NOTE } from '~/composables/tuner/tunerProtocol'

export type CaptureStatus = 'idle' | 'listening' | 'holding' | 'captured' | 'loud'

// Below this the input is silence, not a note (linear 0..1).
const LISTEN_LEVEL = 0.02

export interface StatusSignals {
  stage: Ref<string>
  heard: Ref<number>
  reeds: Ref<{ length: number }>
  locked: Ref<boolean>
  state: Ref<string>
  inputLevel: Ref<number>
  note: Ref<number>
}

// Precedence: a caught note wins, then the lock-hold, then a bare sound.
export function captureStatus(s: StatusSignals): ComputedRef<CaptureStatus> {
  return computed<CaptureStatus>(() => {
    if (s.heard.value) return 'captured'
    if (s.reeds.value.length > 0 && !s.locked.value) return 'holding'
    if (s.state.value === 'tooLoud') return 'loud'
    // Sweep pins the note, so read the level; range keys off a tracked note with no reeds.
    const listening = s.stage.value === 'sweep'
      ? s.state.value === 'running' && s.inputLevel.value > LISTEN_LEVEL
      : s.note.value !== AUTO_NOTE && s.reeds.value.length === 0
    return listening ? 'listening' : 'idle'
  })
}
