<!-- app/components/shell/StatusActions.vue -->
<template>
  <!-- The window bar's own keys: the downloads page in the web edition, Help and About. -->
  <div class="status__actions">
    <a
      v-if="WEB"
      class="status__action"
      :href="DOWNLOADS_URL"
      target="_blank"
      rel="noopener"
      :title="t('web.downloads.title')"
    >
      <v-icon
        icon="mdi-download-outline"
        class="status__action-icon"
      />
      <span class="status__action-label">{{ t('web.downloads.title') }}</span>
    </a>
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
      <span class="status__action-label">{{ t('about.helpTitle') }}</span>
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
      <span class="status__action-label">{{ t('about.aboutTitle') }}</span>
    </button>

    <ShellHelpDialog v-model="helpOpen" />
    <ShellAboutDialog v-model="aboutOpen" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { DOWNLOADS_URL, WEB } from '~/utils/features'

const { t } = useI18n()
const aboutOpen = ref(false)
const helpOpen = ref(false)
</script>

<style scoped>
.status__actions {
  align-items: center;
  display: flex;
  flex: 0 0 auto;
  gap: var(--gap-tight);
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
  font-size: 11px;
  gap: var(--gap-tight);
  line-height: 1;
  padding: var(--gap);
  text-decoration: none;
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
  font-size: 14px;
}

@media (max-width: 959px), (max-height: 559px) {
  .status__action-label { display: none; }
  .status__action { padding: 4px 8px; }
}
</style>
