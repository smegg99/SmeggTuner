// Holds the settled reading capturable for a grace window past the lock drop; a consumed lock's tail must not re-arm Capture.
import { computed, effectScope, ref, watch } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'
import { useTuner } from '~/composables/useTuner'

const consumedLock = ref(false)
// Last note the detector was sure of; shown greyed while it waits for the next.
const lastHeard = ref(0)
// The latched reading, kept capturable for captureMs after the lock drops.
const graceNote = ref(0)
let graceTimer: ReturnType<typeof setTimeout> | null = null
let captureMs = () => 2000

function clearGrace() {
  graceNote.value = 0
  if (graceTimer !== null) {
    clearTimeout(graceTimer)
    graceTimer = null
  }
}

function startGrace(latched: number) {
  clearGrace()
  if (!latched) return
  graceNote.value = latched
  graceTimer = setTimeout(() => {
    graceNote.value = 0
    graceTimer = null
  }, captureMs())
}

// Detached, so the watches outlive whichever component mounted first.
const scope = effectScope(true)
let watching = false

export function useCaptureLatch() {
  const { note, locked } = useTuner()
  const { config } = useConfigSync()
  captureMs = () => config.tuner?.calibration_capture_ms || 2000

  if (!watching) {
    watching = true
    scope.run(() => {
      watch(locked, (isLocked) => {
        if (isLocked) {
          clearGrace()
        }
        else {
          if (!consumedLock.value) startGrace(lastHeard.value)
          consumedLock.value = false
        }
      })
      watch([locked, note], ([isLocked, n]) => {
        if (isLocked && n) lastHeard.value = n
      })
    })
  }

  // The live lock, or the grace latch after it drops; a spent lock counts as neither.
  const heard = computed(() => {
    if (consumedLock.value) return 0
    if (locked.value) return note.value
    return graceNote.value
  })
  const shown = computed(() => heard.value || lastHeard.value)

  // Mark the lock consumed; the next step waits for a new note.
  function consume() {
    consumedLock.value = true
    clearGrace()
  }

  function resetLatch() {
    consumedLock.value = false
    lastHeard.value = 0
    clearGrace()
  }

  return { heard, shown, consume, resetLatch }
}
