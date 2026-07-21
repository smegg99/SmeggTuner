<template>
  <v-app v-cloak>
    <ShellAppWindow />
  </v-app>
</template>

<script setup lang="ts">
import * as LoggerService from '~~bindings/github.com/smegg99/s99wails/services/logger/service.js'
import { installFrontendLogBridge } from '~/composables/s99wails'
import { useConfigSync } from '~/composables/useConfigSync'
import { useThemeSync } from '~/composables/useThemeSync'

// The window is mounted above the router so a route change never tears down a running engine's subscriptions; rooms are switched by state, not routes.
useConfigSync()
useThemeSync()

// Routes uncaught errors and unhandled rejections to the Go log; installed for the window's life, a no-op outside the desktop.
installFrontendLogBridge(LoggerService)
</script>

<style>
[v-cloak] {
  display: none;
}

/* v-app lays itself out around app bars and drawers. There are none. */
.v-application__wrap {
  min-height: 0;
}
</style>
