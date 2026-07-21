<template>
  <button
    type="button"
    class="key"
    :class="{
      'key--active': active,
      'key--icon': !label,
      'key--fixed': fixed !== undefined,
      'key--tone': !!tone,
    }"
    :style="keyStyle"
    :disabled="disabled"
    :title="title ?? label"
    :data-icon="icon"
    :aria-label="title ?? label"
    :aria-pressed="active"
  >
    <v-icon
      v-if="icon"
      :icon="icon"
      class="key__icon"
    />

    <!-- label plus a hidden ghost of every other face; the key sizes to the widest so it never moves when it flips -->
    <span
      v-if="label"
      class="key__label"
    >
      <span class="key__text">{{ label }}</span>
      <span
        v-for="ghost in ghosts"
        :key="ghost"
        class="key__ghost"
        aria-hidden="true"
      >{{ ghost }}</span>
    </span>

    <v-icon
      v-if="caret"
      icon="mdi-menu-down"
      class="key__caret"
    />
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { buildKeyStyle, ghostFaces, type KeyTone } from '~/utils/toolKey'

// Not a v-btn (its height is unreachable behind density props).
const props = defineProps<{
  icon?: string
  label?: string
  active?: boolean
  disabled?: boolean
  title?: string
  caret?: boolean
  /** the other faces this key can show; sized for all at once */
  sizeFor?: string | readonly string[]
  /** an explicit width in cqw, for a label that is data */
  fixed?: number
  /** colour tone: the mark at rest, the whole fill when active */
  tone?: KeyTone
}>()

const keyStyle = computed(() => buildKeyStyle(props.fixed, props.tone))
const ghosts = computed(() => ghostFaces(props.sizeFor))
</script>

<style scoped>
.key {
  align-items: center;
  background: rgb(var(--v-theme-raised));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  color: rgb(var(--v-theme-ink2));
  display: inline-flex;
  flex: 0 0 auto;
  font-family: var(--font-sans);
  font-size: 1.75cqh;
  font-weight: 500;
  gap: 0.45cqw;
  height: 3.9cqh;
  justify-content: center;
  line-height: 1;
  min-width: 0;
  padding-inline: 0.85cqw;
  white-space: nowrap;
}

.key:hover:not(:disabled) {
  color: rgb(var(--v-theme-ink));
}

.key:focus-visible {
  outline: 2px solid rgb(var(--v-theme-ink));
  outline-offset: 1px;
}

.key:disabled {
  cursor: default;
  opacity: 0.38;
}

/* Active: fills with accent; border kept transparent (not removed) so it keeps its size. */
.key--active {
  background: rgb(var(--v-theme-accent));
  border-color: transparent;
  color: rgb(var(--v-theme-on-accent));
}

.key--active:hover:not(:disabled) {
  background: rgb(var(--v-theme-accent));
  border-color: transparent;
  color: rgb(var(--v-theme-on-accent));
  filter: brightness(1.1);
}

/* Toned key uses --key-tone; declared after active so it wins the fill. */
.key--tone {
  color: rgb(var(--key-tone));
}

.key--tone:hover:not(:disabled) {
  color: rgb(var(--key-tone));
  filter: brightness(1.1);
}

.key--tone.key--active {
  background: rgb(var(--key-tone));
  border-color: transparent;
  color: rgb(var(--v-theme-bg));
}

.key--tone.key--active:hover:not(:disabled) {
  background: rgb(var(--key-tone));
  border-color: transparent;
  color: rgb(var(--v-theme-bg));
  filter: brightness(1.1);
}

.key:active:not(:disabled) {
  background: rgb(var(--v-theme-sunk));
}

.key--icon {
  padding-inline: 0;
  width: 3.9cqh;
}

.key__label {
  display: grid;
  justify-items: center;
  min-width: 0;
}

.key__label > * {
  grid-area: 1 / 1;
  min-width: 0;
}

.key__ghost {
  visibility: hidden;
}

/* Declared width: justify-items back to stretch so a long name ellipsises inside. */
.key--fixed {
  width: var(--key-w);
}

.key--fixed .key__label {
  justify-items: stretch;
  max-width: 100%;
}

.key--fixed .key__text {
  overflow: hidden;
  text-align: start;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.key__icon {
  font-size: 2.2cqh;
}

.key__caret {
  font-size: 2cqh;
  margin-inline-end: -0.25cqw;
  opacity: 0.7;
}
</style>
