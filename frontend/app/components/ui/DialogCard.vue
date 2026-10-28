<!-- app/components/ui/DialogCard.vue -->
<template>
  <UiPanel
    :title="title"
    class="dlg"
    flush
  >
    <template #head>
      <h2 class="dlg__title">
        {{ title }}
      </h2>
      <v-btn
        class="dlg__close"
        icon="mdi-close"
        variant="text"
        :width="32"
        :height="32"
        :min-height="32"
        :disabled="busy"
        :title="t('common.close')"
        :aria-label="t('common.close')"
        @click="emit('close')"
      />
    </template>
    <div class="dlg__body">
      <slot />
    </div>
    <template
      v-if="$slots.actions"
      #footer
    >
      <div class="dlg__foot">
        <slot name="actions" />
      </div>
    </template>
  </UiPanel>
</template>

<script setup lang="ts">
defineProps<{ title: string, busy?: boolean }>()
const emit = defineEmits<{ close: [] }>()
const { t } = useI18n()
</script>

<style scoped>
.dlg {
  background: rgb(var(--v-theme-chrome));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: var(--app-radius-panel);
  max-height: min(720px, calc(100dvh - 32px));
  min-height: 0;
  overflow: hidden;
}
/* As tall as the main bar at every window size (useBarHeight), so a dialog's head lines up with it. */
.dlg :deep(.panel__head) {
  background: rgb(var(--v-theme-chrome2));
  border-bottom: 1px solid rgb(var(--v-theme-lineSoft));
  box-sizing: border-box;
  flex-wrap: nowrap;
  gap: 16px;
  height: var(--app-bar-h, 48px);
  min-height: 0;
  padding: 0 12px 0 16px;
}
.dlg__title {
  color: rgb(var(--v-theme-ink));
  flex: 1 1 auto;
  min-width: 0;
  font-size: 16px;
  font-weight: 600;
  line-height: 1.35;
  margin: 0;
}
.dlg .dlg__close {
  flex: 0 0 auto;
  color: rgb(var(--v-theme-ink2));
}
.dlg :deep(.panel__body) { overflow: hidden; }
.dlg__body {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 16px;
  font-size: 13px;
  line-height: 1.5;
}
.dlg :deep(.panel__foot) {
  display: block;
  border-top: 1px solid rgb(var(--v-theme-lineSoft));
  padding: 12px 16px;
}
.dlg__foot {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  align-items: center;
  gap: 8px;
}
</style>
