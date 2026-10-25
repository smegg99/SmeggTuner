import { computed, ref } from 'vue'
import * as RecordService from '~~bindings/smegg.me/smeggtuner/backend/services/record/service.js'
import * as TunerService from '~~bindings/smegg.me/smeggtuner/backend/services/tuner/service.js'
import { buildBassSteps, buildSteps, isRecorded, isRecordedBass, nextUndone } from '~/composables/calibrationSteps'
import type { Step } from '~/composables/calibrationSteps'
import { useCalibrationRange } from '~/composables/calibrationRange'
import { captureStatus } from '~/composables/calibrationStatus'
import { useCaptureLatch } from '~/composables/captureLatch'
import { useRecord } from '~/composables/useRecord'
import { useSessions } from '~/composables/useSessions'
import { AUTO_NOTE, useTuner } from '~/composables/useTuner'

export type Stage = 'range' | 'setup' | 'sweep' | 'complete'

const stage = ref<Stage>('range')

const register = ref('')
// The sweep walks one keyboard at a time; a bass sweep steps the twelve buttons instead of keys.
const bassSweep = ref(false)
const steps = ref<Step[]>([])
const stepIndex = ref(0)

export function useCalibration() {
  const { active } = useSessions()
  const sessions = useSessions()
  const { setManualNote, state, inputLevel, note, reeds, lock } = useTuner()
  const record = useRecord()
  const { heard, shown, consume, resetLatch } = useCaptureLatch()
  const { phase, lo, hi, saving, backwards, resetRange, capture, backRange, saveRange } = useCalibrationRange(heard, consume, resetLatch)

  const instrument = computed(() => active.value?.instrument)

  // The range the sweep walks: learned now, or what the instrument already knew.
  const sweepLo = computed(() => lo.value || instrument.value?.lo || 0)
  const sweepHi = computed(() => hi.value || instrument.value?.hi || 0)

  const registers = computed(() => instrument.value?.registers ?? [])

  const currentStep = computed<Step | null>(() => steps.value[stepIndex.value] ?? null)

  // register.value is '' for a no-register sweep, and for the whole (or fixed) bass machine.
  const isDone = (s: Step) => (s.pc !== undefined
    ? isRecordedBass(record.table.value?.rows ?? [], s.pc, register.value)
    : isRecorded(record.table.value?.rows ?? [], s.note, register.value))
  const captured = computed(() => currentStep.value ? isDone(currentStep.value) : false)

  const capturedCount = computed(() => record.table.value?.rows.length ?? 0)

  const progress = computed(() => ({ done: stepIndex.value, total: steps.value.length }))

  function applyStep() {
    const s = currentStep.value
    if (!s) return
    // A bass button's octave is the machine's own; the note is detected, never pinned.
    void setManualNote(s.pc !== undefined ? AUTO_NOTE : s.note)
  }

  async function begin(reg: string, bass = false) {
    register.value = reg
    bassSweep.value = bass
    if (bass) {
      steps.value = buildBassSteps()
      await sessions.setBass(true)
      await sessions.setBassRegister(reg)
    }
    else {
      steps.value = buildSteps(sweepLo.value, sweepHi.value)
      await sessions.setBass(false).catch(() => undefined)
      await sessions.setRegister(reg)
    }
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
    await leave()
    stage.value = 'complete'
  }

  // Hand the pinned note back on unmount; re-entry resumes from the readings on disk.
  async function leave() {
    await RecordService.SetArmed(false)
    await TunerService.StopTone()
    await TunerService.SetManualNote(AUTO_NOTE)
  }

  async function back(): Promise<boolean> {
    if (stage.value === 'sweep' || stage.value === 'complete') {
      await leave()
      stage.value = 'setup'
      return true
    }
    if (stage.value === 'range' && phase.value !== 'low') {
      backRange()
      return true
    }
    return false
  }

  // Start over: range if the keyboard is unknown, else the register pick.
  function reset() {
    resetRange()
    register.value = ''
    bassSweep.value = false
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

  // Holding means slots sound but the engine shows no lock yet (Lock.Visible), the edge a take is recorded on.
  const sounding = computed(() => reeds.value.filter(r => r.present))
  const locked = computed(() => lock.value.visible)
  const status = captureStatus({ stage, heard, reeds: sounding, locked, state, inputLevel, note })

  return {
    instrument,
    stage: computed(() => stage.value),

    phase,
    lo,
    hi,
    heard,
    shown,
    liveNote: note,
    backwards,
    saving,
    capture,
    backRange,
    saveRange,
    toSetup,

    registers,
    register: computed(() => register.value),
    bassSweep: computed(() => bassSweep.value),
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
    back,

    status,
    reset,
  }
}
