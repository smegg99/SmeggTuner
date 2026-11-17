<!-- app/components/settings/SettingsView.vue -->
<template>
  <!-- Untitled: the main bar names the room. -->
  <UiPanel
    class="settings tuner-panel"
    flush
  >
    <!-- Grouped by what each setting changes: the input, how a note is found, how a reading is shown, the app's
         look. The columns wrap to two, then one, as the window narrows; a compact window lets the sections
         flow into columns of even height instead, so none ends over an empty half. -->
    <div class="settings__body">
      <div class="settings__sections">
        <div class="settings__column">
          <UiFormSection
            :title="t('settings.groups.microphone')"
            class="tuner-section settings__section"
          >
            <SettingsAudioDeviceSelect />
            <SettingsAudioFilterToggles />
            <SettingsAudioSoundCardCalibration v-if="!WEB" />
            <SettingsAudioAutoVolumeToggle v-if="!WEB" />
          </UiFormSection>
          <!-- A compact window's bar has no room for the output level; it is set here. -->
          <UiFormSection
            v-if="compact && !WEB"
            :title="t('settings.groups.output')"
            class="tuner-section settings__section"
          >
            <SettingsAudioOutputLevel />
          </UiFormSection>
          <UiFormSection
            :title="t('settings.groups.detection')"
            class="tuner-section settings__section"
          >
            <SettingsTunerDetectionToggle />
            <SettingsTunerTransposeField />
            <SettingsTunerReadingToggles />
          </UiFormSection>
        </div>
        <div class="settings__column">
          <UiFormSection
            :title="t('settings.groups.readings')"
            class="tuner-section settings__section"
          >
            <SettingsTunerToleranceFields />
            <SettingsTunerUnitToggle />
            <SettingsTunerBeatUnitToggle />
            <!-- A target is a session's goal curve; without the workshop there is none to read against. -->
            <SettingsTunerErrorReferenceToggle v-if="WORKSHOP" />
            <SettingsTunerScaleNamingSelect />
          </UiFormSection>
          <UiFormSection
            v-if="WORKSHOP"
            :title="t('settings.groups.report')"
            class="tuner-section settings__section"
          >
            <SettingsReportLetterhead />
          </UiFormSection>
        </div>
        <div class="settings__column">
          <UiFormSection
            :title="t('settings.groups.appearance')"
            class="tuner-section settings__section"
          >
            <SettingsAppearanceThemeButtonGroup />
            <SettingsAppearanceAccentControl />
            <SettingsAppearanceLanguageSelect />
            <SettingsAppearanceTrayToggle v-if="!WEB" />
            <SettingsAppearanceUpdateToggle v-if="!WEB" />
            <!-- Only a debug build has any: utils/debugSettings.ts. -->
            <component
              :is="debugSettings"
              v-if="debugSettings"
            />
          </UiFormSection>
          <LazyWebDiagnostics v-if="WEB" />
        </div>
      </div>
    </div>
  </UiPanel>
</template>

<script setup lang="ts">
import { debugSettings } from '~/utils/debugSettings'
import { WEB, WORKSHOP } from '~/utils/features'

const { t } = useI18n()
const compact = useCompactLayout()
</script>

<style scoped>
.settings { min-height: 0; min-width: 0; }
.settings__body {
  min-height: 0;
  padding: 16px;
  overflow-y: auto;
}
.settings__sections {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 280px), 1fr));
  align-items: start;
  gap: 24px;
}
.settings__column { display: flex; flex-direction: column; gap: 24px; min-width: 0; }
.settings__section { min-width: 0; min-height: 0; }
/* A select names one of a few things: it does not stretch across a wide column. */
.settings__section :deep(.v-select) { max-width: 404px; }

@media (max-width: 959px), (max-height: 559px) {
  .settings__body { padding: 12px; }
  .settings__sections { display: block; columns: 320px; column-gap: 24px; }
  .settings__column { display: contents; }
  .settings__section { break-inside: avoid; margin: 0 0 24px; }
}
</style>
