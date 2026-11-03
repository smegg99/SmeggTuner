// app/composables/useSignal.ts
import { computed, effectScope, ref, watch } from 'vue'
import * as TunerService from '~~bindings/smegg.me/smeggtuner/backend/services/tuner/service.js'
import { AUTO_NOTE, NOTE_MAX, NOTE_MIN, toErrorKey, useTuner } from '~/composables/useTuner'

// The signal generator: play a note clean with every tone laid over it, so a reed is tuned against
// the app by ear. The note is the reed heard last, or the one the arrows picked after it.
// Module scope, so the panel and the keys agree.
export const MAX_VOICES = 4
export const MAX_CENTS = 100

const STORE = 'smeggtuner.signal'

interface Stored {
  cents: number[]
  reed: boolean
}

// The tones and the waveform are kept per machine; a missing or broken entry is one clean tone.
function load(): Stored {
  const fallback = { cents: [0], reed: true }
  try {
    const s = JSON.parse(localStorage.getItem(STORE) ?? '') as Partial<Stored>
    const cents = (Array.isArray(s.cents) ? s.cents : [])
      .filter(c => typeof c === 'number' && Math.abs(c) <= MAX_CENTS)
      .slice(0, MAX_VOICES)
    return cents.length ? { cents, reed: s.reed !== false } : fallback
  }
  catch {
    return fallback
  }
}

const saved = load()
const cents = ref<number[]>(saved.cents)
const reed = ref(saved.reed)
// The tune room shows the generator in place of the readings while open.
const open = ref(false)
const target = ref(AUTO_NOTE)
// The arrows picked the note; a reed heard later replaces it.
const manual = ref(false)
// The engine was stopped for the tone: it starts again when the tone ends.
const paused = ref(false)
// Play was pressed: the tone sounds, or starts on the first note heard.
const playing = ref(false)
const error = ref('')

watch([cents, reed], () => {
  try {
    localStorage.setItem(STORE, JSON.stringify({ cents: cents.value, reed: reed.value }))
  }
  catch { /* kept for this run only */ }
}, { deep: true })

// Detached: the watch below outlives whichever component asked first.
const scope = effectScope(true)
let started = false

export function useSignal() {
  const { note, manualNote, a4, transpose, running, start, stop: stopEngine } = useTuner()

  // The reed's note: the pinned one, else the one heard last.
  const heard = computed(() => (manualNote.value !== AUTO_NOTE ? manualNote.value : note.value))
  // Listening for a note to play.
  const waiting = computed(() => playing.value && target.value === AUTO_NOTE)

  // A tone's pitch as the engine sounds it: the band table's pitch, A4 and transposition included.
  const hz = (c: number) => a4.value * 2 ** ((target.value + transpose.value - 69) / 12 + c / 1200)

  // Nothing listens while the tone sounds: the engine stops first.
  async function sound() {
    if (running.value) {
      paused.value = true
      await stopEngine()
    }
    try {
      await TunerService.PlaySignal(target.value, cents.value, reed.value)
      error.value = ''
    }
    catch (err) {
      error.value = toErrorKey(err)
      playing.value = false
    }
  }

  // Ends the tone and listens again if the tone stopped the listening; leaving the room passes
  // resume false, as the room stops the engine itself.
  async function stop(resume = true) {
    playing.value = false
    try {
      await TunerService.StopTone()
    }
    catch { /* idempotent; nothing to show */ }
    if (paused.value) {
      paused.value = false
      if (resume) await start()
    }
  }

  // Plays the note: the reed heard last unless the arrows picked one since. With no note yet it
  // listens, and the tone starts on the first note heard.
  async function play() {
    if (!manual.value && heard.value !== AUTO_NOTE) target.value = heard.value
    playing.value = true
    if (target.value !== AUTO_NOTE) await sound()
    else if (!running.value) await start()
  }

  // A change while the tone sounds is heard at once.
  function replay() {
    if (playing.value && target.value !== AUTO_NOTE) void sound()
  }

  // A reed heard while silent becomes the note; a Play still listening takes the first one.
  if (!started) {
    started = true
    scope.run(() => watch(heard, (n) => {
      if (n === AUTO_NOTE || (playing.value && !waiting.value)) return
      target.value = n
      manual.value = false
      if (playing.value) void sound()
    }))
  }

  return {
    open,
    cents,
    reed,
    playing,
    waiting,
    target,
    error,
    hz,

    play,
    stop,
    toggle: () => (playing.value ? stop() : play()),
    // Closing it silences the tone.
    setOpen: (on: boolean) => {
      open.value = on
      if (!on && playing.value) void stop()
    },
    // Picks a note from the one shown, so any note plays whether or not a reed was heard.
    step: (delta: number) => {
      const next = (target.value === AUTO_NOTE ? 69 : target.value) + delta
      if (next < NOTE_MIN || next > NOTE_MAX) return
      target.value = next
      manual.value = true
      replay()
    },
    setCents: (i: number, c: number) => {
      cents.value[i] = Math.round(Math.max(-MAX_CENTS, Math.min(MAX_CENTS, c)) * 10) / 10
      replay()
    },
    addVoice: () => {
      if (cents.value.length >= MAX_VOICES) return
      cents.value.push(0)
      replay()
    },
    removeVoice: (i: number) => {
      if (cents.value.length <= 1) return
      cents.value.splice(i, 1)
      replay()
    },
    setReed: (on: boolean) => {
      reed.value = on
      replay()
    },
  }
}
