<template>
  <!-- Shown while the microphone is not calibrated, or is being calibrated: it asks for the calibration. -->
  <UiToolKey
    v-if="shown"
    class="sound-card-key"
    :style="{ '--sound-card-tone': `rgb(var(--v-theme-${tone}))` }"
    :icon="running ? 'mdi-clock-outline' : 'mdi-clock-alert-outline'"
    :title="t(running ? 'tuner.soundcard.running' : `tuner.soundcard.${status}`)"
    @click="setView('soundcard')"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useShell } from '~/composables/useShell'
import { useSoundCard } from '~/composables/useSoundCard'
import { calibrationTone } from '~/utils/soundCard'

const { t } = useI18n()
const { setView } = useShell()
const { calibration } = useSoundCard()

const running = computed(() => calibration.value?.running ?? false)
const status = computed(() => (calibration.value?.status === 'unreliable' ? 'unreliable' : 'uncalibrated'))
const tone = computed(() => calibrationTone(running.value ? 'calibrating' : status.value))
// The microphone's own state, shown whatever the source, so switching to a recording moves nothing.
const shown = computed(() => calibration.value !== null
  && (running.value || ['uncalibrated', 'unreliable'].includes(calibration.value.status)))
</script>

<style scoped>
.sound-card-key :deep(.v-icon) {
  color: var(--sound-card-tone);
}
</style>
