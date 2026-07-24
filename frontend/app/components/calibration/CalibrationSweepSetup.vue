<template>
  <div class="sw sw--setup">
    <p class="sw__prompt">
      {{ t('calibrate.sweep.pickRegister') }}
    </p>

    <div
      v-if="registers.length"
      class="sw__registers"
    >
      <button
        v-for="r in registers"
        :key="r.name"
        type="button"
        class="sw__reg"
        @click="emit('begin', r.name)"
      >
        <InstrumentRegisterSymbol
          :banks="r.banks"
          class="sw__regsym"
        />
        <span class="sw__regname">{{ r.name }}</span>
      </button>
    </div>

    <!-- The bass side sweeps its twelve buttons: the whole ladder, and each declared switch. -->
    <div
      v-if="bassOptions.length"
      class="sw__registers"
    >
      <button
        v-for="b in bassOptions"
        :key="`bass-${b.register}`"
        type="button"
        class="sw__reg"
        @click="emit('begin', b.register, true)"
      >
        <span class="sw__regname">{{ b.label }}</span>
      </button>
    </div>

    <!-- A solo register's sweep doubles as calibration of that rank's voice: the app learns how
         loud its partials stand, and compound registers can then tell that rank tuned dead onto a
         partial from a rank not sounding at all. -->
    <p
      v-if="hasSolo"
      class="sw__hint"
    >
      {{ t('calibrate.sweep.soloTeaches') }}
    </p>

    <!-- No register described: sweep into numbered columns, like an undescribed session. -->
    <template v-else>
      <p class="sw__hint">
        {{ t('calibrate.sweep.noRegisters') }}
      </p>
      <UiToolKey
        icon="mdi-play"
        tone="success"
        :label="t('calibrate.sweep.startPlain')"
        @click="emit('begin', '')"
      />
    </template>

    <UiToolKey
      icon="mdi-arrow-left"
      :label="t('calibrate.sweep.redoRange')"
      @click="emit('redoRange')"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
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
</script>

<style scoped>
.sw {
  align-items: center;
  display: flex;
  flex-direction: column;
  gap: 2cqh;
  justify-content: center;
  width: 100%;
}

.sw__prompt {
  color: rgb(var(--v-theme-ink));
  font-size: 2.2cqh;
  margin: 0;
  max-width: 44ch;
}

.sw__hint {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.7cqh;
  margin: 0;
  max-width: 40ch;
}

.sw__registers {
  display: flex;
  flex-wrap: wrap;
  gap: 1cqw;
  justify-content: center;
  max-width: 60cqw;
}

.sw__reg {
  align-items: center;
  background: rgb(var(--v-theme-raised));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 0.8cqh;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 0.8cqh;
  padding: 1.6cqh 1.6cqw;
  transition: border-color 120ms, background 120ms;
}

.sw__reg:hover {
  background: rgb(var(--v-theme-chrome2));
  border-color: rgb(var(--v-theme-ink3));
}

.sw__regsym {
  height: 7cqh;
}

.sw__regname {
  color: rgb(var(--v-theme-ink));
  font-size: 1.6cqh;
  font-weight: 600;
}
</style>
