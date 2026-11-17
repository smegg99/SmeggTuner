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
      <ShellDocumentArticle
        ref="article"
        v-model:language="language"
        :kind="kind"
        :title="title"
      />
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
  </v-dialog>
</template>

<script setup lang="ts">
import { useTemplateRef, watch } from 'vue'
import * as BrowserService from '~~bindings/github.com/smegg99/s99wails/v2/services/browser/service.js'
import { documentURL, type DocumentKind } from '~/utils/documents'
import { useDocument } from '~/composables/useDocument'

const { kind } = defineProps<{ kind: DocumentKind }>()
const open = defineModel<boolean>({ required: true })
const { t } = useI18n()
const { language, title, reset } = useDocument(() => kind)
const article = useTemplateRef<{ toTop: () => Promise<void> }>('article')

// Every opening starts at the top, in the interface's language.
watch(open, () => {
  reset()
  void article.value?.toTop()
})

function visit(url: string) {
  void BrowserService.OpenURL(url)
}
</script>
