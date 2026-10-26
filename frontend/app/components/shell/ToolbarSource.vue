<template>
  <v-menu
    location="bottom start"
    :min-width="125"
  >
    <template #activator="{ props: menu }">
      <!-- Device/recording names are arbitrary; the key is a declared width and the name ellipsises. -->
      <v-btn
        v-bind="menu"
        class="source-key"
        :title="running ? t('settings.lockedWhileRunning') : label"
        :disabled="loading || running"
        :aria-label="running ? t('settings.lockedWhileRunning') : label"
        :width="125"
        :prepend-icon="current.kind === 'file' ? 'mdi-file-music-outline' : 'mdi-microphone'"
        :data-icon="current.kind === 'file' ? 'mdi-file-music-outline' : 'mdi-microphone'"
        append-icon="mdi-menu-down"
      >
        <span class="source-key__label">{{ label }}</span>
      </v-btn>
    </template>

    <v-list class="tuner-menu">
      <div>
        <!-- System default is the empty device id; no enumerated device carries it, so it's a synthetic entry. -->
        <v-list-item
          :active="onDefault"
          :append-icon="(onDefault) ? 'mdi-check' : undefined"
          :title="t('tuner.source.default')"
          @click="selectMic(SYSTEM_DEFAULT)"
        />

        <v-list-item
          v-for="device in devices"
          :key="device.id"
          :active="current.kind === 'mic' && current.deviceId === device.id"
          :append-icon="(current.kind === 'mic' && current.deviceId === device.id) ? 'mdi-check' : undefined"
          :title="device.name"
          @click="selectMic(device.id)"
        />

        <v-list-item
          v-if="!devices.length"
          disabled
          :title="t('tuner.source.noDevices')"
        />
      </div>

      <div>
        <v-list-item
          v-if="!WEB"
          :active="current.kind === 'file'"
          :append-icon="(current.kind === 'file') ? 'mdi-check' : undefined"
          :title="t('tuner.source.analyze')"
          prepend-icon="mdi-file-music-outline"
          @click="pickFile"
        />
        <v-list-item
          :title="t('tuner.source.refresh')"
          prepend-icon="mdi-refresh"
          @click="refresh"
        />
      </div>
    </v-list>
  </v-menu>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAudioDevices } from '~/composables/useAudioDevices'
import { useTuner } from '~/composables/useTuner'
import { WEB } from '~/utils/features'

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

<style scoped>
.source-key :deep(.v-btn__content) {
  min-width: 0;
}

.source-key__label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
