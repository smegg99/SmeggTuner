<template>
  <div
    class="box"
    :class="{ 'box--beat': box.kind === 'beat', 'box--out': box.out }"
  >
    <div class="box__top">
      <span class="box__label">{{ label }}</span>
      <span class="box__unit">{{ box.unit }}</span>
    </div>

    <!-- "merged": the pair could not be split, so no number is shown (a wrong reed is worse
         than none). "notHeard"/"harmonicOnly": a declared rank the engine did not hear - the
         latter is the blocked-rank shape, where only the lower rank's harmonic stands in the
         band. "idle": nothing sounding. Either way the box keeps its shape. -->
    <div
      class="box__value"
      :class="{ 'box__value--none': box.value === null }"
    >
      <template v-if="box.value !== null">
        {{ signed(box.value) }}
      </template>
      <template v-else-if="box.blank === 'merged'">
        {{ t('tuner.boxes.notSplit') }}
      </template>
      <template v-else-if="box.blank === 'notHeard'">
        {{ t('tuner.boxes.notHeard') }}
      </template>
      <template v-else-if="box.blank === 'harmonicOnly'">
        {{ t('tuner.boxes.harmonicOnly') }}
      </template>
    </div>

    <TunerDeviationBar
      :frac="box.frac"
      :out="box.out"
    />

    <div class="box__foot">
      <template v-if="box.goal !== null">
        {{ t('tuner.boxes.goal') }}
        <b>{{ signed(box.goal) }}</b>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Box } from '~/utils/boxes'

const props = defineProps<{ box: Box }>()

const { t } = useI18n()

// Label text is composed here, not in the model: it's UI language (Polish-first). A box standing
// for a named rank is labeled by it - the treble side's bank (the card's column) or the bass
// side's foot.
const label = computed(() => {
  if (props.box.kind === 'reed') {
    return props.box.rank
      ? t('tuner.boxes.reedBank', { bank: props.box.rank })
      : t('tuner.boxes.reed', { n: props.box.index + 1 })
  }
  return t('tuner.boxes.beat', { a: props.box.index, b: props.box.index + 1 })
})

// Formatting only; the value comes from the DTO.
function signed(value: number | null): string {
  if (value === null) return ''
  return value > 0 ? `+${value.toFixed(1)}` : value.toFixed(1)
}
</script>

<style scoped>
.box {
  background: rgb(var(--v-theme-chrome2));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 0.45cqh;
  display: flex;
  flex: 1 1 0;
  flex-direction: column;
  gap: 0.4cqh;
  min-width: 0;
  padding: 0.65cqh 1.1cqw 0.65cqh;
}

.box--beat {
  background: rgb(var(--v-theme-sunk));
}

/* Rows use fixed heights, not min-heights, so the box can't grow when a goal appears mid-use. */
.box__top {
  align-items: baseline;
  display: flex;
  flex: 0 0 auto;
  gap: 0.5cqw;
  height: 1.9cqh;
}

.box__label {
  color: rgb(var(--v-theme-ink2));
  font-size: 1.55cqh;
  font-weight: 600;
}

.box__unit {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.4cqh;
  margin-left: auto;
}

.box__value {
  align-items: center;
  color: rgb(var(--v-theme-ink));
  display: flex;
  flex: 0 0 auto;
  font-family: var(--font-mono);
  font-size: 3.5cqh;
  font-variant-numeric: tabular-nums;
  font-weight: 500;
  /* Digit height reserved so the row doesn't jump when a note starts. */
  height: 3.5cqh;
  letter-spacing: -0.03em;
  line-height: 1;
  overflow: hidden;
}

.box--out .box__value {
  color: rgb(var(--v-theme-warn));
}

/* The "no number" text borrows the digit's reserved height so the box stays one size. */
.box__value--none {
  color: rgb(var(--v-theme-ink3));
  font-family: var(--font-sans);
  font-size: 1.65cqh;
  font-weight: 400;
  letter-spacing: 0;
  line-height: 1.25;
}

.box__foot {
  align-items: baseline;
  color: rgb(var(--v-theme-ink3));
  display: flex;
  flex: 0 0 auto;
  font-size: 1.35cqh;
  gap: 0.4cqw;
  /* Fixed, not a minimum: the 1.75cqh <b> would otherwise push past a floor and grow the card. */
  height: 2.05cqh;
  overflow: hidden;
}

.box__foot b {
  color: rgb(var(--v-theme-goal));
  font-family: var(--font-mono);
  font-size: 1.75cqh;
  font-variant-numeric: tabular-nums;
  font-weight: 600;
}
</style>
