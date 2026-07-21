<template>
  <UiPanelHead
    :title="t('record.table.label')"
    :help-title="t('record.table.helpTitle')"
    @help="emit('help')"
  >
    <template #chip>
      <UiChip v-if="count">
        {{ t('record.table.notes', count) }}
      </UiChip>
    </template>
  </UiPanelHead>
</template>

<script setup lang="ts">
defineProps<{ count: number }>()

const emit = defineEmits<{ help: [] }>()

const { t } = useI18n()
</script>
