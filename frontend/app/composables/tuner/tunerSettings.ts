import { ref } from 'vue'
import type { Ref } from 'vue'
import type { SettingsDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/tuner/models.js'
import type { PanelMode } from '~/utils/panel'
import { AUTO_NOTE } from './tunerProtocol'

/** What the engine is measuring with, as tuner:settings last said. */
export interface TunerSettings {
  a4: Ref<number>
  reedCount: Ref<number>
  /** reeds of the open session's instrument; 0 = no session, the panel decides */
  sessionReeds: Ref<number>
  mode: Ref<PanelMode>
  reedBoxes: Ref<number[]>
  octaveOffsets: Ref<number[]>
  transpose: Ref<number>
  manualNote: Ref<number>
}

export function tunerSettings(): TunerSettings {
  return {
    a4: ref(0),
    reedCount: ref(1),
    sessionReeds: ref(0),
    mode: ref<PanelMode>('tones'),
    reedBoxes: ref([1]),
    octaveOffsets: ref([12]),
    transpose: ref(0),
    manualNote: ref(AUTO_NOTE),
  }
}

export function toMode(raw: string): PanelMode {
  return raw === 'octaves' || raw === 'chords' ? raw : 'tones'
}

const sameList = (a: readonly number[], b: readonly number[]) => a.length === b.length && a.every((x, i) => x === b[i])

// applySettings takes the service's word for every field; lists are replaced only when they changed.
export function applySettings(dto: SettingsDTO, s: TunerSettings) {
  s.a4.value = dto.a4
  s.reedCount.value = dto.reedCount
  s.sessionReeds.value = dto.sessionReeds
  s.mode.value = toMode(dto.mode)
  const boxes = [...(dto.reedBoxes ?? [])].sort((a, b) => a - b)
  const offsets = [...(dto.octaveOffsets ?? [])].sort((a, b) => a - b)
  if (!sameList(boxes, s.reedBoxes.value)) s.reedBoxes.value = boxes
  if (!sameList(offsets, s.octaveOffsets.value)) s.octaveOffsets.value = offsets
  s.transpose.value = dto.transpose
  s.manualNote.value = dto.manualNote
}
