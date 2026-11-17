<!-- frontend/app/components/shell/DocumentArticle.vue -->
<!-- One bundled document as text: the guide and the privacy policy rendered from Markdown, the license as it is.
     It scrolls inside whatever holds it, a dialog's body or a room's. -->
<template>
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
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import * as BrowserService from '~~bindings/github.com/smegg99/s99wails/v2/services/browser/service.js'
import { documentSource, renderDocument, type DocumentKind, type DocumentLocale } from '~/utils/documents'

const { kind, title } = defineProps<{
  kind: DocumentKind
  /** names a picture that has no caption of its own */
  title: string
}>()
/** The language the document is read in: the interface's, until its own language links change it. */
const language = defineModel<DocumentLocale>('language', { required: true })

const prefix = computed(() => `${kind}-${language.value}-`)
const source = computed(() => documentSource(kind, language.value))
const html = computed(() => kind === 'license' ? '' : renderDocument(source.value, prefix.value))
const article = ref<HTMLElement>()
const imageOpen = ref(false)
const imageURL = ref('')
const imageTitle = ref('')

// The scrolling box around the document: a dialog's body or a room's.
const viewport = () => article.value?.closest<HTMLElement>('.dlg__body, .room__body')

async function toTop() {
  imageOpen.value = false
  await nextTick()
  viewport()?.scrollTo(0, 0)
}
watch([language, () => kind], toTop)
defineExpose({ toTop })

async function followLink(event: MouseEvent) {
  const link = (event.target as Element).closest<HTMLAnchorElement>('a')
  if (!link) return
  event.preventDefault()
  if (link.dataset.documentImage) {
    imageURL.value = link.dataset.documentImage
    imageTitle.value = link.querySelector('img')?.alt || title
    imageOpen.value = true
  }
  else if (link.dataset.documentLocale) {
    language.value = link.dataset.documentLocale as DocumentLocale
  }
  else if (link.getAttribute('href')?.startsWith('#')) {
    const heading = article.value?.querySelector<HTMLElement>(`[id="${decodeURIComponent(link.hash.slice(1))}"]`)
    const box = viewport()
    if (heading && box) {
      const bounds = box.getBoundingClientRect()
      const scale = bounds.height / box.offsetHeight
      const top = box.scrollTop + (heading.getBoundingClientRect().top - bounds.top) / scale - 24
      box.scrollTo({ top: Math.max(0, top) })
      heading.focus({ preventScroll: true })
    }
  }
  else void BrowserService.OpenURL(link.href)
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
/* Under a finger a link is taller than its line; inline padding adds that without moving the text. */
@media (max-width: 959px), (max-height: 559px) {
  .document :deep(a) { padding-block: 8px; }
}
/* A portrait phone has no width for columns: each row is an entry, its leading cells the heading line and
   its last cell the text under it, at the document's own size. The column names say nothing then. */
@media (max-width: 599px) {
  .document :deep(.document-table) { overflow-x: visible; }
  .document :deep(table) { font-size: inherit; }
  .document :deep(table), .document :deep(tbody), .document :deep(tr) { display: block; }
  .document :deep(thead) { display: none; }
  .document :deep(tr) { padding-block: 10px; border-top: 1px solid rgb(var(--v-theme-lineSoft)); }
  .document :deep(td) { display: inline; padding: 0; border: 0; }
  .document :deep(td:not(:last-child)) { color: rgb(var(--v-theme-ink)); font-weight: 600; margin-inline-end: 8px; }
  .document :deep(td:last-child) { display: block; margin-top: 4px; }
}
</style>
