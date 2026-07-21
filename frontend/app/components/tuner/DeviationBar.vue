<template>
  <div class="dev">
    <span class="dev__axis" />
    <span class="dev__tol dev__tol--low" />
    <span class="dev__tol dev__tol--high" />
    <span class="dev__zero" />

    <span
      v-if="frac !== null"
      class="dev__bar"
      :class="{ 'dev__bar--out': out }"
      :style="geometry"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

// Notches sit at 37.5% and 62.5% because the track is 4 tolerances wide (utils/boxes.ts),
// which lets cents and hertz read on the same scale.
const props = defineProps<{
  // 0..1 along the track; null means no reading (not zero, which is a real reading at the flat end).
  frac: number | null
  out: boolean
}>()

const geometry = computed(() => {
  const frac = props.frac ?? 0.5

  return {
    left: `${Math.min(0.5, frac) * 100}%`,
    width: `${Math.abs(frac - 0.5) * 100}%`,
  }
})
</script>

<style scoped>
.dev {
  height: 2.6cqh;
  margin-top: auto;
  position: relative;
}

.dev__axis {
  background: rgba(var(--v-theme-neutral), 0.22);
  height: 1px;
  left: 0;
  position: absolute;
  right: 0;
  top: 50%;
}

.dev__zero {
  background: rgba(var(--v-theme-neutral), 0.5);
  bottom: 0;
  left: 50%;
  position: absolute;
  top: 0;
  transform: translateX(-50%);
  width: 1px;
}

.dev__tol {
  background: rgba(var(--v-theme-neutral), 0.35);
  bottom: 22%;
  position: absolute;
  top: 22%;
  transform: translateX(-50%);
  width: 1px;
}

.dev__tol--low { left: 37.5%; }
.dev__tol--high { left: 62.5%; }

.dev__bar {
  background: rgb(var(--v-theme-ink));
  bottom: 28%;
  position: absolute;
  top: 28%;
}

.dev__bar--out {
  background: rgb(var(--v-theme-warn));
}
</style>
