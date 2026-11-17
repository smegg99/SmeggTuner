<!-- app/components/shell/HelpView.vue -->
<!-- The guide as a room of its own: a compact window has no place for a dialog over it, and a room is
     left the way every other one is, by its neighbour's key. -->
<template>
  <ShellRoomPage>
    <template #action>
      <UiToolKey
        icon="mdi-open-in-new"
        :label="t('common.about.onlineVersion')"
        @click="visit"
      />
    </template>
    <ShellDocumentArticle
      v-model:language="language"
      kind="guide"
      :title="title"
    />
  </ShellRoomPage>
</template>

<script setup lang="ts">
import * as BrowserService from '~~bindings/github.com/smegg99/s99wails/v2/services/browser/service.js'
import { useDocument } from '~/composables/useDocument'
import { documentURL } from '~/utils/documents'

const { t } = useI18n()
const { language, title } = useDocument('guide')

function visit() {
  void BrowserService.OpenURL(documentURL('guide', language.value))
}
</script>
