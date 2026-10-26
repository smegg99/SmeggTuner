<!-- frontend/app/components/shell/AboutDialog.vue -->
<template>
  <v-dialog
    v-model="open"
    :aria-label="t('about.aboutTitle')"
    width="var(--dlg-md)"
  >
    <UiDialogCard
      :title="t('about.aboutTitle')"
      @close="open = false"
    >
      <div class="about">
        <ShellAppMascot
          class="about__mascot"
          :label="t('about.mascot')"
        />
        <h1 class="about__title">
          <ShellAppLogo
            class="about__logo"
            :title="t('about.appName')"
          />
        </h1>

        <p class="about__tagline">
          {{ t('about.tagline') }}
        </p>
        <p class="about__publisher">
          {{ t('about.publisher') }}
        </p>

        <dl class="about__meta">
          <div class="about__meta-row">
            <dt>{{ t('about.version') }}</dt>
            <dd>{{ version || t('about.untagged') }}</dd>
          </div>
          <div
            v-if="commit"
            class="about__meta-row"
          >
            <dt>{{ t('about.commit') }}</dt>
            <dd>{{ commit }}</dd>
          </div>
        </dl>

        <div class="about__links">
          <a
            v-for="link in links"
            :key="link.url"
            class="about__link"
            :href="link.url"
            @click.prevent="visit(link.url)"
          >
            <v-icon
              :icon="link.icon"
              class="about__link-icon"
            />
            {{ link.label }}
          </a>
        </div>
      </div>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import * as BrowserService from '~~bindings/github.com/smegg99/s99wails/v2/services/browser/service.js'

const { t } = useI18n()

// External links open in the system browser via s99wails, never inside the webview.
const links = computed(() => [
  { url: 'https://qr2g.it/1R', icon: 'mdi-github', label: 'GitHub' },
  { url: 'https://linkedin.smegg.me', icon: 'mdi-linkedin', label: 'LinkedIn' },
  { url: 'https://smeggtuner.com', icon: 'mdi-web', label: t('common.website') },
  { url: 'https://ko-fi.com/smegg99', icon: 'mdi-coffee-outline', label: t('about.donate') },
])

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
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  justify-content: center;
  padding-block-start: var(--gap);
}

.about__link {
  align-items: center;
  color: rgb(var(--v-theme-accent));
  display: flex;
  font-size: 14px;
  font-weight: 600;
  gap: var(--gap);
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
