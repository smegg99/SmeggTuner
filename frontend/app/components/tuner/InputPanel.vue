<template>
  <UiPanelCanvas
    :title="t('tuner.inputPanel.title')"
    :label="t('tuner.inputPanel.label')"
    :renderer="renderer"
  >
    <template #meta>
      <span v-if="running">
        {{ source }}, <b>{{ level }}</b>
      </span>
    </template>
  </UiPanelCanvas>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useInputRender } from '~/composables/render/useInputRender'
import { useAudioDevices } from '~/composables/useAudioDevices'
import { useTuner } from '~/composables/useTuner'

// With the engine stopped the header shows nothing — never a fabricated "0 dB" / "-inf".
const { t } = useI18n()
const { inputLevel, running } = useTuner()
const { current } = useAudioDevices()

const renderer = useInputRender()

const source = computed(() => current.value.name || t('tuner.source.default'))

// dB is formatting of the backend's measured level; below the floor there is no honest number.
const level = computed(() => {
  const value = inputLevel.value
  if (!(value > 0)) return t('tuner.level.silent')

  return `${(20 * Math.log10(value)).toFixed(0)} dB`
})
</script>
