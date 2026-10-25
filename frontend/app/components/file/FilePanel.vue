<template>
  <UiPanel
    class="file tuner-panel"
    flush
  >
    <template #head>
      <span class="file__name">{{ state.name }}</span>

      <FileRun />
      <FileTransport
        v-if="!compact"
        :with-volume="withVolume"
      />
      <v-menu
        v-else
        :close-on-content-click="false"
      >
        <template #activator="{ props: menu }">
          <UiToolKey
            v-bind="menu"
            icon="mdi-tune-variant"
            :title="t('tuner.controls')"
          />
        </template>
        <div class="file-options">
          <FileTransport :with-volume="withVolume" />
        </div>
      </v-menu>

      <span class="file__sp" />

      <span class="file__at"><b>{{ clock(heard) }}</b> / {{ clock(state.duration) }}</span>
    </template>

    <div class="file__body">
      <FileWave class="file__wave" />
      <FileOverview class="file__map" />
    </div>
  </UiPanel>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useTransport } from '~/composables/useTransport'

defineProps<{
  /** Carry output level in this header when the toolbar has no volume key. */
  withVolume?: boolean
}>()

const { t } = useI18n()
const compact = useCompactLayout()
const { state, heard, refresh } = useTransport()

// m:ss.s - the tenth matters: a reed's attack is over inside one second.
function clock(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = seconds - m * 60

  return `${m}:${s < 10 ? '0' : ''}${s.toFixed(1)}`
}

// The file may have been selected before this panel mounted, so ask once for the missed state.
onMounted(refresh)
</script>

<style scoped>
/* The playback keys are text keys, as the run keys beside them: no outline until hovered or lit. */
.file :deep(.panel__head .key:not(.key--active)) {
  background: transparent;
  border-color: transparent;
}

.file {
  min-height: 0;
}

.file__name {
  color: rgb(var(--v-theme-ink2));
  flex: 0 1 auto;
  font-size: 11px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file__sp {
  flex: 1 1 auto;
}

.file__at {
  color: rgb(var(--v-theme-ink3));
  flex: 0 0 auto;
  font-family: var(--font-mono);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.file__at b {
  color: rgb(var(--v-theme-ink2));
  font-weight: 600;
}

.file__body {
  display: grid;
  grid-template-rows: 1fr 22.5px;
  height: 100%;
  min-height: 0;
}

.file__wave,
.file__map {
  min-height: 0;
}

.file-options { display: flex; flex-wrap: wrap; align-items: center; gap: 12px; padding: 12px; width: min(330px, calc(100vw - 24px)); background: rgb(var(--v-theme-chrome)); border: 1px solid rgb(var(--v-theme-line)); border-radius: var(--app-radius-panel); }
.file-options :deep(.rule) { display: none; }
@media (max-width: 959px), (max-height: 559px) {
  .file__name { flex: 1; min-width: 0; }
  .file__sp { display: none; }
  .file__at { font-size: 10px; }
}
@media (max-height: 619px) {
  .file__body { grid-template-rows: minmax(0, 1fr) 12px; }
}
@media (max-height: 359px) and (min-width: 560px) {
  .file :deep(.panel__head) { min-height: 28px; padding-block: 0; }
}
</style>
