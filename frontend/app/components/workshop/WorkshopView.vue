<template>
  <UiPanelCard class="shop">
    <template #head>
      <span class="panel__title">{{ headTitle }}</span>
      <span class="panel__sp" />
      <span class="panel__meta">{{ headMeta }}</span>
    </template>

    <SessionDetail
      v-if="openSession"
      :id="openSession"
      :src="activeSrc"
    />

    <UiShelf
      v-else-if="section === 'instruments'"
      :empty="instruments.empty.value"
      icon="mdi-piano"
      :lead="t('instrument.empty')"
      :hint="t('instrument.emptyHint')"
    >
      <InstrumentCard
        v-for="i in instruments.list.value"
        :key="i.id"
        :instrument="i"
        :src="instruments.imageOf(i)"
        @edit="id => dialogs?.describe(id)"
        @photo="instruments.setImage"
        @export="exportInstrument"
        @delete="id => dialogs?.askDeleteInstrument(id)"
      />
    </UiShelf>

    <UiShelf
      v-else
      :empty="!list.length"
      icon="mdi-folder-music-outline"
      :lead="t('session.empty')"
      :hint="t('session.emptyHint')"
    >
      <SessionCard
        v-for="s in list"
        :key="s.id"
        :session="s"
        :src="srcOf(s.instrumentId)"
        @open="openIt"
        @export="id => exportSession(id)"
        @delete="id => dialogs?.askDeleteSession(id)"
      />
    </UiShelf>

    <WorkshopDialogs ref="dialogs" />
  </UiPanelCard>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import WorkshopDialogs from './WorkshopDialogs.vue'
import { useInstruments } from '~/composables/useInstruments'
import { useRecord } from '~/composables/useRecord'
import { useSessions } from '~/composables/useSessions'
import { useShell } from '~/composables/useShell'
import { useWorkshop } from '~/composables/useWorkshop'

const { t } = useI18n()
const { section, openSession, openSessionAt } = useShell()
const { seq, last } = useWorkshop()
const { list, active, load, open, importFile, exportFile } = useSessions()
// Aliased: `load` is already this room's session list.
const { load: loadRecorded } = useRecord()
const instruments = useInstruments()

const dialogs = ref<InstanceType<typeof WorkshopDialogs> | null>(null)

// Load on mount, not top-level await: awaiting in setup suspends the room on re-entry and flashes the shelf.
onMounted(async () => {
  await load()
  await instruments.load()
  await loadRecorded()
})

const headTitle = computed(() => {
  if (openSession.value) return active.value?.name ?? t('session.title')
  return section.value === 'instruments' ? t('instrument.title') : t('session.title')
})

const headMeta = computed(() => {
  if (openSession.value) return ''
  return String(section.value === 'instruments' ? instruments.list.value.length : list.value.length)
})

// A session's photo lives with its instrument; a deleted instrument just means no picture.
function srcOf(instrumentId?: string): string {
  const i = instrumentId ? instruments.find(instrumentId) : undefined
  return i ? instruments.imageOf(i) : ''
}

const activeSrc = computed(() => srcOf(active.value?.instrumentId))

// The shared toolbar raises a request; what it means depends on the current shelf.
watch(seq, () => {
  const onShelf = section.value === 'instruments'

  switch (last.value) {
    case 'new':
      if (onShelf) dialogs.value?.describe(null)
      else dialogs.value?.openCreate()
      break

    case 'import':
      if (onShelf) void instruments.importFile()
      else void importFile()
      break

    case 'export':
      if (openSession.value) exportSession(openSession.value)
      break
  }
})

async function openIt(id: string) {
  await open(id)
  openSessionAt(id)
}

function exportSession(id: string) {
  const s = list.value.find(x => x.id === id)
  if (s) exportFile(s.id, s.name)
}

function exportInstrument(id: string) {
  const i = instruments.find(id)
  if (i) instruments.exportFile(i)
}
</script>

<style scoped>
.shop {
  min-height: 0;
  min-width: 0;
}
</style>
