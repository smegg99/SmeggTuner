<template>
  <div class="cal__steps">
    <span
      class="cal__step"
      :class="{ 'cal__step--on': stage === 'range' }"
    >
      <span class="cal__stepno">1</span>
      <span class="cal__steplabel">{{ t('calibrate.range.step') }}</span>
    </span>
    <span class="cal__stepsep" />
    <span
      class="cal__step"
      :class="{ 'cal__step--on': stage !== 'range' }"
    >
      <span class="cal__stepno">2</span>
      <span class="cal__steplabel">{{ t('calibrate.sweep.step') }}</span>
    </span>
  </div>
</template>

<script setup lang="ts">
import type { Stage } from '~/composables/useCalibration'

defineProps<{ stage: Stage }>()

const { t } = useI18n()
</script>

<style scoped>
.cal__steps {
  align-items: center;
  display: flex;
  gap: 0.8cqw;
}

.cal__step {
  align-items: center;
  display: flex;
  gap: 0.7cqw;
  opacity: 0.4;
}

.cal__step--on {
  opacity: 1;
}

.cal__stepsep {
  background: rgb(var(--v-theme-line));
  flex: 0 0 3cqw;
  height: 1px;
}

/* flex: 0 0 keeps it square; a normal flex child would shrink in width only. */
.cal__stepno {
  align-items: center;
  aspect-ratio: 1;
  background: rgb(var(--v-theme-ink2));
  border-radius: 50%;
  color: rgb(var(--v-theme-bg));
  display: flex;
  flex: 0 0 auto;
  font-size: 1.6cqh;
  font-weight: 700;
  height: 3cqh;
  justify-content: center;
  width: 3cqh;
}

.cal__steplabel {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.4cqh;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}
</style>
