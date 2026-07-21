<template>
  <div class="rng">
    <template v-if="phase !== 'done'">
      <p class="rng__prompt">
        {{ t(phase === 'low' ? 'calibrate.range.playLow' : 'calibrate.range.playHigh') }}
      </p>

      <CalibrationReadout
        :note="readoutNote ? noteName(readoutNote) : '--'"
        :sure="!!heard"
        :status="status"
        :progress="lockProgress"
        box-width="22cqh"
        box-height="22cqh"
        note-size="7cqh"
      />

      <div class="rng__keys">
        <UiToolKey
          v-if="phase === 'high'"
          icon="mdi-arrow-left"
          :label="t('common.back')"
          @click="backRange"
        />
        <UiToolKey
          icon="mdi-check"
          :label="t('calibrate.range.capture')"
          :disabled="!heard"
          @click="capture"
        />
      </div>
    </template>

    <template v-else>
      <div class="rng__range">
        <span class="rng__end">
          <span class="rng__endlabel">{{ t('calibrate.range.lowest') }}</span>
          <span class="rng__endnote">{{ noteName(lo) }}</span>
        </span>
        <v-icon
          icon="mdi-arrow-right"
          class="rng__arrow"
        />
        <span class="rng__end">
          <span class="rng__endlabel">{{ t('calibrate.range.highest') }}</span>
          <span class="rng__endnote">{{ noteName(hi) }}</span>
        </span>
      </div>

      <p
        v-if="backwards"
        class="rng__warn"
      >
        {{ t('calibrate.range.backwards') }}
      </p>

      <div class="rng__keys">
        <UiToolKey
          icon="mdi-refresh"
          :label="t('calibrate.range.again')"
          @click="reset"
        />
        <UiToolKey
          icon="mdi-arrow-right"
          :label="t('calibrate.range.next')"
          :disabled="backwards || saving"
          @click="onNext"
        />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import CalibrationReadout from './CalibrationReadout.vue'
import { useCalibration } from '~/composables/useCalibration'
import { useConfigSync } from '~/composables/useConfigSync'
import { noteName as toNoteName } from '~/utils/tuning'

const { t } = useI18n()
const { config } = useConfigSync()
const { phase, lo, hi, heard, shown, liveNote, status, lockProgress, backwards, saving, reset, capture, backRange, saveRange, toSetup } = useCalibration()

const readoutNote = computed(() => heard.value || liveNote.value || shown.value)

function noteName(n: number): string {
  return toNoteName(n, config.tuner?.scale_naming)
}

async function onNext() {
  if (await saveRange()) toSetup()
}
</script>

<style scoped>
.rng {
  align-items: center;
  display: flex;
  flex-direction: column;
  gap: 2cqh;
  justify-content: center;
}

.rng__prompt {
  color: rgb(var(--v-theme-ink));
  font-size: 2.4cqh;
  margin: 0;
  max-width: 40ch;
}

.rng__keys {
  display: flex;
  gap: 0.6cqw;
}

.rng__range {
  align-items: center;
  display: flex;
  gap: 1.5cqw;
}

.rng__end {
  display: flex;
  flex-direction: column;
  gap: 0.4cqh;
}

.rng__endlabel {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.4cqh;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.rng__endnote {
  color: rgb(var(--v-theme-ink));
  font-family: var(--font-mono);
  font-size: 4cqh;
  font-variant-numeric: tabular-nums;
  font-weight: 600;
}

.rng__arrow {
  color: rgb(var(--v-theme-ink3));
  font-size: 3cqh;
}

.rng__warn {
  color: rgb(var(--v-theme-warn));
  font-size: 1.6cqh;
  margin: 0;
}
</style>
