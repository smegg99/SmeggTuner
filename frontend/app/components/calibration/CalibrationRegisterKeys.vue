<!-- app/components/calibration/CalibrationRegisterKeys.vue -->
<template>
  <!-- The registers a sweep can start on, as keys showing their ranks. Its root takes no box, so both
       rows stay items of the setup's column. -->
  <div class="sw__picks">
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
  </div>
</template>

<script setup lang="ts">
import type { Register } from '~/types/session'

defineProps<{ registers: Register[], bassOptions: { register: string, label: string }[] }>()
const emit = defineEmits<{ begin: [name: string, bass?: boolean] }>()
</script>

<style scoped>
.sw__picks { display: contents; }

.sw__registers {
  display: flex;
  flex-wrap: wrap;
  gap: var(--gap);
  justify-content: center;
  max-width: 577.5px;
}

.sw__reg {
  align-items: center;
  background: rgb(var(--v-theme-raised));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 6.4px;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: var(--gap);
  padding: var(--gap-wide);
  transition: border-color 120ms, background 120ms;
}

.sw__reg:hover {
  background: rgb(var(--v-theme-chrome2));
  border-color: rgb(var(--v-theme-ink3));
}

.sw__regsym {
  height: 49px;
}

.sw__regname {
  color: rgb(var(--v-theme-ink));
  font-size: 11px;
  font-weight: 600;
}
</style>
