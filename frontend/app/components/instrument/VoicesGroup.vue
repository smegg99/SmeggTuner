<!-- app/components/instrument/VoicesGroup.vue -->
<template>
  <!-- The treble registers' section of the accordion form: what the switches say, then the switches. -->
  <UiFormSection
    :title="t('instrument.sections.voices')"
    class="tuner-section"
  >
    <template #head>
      <span class="form-section__title section-hint">
        {{ t('instrument.sections.voices') }}
        <v-tooltip
          activator="parent"
          location="top start"
          max-width="340"
          open-delay="300"
        >
          {{ t('instrument.registersHint') }}
        </v-tooltip>
      </span>
    </template>
    <p class="acc__says">
      {{ banks.length ? t('instrument.says', { banks: banks.join(' '), reeds: reedCount }, reedCount) : t('instrument.registersSubtitle') }}
    </p>

    <InstrumentRegisterSection
      v-model="registers"
      :mode="mode"
    />
  </UiFormSection>
</template>

<script setup lang="ts">
import type { Register } from '~/types/session'

const registers = defineModel<Register[]>({ required: true })
defineProps<{ mode: 'all' | 'list' | 'builder', banks: string[], reedCount: number }>()
const { t } = useI18n()
</script>

<style scoped>
/* One line, reserved even when empty, so describing switches never reflows the pane below. */
.acc__says {
  color: rgb(var(--v-theme-ink2));
  font-size: 11px;
  height: 14.5px;
  line-height: 14.5px;
  margin: 0;
  overflow: hidden;
}
</style>
