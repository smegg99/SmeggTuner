import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick, reactive, ref } from 'vue'
import ValueBoxes from '~/components/tuner/ValueBoxes.vue'
import NoteStepper from '~/components/instrument/NoteStepper.vue'
import CalibrationSweepSetup from '~/components/calibration/CalibrationSweepSetup.vue'
import { Config } from '~~bindings/smegg.me/smeggtuner/backend/common/config/models.js'
import { Bank } from '~/types/session'

const config = reactive(new Config())
const active = ref<{ instrument: { tolerance: number, beatTolerance: number }, bench: { reeds: number } } | null>(null)
vi.mock('~/composables/useConfigSync', () => ({ useConfigSync: () => ({ config }) }))
vi.mock('~/composables/useSessions', () => ({ useSessions: () => ({ active }) }))
vi.mock('~/composables/useTuner', () => ({
  NOTE_MIN: 12, NOTE_MAX: 119,
  useTuner: () => ({
    panels: ref(Array.from({ length: 3 }, () => ({ value: 3, target: 0, shown: '3', dimmed: false }))),
    reedErrors: ref([{ reed: 0, error: 3, inTol: true }, { reed: 1, error: 3, inTol: true }]),
    beatErrors: ref([{ low: 0, high: 1, error: 4, inTol: true }]),
    reeds: ref([]), reedCount: ref(2), mode: ref('tones'), reedBoxes: ref([1, 2]), octaveOffsets: ref([0, 0]),
  }),
}))
vi.mock('~/composables/render/useRulerRender', () => ({ useRulerRender: () => ({ draw: vi.fn() }) }))

const stubs = {
  TunerValueBox: { props: ['box'], template: '<span class="box" :data-frac="box.frac" />' },
  UiPanelCanvas: { template: '<section><slot name="meta" /></section>' },
  UiToolKey: true,
  InstrumentRegisterSymbol: true,
  VBtn: { template: '<button><slot /></button>' },
}
const mounted: ReturnType<typeof mount>[] = []

beforeEach(() => {
  config.tuner.tolerance = 1
  config.tuner.beat_tolerance = 3
  config.tuner.deviation_unit = 'reference'
  config.tuner.scale_naming = 'cdefgah'
  active.value = { instrument: { tolerance: 5, beatTolerance: 8 }, bench: { reeds: 2 } }
  vi.stubGlobal('useI18n', () => ({ t: (key: string, params?: { tol?: string }) => params?.tol ?? key }))
})

afterEach(() => {
  mounted.splice(0).forEach(wrapper => wrapper.unmount())
  vi.unstubAllGlobals()
})

describe('effective tuning values in the interface', () => {
  it('scales reed and beat bars with session tolerances', async () => {
    const boxes = mount(ValueBoxes, { global: { stubs } })
    mounted.push(boxes)
    expect(boxes.findAll('.box').map(box => Number(box.attributes('data-frac')))).toEqual([0.575, 0.5625, 0.575])
    active.value = null
    await nextTick()
    expect(Number(boxes.find('.box').attributes('data-frac'))).toBe(0.875)
  })

  it.each([0, -1])('ignores nonpositive session tolerance %s like the backend', (tolerance) => {
    active.value!.instrument = { tolerance, beatTolerance: tolerance }
    const boxes = mount(ValueBoxes, { global: { stubs } })
    mounted.push(boxes)
    expect(Number(boxes.find('.box').attributes('data-frac'))).toBe(0.875)
    expect(Number(boxes.findAll('.box')[1]!.attributes('data-frac'))).toBeCloseTo(0.5 + 4 / 24)
  })

  it('uses the current naming convention in note steppers and reacts to changes', async () => {
    const stepper = mount(NoteStepper, { props: { modelValue: 71 }, global: { stubs } })
    mounted.push(stepper)
    expect(stepper.find('output').text()).toBe('H4')
    config.tuner.scale_naming = 'cdefgab'
    await nextTick()
    expect(stepper.find('output').text()).toBe('B4')
  })
})

describe('calibration register choices', () => {
  it.each([
    { registers: [{ name: 'MM', banks: [Bank.BankM1, Bank.BankM2] }], bassReeds: 0 },
    { registers: [], bassReeds: 3 },
  ])('does not offer an unnamed treble sweep when described choices exist: %j', (props) => {
    const setup = mount(CalibrationSweepSetup, { props, global: { stubs } })
    mounted.push(setup)
    expect(setup.text()).not.toContain('calibrate.sweep.noRegisters')
    expect(setup.text()).not.toContain('calibrate.sweep.startPlain')
    expect(setup.findAll('.sw__reg').length).toBeGreaterThan(0)
  })

  it('offers the generic sweep when there are no register choices', async () => {
    const setup = mount(CalibrationSweepSetup, { props: { registers: [] }, global: { stubs } })
    mounted.push(setup)
    expect(setup.text()).toContain('calibrate.sweep.noRegisters')
    await setup.find('button').trigger('click')
    expect(setup.emitted('begin')).toEqual([['']])
  })

  it('keeps the solo calibration explanation for a single-rank register', () => {
    const setup = mount(CalibrationSweepSetup, { props: { registers: [{ name: 'M', banks: [Bank.BankM1] }] }, global: { stubs } })
    mounted.push(setup)
    expect(setup.text()).toContain('calibrate.sweep.soloTeaches')
    expect(setup.text()).not.toContain('calibrate.sweep.noRegisters')
  })
})
