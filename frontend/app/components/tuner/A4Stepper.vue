<template>
  <!-- Inline: A4 is typed into the field. -->
  <div class="a4">
    <label
      class="a4__field"
      :class="{ 'a4__field--off': !!why }"
      :title="t(why || 'tuner.reference.a4Type')"
    >
      <span class="a4__name">A4</span>
      <input
        v-model="draft"
        class="a4__input"
        inputmode="decimal"
        :disabled="!!why"
        :aria-label="t('tuner.reference.a4Type')"
        @focus="($event.target as HTMLInputElement).select()"
        @keydown.enter="commit"
        @keydown.esc="revert"
        @blur="commit"
      >
    </label>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'
import { useSessions } from '~/composables/useSessions'
import { useTuner } from '~/composables/useTuner'
import { a4Refusal, parseA4 } from '~/utils/reference'

// The engine's A4; with a session open it is the session's.
const { t } = useI18n()
const { config } = useConfigSync()
const { active } = useSessions()
const { a4, mode, setA4 } = useTuner()

const why = computed(() => a4Refusal({
  lockA4: config.tuner?.lock_a4 ?? false,
  session: active.value !== null,
  readings: active.value?.readings ?? 0,
  mode: mode.value,
}))

// The typed value; it follows A4 until the user edits it, and a refused or invalid entry goes back.
const draft = ref(a4.value.toFixed(1))
watch(a4, (hz) => {
  draft.value = hz.toFixed(1)
})

function revert(event?: Event) {
  draft.value = a4.value.toFixed(1)
  ;(event?.target as HTMLInputElement | undefined)?.blur()
}

function commit(event: Event) {
  const hz = parseA4(draft.value)
  if (hz === null) return revert()
  if (hz !== a4.value) void setA4(hz)
  draft.value = hz.toFixed(1)
  if (event.type === 'keydown') (event.target as HTMLInputElement).blur()
}
</script>

<style scoped>
.a4 {
  --s99ab-control-h: 26px;
  align-items: center;
  display: flex;
  justify-content: center;
  min-width: 0;
}

/* Drawn as a key; the number is the field. */
.a4__field {
  align-items: center;
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-edge));
  border-radius: var(--s99ab-radius);
  cursor: text;
  display: flex;
  gap: 4px;
  height: var(--s99ab-control-h);
  padding-inline: 6px;
  width: 80px;
}

.a4__field:focus-within {
  outline: 2px solid rgb(var(--v-theme-accent));
  outline-offset: -1px;
}

.a4__field--off {
  cursor: default;
  opacity: 0.45;
}

.a4__name {
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
}

.a4__input {
  background: none;
  border: 0;
  color: rgb(var(--v-theme-ink));
  font-family: var(--font-mono);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  min-width: 0;
  outline: none;
  padding: 0;
  text-align: right;
  width: 100%;
}
</style>
