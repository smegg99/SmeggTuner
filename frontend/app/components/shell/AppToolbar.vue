<template>
  <div
    ref="bar"
    class="bar"
  >
    <!-- Compact windows keep the input and output here; the tuner gives the note keys a row of its own. -->
    <template v-if="view === 'tune'">
      <div class="bar__top">
        <ShellToolbarTransport />
        <ShellToolbarRecord v-if="WORKSHOP" />

        <div class="bar__divider group-rule" />

        <ShellToolbarSource />
        <ShellToolbarSoundCard v-if="!WEB" />
        <ShellToolbarAudio v-if="compact && !WEB" />
      </div>

      <!-- Wide windows carry the note keys here; so does a landscape phone with the width for them. -->
      <div
        v-if="!compact || barTools"
        class="bar__low bar__low--keys"
      >
        <div class="bar__divider bar__divider--joint group-rule" />

        <ShellToolbarNote />

        <div class="bar__divider group-rule" />

        <ShellToolbarReading />
      </div>
    </template>

    <ShellToolbarPage
      v-if="title"
      :title="title"
    />

    <!-- v-else-if, not v-else: a bare v-else would also catch the settings room, which has no toolbar keys. -->
    <!-- An open session's keys are too wide to share the views' row in portrait: their own row,
         as the tuner's note keys. The shelf's two keys fit beside the views. -->
    <div
      v-else-if="view === 'calibrate'"
      class="bar__low"
    >
      <ShellToolbarCalibrate />
    </div>

    <div
      v-else-if="view === 'workshop'"
      :class="openSession ? 'bar__low' : 'bar__top'"
    >
      <ShellToolbarWorkshop :notes="notes" />
    </div>

    <div
      ref="spacer"
      class="bar__spacer"
    />

    <!-- A landscape phone: square marks, Help and About first and the rooms at the far end. -->
    <ShellToolbarViews
      v-if="!compact || landscape"
      ref="views"
      class="bar__views"
      :labelled="labelled && !compact"
      reading-first
    />
    <!-- Measures the named keys while only their marks show; never seen or reached. -->
    <ShellToolbarViews
      v-if="!compact || landscape"
      ref="probe"
      class="bar__probe"
      aria-hidden="true"
      inert
    />
  </div>
</template>

<script setup lang="ts">
import { useTemplateRef } from 'vue'
import { usePageTitle } from '~/composables/usePageTitle'
import { useShell } from '~/composables/useShell'
import { useBarHeight } from '~/composables/useBarHeight'
import { useWideLandscapeLayout } from '~/composables/useCompactLayout'
import { useViewLabels } from '~/composables/useViewLabels'
import { WEB, WORKSHOP } from '~/utils/features'

defineProps<{
  /** notes on screen; enables the workshop Print key */
  notes?: number
}>()

const { view, openSession } = useShell()
const compact = useCompactLayout()
const landscape = useLandscapeLayout()
const barTools = useWideLandscapeLayout()
const { locale } = useI18n()
const title = usePageTitle()

const bar = useTemplateRef<HTMLElement>('bar')
const spacer = useTemplateRef<HTMLElement>('spacer')
const views = useTemplateRef<{ $el: HTMLElement }>('views')
const probe = useTemplateRef<{ $el: HTMLElement }>('probe')
const labelled = useViewLabels(bar, spacer, views, probe, () => view.value === 'tune', locale)
useBarHeight(bar)
</script>

<style scoped>
.bar {
  container-type: inline-size;
  position: relative;
  align-items: center;
  background: rgb(var(--v-theme-chrome2));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: var(--app-radius-panel);
  display: flex;
  gap: var(--gap);
  min-width: 0;
  overflow: hidden;
  padding: var(--gap);
}

.bar__spacer {
  flex: 1 1 auto;
}

.bar__probe {
  pointer-events: none;
  position: absolute;
  visibility: hidden;
}

.bar__top,
.bar__low {
  display: contents;
}

/* A narrow desktop bar keeps every control and its label, the input picker at full width, and tightens
   round them: keys a little smaller, the reed keys without their caption, a shorter slider, closer groups. */
@media (min-width: 960px) and (min-height: 560px) {
  .bar { flex-wrap: wrap; }

  @container (max-width: 1280px) {
    .bar > * { --s99ab-control-h: 26px; }
    .bar :deep(.modes__label) { display: none; }
    .bar :deep(.tuner-volume) { flex-basis: 44px; width: 44px; }
    .bar :deep(.note) { width: 40px; }
    .bar :deep(.key:not(.key--icon)) { padding-inline: 6px; }
    .bar :deep(.group-rule) { margin-inline: 0; }
  }
}

@media (max-width: 959px), (max-height: 559px) {
  .bar {
    display: flex;
    gap: var(--s99ab-control-gap);
    padding: 4px;
  }

  .bar__top,
  .bar__low {
    align-items: center;
    display: flex;
    gap: var(--s99ab-control-gap);
    min-width: 0;
  }

  .bar__top { flex: 1 1 auto; }
  .bar__low { flex: 1 1 auto; }
  .bar__spacer,
  .bar__divider--joint { display: none; }
  .bar__divider { display: none; }
  .bar :deep(.source-key) { flex: 1 1 auto; width: 0 !important; min-width: 0; }

  /* A bar wider than a portrait phone stops stretching the input picker: it keeps one width and the
     rooms go to the far end. */
  @container (min-width: 520px) {
    .bar__top { flex: 0 1 auto; }
    .bar__low--keys { flex: 0 0 auto; }
    .bar__spacer { display: block; }
    .bar :deep(.source-key) { flex: 0 1 auto; width: 240px !important; }
  }
}
</style>
