<!-- app/components/settings/SettingsView.vue -->
<template>
  <!-- Untitled: the main bar names the room. -->
  <UiPanel
    class="settings tuner-panel"
    flush
  >
    <!-- Grouped by what each setting changes: the input, how a note is found, how a reading is shown, the app's
         look. The columns wrap to two, then one, as the window narrows. -->
    <div class="settings__body">
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
        </UiFormSection>
        <LazyWebDiagnostics v-if="WEB" />
      </div>
    </div>
  </UiPanel>
</template>

<script setup lang="ts">
import { WEB, WORKSHOP } from '~/utils/features'

const { t } = useI18n()
</script>

<style scoped>
.settings { min-height: 0; min-width: 0; }
.settings__body {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 280px), 1fr));
  align-items: start;
  gap: 24px;
  min-height: 0;
  padding: 16px;
  overflow-y: auto;
}
.settings__column { display: flex; flex-direction: column; gap: 24px; min-width: 0; }
.settings__section { min-width: 0; min-height: 0; }
</style>
