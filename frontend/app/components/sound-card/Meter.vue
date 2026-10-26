<!-- app/components/sound-card/Meter.vue -->
<template>
  <!-- Two halves, as the estimator works: the reading settles, then it is held for 20 s. Kept in
       place, unseen, while no calibration runs. -->
  <div
    class="sc__meter"
    :class="{ 'sc__meter--idle': !calibration?.running }"
  >
    <div class="sc__track">
      <div
        v-for="half in halves"
        :key="half.key"
        class="sc__half"
        :class="{ 'sc__half--on': half.on }"
      >
        <div class="sc__rail">
          <div
            class="sc__fill"
            :style="{ width: `${half.fill * 100}%` }"
          />
        </div>
        <span class="sc__halflabel">{{ t(`settings.calibration.screen.bar.${half.key}`) }}</span>
      </div>
    </div>
    <span class="sc__percent">{{ calibration?.progress ?? 0 }} %</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { CalibrationDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/audio/models.js'

const props = defineProps<{ calibration: CalibrationDTO | null }>()
const { t } = useI18n()

// 0..50 the reading settles, 50..100 it is held still.
const halves = computed(() => {
  const p = props.calibration?.progress ?? 0
  const settled = props.calibration?.settled ?? false
  return [
    { key: 'settling', fill: Math.min(p, 50) / 50, on: props.calibration?.status === 'calibrating' && !settled },
    { key: 'confirming', fill: Math.max(0, p - 50) / 50, on: settled },
  ]
})
</script>

<style scoped>
.sc__meter { align-items: flex-start; display: flex; gap: var(--gap-wide); width: 100%; }
.sc__meter--idle { visibility: hidden; }
.sc__track { display: flex; flex: 1 1 auto; gap: 4px; min-width: 0; }
.sc__half { display: flex; flex: 1 1 0; flex-direction: column; gap: 6px; min-width: 0; }

.sc__rail {
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  height: 10px;
  overflow: hidden;
}

.sc__fill {
  background: rgb(var(--v-theme-ink3));
  height: 100%;
  transition: width 0.4s ease-out;
}
.sc__half--on .sc__fill { background: rgb(var(--v-theme-accent)); }

.sc__halflabel {
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-align: start;
  text-transform: uppercase;
  white-space: nowrap;
}
.sc__half--on .sc__halflabel { color: rgb(var(--v-theme-accent)); }

.sc__percent {
  font-size: 14px;
  font-variant-numeric: tabular-nums;
  font-weight: 700;
  line-height: 12px;
  min-width: 4ch;
  text-align: end;
}
</style>
