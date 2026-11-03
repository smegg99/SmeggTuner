<!-- app/components/signal/SignalVoices.vue -->
<template>
  <!-- One row per tone: its offset from the note in cents, typed or stepped by a cent. -->
  <div class="voices">
    <div
      v-for="(c, i) in cents"
      :key="i"
      class="voices__row"
    >
      <UiToolKey
        icon="mdi-minus"
        :title="t('tuner.signal.cents')"
        :disabled="c <= -MAX_CENTS"
        @click="setCents(i, c - 1)"
      />
      <label
        class="voices__field"
        :title="t('tuner.signal.cents')"
      >
        <input
          :value="drafts.get(i) ?? format(c)"
          class="voices__input"
          inputmode="decimal"
          :aria-label="t('tuner.signal.cents')"
          @focus="($event.target as HTMLInputElement).select()"
          @input="drafts.set(i, ($event.target as HTMLInputElement).value)"
          @keydown.enter="($event.target as HTMLInputElement).blur()"
          @keydown.esc="revert(i, $event)"
          @blur="commit(i)"
        >
        <span class="voices__unit">c</span>
      </label>
      <UiToolKey
        icon="mdi-plus"
        :title="t('tuner.signal.cents')"
        :disabled="c >= MAX_CENTS"
        @click="setCents(i, c + 1)"
      />
      <UiToolKey
        icon="mdi-close"
        :title="t('tuner.signal.remove')"
        :disabled="cents.length <= 1"
        @click="removeVoice(i)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { MAX_CENTS, useSignal } from '~/composables/useSignal'

const { t } = useI18n()
const { cents, setCents, removeVoice } = useSignal()

// What is being typed in a row, until Enter or leaving the field commits it.
const drafts = reactive(new Map<number, string>())

function format(c: number): string {
  const text = Number.isInteger(c) ? String(c) : c.toFixed(1)
  return c > 0 ? `+${text}` : text
}

function commit(i: number) {
  const draft = drafts.get(i)
  drafts.delete(i)
  if (draft === undefined) return
  const c = Number.parseFloat(draft.replace(',', '.'))
  if (Number.isFinite(c)) setCents(i, c)
}

function revert(i: number, event: KeyboardEvent) {
  drafts.delete(i)
  ;(event.target as HTMLInputElement).blur()
}
</script>

<style scoped>
.voices {
  display: flex;
  flex-direction: column;
  gap: var(--gap);
}

.voices__row {
  align-items: center;
  display: flex;
  gap: var(--s99ab-control-gap);
}

/* Drawn as a key; the number is the field, as A4's. */
.voices__field {
  align-items: center;
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-edge));
  border-radius: var(--s99ab-radius);
  cursor: text;
  display: flex;
  flex: 1 1 auto;
  gap: 4px;
  height: var(--s99ab-control-h);
  min-width: 0;
  padding-inline: 6px;
}

.voices__field:focus-within {
  outline: 2px solid rgb(var(--v-theme-accent));
  outline-offset: -1px;
}

.voices__input {
  background: none;
  border: 0;
  color: rgb(var(--v-theme-ink));
  font-family: var(--font-mono);
  font-size: 13px;
  font-variant-numeric: tabular-nums;
  min-width: 0;
  outline: none;
  padding: 0;
  text-align: right;
  width: 100%;
}

.voices__unit {
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
}
</style>
