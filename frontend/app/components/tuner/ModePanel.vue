<template>
  <div class="modes">
    <v-select
      v-if="compact"
      class="modes__select"
      :model-value="mode"
      :items="MODES"
      item-title="label"
      item-value="value"
      :aria-label="t('tuner.panel.modeLabel')"
      hide-details
      @update:model-value="v => v !== mode && setMode(v as PanelMode)"
    />
    <UiToolGroup
      v-else
      :model-value="mode"
      :items="MODES"
      :label="t('tuner.panel.modeLabel')"
      @update:model-value="v => v !== mode && setMode(v as PanelMode)"
    />

    <div
      class="modes__reeds"
      role="group"
      :aria-label="t('tuner.panel.reeds')"
    >
      <span class="modes__label">{{ t('tuner.panel.reeds') }}</span>

      <!-- The three keys stay in every mode so switching never moves the row; a key that
         cannot change is lit or greyed, not hidden. -->
      <UiToolKey
        v-for="b in boxes"
        :key="b.box"
        :class="{ 'modes__key--locked': b.locked }"
        :label="String(b.box)"
        :active="b.on"
        :disabled="b.chord"
        :aria-disabled="b.locked || undefined"
        :title="boxTitle(b)"
        @click="toggle(b)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ToolItem } from '@smegg99/s99ab/types'
import { useTuner } from '~/composables/useTuner'
import { PANEL_MODES, octaveOfBox, panelBoxes } from '~/utils/panel'
import type { PanelBox, PanelMode } from '~/utils/panel'

interface ShownBox extends PanelBox {
  /** on and not switchable: the octaves base, or the last box left on */
  locked: boolean
  /** chords mode measures its three tones itself */
  chord: boolean
}

const { t } = useI18n()
const compact = useCompactLayout()
const { mode, reedBoxes, octaveOffsets, setMode, setReedBox, setOctave } = useTuner()

const MODES = computed<ToolItem[]>(() => PANEL_MODES.map(m => ({
  value: m,
  label: t(`tuner.panel.modes.${m}`),
  title: t(`tuner.panel.modeHints.${m}`),
})))

const boxes = computed<ShownBox[]>(() => {
  if (mode.value === 'chords') {
    return [1, 2, 3].map(box => ({ box, on: false, fixed: false, last: false, locked: false, chord: true }))
  }
  return panelBoxes(mode.value, reedBoxes.value, octaveOffsets.value)
    .map(b => ({ ...b, locked: b.fixed || b.last, chord: false }))
})

function boxTitle(b: ShownBox): string {
  if (b.chord) return t('tuner.panel.boxChord')
  if (b.fixed) return t('tuner.panel.boxBase')
  if (b.last) return t('tuner.panel.boxLast')
  if (mode.value === 'octaves') return t('tuner.panel.boxOctave', { box: b.box, octaves: b.box - 1 }, b.box - 1)
  return t('tuner.panel.boxReed', { n: b.box })
}

function toggle(b: ShownBox) {
  if (b.locked || b.chord) return
  if (mode.value === 'octaves') void setOctave(octaveOfBox(b.box), !b.on)
  else void setReedBox(b.box, !b.on)
}
</script>

<style scoped>
.modes {
  align-items: center;
  display: flex;
  justify-content: space-between;
  gap: var(--gap-wide);
  min-width: 0;
}

.modes__reeds { display: flex; align-items: center; gap: var(--s99ab-control-gap); }
.modes__select { flex: 1 1 120px; min-width: 0; max-width: 220px; }
.modes__reeds :deep(.key) { min-width: var(--s99ab-control-h); }

.modes__label {
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
  white-space: nowrap;
}

.modes__key--locked {
  cursor: default;
}

@media (max-width: 380px) {
  .modes__label { display: none; }
}
</style>
