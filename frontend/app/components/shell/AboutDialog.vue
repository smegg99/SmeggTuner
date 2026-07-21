<template>
  <v-dialog
    v-model="open"
    width="var(--dlg-sm)"
  >
    <UiDialogCard
      :title="t('about.aboutTitle')"
      @close="open = false"
    >
      <div class="about">
        <SmeggTunerStacked
          class="about__logo"
          role="img"
          :aria-label="t('about.appName')"
        />

        <p class="about__tagline">
          {{ t('about.tagline') }}
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
            class="about__link"
            :href="REPO"
            @click.prevent="visit(REPO)"
          >
            <v-icon
              icon="mdi-github"
              class="about__link-icon"
            />
            GitHub
          </a>
          <a
            class="about__link"
            :href="SITE"
            @click.prevent="visit(SITE)"
          >
            <v-icon
              icon="mdi-web"
              class="about__link-icon"
            />
            {{ t('common.website') }}
          </a>
          <a
            class="about__link"
            :href="KOFI"
            @click.prevent="visit(KOFI)"
          >
            <v-icon
              icon="mdi-coffee-outline"
              class="about__link-icon"
            />
            {{ t('about.donate') }}
          </a>
        </div>
      </div>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import SmeggTunerStacked from '~/assets/smeggtuner-stacked.svg'
import * as BrowserService from '~~bindings/github.com/smegg99/s99wails/services/browser/service.js'

// External links open in the system browser via s99wails, never inside the webview.
const REPO = 'https://github.com/smegg99/SmeggTuner'
const SITE = 'https://smeggtuner.com'
const KOFI = 'https://ko-fi.com/smegg99'

const { t } = useI18n()

const open = defineModel<boolean>({ required: true })

// The build stamps these (nuxt.config): the nearest tag, and the short commit under it.
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
  gap: 2.4cqh;
  padding: 1.6cqh 0 1cqh;
  text-align: center;
}

.about__logo {
  block-size: 24cqh;
  inline-size: auto;
  max-inline-size: 66%;
}

.about__tagline {
  color: rgb(var(--v-theme-ink2));
  font-size: 2.1cqh;
  font-weight: 500;
  line-height: 1.4;
  margin: 0;
}

.about__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 1cqh 3cqw;
  justify-content: center;
  margin: 0;
}

.about__meta-row {
  align-items: baseline;
  display: flex;
  gap: 0.8cqw;
}

.about__meta-row dt {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.6cqh;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.about__meta-row dd {
  color: rgb(var(--v-theme-ink));
  font-family: var(--font-mono);
  font-size: 2cqh;
  font-variant-numeric: tabular-nums;
  margin: 0;
}

.about__links {
  display: flex;
  flex-wrap: wrap;
  gap: 2.4cqw;
  justify-content: center;
  padding-block-start: 0.6cqh;
}

.about__link {
  align-items: center;
  color: rgb(var(--v-theme-accent));
  display: flex;
  font-size: 2cqh;
  font-weight: 600;
  gap: 0.6cqw;
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
  font-size: 2.6cqh;
}
</style>
