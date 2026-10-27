<template>
  <div class="bass">
    <div
      v-show="mode !== 'builder'"
      class="bass__row"
    >
      <span class="bass__label">{{ t('instrument.bass.voices') }}</span>
      <UiToolGroup
        :model-value="reeds"
        :items="COUNTS"
        :label="t('instrument.bass.voices')"
        @update:model-value="n => setReeds(n as number)"
      />
    </div>

    <!-- No machine declared, no machine to describe: one line saying what the row above is for,
         rather than an empty list, a row of nothing and a name field for a switch that cannot
         exist. -->
    <p
      v-if="reeds === 0"
      class="bass__none"
    >
      {{ t('instrument.bass.hint') }}
    </p>

    <template v-else>
      <InstrumentRegisterList
        v-show="mode !== 'builder'"
        :items="items"
        :empty="t('instrument.bass.fixed')"
        @remove="removeRegister"
      />

      <InstrumentBassBuilder
        v-show="mode !== 'list'"
        v-model="registers"
        :feet="machineFeet"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { bassFeetOf } from '~/utils/feet'
import type { ToolItem } from '@smegg99/s99ab/types'
import type { BassRegister } from '~/types/session'

const reeds = defineModel<number>('reeds', { required: true })
const registers = defineModel<BassRegister[]>('registers', { required: true })

withDefaults(defineProps<{ mode?: 'all' | 'list' | 'builder' }>(), { mode: 'all' })

const { t } = useI18n()

// None, or the counts machines are built with; four and five are the common ones.
const COUNTS: ToolItem[] = [
  { value: 0, label: t('instrument.bass.none') },
  ...[2, 3, 4, 5, 6].map(n => ({ value: n, label: String(n) })),
]

const machineFeet = computed(() => bassFeetOf(reeds.value))

const items = computed(() => registers.value.map(r => ({
  name: r.name,
  feet: r.feet.map(f => `${f}'`).join('+'),
})))

function setReeds(n: number) {
  reeds.value = n
  // A shrunk machine invalidates switches reaching ranks it no longer has.
  registers.value = registers.value.filter(r => r.feet.every(f => bassFeetOf(n).includes(f)))
}

function removeRegister(i: number) {
  registers.value = registers.value.filter((_, n) => n !== i)
}
</script>

<style scoped>
.bass {
  display: flex;
  flex-direction: column;
  gap: var(--gap);
}

.bass__row {
  align-items: center;
  display: flex;
  gap: var(--gap);
}

.bass__label {
  color: rgb(var(--v-theme-ink));
  flex: 1 1 auto;
  font-size: 11.5px;
}

.bass__none {
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
  line-height: 1.5;
  margin: 0;
  max-width: 52ch;
}

@media (max-height: 559px) and (min-width: 560px) {
  .bass { display: grid; grid-template-columns: 1fr 1fr; gap: 6px; align-items: start; }
  .bass__row, .bass :deep(.list) { grid-column: 1 / -1; }
  .bass :deep(.bass__h) { display: none; }
  .bass :deep(.bass__feet) { flex-wrap: wrap; }
}
</style>
