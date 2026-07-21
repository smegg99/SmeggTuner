<template>
  <div class="regs">
    <div class="regs__pane">
      <ul
        v-if="registers.length"
        class="regs__list"
      >
        <li
          v-for="(r, i) in registers"
          :key="`${r.name}-${i}`"
          class="regs__item"
        >
          <span class="regs__sym"><InstrumentRegisterSymbol :banks="r.banks" /></span>
          <span class="regs__feet">{{ feetOf(r.banks) }}</span>
          <span class="regs__name">{{ r.name }}</span>
          <button
            type="button"
            class="regs__kill"
            :title="t('instrument.builder.remove')"
            @click="remove(i)"
          >
            <v-icon
              icon="mdi-close"
              size="x-small"
            />
          </button>
        </li>
      </ul>

      <p
        v-else
        class="regs__empty"
      >
        {{ t('instrument.builder.none') }}
      </p>
    </div>

    <InstrumentRegisterBuilder @add="add" />
  </div>
</template>

<script setup lang="ts">
import { feetOf } from '~/utils/feet'
import type { Register } from '~/types/session'

// Registers held as v-model data; a register with a duplicate name is not re-added.
const registers = defineModel<Register[]>({ required: true })

const { t } = useI18n()

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
  gap: 1cqh;
}

/* Fixed height so adding registers scrolls inside the pane instead of growing the dialog; gutter reserved so the remove key never collides with the scrollbar. */
.regs__pane {
  background: rgb(var(--v-theme-chrome2));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 2px;
  height: 30cqh;
  overflow-y: auto;
  padding: 0.6cqh 0.5cqw;
  scrollbar-gutter: stable;
}

.regs__list {
  display: flex;
  flex-direction: column;
  gap: 0.4cqh;
  list-style: none;
  margin: 0;
  padding: 0;
}

.regs__item {
  align-items: center;
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 2px;
  display: flex;
  gap: 0.8cqw;
  padding: 0.7cqh 0.7cqw;
}

.regs__sym {
  block-size: 5cqh;
  flex: 0 0 auto;
  inline-size: 5cqh;
}

.regs__feet {
  color: rgb(var(--v-theme-ink));
  font-family: var(--font-mono);
  font-size: 1.9cqh;
}

.regs__name {
  color: rgb(var(--v-theme-ink3));
  font-family: var(--font-mono);
  font-size: 1.5cqh;
}

.regs__kill {
  align-items: center;
  background: none;
  border: 0;
  color: rgb(var(--v-theme-ink3));
  cursor: pointer;
  display: flex;
  margin-inline-start: auto;
}

.regs__kill:hover {
  color: rgb(var(--v-theme-error));
}

.regs__empty {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.6cqh;
  line-height: 1.5;
  margin: 0;
}
</style>
