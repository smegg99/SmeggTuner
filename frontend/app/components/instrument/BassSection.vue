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

    <!-- Every block keeps its place whatever is picked, so the dialog never jumps. -->
    <InstrumentRegisterList
      :items="items"
      :empty="t(reeds > 0 ? 'instrument.bass.fixed' : 'instrument.bass.hint')"
      height="14cqh"
      @remove="removeRegister"
    />

    <div class="bass__feet">
      <UiToolKey
        v-for="f in machineFeet"
        :key="f"
        :label="`${f}'`"
        :active="draft.includes(f)"
        @click="toggleFoot(f)"
      />
    </div>

    <div class="bass__add">
      <UiFormField
        v-slot="{ id }"
        class="bass__field"
        :label="t('instrument.bass.registerName')"
      >
        <UiTextInput
          :id="id"
          v-model="draftName"
          :disabled="reeds === 0"
          @enter="addRegister"
        />
      </UiFormField>
      <UiToolKey
        icon="mdi-plus"
        :label="t('instrument.builder.add')"
        :disabled="!draft.length || !draftName.trim()"
        @click="addRegister"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
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
  const name = draftName.value.trim()
  if (!draft.value.length || !name) return
  if (registers.value.some(r => r.name === name)) return
  registers.value = [...registers.value, { name, feet: [...draft.value] }]
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
  gap: 0.8cqh;
}

.bass__row {
  align-items: center;
  display: flex;
  gap: 0.8cqw;
}

.bass__label {
  color: rgb(var(--v-theme-ink));
  flex: 1 1 auto;
  font-size: 1.65cqh;
}

.bass__feet {
  align-items: center;
  display: flex;
  gap: 0.6cqw;
  min-height: 4cqh;
}

.bass__add {
  align-items: flex-end;
  display: flex;
  gap: 0.8cqw;
}

.bass__field {
  flex: 1 1 auto;
  min-width: 0;
}
</style>
