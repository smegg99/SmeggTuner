// The calibration's first stage: the lowest and highest keys are played and captured, then saved to the
// session's accordion. State is module level, as the rest of the calibration's, so leaving keeps it.
import { computed, ref, type Ref } from 'vue'
import * as SessionService from '~~bindings/smegg.me/smeggtuner/backend/services/session/service.js'

export type RangePhase = 'low' | 'high' | 'done'

const phase = ref<RangePhase>('low')
const lo = ref(0)
const hi = ref(0)
const saving = ref(false)

/** heard is the latched key; consume and resetLatch are its latch's. */
export function useCalibrationRange(heard: Readonly<Ref<number>>, consume: () => void, resetLatch: () => void) {
  function resetRange() {
    phase.value = 'low'
    lo.value = 0
    hi.value = 0
    resetLatch()
  }

  function capture() {
    if (!heard.value) return
    if (phase.value === 'low') {
      lo.value = heard.value
      phase.value = 'high'
    }
    else if (phase.value === 'high') {
      hi.value = heard.value
      phase.value = 'done'
    }
    consume()
  }

  const backwards = computed(() => lo.value > 0 && hi.value > 0 && lo.value > hi.value)

  function backRange() {
    if (phase.value === 'done') phase.value = 'high'
    else if (phase.value === 'high') phase.value = 'low'
  }

  async function saveRange(): Promise<boolean> {
    if (backwards.value || !lo.value || !hi.value) return false
    saving.value = true
    try {
      await SessionService.SetKeyboardRange(lo.value, hi.value)
      return true
    }
    finally {
      saving.value = false
    }
  }

  return {
    phase: computed(() => phase.value),
    lo: computed(() => lo.value),
    hi: computed(() => hi.value),
    saving: computed(() => saving.value),
    backwards, resetRange, capture, backRange, saveRange,
  }
}
