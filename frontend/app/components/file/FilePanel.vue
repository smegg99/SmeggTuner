<template>
  <UiPanelCard class="file">
    <template #head>
      <span class="file__name">{{ state.name }}</span>

      <FileRun />
      <span class="file__rule" />
      <FileTransport :with-volume="withVolume" />

      <span class="file__sp" />

      <span class="file__at"><b>{{ clock(heard) }}</b> / {{ clock(state.duration) }}</span>
      <span class="file__meta">{{ selection }}</span>
    </template>

    <div class="file__body">
      <FileWave class="file__wave" />
      <FileOverview class="file__map" />
    </div>
  </UiPanelCard>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useTransport } from '~/composables/useTransport'

defineProps<{
  /** Carry output level in this header when the toolbar has no volume key. */
  withVolume?: boolean
}>()

const { t } = useI18n()
const { state, hasSelection, heard, refresh } = useTransport()

// m:ss.s - the tenth matters: a reed's attack is over inside one second.
function clock(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = seconds - m * 60

  return `${m}:${s < 10 ? '0' : ''}${s.toFixed(1)}`
}

// What the engine may measure; "the whole recording" is a real answer, not silence.
const selection = computed(() =>
  hasSelection.value
    ? t('file.selected', { from: clock(state.from), to: clock(state.to) })
    : t('file.whole'),
)

// The file may have been selected before this panel mounted, so ask once for the missed state.
onMounted(refresh)
</script>

<style scoped>
.file {
  min-height: 0;
}

.file__name {
  color: rgb(var(--v-theme-ink2));
  flex: 0 1 auto;
  font-size: 1.6cqh;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file__sp {
  flex: 1 1 auto;
}

.file__rule {
  background: rgb(var(--v-theme-line));
  block-size: 2.4cqh;
  flex: 0 0 1px;
  margin-inline: 0.35cqw;
}

.file__at,
.file__meta {
  color: rgb(var(--v-theme-ink3));
  flex: 0 0 auto;
  font-family: var(--font-mono);
  font-size: 1.45cqh;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.file__at b {
  color: rgb(var(--v-theme-ink2));
  font-weight: 600;
}

.file__body {
  display: grid;
  grid-template-rows: 1fr 3.2cqh;
  height: 100%;
  min-height: 0;
}

.file__wave,
.file__map {
  min-height: 0;
}
</style>
