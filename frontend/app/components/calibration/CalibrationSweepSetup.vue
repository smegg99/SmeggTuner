<template>
  <div class="sw sw--setup">
    <p
      v-if="!compact || options.length"
      class="sw__prompt"
    >
      {{ t('calibrate.sweep.pickRegister') }}
    </p>

    <CalibrationRegisterKeys
      v-if="!compact"
      :registers="registers"
      :bass-options="bassOptions"
      @begin="(name: string, bass?: boolean) => emit('begin', name, bass)"
    />

    <div
      v-if="compact && options.length"
      class="sw__choice"
    >
      <v-select
        v-model="choice"
        :items="options"
        item-title="label"
        item-value="id"
        :aria-label="t('calibrate.sweep.pickRegister')"
        hide-details
      />
      <UiToolKey
        icon="mdi-play"
        :title="t('calibrate.sweep.startPlain')"
        @click="startChoice"
      />
    </div>
    <!-- A solo register's sweep doubles as calibration of that rank's voice: the app learns how
         loud its partials stand, and compound registers can then tell that rank tuned dead onto a
         partial from a rank not sounding at all. -->
    <p
      v-if="hasSolo && !compact"
      class="sw__hint"
    >
      {{ t('calibrate.sweep.soloTeaches') }}
    </p>

    <!-- No register described: sweep into numbered columns, like an undescribed session. -->
    <template v-if="!registers.length && !bassOptions.length">
      <p class="sw__hint">
        {{ t('calibrate.sweep.noRegisters') }}
      </p>
      <v-btn
        :title="t('calibrate.sweep.startPlain')"
        :aria-label="t('calibrate.sweep.startPlain')"
        class="tone-key"
        style="--key-tone: rgb(var(--v-theme-goal))"
        prepend-icon="mdi-play"
        :data-icon="'mdi-play'"
        @click="emit('begin', '')"
      >
        {{ t('calibrate.sweep.startPlain') }}
      </v-btn>
    </template>

    <div class="sw__actions">
      <UiToolKey
        icon="mdi-arrow-left"
        :label="t('calibrate.sweep.redoRange')"
        :title="t('calibrate.sweep.redoRange')"
        @click="emit('redoRange')"
      />
      <v-menu v-if="compact && hasSolo">
        <template #activator="{ props: menu }">
          <UiToolKey
            v-bind="menu"
            icon="mdi-information-outline"
            :title="t('calibrate.sweep.soloTeaches')"
          />
        </template>
        <p class="sw__explanation">
          {{ t('calibrate.sweep.soloTeaches') }}
        </p>
      </v-menu>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import CalibrationRegisterKeys from './CalibrationRegisterKeys.vue'
import type { BassRegister, Register } from '~/types/session'

const props = defineProps<{
  registers: Register[]
  bassReeds?: number
  bassRegisters?: BassRegister[]
}>()
const emit = defineEmits<{
  begin: [name: string, bass?: boolean]
  redoRange: []
}>()

const { t } = useI18n()

const hasSolo = computed(() =>
  props.registers.some(r => r.banks.length === 1)
  || (props.bassRegisters ?? []).some(r => r.feet.length === 1))

// The whole machine first - the only choice on a fixed bass - then each declared switch.
const bassOptions = computed(() => {
  if (!props.bassReeds) return []
  const out = [{ register: '', label: t('calibrate.sweep.bassAll') }]
  for (const r of props.bassRegisters ?? []) {
    out.push({ register: r.name, label: `${r.name} - ${r.feet.map(f => `${f}'`).join('+')}` })
  }
  return out
})
const compact = useCompactLayout()
const choice = ref(0)
const options = computed(() => [
  ...props.registers.map(r => ({ register: r.name, label: r.name, bass: false })),
  ...bassOptions.value.map(r => ({ ...r, bass: true })),
].map((r, id) => ({ ...r, id })))
function startChoice() {
  const item = options.value[choice.value]
  if (item) emit('begin', item.register, item.bass)
}
</script>

<style scoped>
.sw {
  align-items: center;
  display: flex;
  flex-direction: column;
  gap: var(--gap-wide);
  justify-content: center;
  width: 100%;
}

.sw__prompt {
  color: rgb(var(--v-theme-ink));
  font-size: 15.5px;
  margin: 0;
  max-width: 44ch;
}

.sw__hint {
  color: rgb(var(--v-theme-ink3));
  font-size: 12px;
  margin: 0;
  max-width: 40ch;
}

.sw__choice { display: flex; align-items: center; gap: 8px; width: 100%; max-width: 360px; }
.sw__choice > .v-select { flex: 1; min-width: 0; }
@media (max-width: 959px), (max-height: 559px) {
  .sw { gap: 8px; }
  .sw__prompt { font-size: 13px; }
  .sw__hint { font-size: 11px; }
}
.sw__actions { display: flex; align-items: center; gap: 8px; }
.sw__explanation { padding: 12px; max-width: min(320px, calc(100vw - 24px)); font-size: 12px; background: rgb(var(--v-theme-chrome)); border: 1px solid rgb(var(--v-theme-line)); border-radius: var(--app-radius-panel); }
</style>
