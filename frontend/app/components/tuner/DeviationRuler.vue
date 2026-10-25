<template>
  <!-- No head: the scale is the whole panel, so the needles get its full height. -->
  <UiPanelCanvas
    :label="t('tuner.ruler.label')"
    :renderer="renderer"
    bare
  />
</template>

<script setup lang="ts">
import { useRulerRender } from '~/composables/render/useRulerRender'

const { t } = useI18n()

const renderer = useRulerRender()
</script>
