<!-- app/components/shell/BootScreen.vue -->
<template>
  <!-- S99AB's boot screen with the app's logo; a failure turns it into the error screen in the same place. -->
  <UiCrashScreen
    v-if="failure"
    :report="{ kind: 'startup', app: name, error: failure, text: failure }"
    :actions="{ restart: reload }"
    :message="message"
  >
    <template #logo>
      <ShellAppLogo class="boot__status-logo" />
    </template>
    <template #figure>
      <ShellAppDeadCat />
    </template>
  </UiCrashScreen>
  <UiBootScreen
    v-else
    :app="name"
  >
    <template #logo>
      <ShellAppLogo class="boot__logo" />
    </template>
  </UiBootScreen>
</template>

<script setup lang="ts">
// failure is the technical cause; message is what the user reads instead of the generic startup text.
defineProps<{ failure?: string, message?: string }>()

const { t } = useI18n()
const name = computed(() => t('app.name'))

async function reload() {
  window.location.reload()
}
</script>

<style scoped>
.boot__logo { height: 40px; }
.boot__status-logo { --app-logo-h: 15px; }
</style>
