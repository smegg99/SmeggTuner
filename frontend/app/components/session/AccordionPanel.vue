<template>
  <div class="acc">
    <!-- Not saved by a key: every committed change is written. The one thing that cannot be
         written is called out, because the backend would refuse it. -->
    <p
      v-if="frozen"
      class="acc__frozen"
      :class="{ 'acc__frozen--refused': columnsMoved }"
    >
      {{ columnsMoved ? t('session.accordion.columnsRefused') : t('session.accordion.frozen') }}
    </p>

    <div class="acc__form">
      <section class="acc__col">
        <InstrumentIdentityFields
          v-model:name="form.name"
          :name-error="nameError"
          @submit="save"
        />
        <InstrumentReferenceFields
          v-model:a4-draft="a4Draft"
          v-model:tol-draft="tolDraft"
          v-model:beat-draft="beatDraft"
          :a4-disabled="frozen"
          :a4-note="frozen ? t('session.accordion.a4Locked') : ''"
          @commit-a4="commitA4"
          @commit-tol="commitTol"
          @commit-beat="commitBeat"
        />
        <UiFieldGroup
          dense
          :title="t('instrument.sections.bass')"
          :hint="t('instrument.bass.hint')"
        >
          <InstrumentBassSection
            v-model:reeds="bassReeds"
            v-model:registers="bassRegisters"
          />
        </UiFieldGroup>
      </section>

      <section class="acc__col">
        <UiFieldGroup
          dense
          :title="t('instrument.sections.voices')"
          :hint="t('instrument.registersHint')"
        >
          <p class="acc__says">
            {{ banks.length ? t('instrument.says', { banks: banks.join(' '), reeds: reedCount }, reedCount) : t('instrument.registersSubtitle') }}
          </p>

          <InstrumentRegisterSection v-model="registers" />
        </UiFieldGroup>

        <InstrumentKeyboardFields
          v-model:lo="form.lo"
          v-model:hi="form.hi"
          :range-error="rangeError"
        />
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useInstrumentForm } from '~/components/instrument/useInstrumentForm'
import { useSessionAccordion } from '~/composables/useSessionAccordion'
import { frozenByReadings } from '~/utils/accordion'
import * as SessionService from '~~bindings/smegg.me/smeggtuner/services/session/service.js'

// Long enough that typing a name is one write, short enough that leaving the tab never races it.
const AUTOSAVE_MS = 600

const { t } = useI18n()
const { active, setA4, refresh } = useSessions()
const library = useInstruments()
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

// What is already stored. Loading an accordion sets it, so arriving on the tab is not an edit.
const stored = ref(edited.value)
watch(() => source.value?.id, () => {
  stored.value = edited.value
}, { flush: 'post' })

let timer: ReturnType<typeof setTimeout> | undefined

watch(edited, (now) => {
  if (now === stored.value) return
  if (timer) clearTimeout(timer)
  timer = setTimeout(() => void save(), AUTOSAVE_MS)
})

// Leaving the tab tears the panel down before a field can blur, so this is the one moment the
// number drafts have to be forced in - and the one moment a pending write must go now, not never.
onBeforeUnmount(() => {
  commitA4()
  commitTol()
  commitBeat()
  if (!timer) return
  clearTimeout(timer)
  void save()
})

async function save() {
  timer = undefined
  // A name it cannot be found by, or a backwards keyboard: the field says so, and nothing is
  // written until it is fixed.
  const next = collect()
  if (!next) return
  // The card's columns are frozen once there are readings and the backend refuses them outright,
  // so the panel does not ask; the notice above says why.
  if (columnsMoved.value) return

  const was = edited.value
  // Both copies, because they are the same accordion: the session keeps its own so an
  // exported .stsf still means something on a machine that never saw this one.
  await SessionService.SetInstrument(next.instrument as never)
  if (remembered.value) await library.save({ ...next, id: source.value!.id })
  if (!frozen.value) await setA4(form.a4)
  await refresh()
  stored.value = was
}
</script>

<style scoped>
/* Rows sized to their content, not capped at the pane: capped, a form taller than the pane was
   clipped at the bottom - the switch-name field cut in half - with nothing for overflow-y to
   scroll. Sized to content the grid outgrows .acc, and .acc scrolls. */
.acc {
  align-content: start;
  display: grid;
  gap: 9.6px;
  grid-template-rows: auto auto;
  min-height: 0;
  overflow-y: auto;
  padding: 8px 11px;
}

.acc__frozen {
  color: rgb(var(--v-theme-ink3));
  flex: 1 1 auto;
  font-size: 11.6px;
  margin: 0;
  max-width: 60ch;
}

.acc__sp {
  flex: 1 1 auto;
}

.acc__form {
  display: grid;
  gap: 14.4px 30.8px;
  grid-template-columns: 1fr 1.1fr;
}

.acc__col {
  display: flex;
  flex-direction: column;
  gap: 9.6px;
  min-width: 0;
}

/* One line, reserved even when empty, so describing switches never reflows the pane below. */
.acc__says {
  color: rgb(var(--v-theme-ink2));
  font-size: 11.6px;
  height: 16.8px;
  line-height: 16.8px;
  margin: 0;
  overflow: hidden;
}
</style>
