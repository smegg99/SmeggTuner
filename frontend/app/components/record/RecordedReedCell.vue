<!-- app/components/record/RecordedReedCell.vue -->
<template>
  <td
    class="num"
    :class="{ 'num--out': cell.present && !cell.inTol }"
    :title="cell.blank ? t(`tuner.boxes.${cell.blank}`) : undefined"
    @dblclick.stop="open"
  >
    <input
      v-if="editing"
      ref="field"
      class="num__edit"
      :value="draft"
      :aria-label="label"
      @input="draft = ($event.target as HTMLInputElement).value"
      @keydown.enter.stop.prevent="commit"
      @keydown.esc.stop.prevent="close"
      @blur="editing = false"
      @click.stop
    >
    <button
      v-else-if="editable"
      ref="button"
      type="button"
      class="num__key"
      :title="t('record.table.editHint')"
      :aria-label="`${label}: ${cell.curr}`"
      @click.stop="open"
      @keydown.enter.stop.prevent="open"
      @keydown.space.stop.prevent="open"
    >
      {{ cell.curr }}
    </button>
    <template v-else>
      {{ cell.curr }}
    </template>
  </td>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, useTemplateRef } from 'vue'
import type { ReedCell } from '~/utils/recordedRows'
import { useRecordFormat } from '~/composables/useRecordFormat'

const props = defineProps<{ cell: ReedCell, noteName: string }>()

// The typed value, in the display unit; the backend stores it at the slot's own pitch.
const emit = defineEmits<{ edit: [value: number] }>()

const { t } = useI18n()
const { parse } = useRecordFormat()

const label = computed(() => `${props.noteName}: ${t('record.table.reed', { index: props.cell.reed + 1 })}`)
const editable = computed(() => props.cell.present && props.cell.slot >= 0)
const button = useTemplateRef<HTMLButtonElement>('button')
const editing = ref(false)
const draft = ref('')
const field = useTemplateRef<HTMLInputElement>('field')

async function open() {
  if (!editable.value || editing.value) return
  draft.value = props.cell.raw.toFixed(1)
  editing.value = true
  await nextTick()
  field.value?.focus()
  field.value?.select()
}

async function close() {
  editing.value = false
  await nextTick()
  button.value?.focus()
}

function commit() {
  const value = parse(draft.value)
  void close()
  if (value !== null) emit('edit', value)
}
</script>

<style scoped>
td {
  padding: var(--gap);
  white-space: nowrap;
}

.num {
  color: rgb(var(--v-theme-ink2));
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
  text-align: right;
}

.num--out {
  color: rgb(var(--v-theme-warn));
}

.num__key {
  background: none;
  border: 0;
  color: inherit;
  cursor: text;
  font: inherit;
  padding: 0;
}

.num__key:focus-visible {
  outline: 2px solid rgb(var(--v-theme-accent));
  outline-offset: 2px;
}

.num__edit {
  background: rgb(var(--v-theme-raised));
  border: 1px solid rgb(var(--v-theme-ink2));
  border-radius: 2px;
  color: rgb(var(--v-theme-ink));
  font: inherit;
  padding: 0 3px;
  text-align: right;
  width: 6ch;
}
</style>
