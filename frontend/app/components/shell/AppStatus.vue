<template>
  <div class="status">
    <ShellAppLogo class="status__logo" />

    <!-- Rides the window bar so it survives room switches. See useSessionProgress. -->
    <ShellSessionLabel
      class="status__session"
      :progress="progress"
    />

    <span class="status__spacer" />

    <!-- Just the tag; commit and full build stamp are in the About dialog. -->
    <span
      v-if="version"
      class="status__ver"
    >{{ version }}</span>

    <div class="status__actions">
      <button
        type="button"
        class="status__action"
        :title="t('about.helpTitle')"
        @click="helpOpen = true"
      >
        <v-icon
          icon="mdi-help-circle-outline"
          class="status__action-icon"
        />
        {{ t('about.helpTitle') }}
      </button>
      <button
        type="button"
        class="status__action"
        :title="t('about.aboutTitle')"
        @click="aboutOpen = true"
      >
        <v-icon
          icon="mdi-information-outline"
          class="status__action-icon"
        />
        {{ t('about.aboutTitle') }}
      </button>
    </div>

    <ShellHelpDialog v-model="helpOpen" />
    <ShellAboutDialog v-model="aboutOpen" />
  </div>
</template>

<script setup lang="ts">
// Belongs to the window, not a view, so it survives switching between rooms.
import { ref } from 'vue'
import { useSessionProgress } from '~/composables/useSessionProgress'

const { t } = useI18n()

const progress = useSessionProgress()

// The build stamps this (nuxt.config): the nearest tag. The commit is in the About dialog.
const config = useRuntimeConfig()
const version = config.public.version

const aboutOpen = ref(false)
const helpOpen = ref(false)
</script>

<style scoped>
.status {
  align-items: center;
  border-top: 1px solid rgb(var(--v-theme-lineSoft));
  display: flex;
  gap: 1.2cqw;
  padding: 0.9cqh 1.1cqw;
}

.status__logo {
  flex: 0 0 auto;
}

/* Yields width first: when tight, the session name truncates before the keys get pushed off. */
.status__session {
  flex: 0 1 auto;
  min-width: 0;
}

.status__spacer {
  flex: 1 1 auto;
}

.status__ver {
  color: rgb(var(--v-theme-ink3));
  flex: 0 0 auto;
  font-family: var(--font-mono);
  font-size: 1.4cqh;
  font-variant-numeric: tabular-nums;
}

.status__actions {
  align-items: center;
  display: flex;
  flex: 0 0 auto;
  gap: 0.4cqw;
}

/* Opens a panel, not a link: ink, not accent. */
.status__action {
  align-items: center;
  background: none;
  border: 0;
  border-radius: 2px;
  color: rgb(var(--v-theme-ink2));
  cursor: pointer;
  display: flex;
  font-size: 1.6cqh;
  gap: 0.4cqw;
  line-height: 1;
  padding: 0.5cqh 0.7cqw;
}

.status__action:hover {
  background: rgb(var(--v-theme-chrome2));
  color: rgb(var(--v-theme-ink));
}

.status__action:focus-visible {
  outline: 2px solid rgb(var(--v-theme-ink2));
  outline-offset: 1px;
}

.status__action-icon {
  font-size: 2cqh;
}
</style>
