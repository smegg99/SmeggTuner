<template>
  <div class="d-flex align-center ga-1">
    <!-- Open session: the toolbar becomes that session's, rather than stacking a second row. -->
    <template v-if="openSession">
      <UiToolKey
        icon="mdi-arrow-left"
        :label="t('workshop.exit')"
        tone="error"
        :disabled="busy"
        @click="back"
      />

      <span class="rule" />

      <!-- Calibrate: the guided sweep, here because it works on this session's instrument. -->
      <UiToolKey
        icon="mdi-tune-vertical"
        :label="t('calibrate.start')"
        @click="setView('calibrate')"
      />

      <span class="rule" />

      <UiToolKey
        icon="mdi-printer-outline"
        :label="t('workshop.print')"
        :disabled="!notes"
        @click="ask('print')"
      />
      <UiToolKey
        icon="mdi-export"
        :title="t('session.export')"
        @click="ask('export')"
      />
    </template>

    <template v-else>
      <UiToolKey
        icon="mdi-plus"
        :label="t('session.new')"
        @click="ask('new')"
      />
      <UiToolKey
        icon="mdi-import"
        :label="t('workshop.import')"
        @click="ask('import')"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { useShell } from '~/composables/useShell'
import { useSessions } from '~/composables/useSessions'
import { useWorkshop } from '~/composables/useWorkshop'

defineProps<{
  /** Notes in the recording on screen; print is disabled without them. */
  notes?: number
}>()

const { t } = useI18n()
const { openSession, closeSession, setView } = useShell()
const { busy, close } = useSessions()
const { ask } = useWorkshop()

// Back closes the session: leaving it open would keep swallowing readings with nothing on screen.
async function back() {
  await close()
  closeSession()
}
</script>

<style scoped>
.rule {
  background: rgb(var(--v-theme-line));
  flex: 0 0 1px;
  height: 20px;
  margin: 0 3.8px;
}
</style>
