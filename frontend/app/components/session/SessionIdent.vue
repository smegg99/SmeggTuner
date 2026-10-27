<template>
  <div class="ident">
    <SessionPhoto
      :live="canPick"
      :src="imageSrc"
      :alt="instrument"
      @pick="void setPhoto()"
    />

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

    <v-chip class="ident__chip">
      {{ t('session.card.a4', { hz: (session?.a4 ?? 0).toFixed(1) }) }}
    </v-chip>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import SessionPhoto from './SessionPhoto.vue'
import { useSessionAccordion } from '~/composables/useSessionAccordion'
import type { ToolItem } from '@smegg99/s99ab/types'
import type { SessionDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/session/models.js'

const props = defineProps<{
  session: SessionDTO | null
  busy: boolean
}>()

const view = defineModel<'recorded' | 'curve' | 'accordion'>('view', { required: true })

const { t } = useI18n()
const compact = useCompactLayout()
const { remembered, imageSrc, setPhoto, exportIt, remember } = useSessionAccordion()

// The picker and the accordion key belong to the accordion, so they appear with its tab.
const onAccordion = computed(() => view.value === 'accordion')

// A photograph is stored on the shelf entry, so it can only be set once there is one.
const canPick = computed(() => onAccordion.value && remembered.value)

const instrument = computed(() => {
  const i = props.session?.instrument
  return i?.name ?? ''
})

const VIEWS = computed<ToolItem[]>(() => [
  { value: 'recorded', label: t(compact.value ? 'sessionPage.mobile.recorded' : 'sessionPage.tabs.recorded'), icon: 'mdi-table' },
  { value: 'curve', label: t(compact.value ? 'sessionPage.mobile.curve' : 'sessionPage.tabs.curve'), icon: 'mdi-chart-bell-curve-cumulative' },
  { value: 'accordion', label: t('sessionPage.tabs.accordion'), icon: 'mdi-piano' },
])
</script>

<style scoped>
.ident {
  align-items: center;
  border-bottom: 1px solid rgb(var(--v-theme-lineSoft));
  display: flex;
  gap: var(--gap);
  padding: var(--gap);
}

.ident__key {
  flex: 0 0 auto;
}

.ident__m {
  display: flex;
  flex: 0 1 auto;
  flex-direction: column;
  gap: var(--gap-tight);
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
  font-size: 12.5px;
  font-weight: 600;
}

.ident__s {
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
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
  font-size: 12px;
  font-weight: 500;
  height: var(--s99ab-control-h);
  justify-content: center;
  line-height: 1;
  padding: 0 8px;
}
/* Portrait: name and A4 on the first line, the three tabs on their own. */
@media (max-width: 959px), (max-height: 559px) {
  .ident { gap: 4px; }
  .ident__views { flex: 1 1 auto; min-width: 0; }
  .ident__views :deep(.key) { flex: 1 1 0; min-width: 0; padding-inline: 6px; }
  .ident__views :deep(.key__icon) { display: none; }
  .ident__m, .ident__sp, .ident__chip { display: none; }
  .ident__views :deep(.key__label) { overflow: hidden; text-overflow: ellipsis; }
  .ident :deep(.ident__pic:not(.ident__pic--live)) { display: none; }
  .ident__key :deep(.key__label) { display: none; }
}
@media (max-height: 359px) and (min-width: 560px) {
  .ident { padding-block: 2px; }
}
</style>
