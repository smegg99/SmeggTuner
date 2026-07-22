<template>
  <div class="build">
    <div class="build__presets">
      <button
        v-for="p in PRESETS"
        :key="p.name"
        type="button"
        class="preset"
        :title="p.feet"
        @click="emit('add', { name: p.name, banks: [...p.banks] })"
      >
        <span class="preset__sym"><InstrumentRegisterSymbol :banks="p.banks" /></span>
        <span class="preset__feet">{{ p.feet }}</span>
      </button>
    </div>

    <div class="build__custom">
      <span class="build__custom-h">{{ t('instrument.builder.custom') }}</span>

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
          :disabled="!banks.length"
          @click="add"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive } from 'vue'
import { banksOfSymbol, canonicalName, feetOf } from '~/utils/feet'
import type { RegisterDots } from '~/utils/feet'
import type { ToolItem } from '~/components/ui/ToolGroup.vue'
import type { Bank, Register } from '~/types/session'

// Emits a Register (banks + canonical name); picker and card share feet.ts so they cannot disagree.
const emit = defineEmits<{ add: [register: Register] }>()

const { t } = useI18n()

const PRESETS = computed(() => (
  [
    ['L', 'M1', 'M2', 'H'], // 16+8+8+4
    ['L', 'M1', 'M2', 'M3'], // 16+8+8+8
    ['L', 'M1', 'M2'], // 16+8+8
    ['M1', 'M2', 'M3'], // musette
    ['L', 'M1'], // 16+8
    ['M1'], // a lone 8
  ] as Bank[][]
).map(banks => ({ banks, name: canonicalName(banks), feet: feetOf(banks) })))

const COUNTS: ToolItem[] = [0, 1, 2, 3, 4].map(n => ({ value: n, label: String(n) }))

const sym = reactive<RegisterDots>({ high: false, middle: 0, low: false })

const banks = computed<Bank[]>(() => banksOfSymbol(sym))

function add() {
  if (!banks.value.length) return
  emit('add', { name: canonicalName(banks.value), banks: [...banks.value] })
  sym.high = false
  sym.middle = 0
  sym.low = false
}
</script>

<style scoped>
.build {
  display: flex;
  flex-direction: column;
  gap: 1cqh;
}

.build__presets {
  display: grid;
  gap: 0.8cqh 0.8cqw;
  grid-template-columns: repeat(auto-fill, minmax(11cqw, 1fr));
}

.preset {
  align-items: center;
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  color: rgb(var(--v-theme-ink));
  cursor: pointer;
  display: flex;
  gap: 0.6cqw;
  padding: 0.8cqh 0.7cqw;
}

.preset:hover {
  border-color: rgb(var(--v-theme-ink2));
}

.preset__sym {
  block-size: 5cqh;
  flex: 0 0 auto;
  inline-size: 5cqh;
}

.preset__feet {
  font-family: var(--font-mono);
  font-size: 1.7cqh;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.build__custom {
  border-top: 1px solid rgb(var(--v-theme-lineSoft));
  display: flex;
  flex-direction: column;
  gap: 0.9cqh;
  padding-top: 1.2cqh;
}

.build__custom-h {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.35cqh;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.build__body {
  align-items: center;
  display: flex;
  gap: 1.2cqw;
}

.build__sym {
  block-size: 7cqh;
  flex: 0 0 auto;
  inline-size: 7cqh;
}

.build__rows {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 0.55cqh;
  min-width: 0;
}

.build__row {
  align-items: center;
  display: flex;
  gap: 0.8cqw;
}

.build__label {
  color: rgb(var(--v-theme-ink));
  flex: 1 1 auto;
  font-size: 1.65cqh;
}
</style>
