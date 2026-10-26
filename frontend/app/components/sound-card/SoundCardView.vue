<!-- app/components/sound-card/SoundCardView.vue -->
<template>
  <UiPanel
    class="sc tuner-panel"
    flush
  >
    <!-- Untitled: the main bar names the room and the device. Every row below keeps its place in
         every state, so nothing moves between the steps; what does not apply is only hidden. -->
    <div class="sc__body">
      <div class="sc__column">
        <SoundCardSteps :stage="stage" />

        <!-- One tight group between the stepper and the keys, centred in the space they leave. -->
        <div class="sc__middle">
          <!-- The two balls meet and part while the clock is measured; the outcome replaces them. -->
          <div class="sc__figure">
            <UiGoo
              v-if="running"
              :size="200"
            />
            <v-icon
              v-else
              :icon="status === 'calibrated' ? 'mdi-check-circle-outline' : 'mdi-clock-alert-outline'"
              :style="{ color: `rgb(var(--v-theme-${tone}))` }"
              size="72"
            />
          </div>

          <div class="sc__text">
            <p
              class="sc__headline"
              :style="{ color: `rgb(var(--v-theme-${tone}))` }"
            >
              {{ headline }}
            </p>
            <p class="sc__lead">
              {{ lead }}
            </p>
          </div>

          <SoundCardMeter :calibration="calibration" />
          <SoundCardFigures
            :stats="stats"
            :points="trace"
          />

          <p class="sc__error">
            {{ error ? t(error) : '' }}
          </p>
        </div>

        <!-- Two keys, always in these two places: Return or Cancel, and the calibration itself. -->
        <div class="sc__actions">
          <UiToolKey
            v-if="running"
            class="sc__key"
            icon="mdi-close"
            :label="t('settings.calibration.cancel')"
            :title="t('settings.calibration.cancel')"
            @click="cancelCalibration"
          />
          <UiToolKey
            v-else
            class="sc__key"
            icon="mdi-arrow-left"
            :label="t('settings.calibration.screen.return')"
            :title="t('settings.calibration.screen.return')"
            @click="setView(previous === 'soundcard' ? 'tune' : previous)"
          />
          <UiToolKey
            class="sc__key"
            active
            :label="status === 'uncalibrated' || running ? t('settings.calibration.start') : t('settings.calibration.screen.again')"
            :title="t('settings.calibration.startHint')"
            :disabled="!calibration || running"
            @click="startCalibration"
          />
        </div>
      </div>
    </div>
  </UiPanel>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, watchEffect } from 'vue'
import { useAudioDevices } from '~/composables/useAudioDevices'
import { useCalibrationCopy } from '~/composables/useCalibrationCopy'
import { useDeviationUnit } from '~/composables/useDeviationUnit'
import { useShell } from '~/composables/useShell'
import { useSoundCard } from '~/composables/useSoundCard'
import { calibrationTone } from '~/utils/soundCard'

// The sound card's own calibration room. A session keeps running if the view is left; the toolbar
// key and the settings line still show it.
const { t } = useI18n()
const { setView, previous, pageMeta } = useShell()
const { current } = useAudioDevices()
const { calibration, trace, error, startCalibration, cancelCalibration } = useSoundCard()
const { status, stage, headline, lead, stats } = useCalibrationCopy(calibration, trace, useDeviationUnit())

const running = computed(() => calibration.value?.running ?? false)
const tone = computed(() => calibrationTone(status.value))
const device = computed(() => (current.value.kind === 'mic' && current.value.name) || t('tuner.source.default'))
watchEffect(() => {
  pageMeta.value = device.value
})
onBeforeUnmount(() => {
  pageMeta.value = ''
})
</script>

<style scoped>
.sc { min-height: 0; min-width: 0; }

.sc__body {
  display: flex;
  height: 100%;
  justify-content: center;
  min-height: 0;
  overflow-y: auto;
  padding: var(--gap-wide);
}

/* The stepper and the keys hold the ends of a fixed height; the rest is centred between them. Auto
   margins centre the column, and degrade to scrolling from the top in a window shorter than it. */
.sc__column {
  display: grid;
  flex: 0 0 auto;
  grid-template-rows: 24px minmax(0, 1fr) 30px;
  height: 500px;
  justify-items: center;
  margin-block: auto;
  max-width: 560px;
  text-align: center;
  width: 100%;
}

.sc__middle {
  align-items: center;
  display: flex;
  flex-direction: column;
  gap: 12px;
  justify-content: center;
  position: relative;
  width: 100%;
}

.sc__middle > * { flex: 0 0 auto; }
/* One line each, so every state keeps the group's height; the texts are written to fit. */
.sc__text { display: flex; flex-direction: column; gap: 4px; max-width: 100%; }
.sc__text p { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.sc__figure {
  --s99ab-state-kind: var(--v-theme-accent);
  align-items: center;
  display: flex;
  height: 96px;
  justify-content: center;
}

.sc__headline { font-size: 16px; font-weight: 700; line-height: 22px; margin: 0; }
.sc__lead { color: rgb(var(--v-theme-ink2)); font-size: 12.5px; line-height: 19px; margin: 0; }
/* Out of the group's centring, just above the keys, so an empty one weighs nothing. */
.sc__error { bottom: 8px; color: rgb(var(--v-theme-reed)); font-size: 12px; inset-inline: 0; margin: 0; position: absolute; }

.sc__actions { display: grid; gap: var(--gap); grid-template-columns: 140px 140px; }
.sc__key { justify-content: center; width: 100%; }
</style>
