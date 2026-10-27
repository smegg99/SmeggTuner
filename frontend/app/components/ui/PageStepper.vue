<template>
  <nav
    class="pages"
    :aria-label="label"
  >
    <UiToolKey
      icon="mdi-chevron-left"
      :title="t('common.previous')"
      :disabled="modelValue <= 0"
      @click="emit('update:modelValue', modelValue - 1)"
    />
    <span class="pages__label"><slot>{{ label }}</slot></span>
    <span class="pages__count">{{ modelValue + 1 }} / {{ Math.max(1, count) }}</span>
    <UiToolKey
      icon="mdi-chevron-right"
      :title="t('common.next')"
      :disabled="modelValue >= count - 1"
      @click="emit('update:modelValue', modelValue + 1)"
    />
  </nav>
</template>

<script setup lang="ts">
defineProps<{ modelValue: number, count: number, label: string }>()
const emit = defineEmits<{ 'update:modelValue': [number] }>()
const { t } = useI18n()
</script>

<style scoped>
.pages { display: flex; flex: 0 0 auto; align-items: center; gap: 8px; min-width: 0; }
.pages__label { flex: 1 1 auto; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 12px; }
.pages__count { font-size: 11px; color: rgb(var(--v-theme-ink3)); font-variant-numeric: tabular-nums; white-space: nowrap; }
</style>
