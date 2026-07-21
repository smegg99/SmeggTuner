<template>
  <svg
    class="sym"
    viewBox="0 0 40 40"
    aria-hidden="true"
  >
    <circle
      class="sym__ring"
      cx="20"
      cy="20"
      r="18"
    />

    <!-- Two lines divide the circle into three bands: 16' top, 8' middle, 4' bottom. -->
    <line
      class="sym__band"
      x1="4"
      y1="14"
      x2="36"
      y2="14"
    />
    <line
      class="sym__band"
      x1="4"
      y1="26"
      x2="36"
      y2="26"
    />

    <!-- 16', top band -->
    <circle
      v-if="sym.low"
      class="sym__dot"
      cx="20"
      cy="8.5"
      r="2.3"
    />

    <!-- 8', middle band: one to four, spread across the row -->
    <circle
      v-for="x in middleDots"
      :key="x"
      class="sym__dot"
      :cx="x"
      cy="20"
      r="2.3"
    />

    <!-- 4', bottom band -->
    <circle
      v-if="sym.high"
      class="sym__dot"
      cx="20"
      cy="31.5"
      r="2.3"
    />
  </svg>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { symbolOf } from '~/utils/feet'
import type { Bank } from '~/types/session'

// Read-only register symbol: three bands (16'/8'/4'), a dot per rank. RegisterBuilder draws the same thing but interactive.
const props = defineProps<{ banks: readonly Bank[] }>()

const sym = computed(() => symbolOf(props.banks))

// Middle-band dot positions, centred by count (1-4).
const middleDots = computed<number[]>(() => {
  switch (sym.value.middle) {
    case 1: return [20]
    case 2: return [14, 26]
    case 3: return [11, 20, 29]
    case 4: return [9, 16, 24, 31]
    default: return []
  }
})
</script>

<style scoped>
.sym {
  display: block;
  height: 100%;
  width: 100%;
}

.sym__ring {
  fill: none;
  stroke: rgb(var(--v-theme-ink3));
  stroke-width: 1.5;
}

.sym__band {
  stroke: rgb(var(--v-theme-lineSoft));
  stroke-width: 1;
}

.sym__dot {
  fill: rgb(var(--v-theme-ink));
}
</style>
