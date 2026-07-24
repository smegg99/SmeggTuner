<template>
  <div class="regs">
    <InstrumentRegisterList
      :items="items"
      :empty="t('instrument.builder.none')"
      height="10.5cqh"
      @remove="remove"
    >
      <template #symbol="{ index }">
        <InstrumentRegisterSymbol :banks="registers[index]!.banks" />
      </template>
    </InstrumentRegisterList>

    <InstrumentRegisterBuilder @add="add" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { feetOf } from '~/utils/feet'
import type { Register } from '~/types/session'

// Registers held as v-model data; a register with a duplicate name is not re-added.
const registers = defineModel<Register[]>({ required: true })

const { t } = useI18n()

const items = computed(() => registers.value.map(r => ({ name: r.name, feet: feetOf(r.banks) })))

function add(r: Register) {
  if (registers.value.some(x => x.name === r.name)) return
  registers.value = [...registers.value, r]
}

function remove(i: number) {
  registers.value = registers.value.filter((_, n) => n !== i)
}
</script>

<style scoped>
.regs {
  display: flex;
  flex-direction: column;
  gap: 0.8cqh;
}
</style>
