<template>
  <v-dialog
    v-model="open"
    width="var(--dlg-sm)"
  >
    <UiDialogCard
      :title="title"
      :confirm-label="confirmLabel ?? t('common.delete')"
      confirm-color="error"
      :loading="loading"
      @close="open = false"
      @cancel="open = false"
      @confirm="emit('confirm')"
    >
      <p class="dialog-warn">
        {{ body }}
      </p>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  title: string
  body: string
  confirmLabel?: string
  loading?: boolean
}>(), {
  confirmLabel: undefined,
  loading: false,
})

const open = defineModel<boolean>({ required: true })
const emit = defineEmits<{ confirm: [] }>()

const { t } = useI18n()
</script>

<style scoped>
.dialog-warn {
  color: rgb(var(--v-theme-ink2));
  font-size: 1.65cqh;
  line-height: 1.5;
  margin: 0;
}
</style>
