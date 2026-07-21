<template>
  <UiPanelCanvas
    :title="t('tuner.spectrum.title')"
    :label="t('tuner.spectrum.label')"
    :renderer="renderer"
  >
    <template
      v-if="reeds"
      #meta
    >
      <b>{{ t('tuner.spectrum.reeds', { n: reeds }) }}</b>, {{ verdict }}
    </template>
  </UiPanelCanvas>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useSpectrumRender } from '~/composables/render/useSpectrumRender'
import { useTuner } from '~/composables/useTuner'
import { liveReedsDerived, liveReedsUsable } from '~/utils/tuning'

// Header states which of three cases the engine is in: separated, recovered from the beat, or
// not split (per-reed figures are then lobes of one peak and shown nowhere).
const { t } = useI18n()
const { reeds: measured, reedsSeparated, reedsFromBeat } = useTuner()

const renderer = useSpectrumRender()
const reeds = computed(() => measured.value.length)

const verdict = computed(() => {
  if (!liveReedsUsable(reedsSeparated.value, reedsFromBeat.value)) {
    return t('tuner.spectrum.notSplit')
  }

  return liveReedsDerived(reedsSeparated.value, reedsFromBeat.value)
    ? t('tuner.spectrum.fromBeat')
    : t('tuner.spectrum.separated')
})
</script>
