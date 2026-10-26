<!-- app/components/sound-card/Steps.vue -->
<template>
  <!-- The three steps of a calibration; stage 0 is a card never measured, so none is lit. -->
  <div class="sc__steps">
    <template
      v-for="(step, i) in STEPS"
      :key="step"
    >
      <span
        v-if="i > 0"
        class="sc__stepsep"
      />
      <span
        class="sc__step"
        :class="{ 'sc__step--on': i + 1 === stage, 'sc__step--past': i + 1 < stage }"
      >
        <span class="sc__stepno">{{ i + 1 }}</span>
        <span class="sc__steplabel">{{ t(`settings.calibration.screen.steps.${step}`) }}</span>
      </span>
    </template>
  </div>
</template>

<script setup lang="ts">
const STEPS = ['prepare', 'measure', 'done'] as const

defineProps<{ stage: number }>()
const { t } = useI18n()
</script>

<style scoped>
.sc__steps { align-items: center; display: flex; gap: var(--gap); }
.sc__step { align-items: center; display: flex; flex: 0 0 auto; gap: var(--gap); opacity: 0.4; }
.sc__step--on, .sc__step--past { opacity: 1; }
.sc__stepsep { background: rgb(var(--v-theme-line)); flex: 0 0 29px; height: 1px; }

.sc__stepno {
  align-items: center;
  aspect-ratio: 1;
  background: rgb(var(--v-theme-ink2));
  border-radius: 50%;
  color: rgb(var(--v-theme-bg));
  display: flex;
  flex: 0 0 auto;
  font-size: 11px;
  font-weight: 700;
  height: 21px;
  justify-content: center;
  width: 21px;
}
.sc__step--on .sc__stepno { background: rgb(var(--v-theme-accent)); }

.sc__steplabel {
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  white-space: nowrap;
}
</style>
