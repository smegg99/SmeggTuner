<!-- frontend/app/components/shell/AboutContent.vue -->
<!-- Who made the app, which build this is and what there is to read: the body of the About dialog and of
     the About room a compact window has. -->
<template>
  <div class="about">
    <div class="about__identity">
      <ShellAppMascot
        class="about__mascot"
        :label="t('common.about.mascot')"
      />
      <div class="about__intro">
        <ShellAppLogo
          class="about__logo"
          :title="t('common.app.name')"
        />
        <p>{{ t('common.about.tagline') }}</p>
        <a
          class="about__author"
          href="https://smegg.me"
          :aria-label="t('common.about.publisher')"
          :title="t('common.about.publisher')"
          @click.prevent="visit('https://smegg.me')"
        >
          <Smegg99Logo aria-hidden="true" />
        </a>
      </div>
    </div>
    <dl class="about__meta">
      <div><dt>{{ t('common.about.version') }}</dt><dd>{{ version || t('common.about.untagged') }}</dd></div>
      <div v-if="commit">
        <dt>{{ t('common.about.commit') }}</dt><dd>{{ commit }}</dd>
      </div>
    </dl>
    <div class="about__documents">
      <v-btn
        v-for="item in documents"
        :key="item.kind"
        class="about__document"
        variant="text"
        :prepend-icon="item.icon"
        :data-document="item.kind"
        @click="emit('document', item.kind)"
      >
        {{ item.label }}
      </v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import * as BrowserService from '~~bindings/github.com/smegg99/s99wails/v2/services/browser/service.js'
import type { DocumentKind } from '~/utils/documents'
import Smegg99Logo from '../../../../assets/smegg99-wordmark.svg'

const emit = defineEmits<{ document: [kind: DocumentKind] }>()
const { t } = useI18n()
const config = useRuntimeConfig()
const version = config.public.version
const commit = config.public.commit
const documents = computed<{ kind: DocumentKind, icon: string, label: string }[]>(() => [
  { kind: 'guide', icon: 'mdi-book-open-page-variant-outline', label: t('common.about.guide') },
  { kind: 'privacy', icon: 'mdi-shield-account-outline', label: t('common.about.privacy') },
  { kind: 'license', icon: 'mdi-scale-balance', label: t('common.about.license') },
])

function visit(url: string) {
  void BrowserService.OpenURL(url)
}
</script>

<style scoped>
.about { display: grid; gap: 24px; width: 100%; margin: 0; text-align: center; }
.about__identity { display: grid; justify-items: center; gap: 16px; }
.about__mascot { --app-mascot-h: 160px; }
.about__intro { display: grid; justify-items: center; gap: 12px; min-width: 0; }
.about__logo { --app-logo-h: 30px; }
.about__intro p { color: rgb(var(--v-theme-ink2)); margin: 0; }
.about a { color: rgb(var(--v-theme-accent)); text-decoration: none; }
.about a:hover { text-decoration: underline; }
.about a:focus-visible { outline: 2px solid rgb(var(--v-theme-accent)); outline-offset: 4px; }
.about .about__author { display: inline-flex; color: rgb(var(--v-theme-ink)); }
.about__author svg { display: block; width: 136px; height: 24px; }
.about__author :deep([fill="#fff" i]), .about__author :deep([fill="#ffffff" i]) { fill: currentColor; }
.about__meta { display: flex; flex-wrap: wrap; justify-content: center; gap: 8px 24px; margin: 0; }
.about__meta div { display: flex; gap: 8px; align-items: baseline; }
.about__meta dt { color: rgb(var(--v-theme-ink3)); font-size: 12px; }
.about__meta dd { margin: 0; font-family: var(--font-mono); }
.about__documents { display: flex; flex-wrap: wrap; justify-content: center; gap: 12px; }
.about__document { --s99ab-control-h: 36px; --gap: 8px; max-width: 100%; }
.about__document :deep(.v-btn__content) { white-space: normal; line-height: 1.35; }
@media (max-height: 600px) { .about { gap: 16px; } .about__identity { gap: 12px; } .about__mascot { --app-mascot-h: 128px; } }

/* A compact window is worked with a finger: the documents are keys of the control height, one per row. The
   page is a screen, not a scroll: the mascot is as tall as the window's height leaves for it. */
@media (max-width: 959px), (max-height: 559px) {
  .about { gap: 16px; }
  .about__mascot { --app-mascot-h: clamp(64px, calc(100vh - 560px), 160px); }
  .about__documents { display: grid; grid-template-columns: minmax(0, 1fr); gap: var(--gap); }
  /* What to read comes first there; which build this is closes the page. */
  .about__meta { order: 1; }
  .about__document {
    --s99ab-control-h: inherit;
    justify-content: start;
    border: 1px solid rgb(var(--v-theme-edge));
    background: rgb(var(--v-theme-raised));
  }
}
/* A phone on its side has the width for two columns: who made it, then what there is to read. */
@media (max-height: 559px) and (min-width: 560px) {
  .about { grid-template-columns: repeat(2, minmax(0, 1fr)); align-items: center; gap: 8px 24px; }
  .about__identity { gap: 12px; }
  .about__mascot { --app-mascot-h: clamp(56px, calc(100vh - 300px), 160px); }
  .about__identity { grid-area: 1 / 1 / span 2; }
  .about__documents { grid-area: 1 / 2; align-self: end; }
  .about__meta { grid-area: 2 / 2; align-self: start; }
}
</style>
