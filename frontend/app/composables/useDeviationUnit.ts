import { computed } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'
import { deviationUnit } from '~/utils/deviation'

/** The unit every deviation is shown in, from tuner.deviation_unit. */
export function useDeviationUnit() {
  const { config } = useConfigSync()
  return computed(() => deviationUnit(config.tuner?.deviation_unit))
}
