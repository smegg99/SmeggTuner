<template>
  <div class="card__meta">
    <span
      class="card__when"
      :title="t('session.card.updated', { when: updatedFull })"
    >{{ updatedShort }}</span>

    <v-icon
      v-if="session.hasCurve"
      class="card__goal"
      icon="mdi-chart-bell-curve-cumulative"
      :aria-label="t('session.card.hasCurve')"
      :title="t('session.card.hasCurve')"
    />

    <span class="card__sp" />

    <span class="card__a4">{{ t('session.card.a4', { hz: session.a4.toFixed(1) }) }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Summary } from '~~bindings/smegg.me/smeggtuner/core/session/models.js'

const props = defineProps<{
  session: Summary
}>()

const { t, d } = useI18n()

const updatedAt = computed(() => new Date(props.session.updated))
const updatedShort = computed(() => d(updatedAt.value, 'short'))
const updatedFull = computed(() => d(updatedAt.value, 'medium'))
</script>

<style scoped>
.card__meta {
  align-items: center;
  border-top: 1px solid rgb(var(--v-theme-lineSoft));
  color: rgb(var(--v-theme-ink3));
  display: flex;
  font-size: 1.35cqh;
  gap: 0.5cqw;
  padding: 0.45cqh 0.7cqw;
}

.card__when {
  font-variant-numeric: tabular-nums;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card__goal {
  color: rgb(var(--v-theme-goal));
  flex: 0 0 auto;
  font-size: 1.7cqh;
}

.card__sp {
  flex: 1 1 auto;
}

.card__a4 {
  flex: 0 0 auto;
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
}
</style>
