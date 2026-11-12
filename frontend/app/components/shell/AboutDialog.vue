<!-- frontend/app/components/shell/AboutDialog.vue -->
<template>
  <v-dialog
    v-model="open"
    :aria-label="t('common.about.aboutTitle')"
    width="var(--dlg-md)"
  >
    <UiDialogCard
      :title="t('common.about.aboutTitle')"
      @close="open = false"
    >
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
            @click="showDocument(item.kind)"
          >
            {{ item.label }}
          </v-btn>
        </div>
      </div>
      <template #actions>
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
    </UiDialogCard>
    <ShellDocumentDialog
      v-model="documentOpen"
      :kind="documentKind"
    />
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import * as BrowserService from '~~bindings/github.com/smegg99/s99wails/v2/services/browser/service.js'
import type { DocumentKind } from '~/utils/documents'
import Smegg99Logo from '../../../../assets/smegg99-wordmark.svg'

const open = defineModel<boolean>({ required: true })
const { t } = useI18n()
const config = useRuntimeConfig()
const version = config.public.version
const commit = config.public.commit
const documentOpen = ref(false)
const documentKind = ref<DocumentKind>('guide')
const documents = computed<{ kind: DocumentKind, icon: string, label: string }[]>(() => [
  { kind: 'guide', icon: 'mdi-book-open-page-variant-outline', label: t('common.about.guide') },
  { kind: 'privacy', icon: 'mdi-shield-account-outline', label: t('common.about.privacy') },
  { kind: 'license', icon: 'mdi-scale-balance', label: t('common.about.license') },
])
const links = computed(() => [
  { url: 'https://smeggtuner.com', icon: 'mdi-web', label: t('common.website') },
  { url: 'https://github.com/smegg99/SmeggTuner', icon: 'mdi-github', label: 'GitHub' },
  { url: 'mailto:support@smegg.me', icon: 'mdi-email-outline', label: t('common.about.support') },
  { url: 'https://ko-fi.com/smegg99', icon: 'mdi-coffee-outline', label: t('common.about.donate') },
])

function showDocument(kind: DocumentKind) {
  documentKind.value = kind
  documentOpen.value = true
}

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
.about a, .about__links a { color: rgb(var(--v-theme-accent)); text-decoration: none; }
.about a:hover, .about__links a:hover { text-decoration: underline; }
.about a:focus-visible, .about__links a:focus-visible { outline: 2px solid rgb(var(--v-theme-accent)); outline-offset: 4px; }
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
.about__links { display: flex; flex-wrap: wrap; justify-content: center; gap: 12px 24px; width: 100%; }
.about__links a { display: inline-flex; align-items: center; gap: 6px; font-size: 12px; }
@media (max-height: 600px) { .about { gap: 16px; } .about__identity { gap: 12px; } .about__mascot { --app-mascot-h: 128px; } }
</style>
