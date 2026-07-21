<template>
  <UiFormField
    :label="t('settings.accent.label')"
    :hint="t('settings.accent.hint')"
  >
    <div class="accent">
      <UiToolGroup
        :model-value="accent"
        :items="accentItems"
        :label="t('settings.accent.label')"
        @update:model-value="onSelect"
      />

      <button
        v-if="accent === ACCENT_MODE.CUSTOM"
        type="button"
        class="accent__pick"
        :style="{ background: accentColor || '#000000' }"
        :title="t('settings.accent.pick')"
        @click="openPicker"
      />
    </div>
  </UiFormField>

  <v-dialog
    v-model="open"
    width="var(--dlg-sm)"
  >
    <UiDialogCard
      :title="t('settings.accent.pickTitle')"
      :confirm-label="t('common.accept')"
      @close="open = false"
      @cancel="open = false"
      @confirm="accept"
    >
      <!-- Bound to a draft; nothing on the app changes until Accept, Cancel drops it. -->
      <UiColorPicker
        :model-value="draft"
        :presets="ACCENT_PRESETS"
        @update:model-value="draft = $event"
      />
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { ToolItem } from '~/components/ui/ToolGroup.vue'
import { useThemeSync } from '~/composables/useThemeSync'
import { ACCENT_MODE, ACCENT_MODES } from '~/types/config'

const { t } = useI18n()
const { accent, accentColor, setAccentMode, setAccentColor } = useThemeSync()

const open = ref(false)

// Seeded from the live accent on open, committed only on Accept, so dragging previews in the picker.
const draft = ref('')

const ACCENT_PRESETS = [
  '#ef4444', '#f97316', '#f59e0b', '#eab308',
  '#22c55e', '#10b981', '#14b8a6', '#06b6d4',
  '#3b82f6', '#6366f1', '#8b5cf6', '#ec4899',
]

function openPicker() {
  draft.value = accentColor.value || '#000000'
  open.value = true
}

function accept() {
  if (draft.value) setAccentColor(draft.value.slice(0, 7)) // #rrggbb, drop any alpha
  open.value = false
}

function onSelect(value: unknown) {
  const next = ACCENT_MODES.find(item => item.value === value)
  if (next) setAccentMode(next.value)
}

const accentItems = computed<ToolItem[]>(() =>
  ACCENT_MODES.map(m => ({ value: m.value, label: t(m.labelKey), icon: m.icon })),
)
</script>

<style scoped>
.accent {
  align-items: center;
  display: flex;
  gap: 1cqw;
}

/* Square swatch at 3.9cqh to match UiToolKey height. */
.accent__pick {
  aspect-ratio: 1;
  block-size: 3.9cqh;
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  cursor: pointer;
  flex: 0 0 auto;
}
</style>
