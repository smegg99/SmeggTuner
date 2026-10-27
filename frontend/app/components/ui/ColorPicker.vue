<template>
  <div class="cp">
    <div class="cp__top">
      <span
        class="cp__preview"
        :style="{ background: current }"
      />
      <v-text-field
        class="cp__hex"
        :model-value="hexDraft"
        :maxlength="7"
        placeholder="#rrggbb"
        :error="hexDraft !== '' && normalizeHex(hexDraft) === null"
        hide-details
        autocomplete="off"
        aria-label="#rrggbb"
        @update:model-value="onHex"
        @blur="hexDraft = current"
        @keyup.enter="hexDraft = current"
      />
    </div>

    <UiColorChannel
      v-for="ch in channels"
      :key="ch.key"
      :label="ch.label"
      :value="ch.value"
      :max="ch.max"
      :track="ch.track"
      @set="ch.set"
    />

    <div
      v-if="presets.length"
      class="cp__presets"
    >
      <button
        v-for="p in presets"
        :key="p"
        type="button"
        class="cp__swatch"
        :class="{ 'cp__swatch--on': normalizeHex(p) === current }"
        :style="{ background: p }"
        :title="p"
        :aria-label="p"
        @click="applyHex(p)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { hexToHsv, hsvToHex, normalizeHex } from '~/utils/color'

// HSV is the internal working state; hex is derived out. Incoming hex re-seats the sliders only when genuinely different, so our own emit doesn't reset the drag mid-move.
const props = withDefaults(defineProps<{
  modelValue: string
  presets?: string[]
}>(), {
  presets: () => [],
})

const emit = defineEmits<{ 'update:modelValue': [string] }>()

const { t } = useI18n()

const hsv = reactive(hexToHsv(props.modelValue))
const current = computed(() => hsvToHex(hsv))

const hexDraft = ref(current.value)

const channels = computed(() => [
  {
    key: 'h',
    label: t('common.color.hue'),
    value: hsv.h,
    max: 360,
    set: (n: number) => setChannel('h', n),
    track: 'linear-gradient(to right, #f00, #ff0, #0f0, #0ff, #00f, #f0f, #f00)',
  },
  {
    key: 's',
    label: t('common.color.saturation'),
    value: hsv.s,
    max: 100,
    set: (n: number) => setChannel('s', n),
    track: `linear-gradient(to right, ${hsvToHex({ h: hsv.h, s: 0, v: hsv.v })}, ${hsvToHex({ h: hsv.h, s: 100, v: hsv.v })})`,
  },
  {
    key: 'v',
    label: t('common.color.value'),
    value: hsv.v,
    max: 100,
    set: (n: number) => setChannel('v', n),
    track: `linear-gradient(to right, #000, ${hsvToHex({ h: hsv.h, s: hsv.s, v: 100 })})`,
  },
])

function setChannel(key: 'h' | 's' | 'v', n: number) {
  hsv[key] = n
  emit('update:modelValue', current.value)
}

function applyHex(hex: string) {
  const n = normalizeHex(hex)
  if (!n) return
  Object.assign(hsv, hexToHsv(n))
  emit('update:modelValue', current.value)
}

function onHex(value: string) {
  hexDraft.value = value
  applyHex(value)
}

// Field follows the sliders/swatches but not our normalisation, so "3b82f6" isn't reformatted to "#3b82f6" mid-type.
watch(current, (hex) => {
  if (normalizeHex(hexDraft.value) !== hex) hexDraft.value = hex
})

// An outside change re-seats the sliders; our own emit does not.
watch(() => props.modelValue, (hex) => {
  if (normalizeHex(hex) !== current.value) Object.assign(hsv, hexToHsv(hex))
})
</script>

<style scoped>
.cp {
  display: flex;
  flex-direction: column;
  gap: var(--gap);
}

.cp__top {
  align-items: stretch;
  display: flex;
  gap: var(--gap);
}

.cp__preview {
  aspect-ratio: 1;
  block-size: 42px;
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  flex: 0 0 auto;
}

.cp__hex {
  flex: 1 1 auto;
  font-family: var(--font-mono);
}

.cp .cp__hex :deep(.v-field__input) { min-height: 42px; }

.cp__presets {
  display: grid;
  gap: var(--gap);
  grid-auto-columns: 1fr;
  grid-auto-flow: column;
  padding-block-start: var(--gap-tight);
}

.cp__swatch {
  aspect-ratio: 1;
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  cursor: pointer;
  inline-size: 100%;
}

.cp__swatch--on {
  outline: 2px solid rgb(var(--v-theme-ink));
  outline-offset: 1px;
}

.cp__swatch:focus-visible {
  outline: 2px solid rgb(var(--v-theme-ink2));
  outline-offset: 1px;
}
</style>
