<template>
  <div class="build">
    <span class="build__h">{{ t('instrument.builder.title') }}</span>

    <div class="build__body">
      <span class="build__sym"><InstrumentRegisterSymbol :banks="banks" /></span>

      <!-- Rows mirror the engraving top to bottom: 4' high, 8' middle, 16' low. -->
      <div class="build__rows">
        <div class="build__row">
          <span class="build__label">{{ t('instrument.builder.high') }}</span>
          <UiToolKey
            :label="t(sym.high ? 'common.yes' : 'common.no')"
            :active="sym.high"
            @click="sym.high = !sym.high"
          />
        </div>

        <div class="build__row">
          <span class="build__label">{{ t('instrument.builder.middle') }}</span>
          <UiToolGroup
            :model-value="sym.middle"
            :items="COUNTS"
            :label="t('instrument.builder.middle')"
            @update:model-value="n => sym.middle = n as number"
          />
        </div>

        <div class="build__row">
          <span class="build__label">{{ t('instrument.builder.low') }}</span>
          <UiToolKey
            :label="t(sym.low ? 'common.yes' : 'common.no')"
            :active="sym.low"
            @click="sym.low = !sym.low"
          />
        </div>
      </div>

      <UiToolKey
        icon="mdi-plus"
        :label="t('instrument.builder.add')"
        :disabled="!canAdd"
        :title="addHint"
        @click="add"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive } from 'vue'
import { canAddRegister } from '~/utils/accordion'
import { banksOfSymbol, canonicalName } from '~/utils/feet'
import type { RegisterDots } from '~/utils/feet'
import type { ToolItem } from '~/components/ui/ToolGroup.vue'
import type { Bank, Register } from '~/types/session'

const props = defineProps<{ have: readonly Register[] }>()

const emit = defineEmits<{ add: [register: Register] }>()

const { t } = useI18n()

const COUNTS: ToolItem[] = [0, 1, 2, 3, 4].map(n => ({ value: n, label: String(n) }))

const sym = reactive<RegisterDots>({ high: false, middle: 0, low: false })

const banks = computed<Bank[]>(() => banksOfSymbol(sym))
const canAdd = computed(() => canAddRegister(props.have, banks.value))

// Why the key is off, when it is off for a reason worth saying.
const addHint = computed(() => {
  if (canAdd.value || !banks.value.length) return ''
  return t('instrument.builder.already')
})

function add() {
  if (!canAdd.value) return
  emit('add', { name: canonicalName(banks.value), banks: [...banks.value] })
  sym.high = false
  sym.middle = 0
  sym.low = false
}
</script>

<style scoped>
.build {
  border-top: 1px solid rgb(var(--v-theme-lineSoft));
  display: flex;
  flex-direction: column;
  gap: 4.8px;
  padding-top: 7.2px;
}

.build__h {
  color: rgb(var(--v-theme-ink3));
  font-size: 10.8px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.build__body {
  align-items: center;
  display: flex;
  gap: 13.2px;
}

.build__sym {
  block-size: 32px;
  flex: 0 0 auto;
  inline-size: 32px;
}

/* Capped: the panel this sits in is far wider than the dialog it came from, and a label
   stretched across it reads as a gap rather than a row. */
.build__rows {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 3.2px;
  max-width: 286px;
  min-width: 0;
}

.build__row {
  align-items: center;
  display: flex;
  gap: 8.8px;
}

.build__label {
  color: rgb(var(--v-theme-ink));
  flex: 1 1 auto;
  font-size: 13.2px;
}
</style>
