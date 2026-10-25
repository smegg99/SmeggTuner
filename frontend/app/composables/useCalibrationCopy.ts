// What the sound-card calibration room says in each state: the step, headline, line under it and figures.
import { computed, type Ref } from 'vue'
import type { CalibrationDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/audio/models.js'
import type { TracePoint } from '~/composables/useSoundCard'
import type { DeviationUnit } from '~/utils/deviation'
import { clockPitch, signedPercent } from '~/utils/soundCard'

// services/audio's settled bar: a 5 s spread at or under this many ppm.
const TARGET_PPM = 5

export function useCalibrationCopy(
  calibration: Readonly<Ref<CalibrationDTO | null>>,
  trace: Readonly<Ref<TracePoint[]>>,
  unit: Readonly<Ref<DeviationUnit>>,
) {
  const { t } = useI18n()
  const status = computed(() => calibration.value?.status ?? 'uncalibrated')

  // 1 warming up, 2 measuring, 3 an outcome is stored, 0 never run.
  const stage = computed(() => {
    switch (status.value) {
      case 'preparing': return 1
      case 'calibrating': return 2
      case 'calibrated':
      case 'unreliable': return 3
    }
    return 0
  })

  // A clock correction as the pitch error it stands for, in the unit the tuner reads in.
  function pitch(percent: number): string {
    const unitLabel = t(`tuner.boxes.unit.${unit.value === 'cents' ? 'cents' : 'reference'}`)
    return `${clockPitch(percent, unit.value)} ${unitLabel}`
  }

  const headline = computed(() => {
    switch (status.value) {
      case 'preparing': return t('settings.calibration.screen.preparing')
      case 'calibrating': return t('settings.calibration.screen.calibrating')
      case 'calibrated': return t('settings.calibration.screen.calibrated', { percent: signedPercent(calibration.value?.percent ?? 0) })
      case 'unreliable': return t('settings.calibration.screen.unreliable')
    }
    return t('settings.calibration.screen.lead')
  })

  const lead = computed(() => {
    const c = calibration.value
    switch (status.value) {
      case 'preparing': return t('settings.calibration.screen.phase.preparing')
      case 'calibrating': return t(c?.settled ? 'settings.calibration.screen.phase.confirming' : 'settings.calibration.screen.phase.settling')
      case 'calibrated': return t('settings.calibration.screen.pitch', { pitch: pitch(c?.percent ?? 0) })
      case 'unreliable': return t('settings.calibration.screen.retry')
    }
    return t('settings.calibration.screen.how')
  })

  // Warming up has no clock figures yet; the rest are known from the first block. A device not run
  // since the app started has only its stored offset.
  const stats = computed(() => {
    const c = calibration.value
    const run = !!c && (c.running || trace.value.length > 0)
    const measuring = c?.status === 'calibrating'
    const offset = measuring || c?.status === 'calibrated'
    return [
      { key: 'offset', value: c && offset ? `${signedPercent(c.percent)} %, ${pitch(c.percent)}` : '--' },
      { key: 'steadiness', value: c && measuring ? t('settings.calibration.screen.steadiness', { ppm: (c.spread * 1e4).toFixed(1), target: TARGET_PPM }) : '--' },
      { key: 'rate', value: run && c.rate ? `${c.rate} Hz` : '--' },
      { key: 'elapsed', value: run ? `${c.elapsed} s` : '--' },
      { key: 'dropouts', value: run ? String(c.dropouts) : '--' },
    ]
  })

  return { status, stage, headline, lead, stats }
}
