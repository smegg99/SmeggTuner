<template>
  <span class="label">
    <span
      v-if="progress.kind === 'none'"
      class="label__none"
    >{{ t('session.footer.none') }}</span>

    <template v-else>
      <!-- Name shows whenever a session is open, not only while recording. -->
      <span class="label__name">{{ progress.name }}</span>

      <template v-if="progress.kind === 'active'">
        <span class="label__prog">{{ coverage }}</span>
        <span
          v-if="inTune"
          class="label__prog"
        >{{ inTune }}</span>
      </template>
    </template>
  </span>
</template>

<script setup lang="ts">
// Dumb view: useSessionProgress does the counting, this only reads the model.
import { computed } from 'vue'
import type { SessionProgress } from '~/utils/sessionProgress'

const props = defineProps<{ progress: SessionProgress }>()

const { t } = useI18n()

// Plain count when the keyboard size is unknown (total null).
const coverage = computed(() => {
  if (props.progress.kind !== 'active') return ''
  const { done, total } = props.progress
  return total === null
    ? t('session.readings.notes', done)
    : t('session.footer.coverage', { done, total })
})

// In-tune count; denominator is done (the coverage numerator).
const inTune = computed(() => {
  if (props.progress.kind !== 'active' || props.progress.inTune === null) return ''
  return t('session.footer.inTune', { done: props.progress.inTune, total: props.progress.done })
})
</script>

<style scoped>
.label {
  align-items: baseline;
  display: flex;
  flex: 0 1 auto;
  gap: 0.9cqw;
  min-width: 0;
  /* When too narrow, the name truncates and figures clip here rather than spilling over the keys. */
  overflow: hidden;
}

.label__name {
  color: rgb(var(--v-theme-ink2));
  font-size: 1.6cqh;
  font-weight: 600;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.label__prog {
  color: rgb(var(--v-theme-ink3));
  flex: 0 0 auto;
  font-family: var(--font-mono);
  font-size: 1.4cqh;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.label__none {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.6cqh;
}
</style>
