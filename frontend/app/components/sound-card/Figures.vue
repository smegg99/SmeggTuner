<!-- app/components/sound-card/Figures.vue -->
<template>
  <!-- Always shown: the clock's figures above and the device's below the offset trace. -->
  <div class="sc__stats">
    <dl
      v-for="(row, r) in [stats.slice(0, 2), stats.slice(2)]"
      :key="r"
      class="sc__row"
    >
      <div
        v-for="stat in row"
        :key="stat.key"
        class="sc__stat"
      >
        <dt>{{ t(`settings.calibration.screen.stat.${stat.key}`) }}</dt>
        <dd>{{ stat.value }}</dd>
      </div>
    </dl>
    <SoundCardClockTrace
      class="sc__trace"
      :points="points"
    />
  </div>
</template>

<script setup lang="ts">
import type { TracePoint } from '~/composables/useSoundCard'

defineProps<{ stats: { key: string, value: string }[], points: TracePoint[] }>()
const { t } = useI18n()
</script>

<style scoped>
.sc__stats {
  background: rgb(var(--v-theme-chrome2));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: var(--app-radius-panel);
  display: grid;
  gap: 4px;
  grid-template-rows: auto minmax(0, 1fr) auto;
  height: 112px;
  max-width: 100%;
  padding: 6px 12px;
  width: 440px;
}

.sc__row { display: flex; gap: var(--gap-wide); justify-content: space-between; margin: 0; min-width: 0; }
.sc__row:last-of-type { grid-row: 3; }
.sc__trace { grid-row: 2; }
.sc__stat { align-items: baseline; display: flex; gap: 6px; min-width: 0; }

.sc__stat dt {
  color: rgb(var(--v-theme-ink3));
  font-size: 10.5px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  white-space: nowrap;
}

.sc__stat dd {
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
