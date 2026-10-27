<!-- app/components/instrument/BassBuilder.vue -->
<template>
  <!-- Builds the next bass switch from the machine's feet. Its root takes no box, so its parts lay out
       in the section's own flow and grid. -->
  <div class="bass__builder">
    <!-- Same reason as the treble builder's heading: these sit under the switches the machine
         already has, and they build the next one rather than being one. -->
    <span
      class="bass__h"
    >{{ t('instrument.bass.build') }}</span>

    <div
      class="bass__feet"
    >
      <UiToolKey
        v-for="f in feet"
        :key="f"
        :label="`${f}'`"
        :active="draft.includes(f)"
        :disabled="full"
        :title="`${f}'`"
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
        <v-text-field
          :id="id"
          v-model="draftName"
          :disabled="full"
          :maxlength="MAX_SWITCH_NAME"
          hide-details="auto"
          autocomplete="off"
          :aria-label="t('instrument.bass.registerName')"
          @keyup.enter="addRegister()"
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
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { canAddBass } from '~/utils/accordion'
import { MAX_BASS_REGISTERS, MAX_SWITCH_NAME } from '~/types/session'
import type { BassRegister } from '~/types/session'

const registers = defineModel<BassRegister[]>({ required: true })
const props = defineProps<{ feet: number[] }>()
const { t } = useI18n()

const draft = ref<number[]>([])
const draftName = ref('')

// A shrunk machine drops the feet it no longer has from the switch being built.
watch(() => props.feet, (feet) => {
  draft.value = draft.value.filter(f => feet.includes(f))
})

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
</script>

<style scoped>
.bass__builder { display: contents; }

.bass__h {
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.bass__feet {
  align-items: center;
  display: flex;
  gap: var(--gap);
  min-height: 28px;
}

.bass__add {
  display: flex;
  gap: var(--gap);
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
  max-width: 212px;
}
</style>
