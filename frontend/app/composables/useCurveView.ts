import { shallowRef, watch } from 'vue'
import * as SessionService from '~~bindings/smegg.me/smeggtuner/backend/services/session/service.js'
import type { CurveViewDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/session/models.js'
import { useSessions } from '~/composables/useSessions'

// The curve as the engine reads it and the pass's readings on it; refetched whenever the session publishes.
export function useCurveView() {
  const log = useLogger()
  const { active } = useSessions()
  const view = shallowRef<CurveViewDTO | null>(null)
  let seq = 0

  watch(active, async (s) => {
    const mine = ++seq
    if (!s) {
      view.value = null
      return
    }
    try {
      const next = await SessionService.CurveView()
      if (mine === seq) view.value = next
    }
    catch (err) {
      log.error('log.frontendCurveViewFailed', { error: String(err) })
    }
  }, { immediate: true })

  return view
}
