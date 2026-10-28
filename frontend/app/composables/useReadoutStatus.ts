import { onScopeDispose, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'
import type { EngineState } from './tuner/tunerProtocol'

export type ReadoutStatus = EngineState | 'locked' | 'recorded'

/** Paces captions only. Measurement values and engine commands remain immediate. */
export function useReadoutStatus(
  source: MaybeRefOrGetter<ReadoutStatus>,
  timing: { readonly appearAfter: number, readonly holdFor: number },
) {
  const shown = ref<ReadoutStatus>('stopped')
  let timer: ReturnType<typeof setTimeout> | undefined
  let changedAt = 0

  watch(() => toValue(source), (next) => {
    clearTimeout(timer)
    // Never retain a lock, warm-up instruction or running state after it ends.
    if (next === 'stopped' || next === 'deviceLost') {
      shown.value = next
      changedAt = Date.now()
      return
    }
    if (shown.value === 'stopped' || shown.value === 'deviceLost'
      || shown.value === 'initializing'
      || shown.value === 'locked' || shown.value === 'recorded') {
      if (shown.value === next) return
      shown.value = 'running'
      changedAt = 0
    }
    if (next === shown.value) return
    timer = setTimeout(() => {
      shown.value = next
      changedAt = Date.now()
    }, Math.max(timing.appearAfter, changedAt + timing.holdFor - Date.now()))
  }, { immediate: true, flush: 'sync' })

  onScopeDispose(() => clearTimeout(timer))
  return shown
}
