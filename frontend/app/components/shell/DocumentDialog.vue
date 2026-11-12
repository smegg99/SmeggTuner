<!-- frontend/app/components/shell/DocumentDialog.vue -->
<template>
  <v-dialog
    v-model="open"
    :aria-label="title"
    content-class="info-dialog"
  >
    <UiDialogCard
      :title="title"
      @close="open = false"
    >
      <pre
        v-if="kind === 'license'"
        class="document-license"
        tabindex="0"
      >{{ source }}</pre>
      <!-- Static repository Markdown; the renderer escapes HTML and permits only bundled images. -->
      <!-- eslint-disable vue/no-v-html -->
      <article
        v-else
        :key="prefix"
        ref="article"
        class="document"
        :lang="language"
        @click="followLink"
        v-html="html"
      />
      <!-- eslint-enable vue/no-v-html -->
      <template #actions>
        <v-btn
          variant="text"
          prepend-icon="mdi-open-in-new"
          @click="visit(documentURL(kind, language))"
        >
          {{ t('common.about.onlineVersion') }}
        </v-btn>
        <v-btn
          variant="tonal"
          @click="open = false"
        >
          {{ t('common.close') }}
        </v-btn>
      </template>
    </UiDialogCard>
    <v-dialog
      v-model="imageOpen"
      :aria-label="imageTitle"
      content-class="info-dialog"
    >
      <UiDialogCard
        :title="imageTitle"
        @close="imageOpen = false"
      >
        <img
          class="document-image"
          :src="imageURL"
          :alt="imageTitle"
        >
      </UiDialogCard>
    </v-dialog>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import * as BrowserService from '~~bindings/github.com/smegg99/s99wails/v2/services/browser/service.js'
import { documentSource, documentURL, renderDocument, type DocumentKind, type DocumentLocale } from '~/utils/documents'

const { kind } = defineProps<{ kind: DocumentKind }>()
const open = defineModel<boolean>({ required: true })
const { t, locale } = useI18n()
const language = ref<DocumentLocale>(locale.value === 'pl' ? 'pl' : 'en')
const title = computed(() => ({
  guide: t('common.about.guide'),
  privacy: t('common.about.privacy'),
  license: t('common.about.license'),
})[kind])
const prefix = computed(() => `${kind}-${language.value}-`)
const source = computed(() => documentSource(kind, language.value))
const html = computed(() => kind === 'license' ? '' : renderDocument(source.value, prefix.value))
const article = ref<HTMLElement>()
const imageOpen = ref(false)
const imageURL = ref('')
const imageTitle = ref('')

watch([open, locale, () => kind], async () => {
  language.value = locale.value === 'pl' ? 'pl' : 'en'
  imageOpen.value = false
  await nextTick()
  article.value?.closest('.dlg__body')?.scrollTo(0, 0)
})

function visit(url: string) {
  void BrowserService.OpenURL(url)
}

async function followLink(event: MouseEvent) {
  const link = (event.target as Element).closest<HTMLAnchorElement>('a')
  if (!link) return
  event.preventDefault()
  if (link.dataset.documentImage) {
    imageURL.value = link.dataset.documentImage
    imageTitle.value = link.querySelector('img')?.alt || title.value
    imageOpen.value = true
  }
  else if (link.dataset.documentLocale) {
    language.value = link.dataset.documentLocale as DocumentLocale
    await nextTick()
    article.value?.closest('.dlg__body')?.scrollTo(0, 0)
  }
  else if (link.getAttribute('href')?.startsWith('#')) {
    const heading = article.value?.querySelector<HTMLElement>(`[id="${decodeURIComponent(link.hash.slice(1))}"]`)
    const viewport = article.value?.closest<HTMLElement>('.dlg__body')
    if (heading && viewport) {
      const bounds = viewport.getBoundingClientRect()
      const scale = bounds.height / viewport.offsetHeight
      const top = viewport.scrollTop + (heading.getBoundingClientRect().top - bounds.top) / scale - 24
      viewport.scrollTo({ top: Math.max(0, top) })
      heading.focus({ preventScroll: true })
    }
  }
  else visit(link.href)
}
</script>

<style scoped>
.document { color: rgb(var(--v-theme-ink2)); line-height: 1.65; overflow-wrap: anywhere; }
.document :deep(h2), .document :deep(h3) { color: rgb(var(--v-theme-ink)); font-size: 16px; line-height: 1.4; margin: 24px 0 12px; }
.document :deep(h2:first-child), .document :deep(h3:first-child) { margin-top: 0; }
.document :deep(p), .document :deep(ul), .document :deep(ol) { margin: 0 0 12px; }
.document :deep(ul), .document :deep(ol) { padding-inline-start: 24px; }
.document :deep(li) { margin-block: 6px; }
.document :deep(strong) { color: rgb(var(--v-theme-ink)); }
.document :deep(a) { color: rgb(var(--v-theme-accent)); text-underline-offset: 3px; }
.document :deep(img), .document-image { display: block; max-width: 100%; height: auto; }
.document :deep(img) { border: 1px solid rgb(var(--v-theme-line)); border-radius: var(--app-radius-panel); }
.document :deep(code) { font-family: var(--font-mono); font-size: 0.95em; }
.document :deep(.document-table) { max-width: 100%; overflow-x: auto; margin-bottom: 16px; }
.document :deep(table) { width: 100%; border-collapse: collapse; font-size: 12px; }
.document :deep(th), .document :deep(td) { padding: 8px; border: 1px solid rgb(var(--v-theme-lineSoft)); text-align: start; vertical-align: top; }
.document :deep(th) { color: rgb(var(--v-theme-ink)); background: rgb(var(--v-theme-chrome2)); }
.document :deep(a:focus-visible), .document :deep([tabindex]:focus-visible) { outline: 2px solid rgb(var(--v-theme-accent)); outline-offset: 2px; }
.document-license { white-space: pre-wrap; overflow-wrap: anywhere; font: 12px/1.6 var(--font-mono); margin: 0; }
</style>
