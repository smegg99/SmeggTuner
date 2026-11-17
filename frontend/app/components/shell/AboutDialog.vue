<!-- frontend/app/components/shell/AboutDialog.vue -->
<template>
  <v-dialog
    v-model="open"
    :aria-label="t('common.about.aboutTitle')"
    content-class="about-dialog"
  >
    <UiDialogCard
      :title="t('common.about.aboutTitle')"
      @close="open = false"
    >
      <ShellAboutContent @document="showDocument" />
      <template #actions>
        <ShellAboutLinks />
      </template>
    </UiDialogCard>
    <ShellDocumentDialog
      v-model="documentOpen"
      :kind="documentKind"
    />
  </v-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { DocumentKind } from '~/utils/documents'

const open = defineModel<boolean>({ required: true })
const { t } = useI18n()
const documentOpen = ref(false)
const documentKind = ref<DocumentKind>('guide')

function showDocument(kind: DocumentKind) {
  documentKind.value = kind
  documentOpen.value = true
}
</script>
