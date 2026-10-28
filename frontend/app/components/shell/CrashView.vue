<!-- app/components/shell/CrashView.vue -->
<template>
  <v-app>
    <UiCrashScreen
      v-if="report"
      :report="report"
      :actions="crashActions"
      @title="retitle"
    >
      <template #logo>
        <ShellAppLogo class="crash__logo" />
      </template>
      <template #figure>
        <ShellAppDeadCat />
      </template>
    </UiCrashScreen>
  </v-app>
</template>

<script setup lang="ts">
import { useTheme } from 'vuetify'
import { useAppearance } from '@smegg99/s99ab/appearance'
import { crashActions, onCrashReport, type CrashReport } from '~/composables/s99wails'
import { applyInstrumentTheme, TOKENS } from '~/theme/tokens'

// No config service runs here, so the screen follows the system scheme and the language the report carries.
const vuetify = useTheme()
const dark = typeof matchMedia === 'function' && matchMedia('(prefers-color-scheme: dark)').matches
for (const name of ['light', 'dark'] as const) {
  applyInstrumentTheme(vuetify.themes.value[name]!, name, TOKENS[name].accent)
}
useAppearance(() => ({ theme: 'smeggtuner.theme' })).setMode(dark ? 'dark' : 'light', undefined, 'none')

const { locale, setLocale } = useI18n()
useHead(() => ({ htmlAttrs: { lang: locale.value } }))
const report = ref<CrashReport>()
function retitle(phrase: string) {
  void crashActions.setTitle(phrase)
}
// A hang's report changes while the screen is open: it gains the stacks once the app is closed.
const stopReports = onCrashReport((next) => {
  report.value = next
})
onBeforeUnmount(stopReports)
onMounted(async () => {
  const loaded = await crashActions.report()
  if ((loaded.lang === 'pl' || loaded.lang === 'en') && loaded.lang !== locale.value) await setLocale(loaded.lang)
  report.value ??= loaded
})
</script>

<style scoped>
.crash__logo { --app-logo-h: 15px; }
</style>
