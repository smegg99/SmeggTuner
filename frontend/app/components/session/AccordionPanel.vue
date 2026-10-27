<!-- app/components/session/AccordionPanel.vue -->
<template>
  <div
    class="acc"
    :class="{ 'acc--compact': compact }"
  >
    <v-select
      v-if="compact"
      v-model="section"
      :items="sections"
      item-title="label"
      item-value="value"
      :aria-label="t('sessionPage.tabs.accordion')"
      hide-details
    />
    <!-- Not saved by a key: every committed change is written. The one thing that cannot be
         written is called out, because the backend would refuse it. -->
    <p
      v-if="frozen && !compact"
      class="acc__frozen"
      :class="{ 'acc__frozen--refused': columnsMoved }"
    >
      {{ columnsMoved ? t('session.accordion.columnsRefused') : t('session.accordion.frozen') }}
    </p>

    <v-menu v-if="compact && frozen">
      <template #activator="{ props: menu }">
        <UiToolKey
          v-bind="menu"
          class="acc__locked"
          icon="mdi-lock-outline"
          :label="t('session.accordion.frozenShort')"
          :title="t('session.accordion.frozenShort')"
        />
      </template>
      <p class="acc__explanation">
        {{ columnsMoved ? t('session.accordion.columnsRefused') : t('session.accordion.frozen') }}
      </p>
    </v-menu>
    <div class="acc__form">
      <section class="acc__col">
        <InstrumentIdentityFields
          v-show="visible('identity')"
          v-model:name="form.name"
          :name-error="nameError"
          @submit="save"
        />
        <InstrumentReferenceFields
          v-show="visible('reference') || visible('tolerance')"
          v-model:a4-draft="a4Draft"
          v-model:tol-draft="tolDraft"
          v-model:beat-draft="beatDraft"
          :mode="compact ? (section === 'reference' ? 'pitch' : 'tolerance') : 'all'"
          :a4-disabled="frozen"
          :a4-note="frozen ? t('session.accordion.a4Locked') : ''"
          @commit-a4="commitA4"
          @commit-tol="commitTol"
          @commit-beat="commitBeat"
        />
        <InstrumentBassGroup
          v-show="visible('bass') || visible('bassBuilder')"
          v-model:reeds="bassReeds"
          v-model:registers="bassRegisters"
          :mode="compact ? (section === 'bass' ? 'list' : 'builder') : 'all'"
        />
      </section>

      <section class="acc__col">
        <InstrumentVoicesGroup
          v-show="visible('voices') || visible('registerBuilder')"
          v-model="registers"
          :banks="banks"
          :reed-count="reedCount"
          :mode="compact ? (section === 'voices' ? 'list' : 'builder') : 'all'"
        />

        <InstrumentKeyboardFields
          v-show="visible('keyboard')"
          v-model:lo="form.lo"
          v-model:hi="form.hi"
          :range-error="rangeError"
        />
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useInstrumentForm } from '~/components/instrument/useInstrumentForm'
import { useAccordionAutosave } from '~/components/session/useAccordionAutosave'
import { useSessionAccordion } from '~/composables/useSessionAccordion'
import { frozenByReadings } from '~/utils/accordion'

const compact = useCompactLayout()
const section = ref('identity')
const sections = computed(() => [
  ...['identity', 'reference', 'voices', 'keyboard', 'bass'].map(value => ({ value, label: t(`instrument.sections.${value}`) })),
  { value: 'registerBuilder', label: t('instrument.builder.title') },
  { value: 'bassBuilder', label: t('instrument.bass.build') },
  { value: 'tolerance', label: t('instrument.tolerance') },
])
const visible = (value: string) => !compact.value || section.value === value

const { t } = useI18n()
const { active } = useSessions()
const { source, remembered } = useSessionAccordion()

const {
  form, nameError, a4Draft, tolDraft, beatDraft,
  registers, bassReeds, bassRegisters, banks, reedCount, rangeError,
  commitA4, commitTol, commitBeat, collect,
} = useInstrumentForm(source)

const frozen = computed(() => frozenByReadings(active.value?.readings ?? 0))

// A switch nobody had described yet moves no column and stays welcome; a rank added or
// dropped re-files readings nobody re-measured, which services/session refuses outright.
const columnsMoved = computed(() => {
  const had = active.value?.instrument
  if (!frozen.value || !had) return false
  return reedCount.value !== had.reedCount
    || banks.value.join(' ') !== (had.banks ?? []).join(' ')
    || (bassReeds.value || 0) !== (had.bassReeds ?? 0)
})

// Everything the panel may write, as one value. A change to any of it schedules a save; drafts
// are absent on purpose, since a half-typed 44 commits to 440 before it lands here.
const edited = computed(() => JSON.stringify([
  form.name, form.lo, form.hi, form.a4, form.tolerance, form.beatTolerance,
  registers.value, bassReeds.value, bassRegisters.value,
]))

const { save } = useAccordionAutosave({
  edited,
  sourceID: () => source.value?.id,
  collect,
  commitDrafts: () => {
    commitA4()
    commitTol()
    commitBeat()
  },
  columnsMoved,
  nameError,
  libraryID: () => (remembered.value ? source.value?.id : undefined),
})
</script>

<style scoped>
/* Rows sized to their content, not capped at the pane: capped, a form taller than the pane was
   clipped at the bottom - the switch-name field cut in half - with nothing for overflow-y to
   scroll. Sized to content the grid outgrows .acc, and .acc scrolls. */
.acc {
  align-content: start;
  display: grid;
  gap: var(--gap);
  grid-template-rows: auto auto;
  min-height: 0;
  overflow-y: auto;
  padding: 7px 9.5px;
}

.acc__frozen {
  color: rgb(var(--v-theme-ink3));
  flex: 1 1 auto;
  font-size: 11px;
  margin: 0;
  max-width: 60ch;
}

.acc__form {
  display: grid;
  gap: var(--gap-wide) var(--gap-wide);
  grid-template-columns: 1fr 1.1fr;
}

@media (max-width: 720px) {
  .acc__form { grid-template-columns: minmax(0, 1fr); }
}

.acc__col {
  display: flex;
  flex-direction: column;
  gap: var(--gap);
  min-width: 0;
}

.acc--compact { display: flex; flex-direction: column; overflow: hidden; }
.acc--compact > .v-select { flex: 0 0 auto; }
.acc--compact .acc__form { display: block; min-height: 0; min-width: 0; }
.acc--compact .acc__frozen { flex: 0 0 auto; }
.acc--compact :deep(.form-section) { margin-top: 0; }
.acc--compact .acc__col { display: contents; }
.acc--compact :deep(.form-section__head) { display: none; }
.acc--compact :deep(.acc__says) { display: none; }

@media (max-height: 559px) and (min-width: 560px) {
  .acc--compact { display: grid; grid-template-columns: 150px minmax(0, 1fr); grid-template-rows: auto 1fr; align-content: stretch; }
  .acc--compact > .v-select { grid-column: 1; }
  .acc--compact .acc__frozen, .acc--compact .acc__locked { grid-column: 1; grid-row: 2; align-self: start; }
  .acc--compact .acc__form { grid-column: 2; grid-row: 1 / -1; }
}
.acc__explanation { max-width: min(340px, calc(100vw - 24px)); padding: 12px; background: rgb(var(--v-theme-chrome)); border: 1px solid rgb(var(--v-theme-line)); border-radius: var(--app-radius-panel); font-size: 12px; }
</style>
