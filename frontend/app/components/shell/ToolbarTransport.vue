<template>
  <!-- The microphone's run key. It stays, disabled, while a recording is the source, so the bar never shifts;
       a recording plays from its own card (FileRun). Icon-only so the key keeps a constant width. -->
  <v-btn
    :disabled="starting || file"
    :title="runTitle"
    :aria-label="runTitle"
    class="tone-key"
    :style="{ '--key-tone': `rgb(var(--v-theme-${stopping ? 'reed' : 'goal'}))` }"
    :icon="stopping ? 'mdi-stop' : 'mdi-play'"
    :data-icon="stopping ? 'mdi-stop' : 'mdi-play'"
    @click="running ? stop() : begin()"
  />

  <!-- Asked once per device and run: a card never calibrated measures, but uncorrected. -->
  <v-dialog
    v-model="asking"
    :aria-label="t('settings.calibration.askTitle')"
    width="var(--dlg-sm)"
  >
    <UiDialogCard
      :title="t('settings.calibration.askTitle')"
      @close="asking = false"
    >
      <p class="ask">
        {{ t('settings.calibration.askBody') }}
      </p>
      <template #actions>
        <v-btn
          @click="continueUncorrected"
        >
          {{ t('settings.calibration.askContinue') }}
        </v-btn><v-btn
          color="accent"
          @click="calibrateNow"
        >
          {{ t('settings.calibration.askCalibrate') }}
        </v-btn>
      </template>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useTuner } from '~/composables/useTuner'
import { useAudioDevices } from '~/composables/useAudioDevices'
import { useShell } from '~/composables/useShell'
import { useSoundCard } from '~/composables/useSoundCard'

// The engine's state is stated in one place; this is it for the microphone.
const { t } = useI18n()
const { running, starting, start, stop } = useTuner()
const { current } = useAudioDevices()
const { setView } = useShell()
const { needsAsking, startCalibration } = useSoundCard()

const file = computed(() => current.value.kind === 'file')
// A running microphone offers Stop; anything else shows Play, a recording's disabled.
const stopping = computed(() => running.value && !file.value)
const runTitle = computed(() => {
  if (file.value) return t('tuner.transport.fileHint')
  return running.value ? t('tuner.transport.stop') : t('tuner.transport.start')
})
const asking = ref(false)

async function begin() {
  if (await needsAsking()) {
    asking.value = true
    return
  }
  await start()
}

function continueUncorrected() {
  asking.value = false
  void start()
}

async function calibrateNow() {
  asking.value = false
  if (await setView('soundcard')) await startCalibration()
}
</script>

<style scoped>
.ask {
  color: rgb(var(--v-theme-ink2));
  font-size: 11.5px;
  line-height: 1.5;
  margin: 0;
}
</style>
