import { computed } from 'vue'
import { useSessions } from '~/composables/useSessions'
import { useRecord } from '~/composables/useRecord'
import { sessionProgress } from '~/utils/sessionProgress'
import type { ProgressSession, SessionProgress } from '~/utils/sessionProgress'
import type { TakeRow } from '~/types/record'

// Counting is done by the pure sessionProgress(); the casts bridge the record DTOs (numeric-enum note) to the label's shapes, same as SessionDetail.
export function useSessionProgress() {
  const { active } = useSessions()
  const { table, recording } = useRecord()

  return computed<SessionProgress>(() => sessionProgress(
    active.value as unknown as ProgressSession | null,
    (table.value?.rows ?? []) as unknown as TakeRow[],
    recording.value,
  ))
}
