<!-- app/components/shell/BootScreen.vue -->
<template>
  <!-- The logo over the goo, as the sound-card calibration draws it, while the app starts; a failure
       replaces the goo and says why. -->
  <main class="boot">
    <ShellAppLogo class="boot__logo" />
    <div class="boot__figure">
      <UiGoo
        v-if="!failure"
        :size="200"
      />
      <v-icon
        v-else
        icon="mdi-alert-circle-outline"
        size="72"
        class="boot__failed"
        :title="failure"
      />
    </div>
    <p
      v-if="failure"
      class="boot__error"
      role="alert"
    >
      <slot />
    </p>
  </main>
</template>

<script setup lang="ts">
// failure is the technical cause, kept in the icon's tip; the slot is what the user reads.
defineProps<{ failure?: string }>()
</script>

<style scoped>
.boot {
  align-content: center;
  background: rgb(var(--v-theme-bg));
  display: grid;
  gap: 12px;
  inset: 0;
  justify-items: center;
  padding: 16px;
  position: fixed;
  z-index: 3000;
}
.boot__figure {
  --s99ab-state-kind: var(--v-theme-accent);
  align-items: center;
  display: flex;
  height: 96px;
  justify-content: center;
}
.boot__logo { height: 40px; }
.boot__failed { color: rgb(var(--v-theme-error)); }
.boot__error { color: rgb(var(--v-theme-error)); font-size: 13px; margin: 0; max-width: 420px; text-align: center; }
</style>
