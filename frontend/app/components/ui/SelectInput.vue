<template>
  <v-menu :disabled="disabled">
    <template #activator="{ props: menu }">
      <button
        :id="id"
        v-bind="menu"
        type="button"
        class="select"
        :class="{ 'select--disabled': disabled }"
        :disabled="disabled"
      >
        <span class="select__value">{{ current?.label ?? placeholder }}</span>
        <v-icon
          icon="mdi-menu-down"
          class="select__caret"
        />
      </button>
    </template>

    <!-- our sheet, not v-list: it ellipsises long names -->
    <UiMenuSheet>
      <UiMenuItem
        v-for="item in items"
        :key="String(item.value)"
        :label="item.label"
        :active="item.value === modelValue"
        @click="emit('update:modelValue', item.value)"
      />
    </UiMenuSheet>
  </v-menu>
</template>

<script setup lang="ts">
import { computed } from 'vue'

// Not v-select (an unreshapeable Material field). label, not title: matches ToolItem/MenuItem.
export interface SelectItem {
  value: string
  label: string
}

const props = defineProps<{
  modelValue: string
  items: readonly SelectItem[]
  placeholder?: string
  disabled?: boolean
  id?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [string]
}>()

const current = computed(() => props.items.find(item => item.value === props.modelValue))
</script>

<style scoped>
.select {
  align-items: center;
  background: rgb(var(--v-theme-well));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  color: rgb(var(--v-theme-ink));
  display: flex;
  font-family: var(--font-sans);
  font-size: 1.75cqh;
  gap: 0.5cqw;
  height: 4.4cqh;
  min-width: 0;
  padding-inline: 0.7cqw;
  text-align: start;
  width: 100%;
}

.select:hover:not(:disabled) {
  border-color: rgb(var(--v-theme-ink3));
}

.select:focus-visible {
  border-color: rgb(var(--v-theme-ink2));
  outline: none;
}

.select--disabled {
  cursor: default;
  opacity: 0.38;
}

.select__value {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.select__caret {
  color: rgb(var(--v-theme-ink3));
  flex: 0 0 auto;
  font-size: 2cqh;
}
</style>
