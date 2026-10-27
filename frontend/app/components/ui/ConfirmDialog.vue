<!-- app/components/ui/ConfirmDialog.vue -->
<template>
  <!-- A small destructive confirmation: what will be lost, Cancel, and the action in the error colour. -->
  <v-dialog
    v-model="open"
    :aria-label="title"
    width="var(--dlg-sm)"
  >
    <UiDialogCard
      :title="title"
      @close="open = false"
    >
      <p class="dialog-warn">
        {{ message }}
      </p>
      <template #actions>
        <v-btn
          @click="open = false"
        >
          {{ t('common.cancel') }}
        </v-btn><v-btn
          color="error"
          @click="emit('confirm')"
        >
          {{ t('common.delete') }}
        </v-btn>
      </template>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
const open = defineModel<boolean>({ required: true })
defineProps<{ title: string, message: string }>()
const emit = defineEmits<{ confirm: [] }>()
const { t } = useI18n()
</script>
