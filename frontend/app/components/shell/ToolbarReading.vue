<template>
  <div class="d-flex align-center ga-1">
    <UiToolKey
      icon="mdi-snowflake"
      :active="frozen"
      :disabled="!canFreeze"
      :title="t('tuner.freeze.hint')"
      @click="setFrozen(!frozen)"
    />

    <!-- Which keyboard the bench faces; only an instrument with a declared bass machine has two. -->
    <UiToolKey
      v-if="hasBass"
      :label="t(bench?.bass ? 'tuner.side.bass' : 'tuner.side.treble')"
      :title="t('tuner.side.hint')"
      :disabled="advancing"
      @click="setBass(!bench?.bass)"
    />

    <!-- The bass switch pulled; a fixed machine offers only the whole ladder. -->
    <v-menu
      v-if="bench?.bass"
      :disabled="advancing || !bassRegisters.length"
    >
      <template #activator="{ props: menu }">
        <UiToolKey
          v-bind="menu"
          :label="bassLabel"
          :title="t('tuner.bassRegister.hint')"
          :fixed="REGISTER_W"
          :disabled="advancing || !bassRegisters.length"
          :caret="bassRegisters.length > 0"
        />
      </template>

      <UiMenuSheet>
        <UiMenuItem
          :label="t('tuner.bassRegister.all')"
          :active="!bench?.bassRegister"
          @click="setBassRegister('')"
        />
        <UiMenuItem
          v-for="r in bassRegisters"
          :key="r.name"
          :label="`${r.name} - ${r.feet.map(f => `${f}'`).join('+')}`"
          :active="r.name === bench?.bassRegister"
          @click="setBassRegister(r.name)"
        />
      </UiMenuSheet>
    </v-menu>

    <!-- The register can't be derived from audio; it's asked, and it also sets the reed count. -->
    <v-menu
      v-else-if="registers.length"
      :disabled="advancing"
    >
      <template #activator="{ props: menu }">
        <UiToolKey
          v-bind="menu"
          :label="registerLabel"
          :title="advancing ? t('tuner.register.lockedPlaying') : t('tuner.register.hint')"
          :fixed="REGISTER_W"
          :disabled="advancing"
          caret
        />
      </template>

      <UiMenuSheet>
        <UiMenuItem
          v-for="r in registers"
          :key="r.name"
          :label="`${r.name} - ${t('tuner.reeds.count', { n: r.banks.length })}`"
          :active="r.name === bench?.register"
          @click="setRegister(r.name)"
        />
      </UiMenuSheet>
    </v-menu>

    <!-- No registers described: keep the control disabled rather than let it vanish. -->
    <UiToolKey
      v-else-if="active"
      :label="registerLabel"
      :title="t('tuner.register.hint')"
      :fixed="REGISTER_W"
      disabled
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useSessions } from '~/composables/useSessions'
import { useTransport } from '~/composables/useTransport'
import { useTuner } from '~/composables/useTuner'

const { t } = useI18n()
const { frozen, setFrozen, canFreeze, reedCount } = useTuner()
const { active, setRegister, setBass, setBassRegister } = useSessions()
// A file records one register; the switch is locked mid-playback so its reeds aren't re-filed.
const { advancing } = useTransport()

const bench = computed(() => active.value?.bench)
const registers = computed(() => active.value?.instrument?.registers ?? [])
const hasBass = computed(() => (active.value?.instrument?.bassReeds ?? 0) > 0)
const bassRegisters = computed(() => active.value?.instrument?.bassRegisters ?? [])

// Declared width: a register name is arbitrary data, so it's bounded outright and ellipsised.
const REGISTER_W = 9

// The pulled register, or the reed count the engine is resolving when none is described.
const registerLabel = computed(() =>
  bench.value?.register || t('tuner.reeds.count', { n: bench.value?.reeds ?? reedCount.value }),
)

// The pulled bass switch, or the whole ladder's feet on a fixed machine.
const bassLabel = computed(() =>
  bench.value?.bassRegister
  || (bench.value?.bassFeet ?? []).map(f => `${f}'`).join('+')
  || t('tuner.side.bass'),
)
</script>
