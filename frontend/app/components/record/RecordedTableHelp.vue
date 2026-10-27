<template>
  <v-dialog
    v-model="open"
    :aria-label="t('record.table.helpTitle')"
    width="var(--dlg-sm)"
  >
    <UiDialogCard
      :title="t('record.table.helpTitle')"
      @close="open = false"
    >
      <div class="help">
        <p class="help__lead">
          {{ t('record.table.unitHint') }}
        </p>

        <dl class="help__marks">
          <div class="help__mark">
            <dt>{{ t('record.table.manual') }}</dt>
            <dd>{{ t('record.table.manualHint') }}</dd>
          </div>
        </dl>
      </div>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
const open = defineModel<boolean>({ default: false })
const { t } = useI18n()
</script>

<style scoped>
.help {
  display: flex;
  flex-direction: column;
  gap: var(--gap);
}

.help__lead {
  color: rgb(var(--v-theme-ink));
  font-size: 12px;
  margin: 0;
}

.help__marks {
  display: flex;
  flex-direction: column;
  gap: var(--gap);
  margin: 0;
}

.help__mark dt {
  color: rgb(var(--v-theme-ink));
  font-size: 11px;
  font-weight: 700;
}

.help__mark dd {
  color: rgb(var(--v-theme-ink2));
  font-size: 11px;
  margin: 1.6px 0 0;
}
</style>
