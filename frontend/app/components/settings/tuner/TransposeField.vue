<template>
  <UiFormField
    :label="t('tuner.reference.transpose')"
    :hint="t(why || 'tuner.reference.transposeHint')"
  >
    <div class="tr">
      <UiToolKey
        icon="mdi-minus"
        :disabled="!!why || transpose <= 0"
        :title="t(why || 'tuner.reference.transposeDown')"
        @click="setTranspose(stepTranspose(transpose, -1))"
      />
      <output class="tr__value">{{ transpose > 0 ? `+${transpose}` : '0' }}</output>
      <UiToolKey
        icon="mdi-plus"
        :disabled="!!why || transpose >= TRANSPOSE_MAX"
        :title="t(why || 'tuner.reference.transposeUp')"
        @click="setTranspose(stepTranspose(transpose, 1))"
      />
    </div>
  </UiFormField>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useSessions } from '~/composables/useSessions'
import { useTuner } from '~/composables/useTuner'
import { TRANSPOSE_MAX, stepTranspose, transposeRefusal } from '~/utils/reference'

// How many semitones the instrument's keys sound above their names; tones mode only, or a session's.
const { t } = useI18n()
const { active } = useSessions()
const { transpose, mode, setTranspose } = useTuner()

const why = computed(() => transposeRefusal({
  lockA4: false,
  session: active.value !== null,
  readings: active.value?.readings ?? 0,
  mode: mode.value,
}))
</script>

<style scoped>
.tr {
  align-items: center;
  display: flex;
  gap: var(--s99ab-control-gap);
}

.tr__value {
  font-family: var(--font-mono);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  text-align: center;
  width: 32px;
}
</style>
