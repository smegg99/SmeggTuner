<template>
  <div class="cp">
    <div class="cp__top">
      <span
        class="cp__preview"
        :style="{ background: current }"
      />
      <UiTextInput
        class="cp__hex"
        :model-value="hexDraft"
        :maxlength="7"
        placeholder="#rrggbb"
        :invalid="hexDraft !== '' && normalizeHex(hexDraft) === null"
        @update:model-value="onHex"
        @commit="hexDraft = current"
      />
    </div>

    <div
      v-for="ch in channels"
      :key="ch.key"
      class="cp__row"
    >
      <span class="cp__name">{{ ch.label }}</span>
      <span
        class="cp__slider"
        :style="{ background: ch.track }"
      >
        <span
          class="cp__thumb"
          :style="{ insetInlineStart: `${(ch.value / ch.max) * 100}%` }"
        />
        <input
          class="cp__input"
          type="range"
          min="0"
          :max="ch.max"
          step="1"
          :value="ch.value"
          :aria-label="ch.label"
          @input="ch.set(Number(($event.target as HTMLInputElement).value))"
        >
      </span>
    </div>

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
  gap: 1.4cqh;
}

.cp__top {
  align-items: stretch;
  display: flex;
  gap: 1cqw;
}

.cp__preview {
  aspect-ratio: 1;
  block-size: 6cqh;
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  flex: 0 0 auto;
}

.cp__hex {
  flex: 1 1 auto;
  font-family: var(--font-mono);
}

.cp__row {
  align-items: center;
  display: flex;
  gap: 1cqw;
}

.cp__name {
  color: rgb(var(--v-theme-ink2));
  flex: 0 0 auto;
  font-size: 1.5cqh;
  inline-size: 9cqw;
}

.cp__slider {
  block-size: 2.4cqh;
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  flex: 1 1 auto;
  position: relative;
}

/* mix-blend difference inverts the thumb line against any colour under it, so it reads on every band */
.cp__thumb {
  background: #fff;
  block-size: 100%;
  inline-size: 2px;
  inset-block-start: 0;
  mix-blend-mode: difference;
  pointer-events: none;
  position: absolute;
  transform: translateX(-50%);
}

.cp__input {
  block-size: 100%;
  cursor: pointer;
  inline-size: 100%;
  inset: 0;
  margin: 0;
  opacity: 0;
  position: absolute;
}

.cp__input:focus-visible {
  outline: 2px solid rgb(var(--v-theme-ink));
  outline-offset: 2px;
}

.cp__presets {
  display: grid;
  gap: 0.6cqw;
  grid-auto-columns: 1fr;
  grid-auto-flow: column;
  padding-block-start: 0.4cqh;
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
