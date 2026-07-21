<!-- Box/note/bar dimensions come in as props so each parent keeps its own sizing. -->
<template>
  <div
    class="cr__note-wrap"
    :style="{ width: boxWidth, height: boxHeight }"
  >
    <output
      class="cr__note"
      :class="{ 'cr__note--sure': sure }"
      :style="{ fontSize: noteSize }"
    >{{ note }}</output>
  </div>

  <p class="cr__status">
    {{ statusText }}
  </p>
  <div
    class="cr__bar"
    :style="{ width: boxWidth }"
  >
    <div
      class="cr__bar-fill"
      :style="{ width: `${barPct}%` }"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { CaptureStatus } from '~/composables/calibrationStatus'

const props = defineProps<{
  note: string
  sure: boolean
  status: CaptureStatus
  progress: number
  boxWidth: string
  boxHeight: string
  noteSize: string
}>()

const { t } = useI18n()

const statusText = computed(() => t(`calibrate.status.${props.status === 'idle' ? 'play' : props.status}`))
// The bar only means anything mid-hold; empty otherwise.
const barPct = computed(() => (props.status === 'holding' ? Math.round(props.progress * 100) : 0))
</script>

<style scoped>
.cr__note-wrap {
  align-items: center;
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 0.6cqh;
  display: flex;
  justify-content: center;
  transition: border-color 120ms;
}

.cr__note {
  color: rgb(var(--v-theme-ink3));
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
  font-weight: 600;
}

.cr__note--sure {
  color: rgb(var(--v-theme-ink));
}

.cr__status {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.6cqh;
  margin: 0;
  min-height: 1.6cqh;
}

.cr__bar {
  background: rgb(var(--v-theme-well));
  border-radius: 999px;
  height: 0.7cqh;
  overflow: hidden;
}

.cr__bar-fill {
  background: rgb(var(--v-theme-ink2));
  height: 100%;
  transition: width 160ms;
}
</style>
