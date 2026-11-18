import { computed } from 'vue'
import type { ToolItem } from '@smegg99/s99ab/types'
import { useConfigSync } from '~/composables/useConfigSync'
import { deviationUnit } from '~/utils/deviation'

/** The unit every deviation is shown in, from tuner.deviation_unit. */
export function useDeviationUnit() {
  const { config } = useConfigSync()
  return computed(() => deviationUnit(config.tuner?.deviation_unit))
}

/** Beat readouts use either hertz or the selected reed deviation unit. */
export function useBeatUnitItems() {
  const { t } = useI18n()
  const deviation = useDeviationUnit()
  return computed<ToolItem[]>(() => [
    { value: 'hz', label: t('tuner.unit.hz') },
    { value: 'reference', label: t(`tuner.unit.${deviation.value}`) },
  ])
}
