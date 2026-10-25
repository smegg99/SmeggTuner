<template>
  <UiPanelCanvas
    ref="panel"
    :title="t('tuner.notes.title')"
    :label="t('tuner.notes.label')"
    :renderer="renderer"
    interactive
    @pointerdown="onPress"
    @pointermove="onMove"
    @pointerleave="onLeave"
    @pointerup="onRelease"
    @pointercancel="onRelease"
  >
    <template #meta>
      <span v-if="hovered">
        <b>{{ hovered }}</b>, {{ t('tuner.notes.clickToPin') }}
      </span>
      <span v-else>
        {{ t('tuner.notes.range') }}, {{ t('tuner.notes.tracking') }}
        <b>{{ noteName || '--' }}</b>
      </span>
    </template>
  </UiPanelCanvas>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, useTemplateRef } from 'vue'
import { bandAt, useNotesRender } from '~/composables/render/useNotesRender'
import { useConfigSync } from '~/composables/useConfigSync'
import { useOutputLevel } from '~/composables/useOutputLevel'
import { NOTE_MIN, useTuner } from '~/composables/useTuner'
import { noteName as nameOf } from '~/utils/tuning'

const { t } = useI18n()
const { config } = useConfigSync()
const { noteName, setManualNote, mode, sessionReeds, playTone, stopTone } = useTuner()
const { muted } = useOutputLevel()

// A press sounds its note unless the output is muted; muted, no tone is asked for at all, so the
// recorder is not held off by a tone nobody hears.
function press(note: number) {
  if (!muted.value) void playTone(note)
}

function release() {
  void stopTone()
}

const panel = useTemplateRef('panel')
const hover = ref(-1)

// Band under a held finger, or -1.
const held = ref(-1)

const renderer = useNotesRender(hover, held)

// The canvas host from the press. A drag's moves come from a window listener (trackPointer), whose
// currentTarget is the window, so they are measured against this.
let surface: HTMLElement | null = null

const hovered = computed(() => (hover.value < 0
  ? ''
  : nameOf(hover.value + NOTE_MIN, config.tuner?.scale_naming)))

function bandOf(event: PointerEvent | MouseEvent): number {
  const host = event.currentTarget instanceof HTMLElement ? event.currentTarget : surface
  if (!host) return -1

  return bandAt(event.clientX, host.getBoundingClientRect())
}

function onMove(event: PointerEvent) {
  const band = bandOf(event)

  // Dragging with the finger down sweeps the pinned note and tone; TonePlayer retargets a
  // running tone without a restart gap (core/audio) so a sweep glissandos.
  if (held.value >= 0 && band >= 0 && band !== held.value) {
    held.value = band

    const note = band + NOTE_MIN
    if (!chords.value) void setManualNote(note)
    press(note)
    panel.value?.paint()
  }

  if (band === hover.value) return

  hover.value = band
  panel.value?.paint()
}

function onLeave() {
  hover.value = -1

  // A pointer leaving mid-hold never fires release, so stop the tone here or it drones into the mic.
  onRelease()
  panel.value?.paint()
}

// Chords picks its own notes; services/tuner refuses a pin there, so a press only sounds the tone.
const chords = computed(() => !sessionReeds.value && mode.value === 'chords')

// pointerdown, not click: a tone heard on release arrives too late.
function onPress(event: PointerEvent) {
  if (event.currentTarget instanceof HTMLElement) surface = event.currentTarget
  const band = bandOf(event)
  if (band < 0) return

  held.value = band
  panel.value?.paint()
  const note = band + NOTE_MIN

  // Pinning is a manual selection; the detector no longer owns the note.
  if (!chords.value) void setManualNote(note)
  press(note)
}

// Release, cancel, or leave - each ends the tone.
function onRelease() {
  const wasHeld = held.value >= 0
  held.value = -1
  release()
  if (wasHeld) panel.value?.paint()
}

// Window losing focus (alt-tab) with a band held never releases the pointer, so stop the tone on blur.
onMounted(() => window.addEventListener('blur', onRelease))
onBeforeUnmount(() => {
  window.removeEventListener('blur', onRelease)
  onRelease()
})
</script>
