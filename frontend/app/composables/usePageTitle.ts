// The name a room gets in the main bar; Tune's bar is full of its keys, and the workshop names its own panels.
import { computed } from 'vue'
import { useShell } from '~/composables/useShell'
import type { View } from '~/composables/useShell'

const TITLES: Partial<Record<View, string>> = {
  settings: 'common.routes.settings',
  soundcard: 'settings.calibration.label',
  calibrate: 'tuner.calibrate.title',
}

export function usePageTitle() {
  const { view } = useShell()
  const { t } = useI18n()
  return computed(() => {
    const key = TITLES[view.value]
    return key ? t(key) : ''
  })
}
