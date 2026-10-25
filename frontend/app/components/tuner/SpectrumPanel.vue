<template>
  <UiPanelCanvas
    :title="t('tuner.spectrum.title')"
    :label="t('tuner.spectrum.label')"
    :renderer="renderer"
  >
    <!-- With a session open its register decides what is measured; the panel only says so. -->
    <template
      v-if="sessionReeds"
      #meta
    >
      {{ t('tuner.panel.bySession', { n: reedCount }, reedCount) }}
    </template>
  </UiPanelCanvas>
</template>

<script setup lang="ts">
import { useSpectrumRender } from '~/composables/render/useSpectrumRender'
import { useTuner } from '~/composables/useTuner'

const { t } = useI18n()
const { reedCount, sessionReeds } = useTuner()

const renderer = useSpectrumRender()
</script>
