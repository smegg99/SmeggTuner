<!-- Box and note are sized by the parent through --cr-w / --cr-h / --cr-note rather than by
     props: an inline style cannot be overridden, and a short window has to shrink this. -->
<template>
  <div class="cr__note-wrap">
    <output
      class="cr__note"
      :class="{ 'cr__note--sure': sure }"
    >{{ note }}</output>
  </div>

  <p class="cr__status">
    {{ statusText }}
  </p>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { CaptureStatus } from '~/composables/calibrationStatus'

const props = defineProps<{
  note: string
  sure: boolean
  status: CaptureStatus
}>()

const { t } = useI18n()

const statusText = computed(() => t(`calibrate.status.${props.status === 'idle' ? 'play' : props.status}`))
</script>

<style scoped>
.cr__note-wrap {
  align-items: center;
  block-size: var(--cr-h);
  inline-size: var(--cr-w);
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 4.8px;
  display: flex;
  justify-content: center;
  transition: border-color 120ms;
}

.cr__note {
  color: rgb(var(--v-theme-ink3));
  font-family: var(--font-mono);
  font-size: var(--cr-note);
  font-variant-numeric: tabular-nums;
  font-weight: 600;
}

.cr__note--sure {
  color: rgb(var(--v-theme-ink));
}

.cr__status {
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
  margin: 0;
  min-height: 11px;
}
</style>
