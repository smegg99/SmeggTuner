<template>
  <div
    class="note"
    :class="{ 'note--quiet': phase === 'quiet', 'note--play': phase === 'play' }"
  >
    <div
      v-if="phase !== 'normal'"
      class="note__glyph note__glyph--cue"
    >
      <v-icon
        :icon="phase === 'quiet' ? 'mdi-hand-back-right' : 'mdi-play'"
        size="6cqh"
      />
    </div>
    <div
      v-else-if="pitchClass"
      class="note__glyph"
    >
      {{ pitchClass }}<i v-if="octave">{{ octave }}</i>
    </div>

    <div class="note__side">
      <span
        v-if="phase !== 'normal'"
        class="note__cue"
      >{{ phase === 'quiet' ? t('tuner.warmup.quiet') : t('tuner.warmup.play') }}</span>
      <template v-else>
        <span
          v-if="locked"
          class="note__lock"
        >
          <v-icon
            icon="mdi-lock"
            size="small"
          />
          {{ t('tuner.panel.locked') }}
        </span>
        <span
          v-else-if="running"
          class="note__state"
        >
          {{ t(`tuner.state.${state}`) }}
        </span>

        <span
          v-if="scalePitch > 0"
          class="note__hz"
        >{{ scalePitch.toFixed(2) }} Hz</span>
      </template>

      <span class="note__sub">
        A4 {{ a4.toFixed(1) }} Hz, {{ manual ? t('tuner.mode.manual') : t('tuner.mode.auto') }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { AUTO_NOTE, useTuner } from '~/composables/useTuner'
import { useAudioDevices } from '~/composables/useAudioDevices'

const { t } = useI18n()
const { noteName, locked, running, state, scalePitch, a4, manualNote } = useTuner()
const { current } = useAudioDevices()

// File source skips the warm-up cue (it's for the live mic only).
const isFile = computed(() => current.value.kind === 'file')

const manual = computed(() => manualNote.value !== AUTO_NOTE)

// Octave is the trailing digits of the note name, split off to style it small.
const pitchClass = computed(() => noteName.value.replace(/-?\d+$/, ''))
const octave = computed(() => noteName.value.match(/-?\d+$/)?.[0] ?? '')

// Warm-up phase maps off engine state: 'quiet'='initializing', 'play'=brief flash after.
const PLAY_MS = 1000
const playing = ref(false)
let playTimer: ReturnType<typeof setTimeout> | null = null

function clearPlayTimer() {
  if (playTimer !== null) {
    clearTimeout(playTimer)
    playTimer = null
  }
}

watch(state, (now, prev) => {
  if (isFile.value) return
  // Leaving warm-up flashes "now play" unless a note is already shown; any other transition drops it.
  if (prev === 'initializing' && now === 'running') {
    if (pitchClass.value) return
    clearPlayTimer()
    playing.value = true
    playTimer = setTimeout(() => {
      playing.value = false
      playTimer = null
    }, PLAY_MS)
    return
  }
  if (now !== 'running') {
    clearPlayTimer()
    playing.value = false
  }
})

onBeforeUnmount(clearPlayTimer)

const phase = computed<'quiet' | 'play' | 'normal'>(() => {
  if (isFile.value) return 'normal'
  if (state.value === 'initializing') return 'quiet'
  if (playing.value) return 'play'
  return 'normal'
})
</script>

<style scoped>
.note {
  align-items: center;
  background: rgb(var(--v-theme-chrome2));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 0.45cqh;
  display: flex;
  gap: 1.3cqw;
  min-width: 0;
  padding: 0 1.4cqw;
}

.note--quiet {
  background: rgba(var(--v-theme-warn), 0.14);
  border-color: rgb(var(--v-theme-warn));
}

.note--play {
  background: rgba(var(--v-theme-goal), 0.14);
  border-color: rgb(var(--v-theme-goal));
}

.note__glyph--cue {
  color: rgb(var(--v-theme-warn));
}

.note--play .note__glyph--cue {
  color: rgb(var(--v-theme-goal));
}

.note__cue {
  color: rgb(var(--v-theme-warn));
  font-size: 1.9cqh;
  font-weight: 700;
}

.note--play .note__cue {
  color: rgb(var(--v-theme-goal));
}

.note__glyph {
  align-items: baseline;
  color: rgb(var(--v-theme-ink));
  display: flex;
  flex: 0 0 auto;
  font-size: 9.2cqh;
  font-weight: 300;
  letter-spacing: -0.03em;
  line-height: 1;
}

.note__glyph i {
  color: rgb(var(--v-theme-ink3));
  font-size: 0.38em;
  font-style: normal;
  font-weight: 400;
  letter-spacing: 0;
}

.note__side {
  display: flex;
  flex-direction: column;
  gap: 0.5cqh;
  min-width: 0;
}

.note__lock {
  align-items: center;
  color: rgb(var(--v-theme-ink));
  display: flex;
  font-size: 1.75cqh;
  font-weight: 700;
  gap: 0.4cqw;
}

.note__state {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.65cqh;
}

.note__hz {
  color: rgb(var(--v-theme-ink));
  font-family: var(--font-mono);
  font-size: 2.35cqh;
  font-variant-numeric: tabular-nums;
}

.note__sub {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.65cqh;
}
</style>
