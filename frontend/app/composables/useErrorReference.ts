import { computed } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'
import { useSessions } from '~/composables/useSessions'
import type { ErrorReference } from '~/types/config'
import {
  DEFAULT_BEAT_TOLERANCE,
  DEFAULT_ERROR_REFERENCE,
  DEFAULT_TOLERANCE,
  ERROR_REFERENCE,
} from '~/types/config'
import type { BeatError, ReedError } from '~~bindings/smegg.me/smeggtuner/backend/core/target/models.js'

// scale and goal are one measurement shown two ways: both arrive from the backend (core/target computes Curr/Goal/Error together), so switching convention picks a column, never recomputes.
// With no goal curve Goal is zero and both conventions read the same.
export function useErrorReference() {
  const { config } = useConfigSync()
  const { active } = useSessions()

  const reference = computed<ErrorReference>(() =>
    config.tuner?.error_reference === ERROR_REFERENCE.GOAL
      ? ERROR_REFERENCE.GOAL
      : DEFAULT_ERROR_REFERENCE,
  )

  const toGoal = computed(() => reference.value === ERROR_REFERENCE.GOAL)

  // In-tune band uses the instrument's tolerances when the session states them, else the app default - the same precedence the backend judges inTol by (core/session.Instrument.Tolerances), or the band and the verdict disagree.
  const tolerance = computed(() =>
    [active.value?.instrument?.tolerance, config.tuner?.tolerance].find(value => value !== undefined && value > 0) ?? DEFAULT_TOLERANCE)
  const beatTolerance = computed(() =>
    [active.value?.instrument?.beatTolerance, config.tuner?.beat_tolerance].find(value => value !== undefined && value > 0) ?? DEFAULT_BEAT_TOLERANCE)

  /** The number to display, in reference units. */
  function show(e: ReedError | BeatError): number {
    return toGoal.value ? e.error : e.curr
  }

  /** The number to drive it to; zero in the goal convention. */
  function drive(e: ReedError | BeatError): number {
    return toGoal.value ? 0 : e.goal
  }

  /** show() in hertz. */
  function showHz(e: ReedError | BeatError): number {
    return toGoal.value ? e.errorHz : e.currHz
  }

  function driveHz(e: ReedError | BeatError): number {
    return toGoal.value ? 0 : e.goalHz
  }

  return { reference, toGoal, tolerance, beatTolerance, show, drive, showHz, driveHz }
}
