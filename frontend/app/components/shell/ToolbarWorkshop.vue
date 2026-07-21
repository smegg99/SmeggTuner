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
      <UiToolGroup
        :model-value="section"
        :items="SECTIONS"
        :label="t('workshop.label')"
        @update:model-value="s => setSection(s as Section)"
      />

      <span class="rule" />

      <!-- size-for keeps the key as wide as the wider of its two labels, so flipping the shelf doesn't resize it. -->
      <UiToolKey
        icon="mdi-plus"
        :label="onShelf ? t('instrument.new') : t('session.new')"
        :size-for="onShelf ? t('session.new') : t('instrument.new')"
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
import { computed } from 'vue'
import type { ToolItem } from '~/components/ui/ToolGroup.vue'
import { useShell } from '~/composables/useShell'
import { useSessions } from '~/composables/useSessions'
import { useWorkshop } from '~/composables/useWorkshop'
import type { Section } from '~/composables/useShell'

defineProps<{
  /** Notes in the recording on screen; print is disabled without them. */
  notes?: number
}>()

const { t } = useI18n()
const { section, setSection, openSession, closeSession, setView } = useShell()
const { busy, close } = useSessions()
const { ask } = useWorkshop()

// Back closes the session: leaving it open would keep swallowing readings with nothing on screen.
async function back() {
  await close()
  closeSession()
}

const SECTIONS = computed<ToolItem[]>(() => [
  { value: 'sessions', label: t('session.title'), icon: 'mdi-folder-music-outline' },
  { value: 'instruments', label: t('instrument.title'), icon: 'mdi-piano' },
])

const onShelf = computed(() => section.value === 'instruments')
</script>

<style scoped>
.rule {
  background: rgb(var(--v-theme-line));
  flex: 0 0 1px;
  height: 2.5cqh;
  margin: 0 0.35cqw;
}
</style>
