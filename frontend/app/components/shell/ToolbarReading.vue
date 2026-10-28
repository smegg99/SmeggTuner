<template>
  <div class="key-row">
    <UiToolKey
      icon="mdi-snowflake"
      :active="hold"
      :title="t('tuner.hold.hint')"
      @click="setHold(!hold)"
    />

    <!-- Relearns the room's background noise; a microphone learns it on its first run and keeps it after. -->
    <UiToolKey
      icon="mdi-ear-hearing"
      :disabled="!running || file || learning"
      :title="t('tuner.input.relearnHint')"
      @click="relearn"
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
        <v-btn
          v-bind="menu"
          :title="t('tuner.bassRegister.hint')"
          :disabled="advancing || !bassRegisters.length"
          :aria-label="t('tuner.bassRegister.hint')"
          :width="(REGISTER_W) * 11"
          append-icon="mdi-menu-down"
        >
          {{ bassLabel }}
        </v-btn>
      </template>

      <v-list class="tuner-menu">
        <v-list-item
          :active="!bench?.bassRegister"
          :append-icon="(!bench?.bassRegister) ? 'mdi-check' : undefined"
          :title="t('tuner.bassRegister.all')"
          @click="setBassRegister('')"
        />
        <v-list-item
          v-for="r in bassRegisters"
          :key="r.name"
          :active="r.name === bench?.bassRegister"
          :append-icon="(r.name === bench?.bassRegister) ? 'mdi-check' : undefined"
          :title="`${r.name} - ${r.feet.map(f => `${f}'`).join('+')}`"
          @click="setBassRegister(r.name)"
        />
      </v-list>
    </v-menu>

    <!-- The register can't be derived from audio; it's asked, and it also sets the reed count. -->
    <v-menu
      v-else-if="registers.length"
      :disabled="advancing"
    >
      <template #activator="{ props: menu }">
        <v-btn
          v-bind="menu"
          :title="advancing ? t('tuner.register.lockedPlaying') : t('tuner.register.hint')"
          :disabled="advancing"
          :aria-label="advancing ? t('tuner.register.lockedPlaying') : t('tuner.register.hint')"
          :width="(REGISTER_W) * 11"
          append-icon="mdi-menu-down"
        >
          {{ registerLabel }}
        </v-btn>
      </template>

      <v-list class="tuner-menu">
        <v-list-item
          v-for="r in registers"
          :key="r.name"
          :active="r.name === bench?.register"
          :append-icon="(r.name === bench?.register) ? 'mdi-check' : undefined"
          :title="`${r.name} - ${t('tuner.reeds.count', { n: r.banks.length })}`"
          @click="setRegister(r.name)"
        />
      </v-list>
    </v-menu>

    <!-- No registers described: keep the control disabled rather than let it vanish. -->
    <v-btn
      v-else-if="active"
      :title="t('tuner.register.hint')"
      disabled
      :aria-label="t('tuner.register.hint')"
      :width="(REGISTER_W) * 11"
    >
      {{ registerLabel }}
    </v-btn>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAudioDevices } from '~/composables/useAudioDevices'
import { useSessions } from '~/composables/useSessions'
import { useTransport } from '~/composables/useTransport'
import { useTuner } from '~/composables/useTuner'

const { t } = useI18n()
const { hold, setHold, reedCount, running, relearn, state } = useTuner()
// The engine reports initializing while it learns the background; a second press would restart it.
const learning = computed(() => state.value === 'initializing')
const { current } = useAudioDevices()
// A recording has no room to learn: its background is never learned.
const file = computed(() => current.value.kind === 'file')
const { active, setRegister, setBass, setBassRegister } = useSessions()
// A file records one register; the switch is locked mid-playback so its reeds aren't re-filed.
const { advancing } = useTransport()

const bench = computed(() => active.value?.bench)
const registers = computed(() => active.value?.instrument?.registers ?? [])
const hasBass = computed(() => (active.value?.instrument?.bassReeds ?? 0) > 0)
const bassRegisters = computed(() => active.value?.instrument?.bassRegisters ?? [])

// Declared width: a register name is arbitrary data, so it's bounded outright and ellipsised.
const REGISTER_W = 8

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
