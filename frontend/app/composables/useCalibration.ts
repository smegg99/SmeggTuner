import { computed, ref } from 'vue'
import * as SessionService from '~~bindings/smegg.me/smeggtuner/services/session/service.js'
import { buildSteps, isRecorded, nextUndone } from '~/composables/calibrationSteps'
import type { Step } from '~/composables/calibrationSteps'
import { captureStatus } from '~/composables/calibrationStatus'
import { useCaptureLatch } from '~/composables/captureLatch'
import { useRecord } from '~/composables/useRecord'
import { useSessions } from '~/composables/useSessions'
import { AUTO_NOTE, useTuner } from '~/composables/useTuner'

export type Stage = 'range' | 'setup' | 'sweep' | 'complete'
export type RangePhase = 'low' | 'high' | 'done'

const stage = ref<Stage>('range')

const phase = ref<RangePhase>('low')
const lo = ref(0)
const hi = ref(0)
const saving = ref(false)

const register = ref('')
const steps = ref<Step[]>([])
const stepIndex = ref(0)

export function useCalibration() {
  const { active } = useSessions()
  const sessions = useSessions()
  const { setManualNote, state, inputLevel, note, reeds, locked, lockProgress } = useTuner()
  const record = useRecord()
  const { heard, shown, consume, resetLatch } = useCaptureLatch()

  const instrument = computed(() => active.value?.instrument)

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

  // The range the sweep walks: learned now, or what the instrument already knew.
  const sweepLo = computed(() => lo.value || instrument.value?.lo || 0)
  const sweepHi = computed(() => hi.value || instrument.value?.hi || 0)

  const registers = computed(() => instrument.value?.registers ?? [])

  const currentStep = computed<Step | null>(() => steps.value[stepIndex.value] ?? null)

  // register.value is '' for a no-register sweep.
  const isDone = (s: Step) => isRecorded(record.table.value?.rows ?? [], s.note, register.value)
  const captured = computed(() => currentStep.value ? isDone(currentStep.value) : false)

  const capturedCount = computed(() => record.table.value?.rows.length ?? 0)

  const progress = computed(() => ({ done: stepIndex.value, total: steps.value.length }))

  function applyStep() {
    const s = currentStep.value
    if (!s) return
    void setManualNote(s.note)
  }

  async function begin(reg: string) {
    register.value = reg
    steps.value = buildSteps(sweepLo.value, sweepHi.value)

    await sessions.setRegister(reg)
    await record.setArmed(true)
    // Load readings first so a resumed sweep skips only voices it really has.
    await record.load()
    stepIndex.value = nextUndone(steps.value, 0, isDone)
    stage.value = 'sweep'
    if (stepIndex.value >= steps.value.length) {
      void finish()
      return
    }
    applyStep()
  }

  function advance() {
    const next = nextUndone(steps.value, stepIndex.value + 1, isDone)
    if (next >= steps.value.length) {
      void finish()
      return
    }
    stepIndex.value = next
    applyStep()
  }

  const skip = advance

  async function finish() {
    void setManualNote(AUTO_NOTE)
    stage.value = 'complete'
  }

  // Hand the pinned note back on unmount; re-entry resumes from the readings on disk.
  function leave() {
    void setManualNote(AUTO_NOTE)
  }

  // Start over: range if the keyboard is unknown, else the register pick.
  function reset() {
    resetRange()
    register.value = ''
    steps.value = []
    stepIndex.value = 0
    stage.value = instrument.value?.lo && instrument.value?.hi ? 'setup' : 'range'
  }

  function toSetup() {
    stage.value = 'setup'
  }

  function redoRange() {
    resetRange()
    stage.value = 'range'
  }

  const status = captureStatus({ stage, heard, reeds, locked, state, inputLevel, note })

  return {
    instrument,
    stage: computed(() => stage.value),

    phase: computed(() => phase.value),
    lo: computed(() => lo.value),
    hi: computed(() => hi.value),
    heard,
    shown,
    liveNote: note,
    backwards,
    saving: computed(() => saving.value),
    capture,
    backRange,
    saveRange,
    toSetup,

    registers,
    register: computed(() => register.value),
    currentStep,
    captured,
    capturedCount,
    progress,
    begin,
    advance,
    skip,
    finish,
    redoRange,
    leave,

    status,
    lockProgress,
    reset,
  }
}
