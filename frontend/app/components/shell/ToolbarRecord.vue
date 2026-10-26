<template>
  <div class="key-row">
    <!-- Warm-up: live but unlit until armed. Never lit and disabled at once; see recordKeys. -->
    <v-btn
      :disabled="keys.light.disabled"
      :title="t(HINTS[keys.light.hint])"
      :aria-label="t(HINTS[keys.light.hint])"
      :aria-pressed="keys.light.lit"
      class="tone-key"
      style="--key-tone: rgb(var(--v-theme-reed))"
      :icon="'mdi-record'"
      :data-icon="'mdi-record'"
      @click="toggleRecording(!armed)"
    />

    <UiToolKey
      icon="mdi-undo"
      :disabled="keys.undo.disabled"
      :title="t('record.undo.hint')"
      @click="undo"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { recordKeys } from '~/utils/record'
import type { RecordHint } from '~/utils/record'
import { useRecord } from '~/composables/useRecord'

// Recording needs a session (a reading must live somewhere); tuning doesn't. Live-state logic: utils/record.
const { t } = useI18n()
const { sessionId, readings, armed, busy, undo, toggleRecording } = useRecord()

// Warm-up gets its own hint, not the sessionless "open a session to record".
const HINTS: Record<RecordHint, string> = {
  recording: 'record.arm.hint',
  warmup: 'record.arm.warmup',
  noSession: 'record.arm.needsSession',
}

const keys = computed(() => recordKeys({
  session: sessionId.value !== '',
  armed: armed.value,
  readings: readings.value,
  busy: busy.value,
}))
</script>
