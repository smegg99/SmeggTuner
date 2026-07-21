<template>
  <SessionDialog
    v-model="creating"
    :saving="busy"
    @submit="onCreate"
  />
  <InstrumentDialog
    v-model="describing"
    :instrument="described"
    :saving="busy"
    @submit="onDescribe"
  />
  <UiConfirmDialog
    v-model="confirmingSessionDelete"
    :title="t('session.deleteTitle')"
    :body="t('session.deleteHint')"
    @confirm="deleteSessionNow"
  />
  <UiConfirmDialog
    v-model="confirmingInstrumentDelete"
    :title="t('instrument.deleteTitle')"
    :body="t('instrument.deleteHint')"
    @confirm="deleteInstrumentNow"
  />
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useInstruments } from '~/composables/useInstruments'
import { useSessions } from '~/composables/useSessions'
import { useShell } from '~/composables/useShell'
import type { InstrumentTemplate } from '~/types/session'
import type { SessionDraft } from '~/components/session/SessionDialog.vue'

const { t } = useI18n()
const { openSession, closeSession } = useShell()
const { busy, create, remove } = useSessions()
const instruments = useInstruments()

// Which dialog is up, as one value: a single value cannot say two are open at once.
type Dialog
  = | { kind: 'create' }
    | { kind: 'describe', instrument: InstrumentTemplate | null }
    | { kind: 'deleteSession', id: string }
    | { kind: 'deleteInstrument', id: string }

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
const describing = openedWhen('describe')
const confirmingSessionDelete = openedWhen('deleteSession')
const confirmingInstrumentDelete = openedWhen('deleteInstrument')

const described = computed(() =>
  dialog.value?.kind === 'describe' ? dialog.value.instrument : null)

function openCreate() {
  dialog.value = { kind: 'create' }
}

function describe(id: string | null) {
  dialog.value = { kind: 'describe', instrument: id ? (instruments.find(id) ?? null) : null }
}

function askDeleteSession(id: string) {
  dialog.value = { kind: 'deleteSession', id }
}

function askDeleteInstrument(id: string) {
  dialog.value = { kind: 'deleteInstrument', id }
}

async function onCreate(draft: SessionDraft) {
  const made = await create(draft)
  if (made) dialog.value = null
}

async function onDescribe(i: InstrumentTemplate) {
  await instruments.save(i)
  dialog.value = null
}

async function deleteSessionNow() {
  if (dialog.value?.kind !== 'deleteSession') return
  const id = dialog.value.id
  dialog.value = null

  if (openSession.value === id) closeSession()
  await remove(id)
}

async function deleteInstrumentNow() {
  if (dialog.value?.kind !== 'deleteInstrument') return
  const id = dialog.value.id
  dialog.value = null
  await instruments.remove(id)
}

defineExpose({ openCreate, describe, askDeleteSession, askDeleteInstrument })
</script>
