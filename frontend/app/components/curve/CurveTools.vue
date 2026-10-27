<template>
  <UiToolKey
    class="tuner-small"
    :label="t('curve.recorded')"
    :active="recorded"
    :title="t('curve.recordedHint')"
    @click="recorded = !recorded"
  />

  <v-menu>
    <template #activator="{ props: menu }">
      <v-btn
        v-bind="menu"
        :title="t('curve.scaleHint')"
        :aria-label="t('curve.scaleHint')"
        size="small"
        class="tuner-small"
        append-icon="mdi-menu-down"
      >
        {{ scaleLabel(scale) }}
      </v-btn>
    </template>
    <v-list class="tuner-menu">
      <v-list-item
        v-for="v in CURVE_SCALES[hz ? 'hz' : unit]"
        :key="v"
        :active="v === scale"
        :append-icon="(v === scale) ? 'mdi-check' : undefined"
        :title="scaleLabel(v)"
        @click="scale = v"
      />
    </v-list>
  </v-menu>

  <UiToolGroup
    :model-value="hz"
    :items="ENTRY_UNITS"
    :label="t('curve.entryUnit')"
    class="tuner-small"
    @update:model-value="v => hz = Boolean(v)"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ToolItem } from '@smegg99/s99ab/types'
import { CURVE_SCALES } from '~/utils/curveUnits'
import type { DeviationUnit } from '~/utils/deviation'

// The graph's own switches: the pass's readings over the list, a fixed scale, and Hz entry.
const props = defineProps<{ unit: DeviationUnit }>()

const recorded = defineModel<boolean>('recorded', { required: true })
const scale = defineModel<number>('scale', { required: true })
const hz = defineModel<boolean>('hz', { required: true })

const { t } = useI18n()

const scaleLabel = (v: number) => (v ? t('curve.scale', { v }) : t('curve.scaleAuto'))

const ENTRY_UNITS = computed<ToolItem[]>(() => [
  { value: false, label: t(`tuner.unit.${props.unit}`) },
  { value: true, label: t('tuner.unit.hz') },
])
</script>
