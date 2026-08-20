<template>
  <div class="bass">
    <div class="bass__row">
      <span class="bass__label">{{ t('instrument.bass.voices') }}</span>
      <UiToolGroup
        :model-value="reeds"
        :items="COUNTS"
        :label="t('instrument.bass.voices')"
        @update:model-value="n => setReeds(n as number)"
      />
    </div>

    <!-- No machine declared, no machine to describe: one line saying what the row above is for,
         rather than an empty list, a row of nothing and a name field for a switch that cannot
         exist. -->
    <p
      v-if="reeds === 0"
      class="bass__none"
    >
      {{ t('instrument.bass.hint') }}
    </p>

    <template v-else>
      <InstrumentRegisterList
        :items="items"
        :empty="t('instrument.bass.fixed')"
        @remove="removeRegister"
      />

      <!-- Same reason as the treble builder's heading: these sit under the switches the machine
           already has, and they build the next one rather than being one. -->
      <span class="bass__h">{{ t('instrument.bass.build') }}</span>

      <div class="bass__feet">
        <UiToolKey
          v-for="f in machineFeet"
          :key="f"
          :label="`${f}'`"
          :active="draft.includes(f)"
          :disabled="full"
          @click="toggleFoot(f)"
        />
      </div>

      <UiFormField
        v-slot="{ id }"
        class="bass__field"
        :label="t('instrument.bass.registerName')"
      >
        <!-- Name and key inside the field, so the key lines up with the input rather than with
             the label above it. -->
        <div class="bass__add">
          <UiTextInput
            :id="id"
            v-model="draftName"
            :disabled="full"
            :maxlength="MAX_SWITCH_NAME"
            @enter="addRegister"
          />
          <UiToolKey
            icon="mdi-plus"
            :label="t('instrument.builder.add')"
            :disabled="!canAdd"
            :title="addHint"
            @click="addRegister"
          />
        </div>
      </UiFormField>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { canAddBass } from '~/utils/accordion'
import { MAX_BASS_REGISTERS, MAX_SWITCH_NAME } from '~/types/session'
import { bassFeetOf } from '~/utils/feet'
import type { ToolItem } from '~/components/ui/ToolGroup.vue'
import type { BassRegister } from '~/types/session'

const reeds = defineModel<number>('reeds', { required: true })
const registers = defineModel<BassRegister[]>('registers', { required: true })

const { t } = useI18n()

// None, or the counts machines are built with; four and five are the common ones.
const COUNTS: ToolItem[] = [
  { value: 0, label: t('instrument.bass.none') },
  ...[2, 3, 4, 5, 6].map(n => ({ value: n, label: String(n) })),
]

const machineFeet = computed(() => bassFeetOf(reeds.value))

const items = computed(() => registers.value.map(r => ({
  name: r.name,
  feet: r.feet.map(f => `${f}'`).join('+'),
})))

const draft = ref<number[]>([])
const draftName = ref('')

// The machine is full: nothing more can be built, so the whole builder goes quiet rather than
// offering a click the backend would refuse.
const full = computed(() => registers.value.length >= MAX_BASS_REGISTERS)
const canAdd = computed(() => canAddBass(registers.value, draft.value, draftName.value))

// Why the key is off, in the words the situation calls for.
const addHint = computed(() => {
  if (canAdd.value) return ''
  if (full.value) return t('instrument.bass.full', { n: MAX_BASS_REGISTERS })
  if (!draft.value.length || !draftName.value.trim()) return ''
  return t('instrument.builder.already')
})

function setReeds(n: number) {
  reeds.value = n
  // A shrunk machine invalidates switches reaching ranks it no longer has.
  registers.value = registers.value.filter(r => r.feet.every(f => bassFeetOf(n).includes(f)))
  draft.value = draft.value.filter(f => bassFeetOf(n).includes(f))
}

function toggleFoot(f: number) {
  draft.value = draft.value.includes(f)
    ? draft.value.filter(x => x !== f)
    : [...draft.value, f].sort((a, b) => b - a)
}

function addRegister() {
  if (!canAdd.value) return
  registers.value = [...registers.value, { name: draftName.value.trim(), feet: [...draft.value] }]
  draft.value = []
  draftName.value = ''
}

function removeRegister(i: number) {
  registers.value = registers.value.filter((_, n) => n !== i)
}
</script>

<style scoped>
.bass {
  display: flex;
  flex-direction: column;
  gap: 6.4px;
}

.bass__row {
  align-items: center;
  display: flex;
  gap: 8.8px;
}

.bass__label {
  color: rgb(var(--v-theme-ink));
  flex: 1 1 auto;
  font-size: 13.2px;
}

.bass__none {
  color: rgb(var(--v-theme-ink3));
  font-size: 11.6px;
  line-height: 1.5;
  margin: 0;
  max-width: 52ch;
}

.bass__h {
  color: rgb(var(--v-theme-ink3));
  font-size: 10.8px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.bass__feet {
  align-items: center;
  display: flex;
  gap: 6.6px;
  min-height: 32px;
}

.bass__add {
  display: flex;
  gap: 8.8px;
}

.bass__add > :first-child {
  flex: 1 1 auto;
  min-width: 0;
}

/* A tool key is 31.2px and a text input 35.2px - two control families, deliberately different.
   Side by side in one row they have to agree, and the input is the one that sets the line. */
.bass__add :deep(.key) {
  height: auto;
}

/* Capped for the same reason as the register builder's rows: the panel is far wider
   than the dialog this came from. */
.bass__field {
  max-width: 242px;
}
</style>
