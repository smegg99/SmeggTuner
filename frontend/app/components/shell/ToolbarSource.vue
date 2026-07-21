<template>
  <v-menu>
    <template #activator="{ props: menu }">
      <!-- Device/recording names are arbitrary; the key is a declared width and the name ellipsises. -->
      <UiToolKey
        v-bind="menu"
        :icon="current.kind === 'file' ? 'mdi-file-music-outline' : 'mdi-microphone'"
        :label="label"
        :title="running ? t('settings.lockedWhileRunning') : label"
        :disabled="loading || running"
        :fixed="13"
        caret
      />
    </template>

    <UiMenuSheet>
      <UiMenuGroup :label="t('tuner.source.devices')">
        <!-- System default is the empty device id; no enumerated device carries it, so it's a synthetic entry. -->
        <UiMenuItem
          :label="t('tuner.source.default')"
          :active="onDefault"
          @click="selectMic(SYSTEM_DEFAULT)"
        />

        <UiMenuItem
          v-for="device in devices"
          :key="device.id"
          :label="device.name"
          :active="current.kind === 'mic' && current.deviceId === device.id"
          @click="selectMic(device.id)"
        />

        <UiMenuItem
          v-if="!devices.length"
          :label="t('tuner.source.noDevices')"
          disabled
        />
      </UiMenuGroup>

      <UiMenuGroup>
        <UiMenuItem
          icon="mdi-file-music-outline"
          :label="t('tuner.source.analyze')"
          :active="current.kind === 'file'"
          @click="pickFile"
        />
        <UiMenuItem
          icon="mdi-refresh"
          :label="t('tuner.source.refresh')"
          @click="refresh"
        />
      </UiMenuGroup>
    </UiMenuSheet>
  </v-menu>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAudioDevices } from '~/composables/useAudioDevices'
import { useTuner } from '~/composables/useTuner'

// Switching source restarts the engine, so the picker is locked while the tuner runs.

// Empty id means "follow the system default"; synthesised here as no enumerated device carries it.
const SYSTEM_DEFAULT = ''

const { t } = useI18n()
const { devices, current, loading, refresh, selectMic, pickFile } = useAudioDevices()
const { running } = useTuner()

const onDefault = computed(() =>
  current.value.kind === 'mic' && current.value.deviceId === SYSTEM_DEFAULT,
)

// The backend leaves the name empty while following the system default; say so in words.
const label = computed(() => current.value.name || t('tuner.source.default'))
</script>
