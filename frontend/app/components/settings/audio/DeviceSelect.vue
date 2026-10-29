<!-- app/components/settings/audio/DeviceSelect.vue -->
<template>
  <UiFormField
    v-slot="{ id }"
    :label="t('settings.audioDeviceLabel')"
    :hint="running ? t('settings.lockedWhileRunning') : t('settings.audioDeviceHint')"
    :error="errorMessage"
  >
    <div class="device">
      <v-select
        :id="id"
        v-model="deviceId"
        v-model:menu="open"
        :items="deviceItems"
        :disabled="running || (loading && !open)"
        item-title="label"
        item-value="value"
        hide-details="auto"
        :aria-label="t('settings.audioDeviceLabel')"
        :list-props="{ class: 'tuner-menu' }"
        @update:menu="onOpen"
      >
        <template #item="{ props: itemProps, item }">
          <v-list-item
            v-bind="itemProps"
            :disabled="loading"
            :append-icon="(item.value === deviceId) ? 'mdi-check' : undefined"
          />
        </template>
      </v-select>

      <UiToolKey
        icon="mdi-refresh"
        :disabled="loading || running"
        :title="t('settings.audioDeviceRefresh')"
        @click="browse"
      />
    </div>
  </UiFormField>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { WEB } from '~/utils/features'
import { useAudioDevices } from '~/composables/useAudioDevices'
import type { AudioDeviceView } from '~/composables/useAudioDevices'
import { useConfigSync } from '~/composables/useConfigSync'
import { useTuner } from '~/composables/useTuner'

// Empty ID means "system default" (services/audio, and a fresh config); no enumerated device
// carries it, so it is synthesized here.
const SYSTEM_DEFAULT = ''

const { t } = useI18n()
const { config } = useConfigSync()
const { devices, current, loading, error, browse, selectMic } = useAudioDevices()
// Locked while running: switching device restarts the run.
const { running } = useTuner()
const open = ref(false)

function onOpen(value: boolean) {
  if (value && WEB) void browse()
}

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

const deviceItems = computed(() =>
  items.value.map(device => ({ value: device.id, label: device.name })),
)
</script>

<style scoped>
.device {
  align-items: center;
  display: flex;
  gap: var(--gap);
  min-width: 0;
}

.device :deep(.select) {
  flex: 1 1 auto;
}
</style>
