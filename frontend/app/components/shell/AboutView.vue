<!-- app/components/shell/AboutView.vue -->
<!-- About as a room of its own in a compact window; see ShellHelpView. The guide it offers is the Help room. -->
<template>
  <ShellRoomPage centred>
    <ShellAboutContent @document="showDocument" />
    <ShellDocumentDialog
      v-model="documentOpen"
      :kind="documentKind"
    />
    <template #footer>
      <ShellAboutLinks />
    </template>
  </ShellRoomPage>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useShell } from '~/composables/useShell'
import type { DocumentKind } from '~/utils/documents'

const { setView } = useShell()
const documentOpen = ref(false)
const documentKind = ref<DocumentKind>('privacy')

function showDocument(kind: DocumentKind) {
  if (kind === 'guide') return void setView('help')
  documentKind.value = kind
  documentOpen.value = true
}
</script>
