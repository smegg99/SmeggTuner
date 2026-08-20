<template>
  <div class="ident">
    <!-- The photograph is the picker, but only where the accordion is edited. Everywhere else it
         is a plain span: no button, no tooltip, no hover, so nothing invites a click that would
         do nothing. It also needs a shelf entry to hang the image on, so it waits for one. -->
    <component
      :is="canPick ? 'button' : 'span'"
      class="ident__pic"
      :class="{ 'ident__pic--live': canPick }"
      :type="canPick ? 'button' : undefined"
      :title="canPick ? t('instrument.photo') : undefined"
      @click="pick"
    >
      <img
        v-if="imageSrc"
        class="ident__img"
        :src="imageSrc"
        :alt="instrument"
      >
      <v-icon
        v-else
        :icon="canPick ? 'mdi-camera-outline' : 'mdi-folder-music-outline'"
        class="ident__none"
      />
    </component>

    <span class="ident__m">
      <span class="ident__n">{{ session?.name }}</span>
      <span class="ident__s">{{ instrument || t('session.card.unnamedInstrument') }}</span>
    </span>

    <span class="ident__sp" />

    <UiToolGroup
      v-model="view"
      class="ident__views"
      :items="VIEWS"
      :label="t('session.title')"
    />

    <!-- The accordion's own actions, beside the tab that edits it. Export and Remember never
         apply at once: one wants a shelf entry, the other is how it gets one. -->
    <UiToolKey
      v-if="onAccordion && remembered"
      class="ident__key"
      icon="mdi-export"
      :title="t('instrument.export')"
      @click="exportIt"
    />
    <UiToolKey
      v-else-if="onAccordion"
      class="ident__key"
      icon="mdi-bookmark-plus-outline"
      :label="t('session.accordion.remember')"
      :title="t('session.accordion.rememberHint')"
      :disabled="busy"
      @click="remember"
    />

    <UiChip class="ident__chip">
      {{ t('session.card.a4', { hz: (session?.a4 ?? 0).toFixed(1) }) }}
    </UiChip>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useSessionAccordion } from '~/composables/useSessionAccordion'
import type { ToolItem } from '~/components/ui/ToolGroup.vue'
import type { SessionDTO } from '~~bindings/smegg.me/smeggtuner/services/session/models.js'

const props = defineProps<{
  session: SessionDTO | null
  busy: boolean
}>()

const view = defineModel<'recorded' | 'curve' | 'accordion'>('view', { required: true })

const { t } = useI18n()
const { remembered, imageSrc, setPhoto, exportIt, remember } = useSessionAccordion()

// The picker and the accordion key belong to the accordion, so they appear with its tab.
const onAccordion = computed(() => view.value === 'accordion')

// A photograph is stored on the shelf entry, so it can only be set once there is one.
const canPick = computed(() => onAccordion.value && remembered.value)

function pick() {
  if (canPick.value) void setPhoto()
}

const instrument = computed(() => {
  const i = props.session?.instrument
  return i?.name ?? ''
})

const VIEWS = computed<ToolItem[]>(() => [
  { value: 'recorded', label: t('sessionPage.tabs.recorded'), icon: 'mdi-table' },
  { value: 'curve', label: t('sessionPage.tabs.curve'), icon: 'mdi-chart-bell-curve-cumulative' },
  { value: 'accordion', label: t('sessionPage.tabs.accordion'), icon: 'mdi-piano' },
])
</script>

<style scoped>
.ident {
  align-items: center;
  border-bottom: 1px solid rgb(var(--v-theme-lineSoft));
  display: flex;
  gap: 7.7px;
  padding: 6.4px 9.9px;
}

.ident__pic {
  align-items: center;
  padding: 0;
  aspect-ratio: 4 / 3;
  background: rgb(var(--v-theme-sunk));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 2px;
  display: flex;
  flex: 0 0 auto;
  height: 40px;
  justify-content: center;
  overflow: hidden;
}

.ident__pic--live {
  cursor: pointer;
}

.ident__pic--live:hover {
  border-color: rgb(var(--v-theme-ink2));
}

.ident__pic--live:hover .ident__none {
  color: rgb(var(--v-theme-ink));
}

.ident__key {
  flex: 0 0 auto;
}

.ident__img {
  height: 100%;
  object-fit: cover;
  width: 100%;
}

.ident__none {
  color: rgb(var(--v-theme-ink3));
  font-size: 19.2px;
}

.ident__m {
  display: flex;
  flex: 0 1 auto;
  flex-direction: column;
  gap: 0.8px;
  min-width: 0;
}

/* A session name is whatever the technician typed. It ellipsises rather than wrapping, or a long
   one folds into a column one word wide and takes the whole header with it. */
.ident__n,
.ident__s {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ident__n {
  font-size: 14.4px;
  font-weight: 600;
}

.ident__s {
  color: rgb(var(--v-theme-ink3));
  font-size: 11.6px;
}

.ident__sp {
  flex: 1 1 auto;
}

.ident__views {
  flex: 0 0 auto;
}

/* Metrics copied from UiToolKey .key so these static labels match the view keys' height. */
.ident__chip {
  align-items: center;
  display: inline-flex;
  font-size: 14px;
  font-weight: 500;
  height: 31.2px;
  justify-content: center;
  line-height: 1;
  padding: 0 9.3px;
}
</style>
