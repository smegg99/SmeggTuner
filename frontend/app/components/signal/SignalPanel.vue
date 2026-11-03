<!-- app/components/signal/SignalPanel.vue -->
<template>
  <!-- The signal generator, under the input in the tune room: play a note clean with the tones below,
       the reed heard last or one picked with the arrows, to tune the reed against by ear. -->
  <UiPanel
    class="sig"
    :title="t('tuner.signal.title')"
  >
    <!-- A failure takes the reference's place, so it needs no line of its own. -->
    <template #meta>
      <span
        v-if="error"
        class="sig__error"
      >{{ t(error) }}</span>
      <template v-else>
        A4 {{ a4.toFixed(1) }} Hz
      </template>
    </template>

    <div class="sig__body">
      <div class="sig__column">
        <!-- One fixed width for the note, every row and Play, so nothing moves as the note or the
             language changes. -->
        <div class="sig__controls">
          <!-- The note between its two keys; the note's column has a fixed width that fits the longest
               name, so the keys never move. -->
          <div class="sig__note">
            <UiToolKey
              icon="mdi-chevron-left"
              :title="t('tuner.signal.down')"
              @click="step(-1)"
            />
            <span
              class="sig__name"
              :class="{
                'sig__name--off': target === AUTO_NOTE,
                'sig__name--muted': waiting,
                'sig__name--on': playing && !waiting,
              }"
            >{{ target === AUTO_NOTE ? '--' : noteName(target, config.tuner?.scale_naming) }}</span>
            <UiToolKey
              icon="mdi-chevron-right"
              :title="t('tuner.signal.up')"
              @click="step(1)"
            />
          </div>
          <!-- Every tone's pitch on one line of fixed height, clipped, so it moves nothing. -->
          <p class="sig__hz">
            {{ pitches }}
          </p>

          <SignalVoices class="sig__voices" />

          <div class="sig__row">
            <UiToolKey
              class="sig__add"
              :label="t('tuner.signal.add')"
              :title="t('tuner.signal.add')"
              :disabled="cents.length >= MAX_VOICES"
              @click="addVoice"
            />
            <span class="group-rule" />
            <UiToolKey
              :label="t('tuner.signal.wave.sine')"
              :title="t('tuner.signal.wave.hint')"
              :active="!reed"
              @click="setReed(false)"
            />
            <UiToolKey
              :label="t('tuner.signal.wave.reed')"
              :title="t('tuner.signal.wave.hint')"
              :active="reed"
              @click="setReed(true)"
            />
          </div>

          <UiToolKey
            class="sig__key"
            active
            :icon="playing ? 'mdi-stop' : 'mdi-play'"
            :label="t(playing ? 'tuner.signal.stop' : 'tuner.signal.play')"
            :title="t(playing ? 'tuner.signal.stopHint' : 'tuner.signal.playHint')"
            :disabled="starting"
            @click="toggle"
          />
        </div>
      </div>
    </div>
  </UiPanel>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, watch } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'
import { useShell } from '~/composables/useShell'
import { MAX_VOICES, useSignal } from '~/composables/useSignal'
import { AUTO_NOTE, useTuner } from '~/composables/useTuner'
import { noteName } from '~/utils/tuning'

const { t } = useI18n()
const { config } = useConfigSync()
const { view } = useShell()
const { a4, starting } = useTuner()
const { cents, reed, playing, waiting, target, error, hz, stop, toggle, step, addVoice, setReed } = useSignal()

const pitches = computed(() =>
  target.value === AUTO_NOTE ? '' : `${cents.value.map(c => hz(c).toFixed(1)).join(', ')} Hz`)

// The tune room stays mounted behind the others, so leaving it has to silence the tone itself; the
// room switch stops the engine, so it is not started again.
watch(view, (room) => {
  if (room !== 'tune' && playing.value) void stop(false)
})
onBeforeUnmount(() => {
  if (playing.value) void stop()
})
</script>

<style scoped>
.sig { min-height: 0; min-width: 0; }

/* Centred in the room it has, never scrolled: four tones fit the smallest window. */
.sig__body {
  display: flex;
  height: 100%;
  justify-content: center;
  min-height: 0;
  overflow: hidden;
  padding: var(--gap-wide);
}

.sig__column {
  align-items: center;
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-block: auto;
  max-width: 560px;
  text-align: center;
  width: 100%;
}

.sig__note {
  align-items: center;
  align-self: center;
  display: grid;
  gap: var(--gap-wide);
  grid-template-columns: var(--s99ab-control-h) 150px var(--s99ab-control-h);
}

.sig__hz {
  color: rgb(var(--v-theme-ink3));
  font-family: var(--font-mono);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
  height: 17px;
  line-height: 17px;
  margin: -6px 0 0;
  overflow: hidden;
  text-align: center;
  white-space: nowrap;
}

.sig__name {
  color: rgb(var(--v-theme-ink));
  font-family: var(--font-mono);
  font-size: 48px;
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  line-height: 1;
  overflow: hidden;
  text-align: center;
  white-space: nowrap;
}

.sig__name--off { color: rgb(var(--v-theme-ink3)); }
.sig__name--on { color: rgb(var(--v-theme-accent)); }
/* Play is listening for the note to sound. */
.sig__name--muted { opacity: 0.45; }

.sig__error { color: rgb(var(--v-theme-reed)); font-weight: 600; }

.sig__controls {
  align-items: stretch;
  display: flex;
  flex-direction: column;
  gap: 10px;
  /* The note picker's width: its two keys, the note and the gaps between them. */
  width: calc(var(--s99ab-control-h) * 2 + 150px + var(--gap-wide) * 2);
}

.sig__row {
  align-items: center;
  display: flex;
  gap: var(--s99ab-control-gap);
}

.sig__add { flex: 1 1 auto; justify-content: center; }
.sig__key { justify-content: center; }

/* A landscape phone has height for neither stack: the note, its pitches and Play on the left, the
   tones and their keys on the right, each column the note picker's width. */
@media (max-height: 559px) {
  .sig__controls {
    column-gap: 24px;
    display: grid;
    grid-template-areas: "note voices" "hz voices" ". voices" "play row";
    grid-template-columns: repeat(2, calc(var(--s99ab-control-h) * 2 + 150px + var(--gap-wide) * 2));
    grid-template-rows: auto auto minmax(0, 1fr) auto;
    width: auto;
  }
  .sig__note { grid-area: note; justify-self: center; }
  .sig__hz { grid-area: hz; }
  .sig__voices { grid-area: voices; }
  .sig__row { grid-area: row; }
  .sig__key { grid-area: play; }
}

/* The shortest phones give the header's line to the tones, as the tuner's plots do. */
@media (max-height: 399px) {
  .sig :deep(.panel__head) { display: none; }
}
</style>
