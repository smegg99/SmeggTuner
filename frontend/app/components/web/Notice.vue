<!-- app/components/web/Notice.vue -->
<template>
  <v-dialog
    v-model="open"
    width="var(--dlg-md)"
    :aria-label="t('web.notice.title')"
  >
    <UiDialogCard
      :title="t('web.notice.title')"
      @close="close"
    >
      <ul class="notice">
        <li>{{ t('web.notice.processing') }}</li>
        <li>{{ t('web.notice.background') }}</li>
        <li>{{ t('web.notice.native') }}</li>
      </ul>
      <template #actions>
        <v-btn
          variant="text"
          prepend-icon="mdi-download-outline"
          :href="DOWNLOADS_URL"
          target="_blank"
          rel="noopener"
        >
          {{ t('web.downloads.title') }}
        </v-btn>
        <v-btn
          color="primary"
          @click="close"
        >
          {{ t('web.notice.ok') }}
        </v-btn>
      </template>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import { DOWNLOADS_URL } from '~/utils/features'
import { markNoticeSeen, noticeSeen } from '~/web/notice'

const { t } = useI18n()
// Opens after the boot screen's cross-fade (WebApp.vue), not on top of it.
const FADE_MS = 1000
const open = ref(false)
let timer: ReturnType<typeof setTimeout> | undefined
onMounted(() => {
  timer = setTimeout(() => {
    open.value = !noticeSeen()
  }, FADE_MS)
})
onBeforeUnmount(() => clearTimeout(timer))

function close() {
  markNoticeSeen()
  open.value = false
}
</script>

<style scoped>
.notice { display: grid; gap: 8px; margin: 0; padding-left: 18px; }
</style>
