<template>
  <UiFormField
    :label="t('settings.toleranceLabel')"
    :hint="t('settings.toleranceHint')"
  >
    <div class="tol">
      <label class="tol__one">
        <span class="tol__cap">{{ t('settings.toleranceReed') }}</span>
        <UiNumberInput
          v-model="reedDraft"
          :suffix="t('settings.toleranceUnit')"
          @commit="commitReed"
        />
      </label>
      <label class="tol__one">
        <span class="tol__cap">{{ t('settings.toleranceBeat') }}</span>
        <UiNumberInput
          v-model="beatDraft"
          :suffix="t('settings.toleranceUnit')"
          @commit="commitBeat"
        />
      </label>
    </div>
  </UiFormField>
</template>

<script setup lang="ts">
import { useConfigSync } from '~/composables/useConfigSync'
import { useDraftNumber } from '~/composables/useDraftNumber'
import { DEFAULT_BEAT_TOLERANCE, DEFAULT_TOLERANCE } from '~/types/config'

// App-wide fallback tolerances, used when no session is open (a session instrument carries its
// own, see core/session.Instrument). Bounds match the schema: above 0, no wider than 50 cents.
const TOL_MIN = 0.1
const TOL_MAX = 50

const { t } = useI18n()
const { config } = useConfigSync()

const { draft: reedDraft, commit: commitReed } = useDraftNumber(
  () => config.tuner?.tolerance ?? DEFAULT_TOLERANCE,
  value => (config.tuner.tolerance = value),
  { min: TOL_MIN, max: TOL_MAX, step: 0.1 },
)

const { draft: beatDraft, commit: commitBeat } = useDraftNumber(
  () => config.tuner?.beat_tolerance ?? DEFAULT_BEAT_TOLERANCE,
  value => (config.tuner.beat_tolerance = value),
  { min: TOL_MIN, max: TOL_MAX, step: 0.1 },
)
</script>

<style scoped>
.tol {
  display: flex;
  gap: 1.2cqw;
}

.tol__one {
  display: flex;
  flex: 1 1 0;
  flex-direction: column;
  gap: 0.3cqh;
  min-width: 0;
}

.tol__cap {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.4cqh;
}
</style>
