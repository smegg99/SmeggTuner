<template>
  <div class="workshop-tools key-row">
    <!-- Open session: the toolbar becomes that session's, rather than stacking a second row. -->
    <template v-if="openSession">
      <v-btn
        :disabled="busy"
        :title="t('workshop.exit')"
        :aria-label="t('workshop.exit')"
        class="tone-key"
        style="--key-tone: rgb(var(--v-theme-reed))"
        prepend-icon="mdi-arrow-left"
        :data-icon="'mdi-arrow-left'"
        @click="back"
      >
        <span class="exit-label">{{ t('workshop.exit') }}</span>
      </v-btn>

      <span class="rule" />

      <!-- Calibrate: the guided sweep, here because it works on this session's instrument. -->
      <UiToolKey
        icon="mdi-tune-vertical"
        :label="t('calibrate.start')"
        :title="t('calibrate.start')"
        @click="setView('calibrate')"
      />

      <span class="rule" />

      <UiToolKey
        icon="mdi-printer-outline"
        :label="t('workshop.print')"
        :disabled="!notes"
        :title="t('workshop.print')"
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
        :title="t('session.new')"
        @click="ask('new')"
      />
      <UiToolKey
        icon="mdi-import"
        :label="t('workshop.import')"
        :title="t('workshop.import')"
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
  if (await close()) closeSession()
}
</script>

<style scoped>
.rule {
  background: rgb(var(--v-theme-line));
  flex: 0 0 1px;
  height: 17.5px;
  margin: 0 var(--gap-tight);
}
@media (max-width: 959px), (max-height: 559px) {
  .rule { display: none; }
  .workshop-tools { width: 100%; justify-content: space-between; }
  .workshop-tools :deep(.key__label), .exit-label { display: none; }
  .workshop-tools :deep(.v-btn__prepend) { margin: 0; }
}
</style>
