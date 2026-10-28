<template>
  <!-- The s99wails error screen loads this frontend too; it gets only the screen, none of the app's services. -->
  <ShellCrashView v-if="crashed" />
  <ShellAppRoot v-else />
</template>

<script setup lang="ts">
import { isCrashWindow } from '~/composables/s99wails'

const crashed = isCrashWindow()
</script>
