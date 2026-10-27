<template>
  <SessionDialog
    v-model="creating"
    :saving="busy"
    @submit="onCreate"
  />
  <v-dialog
    v-model="confirmingSessionDelete"
    :aria-label="t('session.deleteTitle')"
    width="var(--dlg-sm)"
  >
    <UiDialogCard
      :title="t('session.deleteTitle')"
      @close="confirmingSessionDelete = false"
    >
      <p class="dialog-warn">
        {{ t('session.deleteHint') }}
      </p>
      <template #actions>
        <v-btn
          @click="confirmingSessionDelete = false"
        >
          {{ t('common.cancel') }}
        </v-btn><v-btn
          color="error"
          @click="deleteSessionNow"
        >
          {{ t('common.delete') }}
        </v-btn>
      </template>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useSessions } from '~/composables/useSessions'
import { useShell } from '~/composables/useShell'
import type { SessionDraft } from '~/types/session'

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
  // Go already adopted it (backend/services/session/sessions.go, Create -> adopt); the room has to follow,
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
