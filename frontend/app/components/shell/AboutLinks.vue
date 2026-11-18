<!-- frontend/app/components/shell/AboutLinks.vue -->
<!-- Where the app lives outside itself: the foot of the About dialog and of the About room. -->
<template>
  <div class="about__links">
    <a
      v-for="link in links"
      :key="link.url"
      :href="link.url"
      @click.prevent="visit(link.url)"
    >
      <v-icon
        :icon="link.icon"
        size="16"
      />
      {{ link.label }}
    </a>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import * as BrowserService from '~~bindings/github.com/smegg99/s99wails/v2/services/browser/service.js'

const { t, locale } = useI18n()
const links = computed(() => [
  { url: locale.value === 'en' ? 'https://smeggtuner.com/' : `https://smeggtuner.com/${locale.value}/`, icon: 'mdi-web', label: t('common.website') },
  { url: 'https://github.com/smegg99/SmeggTuner', icon: 'mdi-github', label: 'GitHub' },
  { url: 'mailto:support@smegg.me', icon: 'mdi-email-outline', label: t('common.about.support') },
  { url: 'https://ko-fi.com/smegg99', icon: 'mdi-coffee-outline', label: t('common.about.donate') },
])

function visit(url: string) {
  void BrowserService.OpenURL(url)
}
</script>

<style scoped>
.about__links { display: flex; flex-wrap: wrap; justify-content: center; gap: 12px 24px; width: 100%; }
.about__links a { display: inline-flex; align-items: center; gap: 6px; color: rgb(var(--v-theme-accent)); font-size: 12px; text-decoration: none; }
.about__links a:hover { text-decoration: underline; }
.about__links a:focus-visible { outline: 2px solid rgb(var(--v-theme-accent)); outline-offset: 4px; }
/* A compact window is worked with a finger: each link is a key of the control height, apart from the next. */
@media (max-width: 959px), (max-height: 559px) {
  .about__links { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: var(--gap); }
  .about__links a {
    justify-content: center;
    min-height: var(--s99ab-control-h);
    border: 1px solid rgb(var(--v-theme-edge));
    border-radius: var(--s99ab-radius);
    background: rgb(var(--v-theme-raised));
    font-size: 13px;
  }
  .about__links a:hover { text-decoration: none; }
}
@media (max-height: 559px) and (min-width: 560px) {
  .about__links { grid-template-columns: repeat(4, minmax(0, 1fr)); }
}
</style>
