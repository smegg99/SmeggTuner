<template>
  <SessionDialog
    v-model="creating"
    :saving="busy"
    @submit="onCreate"
  />
  <UiConfirmDialog
    v-model="confirmingSessionDelete"
    :title="t('session.deleteTitle')"
    :body="t('session.deleteHint')"
    @confirm="deleteSessionNow"
  />
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useSessions } from '~/composables/useSessions'
import { useShell } from '~/composables/useShell'
import type { SessionDraft } from '~/components/session/SessionDialog.vue'

const { t } = useI18n()
const { openSession, openSessionAt, closeSession } = useShell()
const { busy, create, remove } = useSessions()

// Which dialog is up, as one value: a single value cannot say two are open at once.
type Dialog
  = | { kind: 'create' }
    | { kind: 'deleteSession', id: string }

const dialog = ref<Dialog | null>(null)

function openedWhen(kind: Dialog['kind']) {
  return computed({
    get: () => dialog.value?.kind === kind,
    set: (open: boolean) => {
      if (!open) dialog.value = null
    },
  })
}

const creating = openedWhen('create')
const confirmingSessionDelete = openedWhen('deleteSession')

function openCreate() {
  dialog.value = { kind: 'create' }
}

function askDeleteSession(id: string) {
  dialog.value = { kind: 'deleteSession', id }
}

async function onCreate(draft: SessionDraft) {
  const made = await create(draft)
  if (!made) return

  dialog.value = null
  // Go already adopted it (services/session/sessions.go, Create -> adopt); the room has to follow,
  // or the session that is open records into a shelf nobody is looking at.
  openSessionAt(made.id)
}

async function deleteSessionNow() {
  if (dialog.value?.kind !== 'deleteSession') return
  const id = dialog.value.id
  dialog.value = null

  if (openSession.value === id) closeSession()
  await remove(id)
}

defineExpose({ openCreate, askDeleteSession })
</script>
