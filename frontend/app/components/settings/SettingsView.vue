<template>
  <UiPanelCard class="settings">
    <template #head>
      <span class="panel__title">{{ t('routes.settings') }}</span>
      <span class="panel__sp" />
    </template>

    <!-- Fixed 3-column grid, not auto-fit: columns are declared so groups can be placed to
         balance heights and nothing scrolls. -->
    <div class="settings__body">
      <!-- Audio (tall) paired with the short Behaviour toggles so this column doesn't overflow. -->
      <div class="settings__col">
        <UiFieldGroup
          dense
          :title="t('settings.groups.audio')"
        >
          <SettingsAudioDeviceSelect />
          <SettingsAudioClockPPMField />
          <SettingsAudioFilterToggles />
        </UiFieldGroup>

        <UiFieldGroup
          dense
          :title="t('settings.groups.behaviour')"
        >
          <SettingsTunerBehaviourToggles />
          <SettingsAppearanceTrayToggle />
        </UiFieldGroup>
      </div>

      <!-- Reading is the heavy group, so it gets a column to itself. -->
      <div class="settings__col">
        <UiFieldGroup
          dense
          :title="t('settings.groups.reading')"
        >
          <SettingsTunerUnitToggle />
          <SettingsTunerErrorReferenceToggle />
          <SettingsTunerScaleNamingSelect />
          <SettingsTunerToleranceFields />
        </UiFieldGroup>
      </div>

      <div class="settings__col">
        <UiFieldGroup
          dense
          :title="t('settings.groups.report')"
        >
          <SettingsReportLetterhead />
        </UiFieldGroup>

        <UiFieldGroup
          dense
          :title="t('settings.groups.appearance')"
        >
          <SettingsAppearanceThemeButtonGroup />
          <SettingsAppearanceAccentControl />
          <SettingsAppearanceLanguageSelect />
        </UiFieldGroup>
      </div>
    </div>
  </UiPanelCard>
</template>

<script setup lang="ts">
const { t } = useI18n()
</script>

<style scoped>
.settings {
  min-height: 0;
  min-width: 0;
}

/* Fixed three columns (not auto-fit) so groups can be placed to balance the columns rather than
   wrapping wherever the width breaks. */
.settings__body {
  align-items: start;
  display: grid;
  gap: 12.8px 24.2px;
  grid-template-columns: repeat(3, 1fr);
  min-height: 0;
  /* Safety net for long translations; the layout is built to fit and this should never show. */
  overflow-y: auto;
  padding: 8px 17.6px;
}

.settings__col {
  display: flex;
  flex-direction: column;
  gap: 22.4px;
  min-width: 0;
}
</style>
