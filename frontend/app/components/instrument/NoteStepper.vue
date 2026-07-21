<template>
  <div class="step">
    <UiToolKey
      icon="mdi-chevron-down"
      :disabled="model !== 0 && model <= NOTE_MIN"
      :title="t('instrument.range.down')"
      @click="bump(-1)"
    />

    <!-- Fixed-width readout so stepping between note names doesn't shuffle the arrows either side. -->
    <output
      class="step__note"
      :class="{ 'step__note--unset': model === 0 }"
    >{{ model === 0 ? '--' : noteName(model, naming) }}</output>

    <UiToolKey
      icon="mdi-chevron-up"
      :disabled="model !== 0 && model >= NOTE_MAX"
      :title="t('instrument.range.up')"
      @click="bump(1)"
    />

    <!-- Clear key is hidden (not removed) when unset, so the row stays the same width set or unset. -->
    <span class="step__clear">
      <UiToolKey
        v-if="model !== 0"
        icon="mdi-close"
        :title="t('instrument.range.clear')"
        @click="model = 0"
      />
    </span>
  </div>
</template>

<script setup lang="ts">
import { NOTE_MAX, NOTE_MIN } from '~/composables/useTuner'
import { noteName } from '~/utils/tuning'

// One end of the keyboard. MIDI note number; 0 means unset (core/session: Lo/Hi are 0 until set).
const model = defineModel<number>({ required: true })

defineProps<{ naming?: string }>()

const { t } = useI18n()

// C4: where an unset end starts when first stepped.
const SEED = 60

function bump(delta: number) {
  if (model.value === 0) {
    model.value = SEED
    return
  }
  model.value = Math.min(NOTE_MAX, Math.max(NOTE_MIN, model.value + delta))
}
</script>

<style scoped>
.step {
  align-items: center;
  display: flex;
  gap: 0.4cqw;
}

/* Same sunk readout as ToolbarNote: tabular figures, fixed width. */
.step__note {
  align-items: center;
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  color: rgb(var(--v-theme-ink));
  display: flex;
  flex: 0 0 auto;
  font-family: var(--font-mono);
  font-size: 1.75cqh;
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  height: 3.9cqh;
  justify-content: center;
  line-height: 1;
  width: 5.5cqw;
}

.step__note--unset {
  color: rgb(var(--v-theme-ink3));
  font-weight: 500;
}

/* Reserved slot; never collapses so the v-if clear key moves nothing around it. */
.step__clear {
  align-items: center;
  display: flex;
  flex: 0 0 auto;
  height: 3.9cqh;
  justify-content: center;
  width: 3.9cqh;
}
</style>
