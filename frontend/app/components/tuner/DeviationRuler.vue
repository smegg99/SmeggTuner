<template>
  <UiPanelCanvas
    :title="t('tuner.ruler.title')"
    :label="t('tuner.ruler.label')"
    :renderer="renderer"
  >
    <template #meta>
      {{ t('tuner.ruler.meta', { tol: tolerance.toFixed(1) }) }}
    </template>
  </UiPanelCanvas>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRulerRender } from '~/composables/render/useRulerRender'
import { useConfigSync } from '~/composables/useConfigSync'

const { t } = useI18n()
const { config } = useConfigSync()

const tolerance = computed(() => config.tuner?.tolerance ?? 0)

const renderer = useRulerRender()
</script>
