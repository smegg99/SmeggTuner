<template>
  <v-card class="dlg">
    <div class="dlg__head">
      <div class="dlg__what">
        <h2 class="dlg__title">
          {{ title }}
        </h2>
        <p
          v-if="subtitle"
          class="dlg__sub"
        >
          {{ subtitle }}
        </p>
      </div>

      <button
        type="button"
        class="dlg__close"
        :title="t('common.close')"
        :aria-label="t('common.close')"
        @click="emit('close')"
      >
        <v-icon icon="mdi-close" />
      </button>
    </div>

    <div class="dlg__body">
      <slot />
    </div>

    <div
      v-if="$slots.actions || confirmLabel"
      class="dlg__foot"
    >
      <slot name="actions">
        <v-spacer />

        <v-btn
          :disabled="loading"
          @click="emit('cancel')"
        >
          {{ cancelLabel ?? t('common.cancel') }}
        </v-btn>
        <v-btn
          :color="confirmColor"
          :loading="loading"
          :disabled="confirmDisabled"
          @click="emit('confirm')"
        >
          {{ confirmLabel }}
        </v-btn>
      </slot>
    </div>
  </v-card>
</template>

<script setup lang="ts">
// Shared dialog chrome. Setting confirm-label turns the shared footer on; the actions slot overrides it.
withDefaults(defineProps<{
  title: string
  subtitle?: string
  confirmLabel?: string
  cancelLabel?: string
  confirmColor?: string
  loading?: boolean
  confirmDisabled?: boolean
}>(), {
  subtitle: undefined,
  confirmLabel: undefined,
  cancelLabel: undefined,
  confirmColor: 'accent',
  loading: false,
  confirmDisabled: false,
})

const emit = defineEmits<{ close: [], cancel: [], confirm: [] }>()

const { t } = useI18n()
</script>

<style scoped>
.dlg {
  display: flex;
  flex-direction: column;
  max-height: 752px;
  min-height: 0;
}

/* title and close key share --dlg-glyph so they're one size */
.dlg__head {
  --dlg-glyph: 20.8px;

  align-items: center;
  border-bottom: 1px solid rgb(var(--v-theme-lineSoft));
  display: flex;
  flex: 0 0 auto;
  gap: 11px;
  padding: 9.6px 11px 9.6px 15.4px;
}

.dlg__what {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 2.4px;
  min-width: 0;
}

.dlg__title {
  color: rgb(var(--v-theme-ink));
  font-size: var(--dlg-glyph);
  font-weight: 600;
  line-height: 1.25;
  margin: 0;
  text-wrap: balance;
}

.dlg__sub {
  color: rgb(var(--v-theme-ink3));
  font-size: 12px;
  line-height: 1.4;
  margin: 0;
}

.dlg__close {
  align-items: center;
  background: none;
  block-size: 36.8px;
  border: 0;
  border-radius: 2px;
  color: rgb(var(--v-theme-ink3));
  cursor: pointer;
  display: flex;
  flex: 0 0 auto;
  inline-size: 36.8px;
  justify-content: center;
}

.dlg__close:hover {
  background: rgb(var(--v-theme-row));
  color: rgb(var(--v-theme-ink));
}

.dlg__close:focus-visible {
  outline: 2px solid rgb(var(--v-theme-ink2));
  outline-offset: 1px;
}

.dlg__close :deep(.v-icon) {
  font-size: var(--dlg-glyph);
}

/* Only the body scrolls, so a long dialog keeps its title and Save key in view. */
.dlg__body {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  padding: 12.8px 15.4px;
}

.dlg__foot {
  align-items: center;
  border-top: 1px solid rgb(var(--v-theme-lineSoft));
  display: flex;
  flex: 0 0 auto;
  gap: 6.6px;
  padding: 8.8px 15.4px;
}
</style>
