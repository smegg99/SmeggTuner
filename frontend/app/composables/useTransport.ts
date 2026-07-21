import { computed, ref, shallowRef, watch } from 'vue'
import { Events } from '@wailsio/runtime'
import { usePlayhead } from '~/composables/usePlayhead'
import { useTuner } from '~/composables/useTuner'
import { fit, fitSelection, pan, scrollTo, state, view, zoom } from '~/composables/transportState'
import * as AudioService from '~~bindings/smegg.me/smeggtuner/services/audio/service.js'
import type { PeakDTO, TransportDTO } from '~~bindings/smegg.me/smeggtuner/services/audio/models.js'

// Mirror of Go's file transport (FileSource owns playhead/selection/loop); the zoom window lives in transportState. Every command stores the state Go returns; the view never predicts.

const EVENT_PLAYBACK = 'playback:position'

/** One min/max pair per column, refetched whenever the window or the width changes. */
const peaks = shallowRef<PeakDTO[]>([])

const dragging = ref(false)

/** Speaker mute/volume; Go owns the speaker, these mirror it. */
const muted = ref(false)
const volume = ref(1)

let bound = false

const { land, at, heard } = usePlayhead()

// A plain function, not a computed: called from an event handler and a frame callback, not a render.
let running = () => false

// "moving" is Go's word for "sound leaving the speakers now", not "not paused"; after start/resume the card sits stopped ~0.3s, so a coasting needle would race that silence.
function advancingNow(): boolean {
  return running() && state.available && state.moving && !state.paused && !dragging.value
}

// The only place a Go reply is unpacked, so state never updates with a stale needle.
function adopt(next: TransportDTO) {
  const wasAdvancing = advancingNow()
  // A different name or length means a different file; compared before the assign overwrites it.
  const changedFile = next.name !== state.name || next.duration !== state.duration

  Object.assign(state, next)

  // A reply is an instruction (authoritative): jump, do not ease.
  land(next.position, wasAdvancing, state.to, true)

  // Fit only on a new file (or first, unset window); a seek/pause/loop/stop reply keeps the zoom.
  if (state.available && (changedFile || view.to <= view.from)) fit()
}

async function refreshState() {
  adopt(await AudioService.Transport())
  muted.value = await AudioService.Muted()
  volume.value = await AudioService.Volume()
}

export function useTransport() {
  const tuner = useTuner()
  const { readingAt } = tuner
  running = () => tuner.running.value

  // The needle rides its own ~30/s playback event, since a paused/scrubbed file emits no measurements.
  if (!bound) {
    bound = true
    Events.On(EVENT_PLAYBACK, (ev: { data: { position: number, paused: boolean, moving: boolean } }) => {
      if (dragging.value) return // dragging: do not fight the user for the needle

      state.position = ev.data.position
      state.paused = ev.data.paused
      state.moving = ev.data.moving

      land(ev.data.position, advancingNow(), state.to)
    })

    // On stop, Go rewinds the playhead to the audio actually heard (core/audio: dropQueue); refetch, else the needle stays ~0.3s ahead.
    watch(() => tuner.running.value, (isRunning) => {
      if (!isRunning && state.available) void refreshState()
    })
  }

  const refresh = refreshState

  // Speaker-only gain, downstream of the engine; fire-and-forget because a drag fires ~100x and awaiting each would queue ahead of the needle's events. Go clamps to 0..1.
  const setVolume = (next: number) => {
    volume.value = next
    void AudioService.SetVolume(next)
  }

  // Mute stops only the speakers; the file, engine and needle carry on.
  const setMuted = async (next: boolean) => {
    muted.value = await AudioService.SetMuted(next)
  }

  const seek = async (seconds: number) => adopt(await AudioService.Seek(seconds))
  const setPaused = async (paused: boolean) => adopt(await AudioService.SetPaused(paused))

  // Stop is not Pause: it parks at the top of the selection, not under the finger.
  const stop = async () => {
    await setPaused(true)
    await seek(state.from)
  }
  const setLoop = async (loop: boolean) => adopt(await AudioService.SetLoop(loop))
  const select = async (from: number, to: number) => adopt(await AudioService.SetRange(from, to))

  /** True when a sub-range is selected, not the whole file. */
  const hasSelection = computed(
    () => state.available && (state.from > 0 || state.to < state.duration),
  )

  const clearSelection = () => select(0, 0) // an empty range is the whole file, in Go

  // Audio actually moving now: "not paused" is not "playing" (a stopped engine leaves paused false).
  const advancing = computed(() => tuner.running.value
    && state.available && state.moving && !state.paused && !dragging.value,
  )

  /** Where to draw the needle right now; every consumer uses this. */
  const livePosition = () => at(advancing.value, state.to)

  // Where the on-screen reading was measured, or null when there is nothing honest to point at (playing, on a mic, or across a source change).
  const ghostAt = computed<number | null>(() => {
    if (!state.available) return null
    if (advancing.value) return null

    return readingAt.value
  })

  const loadPeaks = async (columns: number) => {
    if (!state.available || columns <= 0) {
      peaks.value = []
      return
    }
    peaks.value = await AudioService.Peaks(view.from, view.to, Math.round(columns))
  }

  return {
    state,
    view,
    peaks,
    dragging,
    hasSelection,
    muted,
    setMuted,
    advancing,
    livePosition,
    ghostAt,
    heard,
    volume,
    setVolume,

    refresh,
    seek,
    setPaused,
    stop,
    setLoop,
    select,
    clearSelection,
    loadPeaks,

    fit,
    fitSelection,
    zoom,
    pan,
    scrollTo,
  }
}
