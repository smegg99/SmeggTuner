import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick, ref, shallowRef } from 'vue'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import SoundCardView from '~/components/sound-card/SoundCardView.vue'
import SoundCardFigures from '~/components/sound-card/Figures.vue'
import SoundCardMeter from '~/components/sound-card/Meter.vue'
import SoundCardSteps from '~/components/sound-card/Steps.vue'
import ToolKey from '@s99ab/components/primitives/ToolKey.vue'
import Panel from '@s99ab/components/primitives/Panel.vue'
import type { CalibrationDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/audio/models.js'

// The sound card's own room over a known calibration: what it says and offers in each state.
const calibration = shallowRef<CalibrationDTO | null>(null)
const startCalibration = vi.fn()
const cancelCalibration = vi.fn()
const setView = vi.fn()

vi.mock('~/composables/useSoundCard', () => ({
  useSoundCard: () => ({ calibration, trace: shallowRef([]), error: ref(''), startCalibration, cancelCalibration }),
}))
const pageMeta = ref('')
vi.mock('~/composables/useShell', () => ({ useShell: () => ({ setView, previous: ref('settings'), pageMeta }) }))
vi.mock('~/composables/useDeviationUnit', () => ({ useDeviationUnit: () => ref('reference') }))
vi.mock('~/composables/useAudioDevices', () => ({
  useAudioDevices: () => ({ current: ref({ kind: 'mic', deviceId: 'usb', name: 'USB Camera', path: '', loop: false }) }),
}))

const t = (key: string, params?: Record<string, unknown>) =>
  params ? `${key}(${Object.entries(params).map(([k, v]) => `${k}=${String(v)}`).join(',')})` : key
// eslint-disable-next-line @typescript-eslint/no-explicit-any
;(globalThis as any).useI18n = () => ({ t })

const dto = (o: Partial<CalibrationDTO>): CalibrationDTO =>
  ({ deviceId: 'usb', status: 'uncalibrated', percent: 0, elapsed: 0, progress: 0, settled: false, spread: 0, rate: 0,
    dropouts: 0, running: false, error: '', ...o }) as CalibrationDTO

async function view(state: Partial<CalibrationDTO>) {
  calibration.value = dto(state)
  const wrapper = mount(SoundCardView, {
    global: {
      plugins: [createVuetify({ components })],
      components: { UiToolKey: ToolKey, UiPanel: Panel, SoundCardSteps, SoundCardMeter, SoundCardFigures },
      stubs: { UiGoo: { template: '<svg class="goo" />' }, VIcon: { template: '<i class="icon" />' }, SoundCardClockTrace: { template: '<div class="trace" />' } },
    },
  })
  await nextTick()
  return wrapper
}

describe('the sound card calibration room', () => {
  it('names the device and offers to calibrate a card never measured', async () => {
    const w = await view({ status: 'uncalibrated' })
    expect(pageMeta.value).toBe('USB Camera')
    expect(w.text()).toContain('settings.calibration.screen.lead')
    expect(w.find('.goo').exists()).toBe(false)
    // The figures box stays in place with nothing measured yet.
    expect(w.find('.trace').exists()).toBe(true)
    expect(w.findAll('.sc__stat dd').map(d => d.text())).toEqual(['--', '--', '--', '--', '--'])
    const keys = w.findAll('button')
    expect(keys.map(k => k.text())).toEqual(['settings.calibration.screen.return', 'settings.calibration.start'])
    await keys[1]!.trigger('click')
    expect(startCalibration).toHaveBeenCalled()
  })

  it('shows the blobs, both halves of the bar and the clock figures while measuring', async () => {
    const w = await view({ status: 'calibrating', running: true, elapsed: 31, progress: 75, percent: 0.012,
      settled: true, spread: 0.00032, rate: 48000, dropouts: 2 })
    expect(w.find('.goo').exists()).toBe(true)
    expect(w.text()).toContain('settings.calibration.screen.calibrating')
    expect(w.text()).toContain('settings.calibration.screen.phase.confirming')
    expect(w.find('.sc__step--on').text()).toContain('settings.calibration.screen.steps.measure')
    expect(w.findAll('.sc__fill').map(f => f.attributes('style'))).toEqual(['width: 100%;', 'width: 50%;'])
    expect(w.find('.sc__half--on').text()).toBe('settings.calibration.screen.bar.confirming')
    expect(w.findAll('.sc__stat dd').map(d => d.text())).toEqual([
      '+0.012 %, +0.20 tuner.boxes.unit.reference',
      'settings.calibration.screen.steadiness(ppm=3.2,target=5)',
      '48000 Hz', '31 s', '2',
    ])
    // The calibrate key keeps its place, disabled, beside Cancel.
    const keys = w.findAll('button')
    expect(keys.map(k => k.text())).toEqual(['settings.calibration.cancel', 'settings.calibration.start'])
    expect(keys[1]!.attributes('disabled')).toBeDefined()
    await keys[0]!.trigger('click')
    expect(cancelCalibration).toHaveBeenCalled()
  })

  it('has no clock figures while warming up, only the device and the time', async () => {
    const w = await view({ status: 'preparing', running: true, elapsed: 3, progress: 5, rate: 44100 })
    expect(w.text()).toContain('settings.calibration.screen.preparing')
    expect(w.findAll('.sc__stat dd').map(d => d.text())).toEqual(['--', '--', '44100 Hz', '3 s', '0'])
  })

  it('states the stored correction as a pitch error and returns where it came from', async () => {
    const w = await view({ status: 'calibrated', percent: 0.006 })
    expect(w.text()).toContain('settings.calibration.screen.calibrated(percent=+0.006)')
    expect(w.text()).toContain('settings.calibration.screen.pitch(pitch=+0.10 tuner.boxes.unit.reference)')
    expect(w.find('.sc__step--on').text()).toContain('settings.calibration.screen.steps.done')
    expect(w.find('.sc__stat dd').text()).toBe('+0.006 %, +0.10 tuner.boxes.unit.reference')
    const keys = w.findAll('button')
    expect(keys.map(k => k.text())).toEqual(['settings.calibration.screen.return', 'settings.calibration.screen.again'])
    await keys[0]!.trigger('click')
    expect(setView).toHaveBeenCalledWith('settings')
  })

  it('says an unreliable clock stays uncorrected and offers another go', async () => {
    const w = await view({ status: 'unreliable' })
    expect(w.text()).toContain('settings.calibration.screen.unreliable')
    expect(w.text()).toContain('settings.calibration.screen.retry')
    expect(w.findAll('button').map(k => k.text())).toContain('settings.calibration.screen.again')
  })
})
