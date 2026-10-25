<template>
  <UiPanelCanvas
    :title="t('tuner.inputPanel.title')"
    :label="t('tuner.inputPanel.label')"
    :renderer="renderer"
  >
    <template #meta>
      <span
        v-if="running && unreliable"
        class="warn"
      >{{ t('tuner.input.unreliable') }}</span>
      <!-- Blinks on alternate analyses, as the reference does; never while automatic volume steers the input. -->
      <span
        v-if="running && clipping"
        class="warn"
        :class="{ 'warn--off': !tooStrong }"
      >{{ t('tuner.input.tooStrong') }}</span>
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
import { useSoundCard } from '~/composables/useSoundCard'
import { useTuner } from '~/composables/useTuner'
import { tooStrongShown } from '~/utils/soundCard'

// With the engine stopped the header shows nothing - never a fabricated "0 dB" / "-inf".
const { t } = useI18n()
const { inputLevel, running, unreliable, state, blink } = useTuner()
const { current } = useAudioDevices()
const { steering } = useSoundCard()

const clipping = computed(() => state.value === 'tooLoud' && !steering.value)
const tooStrong = computed(() => tooStrongShown({ state: state.value, autoVolume: steering.value, blink: blink.value }))

const renderer = useInputRender()

const source = computed(() => current.value.name || t('tuner.source.default'))

// dB is formatting of the backend's measured level; below the floor there is no honest number.
const level = computed(() => {
  const value = inputLevel.value
  if (!(value > 0)) return t('tuner.level.silent')

  return `${(20 * Math.log10(value)).toFixed(0)} dB`
})
</script>

<style scoped>
.warn {
  color: rgb(var(--v-theme-reed));
  font-weight: 600;
  margin-inline-end: var(--gap);
}

/* Blinking off keeps its place, so the header text beside it does not jump. */
.warn--off {
  visibility: hidden;
}
</style>
