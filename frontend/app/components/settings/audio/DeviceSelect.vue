<template>
  <UiFormField
    v-slot="{ id }"
    :label="t('settings.audioDeviceLabel')"
    :hint="running ? t('settings.lockedWhileRunning') : t('settings.audioDeviceHint')"
    :error="errorMessage"
  >
    <div class="device">
      <UiSelectInput
        :id="id"
        v-model="deviceId"
        :items="deviceItems"
        :disabled="loading || running"
      />

      <UiToolKey
        icon="mdi-refresh"
        :disabled="loading || running"
        :title="t('settings.audioDeviceRefresh')"
        @click="refresh"
      />
    </div>
  </UiFormField>
</template>

<script setup lang="ts">
import type { SelectItem } from '~/components/ui/SelectInput.vue'
import { computed } from 'vue'
import { useAudioDevices } from '~/composables/useAudioDevices'
import type { AudioDeviceView } from '~/composables/useAudioDevices'
import { useConfigSync } from '~/composables/useConfigSync'
import { useTuner } from '~/composables/useTuner'

// Empty ID means "system default" (services/audio, and a fresh config); no enumerated device
// carries it, so it is synthesized here.
const SYSTEM_DEFAULT = ''

const { t } = useI18n()
const { config } = useConfigSync()
const { devices, current, loading, error, refresh, selectMic } = useAudioDevices()
// Locked while running: switching device restarts the run.
const { running } = useTuner()

const items = computed<AudioDeviceView[]>(() => [
  { id: SYSTEM_DEFAULT, name: t('settings.audioDeviceSystemDefault'), default: false },
  ...devices.value,
])

const errorMessage = computed(() => (error.value ? t(error.value) : undefined))

// Shows what the backend holds, not the config: a saved-but-unplugged device (rejected at startup)
// isn't shown active; the config stands in only while a recording is selected.
const deviceId = computed<string>({
  get: () => (current.value.kind === 'mic'
    ? current.value.deviceId
    : config.audio?.device_id ?? SYSTEM_DEFAULT),
  set: (id: string) => {
    void select(id)
  },
})

// selectMic leaves `current` untouched if the device is gone, so persisting `current` (not the
// requested ID) keeps a dead device out of the config.
async function select(id: string) {
  await selectMic(id)
  if (current.value.kind === 'mic') {
    config.audio.device_id = current.value.deviceId
  }
}

const deviceItems = computed<SelectItem[]>(() =>
  items.value.map(device => ({ value: device.id, label: device.name })),
)
</script>

<style scoped>
.device {
  align-items: center;
  display: flex;
  gap: 0.5cqw;
  min-width: 0;
}

.device :deep(.select) {
  flex: 1 1 auto;
}
</style>
