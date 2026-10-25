<template>
  <div
    class="box"
    :class="{ 'box--beat': box.kind === 'beat', 'box--out': box.out, 'box--dim': box.dimmed }"
  >
    <div class="box__top">
      <span class="box__label">{{ label }}</span>
      <span class="box__unit">{{ t(`tuner.boxes.unit.${box.unit}`) }}</span>
    </div>

    <!-- "notHeard": a declared slot the engine did not hear. "idle": nothing sounding. Either way
         the box keeps its shape. -->
    <div
      class="box__value"
      :class="{ 'box__value--none': box.value === null }"
    >
      <template v-if="box.value !== null">
        {{ format(box.value) }}
      </template>
      <template v-else-if="box.blank === 'notHeard' || box.blank === 'hzOnly'">
        {{ t(`tuner.boxes.${box.blank}`) }}
      </template>
    </div>

    <TunerDeviationBar
      :frac="box.frac"
      :out="box.out"
    />

    <div class="box__foot">
      <template v-if="box.goal !== null">
        {{ t('tuner.boxes.goal') }}
        <b>{{ format(box.goal) }}</b>
      </template>
      <!-- A chord tone's note name, where a reed would show its goal. -->
      <b v-else-if="box.text">{{ box.text }}</b>
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
  const b = props.box
  if (b.kind === 'reed') {
    if (b.rank) return t('tuner.boxes.reedBank', { bank: b.rank })
    return t(b.tone ? 'tuner.boxes.tone' : 'tuner.boxes.reed', { n: b.number ?? b.index + 1 })
  }
  const [lo, hi] = b.pair ?? [b.index, b.index + 1]
  return t('tuner.boxes.beat', { a: lo, b: hi })
})

// Formatting only, at the readout's own precision; the value comes from the DTO.
function format(value: number): string {
  const text = value.toFixed(props.box.decimals)
  return props.box.signed && value > 0 ? `+${text}` : text
}
</script>

<style scoped>
.box {
  container-type: inline-size;
  background: rgb(var(--v-theme-chrome2));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: var(--app-radius-panel);
  display: flex;
  flex: 1 1 0;
  flex-direction: column;
  gap: var(--gap-tight);
  min-width: 0;
  padding: var(--gap);
}

.box--beat {
  background: rgb(var(--v-theme-sunk));
}

/* Rows use fixed heights, not min-heights, so the box can't grow when a goal appears mid-use. */
.box__top {
  align-items: baseline;
  display: flex;
  flex: 0 0 auto;
  gap: var(--gap);
  height: 13.5px;
}

.box__label {
  color: rgb(var(--v-theme-ink2));
  font-size: 11px;
  font-weight: 600;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.box__unit {
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
  margin-left: auto;
}

/* Five boxes on a phone: the label keeps the room, the unit is in the ruler's header anyway. */
@container (max-width: 90px) {
  .box__unit { display: none; }
}

.box__value {
  align-items: center;
  color: rgb(var(--v-theme-ink));
  display: flex;
  flex: 0 0 auto;
  font-family: var(--font-mono);
  font-size: 24.5px;
  font-variant-numeric: tabular-nums;
  font-weight: 500;
  /* Digit height reserved so the row doesn't jump when a note starts. */
  height: 24.5px;
  letter-spacing: -0.03em;
  line-height: 1;
  overflow: hidden;
}

.box--out .box__value {
  color: rgb(var(--v-theme-warn));
}

/* No note: the reference keeps the last numbers and dims them. */
.box--dim .box__value {
  color: rgb(var(--v-theme-ink3));
}

/* The "no number" text borrows the digit's reserved height so the box stays one size. */
.box__value--none {
  color: rgb(var(--v-theme-ink3));
  font-family: var(--font-sans);
  font-size: 11.5px;
  font-weight: 400;
  letter-spacing: 0;
  line-height: 1.25;
}

.box__foot {
  align-items: baseline;
  color: rgb(var(--v-theme-ink3));
  display: flex;
  flex: 0 0 auto;
  font-size: 11px;
  gap: var(--gap-tight);
  /* Fixed, not a minimum: the 14px <b> would otherwise push past a floor and grow the card. */
  height: 16.4px;
  overflow: hidden;
}

.box__foot b {
  color: rgb(var(--v-theme-goal));
  font-family: var(--font-mono);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  font-weight: 600;
}
@container (max-width: 100px) {
  .box__value { font-size: 18px; }
  .box__foot { font-size: 10px; }
  .box__foot b { font-size: 11px; }
}
</style>
