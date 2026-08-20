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
  gap: 19.2px;
  padding: 12.8px 0 8px;
  text-align: center;
}

.about__logo {
  block-size: 192px;
  inline-size: auto;
  max-inline-size: 66%;
}

.about__tagline {
  color: rgb(var(--v-theme-ink2));
  font-size: 16.8px;
  font-weight: 500;
  line-height: 1.4;
  margin: 0;
}

.about__publisher {
  color: rgb(var(--v-theme-ink3));
  font-size: 13.6px;
  font-weight: 600;
  margin: -9.6px 0 0;
}

.about__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 33px;
  justify-content: center;
  margin: 0;
}

.about__meta-row {
  align-items: baseline;
  display: flex;
  gap: 8.8px;
}

.about__meta-row dt {
  color: rgb(var(--v-theme-ink3));
  font-size: 12.8px;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.about__meta-row dd {
  color: rgb(var(--v-theme-ink));
  font-family: var(--font-mono);
  font-size: 16px;
  font-variant-numeric: tabular-nums;
  margin: 0;
}

.about__links {
  display: flex;
  flex-wrap: wrap;
  gap: 26.4px;
  justify-content: center;
  padding-block-start: 4.8px;
}

.about__link {
  align-items: center;
  color: rgb(var(--v-theme-accent));
  display: flex;
  font-size: 16px;
  font-weight: 600;
  gap: 6.6px;
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
  font-size: 20.8px;
}
</style>
