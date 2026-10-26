<!-- app/components/shell/AppStatus.vue -->
<template>
  <div class="status">
    <ShellAppLogo class="status__logo" />

    <!-- Rides the window bar so it survives room switches. See useSessionProgress. -->
    <ShellSessionLabel
      v-if="WORKSHOP"
      class="status__session"
      :progress="progress"
    />

    <!-- The tuner's last refusal or failure, where the reference opens a message box; it clears itself. -->
    <span
      v-if="errorText"
      class="status__error"
      role="status"
      :title="errorText"
    >{{ errorText }}</span>

    <span class="status__spacer" />

    <span
      v-if="WEB"
      class="status__experimental"
    >{{ t('web.experimental') }}</span>

    <!-- Just the tag; commit and full build stamp are in the About dialog. -->
    <span
      v-if="version"
      class="status__ver"
    >{{ version }}</span>

    <ShellStatusActions />
  </div>
</template>

<script setup lang="ts">
import { WEB, WORKSHOP } from '~/utils/features'
// Belongs to the window, not a view, so it survives switching between rooms.
import { computed, onBeforeUnmount, watch } from 'vue'
import { useSessionProgress } from '~/composables/useSessionProgress'
import { error as sessionError } from '~/composables/sessionStore'

const ERROR_SHOWN_MS = 6000

const { t } = useI18n()

// Cleared after a while so the same refusal shows again the next time it happens.
const { error } = useTuner()
const errorText = computed(() => {
  const key = sessionError.value || error.value
  return key ? t(key) : ''
})
let errorTimer: ReturnType<typeof setTimeout> | undefined
watch(error, (key) => {
  clearTimeout(errorTimer)
  if (key) {
    errorTimer = setTimeout(() => {
      error.value = ''
    }, ERROR_SHOWN_MS)
  }
})
onBeforeUnmount(() => clearTimeout(errorTimer))

const progress = useSessionProgress()

// The build reads this from VERSION. The commit is in the About dialog.
const config = useRuntimeConfig()
const version = config.public.version
</script>

<style scoped>
.status {
  align-items: center;
  display: flex;
  gap: var(--s99ab-control-gap);
  min-width: 0;
  color: rgb(var(--v-theme-ink3));
  font-size: var(--s99ab-meta-size);
  overflow: hidden;
  white-space: nowrap;
}

.status__logo {
  --app-logo-h: 15px;
  flex: 0 0 auto;
}

/* Yields width first: when tight, the session name truncates before the keys get pushed off. */
.status__session {
  flex: 0 1 auto;
  min-width: 0;
}

.status__spacer {
  flex: 1 1 auto;
}

.status__error {
  color: rgb(var(--v-theme-error));
  flex: 0 1 auto;
  font-size: 11px;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status__experimental {
  border: 1px solid rgb(var(--v-theme-warning));
  border-radius: 3px;
  color: rgb(var(--v-theme-warning));
  flex: 0 0 auto;
  font-size: 10px;
  letter-spacing: 0.04em;
  padding: 0 4px;
  text-transform: uppercase;
}

.status__ver {
  color: rgb(var(--v-theme-ink3));
  flex: 0 0 auto;
  font-family: var(--font-mono);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
}

@media (max-width: 959px), (max-height: 559px) {
  .status__ver { display: none; }

  .status__logo { --app-logo-h: 12px; }
  .status__session { flex: 1 1 auto; }
  .status__spacer { display: none; }
}
</style>
