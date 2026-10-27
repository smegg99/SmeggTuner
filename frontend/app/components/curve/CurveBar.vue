<!-- app/components/curve/CurveBar.vue -->
<template>
  <!-- The curve editor's keys: the rank and note being set, its value, and the curve's own keys. -->
  <div class="bar">
    <UiToolGroup
      v-if="!compact && rankItems.length > 1"
      :model-value="reed"
      :items="rankItems"
      :label="t('curve.rank')"
      @update:model-value="v => reed = v as number"
    />

    <InstrumentNoteStepper v-model="note" />

    <v-text-field
      v-model="value"
      class="bar__value tuner-number"
      :suffix="t(`tuner.unit.${entry}`)"
      :aria-label="t('curve.entryUnit')"
      hide-details="auto"
      autocomplete="off"
      inputmode="decimal"
    />

    <UiToolKey
      icon="mdi-check"
      :label="compact ? undefined : t('curve.set')"
      :disabled="!note || busy"
      :title="t('curve.setHint')"
      @click="emit('set')"
    />
    <UiToolKey
      icon="mdi-close"
      :label="compact ? undefined : t('curve.remove')"
      :disabled="!canRemove || busy"
      :title="t('curve.removeHint')"
      @click="emit('remove')"
    />

    <span
      v-if="!compact"
      class="bar__sp"
    />
    <v-menu
      v-if="compact"
      :close-on-content-click="false"
    >
      <template #activator="{ props: menu }">
        <UiToolKey
          v-bind="menu"
          icon="mdi-tune-variant"
          :title="t('tuner.controls')"
        />
      </template>
      <div class="curve-options">
        <div class="curve-options__row">
          <CurveTools
            v-model:recorded="recorded"
            v-model:scale="scale"
            v-model:hz="hz"
            :unit="unit"
          />
        </div>
        <UiToolGroup
          v-if="rankItems.length > 1"
          :model-value="reed"
          :items="rankItems"
          :label="t('curve.rank')"
          @update:model-value="v => reed = v as number"
        />
        <UiToolKey
          :label="t('curve.interpolate')"
          :active="interpolate"
          :title="t('curve.interpolateHint')"
          @click="emit('interpolate', !interpolate)"
        />
        <div class="curve-options__row">
          <CurveMenu />
          <UiToolKey
            icon="mdi-delete-outline"
            :label="t('curve.clearAll')"
            :disabled="!canClear || busy"
            :title="t('curve.clearAllHint')"
            @click="emit('clear')"
          />
        </div>
      </div>
    </v-menu>

    <UiToolKey
      v-if="!compact"
      :label="t('curve.interpolate')"
      :active="interpolate"
      :title="t('curve.interpolateHint')"
      @click="emit('interpolate', !interpolate)"
    />
    <CurveMenu v-if="!compact" />
    <UiToolKey
      v-if="!compact"
      icon="mdi-delete-outline"
      :label="t('curve.clearAll')"
      :disabled="!canClear || busy"
      :title="t('curve.clearAllHint')"
      @click="emit('clear')"
    />
  </div>
</template>

<script setup lang="ts">
import CurveMenu from './CurveMenu.vue'
import CurveTools from './CurveTools.vue'
import type { ToolItem } from '@smegg99/s99ab/types'
import type { DeviationUnit } from '~/utils/deviation'
import type { EntryUnit } from '~/utils/curveUnits'

const reed = defineModel<number>('reed', { required: true })
const note = defineModel<number>('note', { required: true })
const value = defineModel<number>('value', { required: true })
const recorded = defineModel<boolean>('recorded', { required: true })
const scale = defineModel<number>('scale', { required: true })
const hz = defineModel<boolean>('hz', { required: true })
defineProps<{
  unit: DeviationUnit
  entry: EntryUnit
  rankItems: ToolItem[]
  interpolate: boolean
  canRemove: boolean
  canClear: boolean
  busy: boolean
}>()
const emit = defineEmits<{ set: [], remove: [], interpolate: [boolean], clear: [] }>()

const { t } = useI18n()
const compact = useCompactLayout()
</script>

<style scoped>
/* Matches ShellAppToolbar; no-wrap so Set/Remove never jump to a second line at some widths. */
.bar {
  align-items: center;
  background: rgb(var(--v-theme-chrome2));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 3.6px;
  display: flex;
  gap: var(--gap);
  min-width: 0;
  overflow: hidden;
  padding: var(--gap);
}

.bar__value {
  flex: 0 0 auto;
  width: 86.5px;
}

.bar__sp {
  flex: 1 1 auto;
}

/* Portrait: entry keys on one row, the curve's own keys on the next. */
@media (max-width: 720px) {
  .bar { flex-wrap: wrap; row-gap: var(--gap); }
  .bar__sp { flex-basis: 100%; height: 0; }
}

.curve-options { background: rgb(var(--v-theme-chrome)); border: 1px solid rgb(var(--v-theme-line)); border-radius: var(--app-radius-panel); padding: 12px; display: flex; flex-direction: column; gap: 12px; max-width: calc(100vw - 24px); }
.curve-options__row { display: flex; align-items: center; gap: 6px; }
@media (max-width: 959px), (max-height: 559px) {
  .bar { flex-wrap: nowrap; gap: 4px; padding: 4px; }
  .bar__value { width: 64px; }
  .bar :deep(.step) { flex: 1; }
}
@media (max-width: 559px) {
  .bar { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); }
  .bar :deep(.step) { grid-column: 1 / 3; }
  .bar__value { grid-column: 3; width: 100%; }
  .bar > .key { width: 100%; }
}
</style>
