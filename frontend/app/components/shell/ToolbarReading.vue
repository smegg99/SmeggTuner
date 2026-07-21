<template>
  <div class="d-flex align-center ga-1">
    <UiToolKey
      icon="mdi-snowflake"
      :active="frozen"
      :disabled="!canFreeze"
      :title="t('tuner.freeze.hint')"
      @click="setFrozen(!frozen)"
    />

    <!-- The register can't be derived from audio; it's asked, and it also sets the reed count. -->
    <v-menu
      v-if="registers.length"
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
const { active, setRegister } = useSessions()
// A file records one register; the switch is locked mid-playback so its reeds aren't re-filed.
const { advancing } = useTransport()

const bench = computed(() => active.value?.bench)
const registers = computed(() => active.value?.instrument?.registers ?? [])

// Declared width: a register name is arbitrary data, so it's bounded outright and ellipsised.
const REGISTER_W = 9

// The pulled register, or the reed count the engine is resolving when none is described.
const registerLabel = computed(() =>
  bench.value?.register || t('tuner.reeds.count', { n: bench.value?.reeds ?? reedCount.value }),
)
</script>
