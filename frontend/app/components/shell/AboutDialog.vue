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
        <ShellAppMascot
          class="about__mascot"
          :label="t('common.about.mascot')"
        />
        <h1 class="about__title">
          <ShellAppLogo
            class="about__logo"
            :title="t('common.app.name')"
          />
        </h1>

        <p class="about__tagline">
          {{ t('common.about.tagline') }}
        </p>
        <p class="about__publisher">
          {{ t('common.about.publisher') }}
        </p>

        <dl class="about__meta">
          <div class="about__meta-row">
            <dt>{{ t('common.about.version') }}</dt>
            <dd>{{ version || t('common.about.untagged') }}</dd>
          </div>
          <div
            v-if="commit"
            class="about__meta-row"
          >
            <dt>{{ t('common.about.commit') }}</dt>
            <dd>{{ commit }}</dd>
          </div>
        </dl>

        <div class="about__links">
          <div
            v-for="(row, index) in linkRows"
            :key="index"
            class="about__link-row"
          >
            <a
              v-for="link in row"
              :key="link.url"
              class="about__link"
              :href="link.url"
              :aria-label="link.url === authorUrl ? t('common.about.authorWebsite') : undefined"
              @click.prevent="visit(link.url)"
            >
              <AuthorMark
                v-if="link.url === authorUrl"
                class="about__author-mark"
                aria-hidden="true"
              />
              <v-icon
                v-else
                :icon="link.icon"
                class="about__link-icon"
              />
              {{ link.label }}
            </a>
            <v-dialog
              v-if="index === linkRows.length - 1"
              v-model="licenseOpen"
              :aria-label="t('common.about.license')"
              width="640"
            >
              <template #activator="{ props: licenseActivator }">
                <button
                  v-bind="licenseActivator"
                  type="button"
                  class="about__link"
                >
                  <v-icon
                    icon="mdi-scale-balance"
                    class="about__link-icon"
                  />
                  {{ t('common.about.license') }}
                </button>
              </template>
              <UiDialogCard
                :title="t('common.about.license')"
                @close="licenseOpen = false"
              >
                <pre
                  class="license-text"
                  tabindex="0"
                  role="region"
                  :aria-label="t('common.about.license')"
                >{{ licenseText }}</pre>
              </UiDialogCard>
            </v-dialog>
          </div>
        </div>
      </div>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import * as BrowserService from '~~bindings/github.com/smegg99/s99wails/v2/services/browser/service.js'
import AuthorMark from '../../../../assets/smegg99.svg'
import licenseText from '../../../../LICENSE?raw'

const { t } = useI18n()
const licenseOpen = ref(false)
const authorUrl = 'https://smegg.me'

// External links open in the system browser via s99wails, never inside the webview.
const linkRows = computed(() => [[
  { url: 'https://qr2g.it/1R', icon: 'mdi-github', label: 'GitHub' },
  { url: 'https://linkedin.smegg.me', icon: 'mdi-linkedin', label: 'LinkedIn' },
  { url: authorUrl, icon: '', label: 'smegg.me' },
], [
  { url: 'https://smeggtuner.com', icon: 'mdi-web', label: t('common.website') },
  { url: 'https://ko-fi.com/smegg99', icon: 'mdi-coffee-outline', label: t('common.about.donate') },
]])

const open = defineModel<boolean>({ required: true })

// The build stamps these (nuxt.config): VERSION and the current short commit.
const config = useRuntimeConfig()
const version = config.public.version
const commit = config.public.commit

function visit(url: string) {
  void BrowserService.OpenURL(url)
}
</script>

<style scoped>
.about {
  align-items: center;
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: var(--gap);
  text-align: center;
}

.about__mascot {
  --app-mascot-h: 128px;
}

.about__title {
  margin: 0;
  line-height: 0;
}

.about__logo {
  --app-logo-h: 22px;
}

.about__tagline {
  color: rgb(var(--v-theme-ink2));
  font-size: 14.5px;
  font-weight: 500;
  line-height: 1.4;
  margin: 0;
}

.about__author-mark {
  flex-shrink: 0;
  height: 18px;
  width: 18px;
}

.license-text {
  font-family: var(--font-mono);
  font-size: 12px;
  margin: 0;
  max-height: min(480px, calc(100dvh - 160px));
  overflow: auto;
  overflow-wrap: anywhere;
  overscroll-behavior: contain;
  scrollbar-gutter: stable;
  white-space: pre-wrap;
}

.license-text:focus-visible {
  outline: 2px solid rgb(var(--v-theme-accent));
  outline-offset: 2px;
}

.about__publisher {
  color: rgb(var(--v-theme-ink3));
  font-size: 12px;
  font-weight: 600;
  margin: 0;
}

.about__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--gap) var(--gap-wide);
  justify-content: center;
  margin: 0;
}

.about__meta-row {
  align-items: baseline;
  display: flex;
  gap: var(--gap);
}

.about__meta-row dt {
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.about__meta-row dd {
  color: rgb(var(--v-theme-ink));
  font-family: var(--font-mono);
  font-size: 14px;
  font-variant-numeric: tabular-nums;
  margin: 0;
}

.about__links {
  display: grid;
  gap: 12px;
  padding-block-start: var(--gap);
}

.about__link-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  justify-content: center;
}

.about__link {
  align-items: center;
  background: none;
  border: 0;
  cursor: pointer;
  color: rgb(var(--v-theme-accent));
  display: flex;
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  gap: var(--gap);
  line-height: inherit;
  padding: 0;
  text-decoration: none;
}

.about__link:hover {
  filter: brightness(1.15);
}

.about__link:focus-visible {
  outline: 2px solid rgb(var(--v-theme-accent));
  outline-offset: 2px;
}

.about__link-icon {
  font-size: 18px;
}
</style>
