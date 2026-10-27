<template>
  <UiPanel
    class="shop tuner-panel"
    :class="{ 'shop--session': compact && openSession }"
    flush
  >
    <template #head>
      <span class="panel__title">{{ headTitle }}</span>
      <span class="panel__sp" />
      <span class="panel__meta">{{ headMeta }}</span>
    </template>

    <SessionDetail
      v-if="openSession"
      :id="openSession"
      :key="openSession"
    />

    <UiShelf
      v-else
      ref="shelf"
      :class="{ shop__paged: compact }"
      :empty="!list.length"
      icon="mdi-folder-music-outline"
      :lead="t('session.empty')"
      :hint="t('session.emptyHint')"
    >
      <SessionCard
        v-for="s in shownSessions"
        :key="s.id"
        :session="s"
        :src="srcOf(s.instrumentId)"
        @open="openIt"
        @export="id => exportSession(id)"
        @delete="id => dialogs?.askDeleteSession(id)"
      />
    </UiShelf>

    <WorkshopDialogs ref="dialogs" />
    <template
      v-if="compact && !openSession && list.length"
      #footer
    >
      <UiPageStepper
        v-model="page"
        :count="pages"
        :label="t('session.title')"
        class="shop__pages"
      />
    </template>
  </UiPanel>
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
const { openSession, openSessionAt } = useShell()
const { seq, last } = useWorkshop()
const { list, active, load, open, importFile, exportFile } = useSessions()
// Aliased: `load` is already this room's session list.
const { load: loadRecorded } = useRecord()
const instruments = useInstruments()
const compact = usePagedLayout()
const landscape = useLandscapeLayout()
const shelf = ref<{ $el: HTMLElement } | null>(null)
const shelfElement = computed(() => shelf.value?.$el ?? null)
const pageSize = usePageSize(shelfElement, () => landscape.value ? 102 : 152)
const page = ref(0)
const pages = computed(() => Math.max(1, Math.ceil(list.value.length / pageSize.value)))
watch(pages, (count) => {
  page.value = Math.min(page.value, count - 1)
})
const shownSessions = computed(() => compact.value
  ? list.value.slice(page.value * pageSize.value, (page.value + 1) * pageSize.value)
  : list.value)

const dialogs = ref<InstanceType<typeof WorkshopDialogs> | null>(null)

// Load on mount, not top-level await: awaiting in setup suspends the room on re-entry and flashes the shelf.
// The accordions come too: the cards show their photographs and the new-session picker lists them.
onMounted(async () => {
  await load()
  await instruments.load()
  await loadRecorded()
})

const headTitle = computed(() =>
  openSession.value ? (active.value?.name ?? t('session.title')) : t('session.title'),
)

const headMeta = computed(() => (openSession.value ? '' : String(list.value.length)))

// A session's photo lives with its accordion; a forgotten accordion just means no picture.
function srcOf(instrumentId?: string): string {
  const i = instrumentId ? instruments.find(instrumentId) : undefined
  return i ? instruments.imageOf(i) : ''
}

// The shared toolbar raises a request; this room is the only shelf now.
watch(seq, () => {
  switch (last.value) {
    case 'new':
      dialogs.value?.openCreate()
      break

    case 'import':
      void importFile()
      break

    case 'export':
      if (openSession.value) exportSession(openSession.value)
      break
  }
})

async function openIt(id: string) {
  if (await open(id)) openSessionAt(id)
}

function exportSession(id: string) {
  const s = list.value.find(x => x.id === id)
  if (s) exportFile(s.id, s.name)
}
</script>

<style scoped>
.shop {
  min-height: 0;
  min-width: 0;
}
.shop__pages { width: 100%; }
.shop__paged { overflow: hidden !important; }
.shop__paged :deep(.shelf__grid) { grid-template-columns: minmax(0, 1fr); }
.shop__paged :deep(.card) { height: 146px; }
.shop--session :deep(> .panel__head) { display: none; }
@media (max-height: 559px) and (min-width: 560px) {
  .shop__paged :deep(.card) { display: grid; grid-template-columns: minmax(0, 1fr) auto; grid-template-rows: 1fr auto; height: 96px; }
  .shop__paged :deep(.card__open) { grid-column: 1; grid-row: 1; }
  .shop__paged :deep(.card__keys) { grid-column: 2; grid-row: 1; border-top: 0; }
  .shop__paged :deep(.card__meta) { grid-column: 1 / -1; grid-row: 2; }
}
</style>
