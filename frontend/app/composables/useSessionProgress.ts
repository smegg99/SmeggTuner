import { computed } from 'vue'
import { useSessions } from '~/composables/useSessions'
import { useRecord } from '~/composables/useRecord'
import { sessionProgress } from '~/utils/sessionProgress'
import type { ProgressSession, SessionProgress } from '~/utils/sessionProgress'

// Counting is done by the pure sessionProgress(); the cast bridges the session DTO to the label's shape.
export function useSessionProgress() {
  const { active } = useSessions()
  const { table, recording } = useRecord()

  return computed<SessionProgress>(() => sessionProgress(
    active.value as unknown as ProgressSession | null,
    table.value?.rows ?? [],
    recording.value,
  ))
}
