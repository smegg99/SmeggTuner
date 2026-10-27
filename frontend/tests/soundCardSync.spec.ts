import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { reactive } from 'vue'
import { AutoVolumeDTO, CalibrationDTO, CalibrationStatus } from '~~bindings/smegg.me/smeggtuner/backend/services/audio/models.js'
import FormField from '~/components/ui/FormField.vue'

const failedKey = 'settings.calibration.saveFailed'
const backend = vi.hoisted(() => ({
  state: vi.fn(), volume: vi.fn(), start: vi.fn(), cancel: vi.fn(),
  events: new Map<string, (event: { data: CalibrationDTO }) => void>(),
}))
let current: { value: { kind: string, deviceId: string } }
let wrapper: ReturnType<typeof mount> | undefined

vi.mock('~/composables/useAudioDevices', () => ({ useAudioDevices: () => ({ current }) }))
vi.mock('~/composables/useConfigSync', () => ({ useConfigSync: () => ({ config: { audio: { device_id: 'a' } } }) }))
vi.mock('~/composables/useTuner', () => ({ toErrorKey: () => 'tuner.error.deviceGone' }))
vi.mock('~~bindings/smegg.me/smeggtuner/backend/services/audio/service.js', () => ({
  CalibrationState: backend.state, AutoVolume: backend.volume,
  StartCalibration: backend.start, CancelCalibration: backend.cancel,
  Correction: async () => ({ missing: true, percent: 0 }),
}))
vi.mock('@wailsio/runtime', async original => ({
  ...await original<typeof import('@wailsio/runtime')>(),
  Events: { On: (name: string, callback: (event: { data: CalibrationDTO }) => void) => backend.events.set(name, callback) },
}))

function state(overrides: Partial<CalibrationDTO> = {}) {
  return new CalibrationDTO({ deviceId: 'a', status: CalibrationStatus.CalibrationUncalibrated, ...overrides })
}

function pending() {
  let resolve!: (value: CalibrationDTO) => void
  const promise = new Promise<CalibrationDTO>((done) => {
    resolve = done
  })
  return { promise, resolve }
}

function emit(value: CalibrationDTO) {
  backend.events.get('audio:calibration')!({ data: value })
}

async function startView() {
  const { useSoundCard } = await import('~/composables/useSoundCard')
  const soundCard = useSoundCard()
  const { default: SoundCardCalibration } = await import('~/components/settings/audio/SoundCardCalibration.vue')
  wrapper = mount(SoundCardCalibration, {
    global: { components: { UiFormField: FormField }, stubs: { UiToolKey: true, VTooltip: true } },
  })
  await flushPromises()
  return soundCard
}

beforeEach(() => {
  vi.resetModules()
  backend.events.clear()
  current = reactive({ value: { kind: 'mic', deviceId: 'a' } })
  backend.state.mockReset().mockImplementation(async (id: string) => state({ deviceId: id || 'a' }))
  backend.volume.mockReset().mockImplementation(async (id: string) => new AutoVolumeDTO({ deviceId: id || 'a' }))
  backend.start.mockReset().mockResolvedValue(undefined)
  backend.cancel.mockReset().mockResolvedValue(undefined)
  vi.stubGlobal('useLogger', () => ({ error: vi.fn() }))
  vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
})

afterEach(() => {
  wrapper?.unmount()
  wrapper = undefined
  vi.unstubAllGlobals()
})

describe('sound-card calibration failure display', () => {
  it('renders a final persistence error and clears it when a new run starts successfully', async () => {
    const card = await startView()
    emit(state({ error: failedKey }))
    await flushPromises()
    expect(card.error.value).toBe(failedKey)
    expect(wrapper!.get('.field__error').text()).toBe(failedKey)
    expect(card.calibration.value?.running).toBe(false)

    backend.start.mockImplementationOnce(async () => {
      emit(state({ status: CalibrationStatus.CalibrationPreparing, running: true }))
    })
    backend.state.mockResolvedValueOnce(state({ status: CalibrationStatus.CalibrationPreparing, running: true }))
    await card.startCalibration()
    await flushPromises()
    expect(card.error.value).toBe('')
    expect(wrapper!.find('.field__error').exists()).toBe(false)
  })

  it('renders a remembered failure from the initial state query', async () => {
    backend.state.mockResolvedValueOnce(state({ error: failedKey }))
    const card = await startView()
    expect(card.error.value).toBe(failedKey)
    expect(wrapper!.get('.field__error').text()).toBe(failedKey)
  })

  it.each(['a', ''])('keeps a newer failure while the initial read resolves selected device %j', async (deviceId) => {
    current.value.deviceId = deviceId
    const read = pending()
    backend.state.mockReturnValueOnce(read.promise)
    const card = await startView()
    emit(state({ error: failedKey }))
    // An unrelated device must not hide the relevant event in the pending-read buffer.
    emit(state({ deviceId: 'other', status: CalibrationStatus.CalibrationCalibrated }))
    read.resolve(state({ status: CalibrationStatus.CalibrationPreparing, running: true }))
    await flushPromises()
    expect(card.calibration.value).toMatchObject({ deviceId: 'a', running: false, error: failedKey })
    expect(wrapper!.get('.field__error').text()).toBe(failedKey)
  })

  it('keeps the final failure over an older post-start refresh', async () => {
    const card = await startView()
    const read = pending()
    backend.state.mockReturnValueOnce(read.promise)
    const starting = card.startCalibration()
    await flushPromises()
    emit(state({ error: failedKey }))
    read.resolve(state({ status: CalibrationStatus.CalibrationCalibrating, running: true }))
    await starting
    await flushPromises()
    expect(card.calibration.value?.running).toBe(false)
    expect(wrapper!.get('.field__error').text()).toBe(failedKey)
  })

  it('ignores a superseded device read and errors emitted for another device', async () => {
    const first = pending()
    backend.state.mockReturnValueOnce(first.promise)
    const card = await startView()
    current.value.deviceId = 'b'
    await flushPromises()
    emit(state({ deviceId: 'b', error: failedKey }))
    emit(state({ deviceId: 'a' }))
    first.resolve(state())
    await flushPromises()
    expect(card.calibration.value?.deviceId).toBe('b')
    expect(wrapper!.get('.field__error').text()).toBe(failedKey)

    current.value.deviceId = 'c'
    await flushPromises()
    expect(card.calibration.value?.deviceId).toBe('c')
    expect(card.error.value).toBe('')
  })

  it.each([
    [failedKey, failedKey],
    ['device gone', 'tuner.error.deviceGone'],
  ])('preserves a synchronous command failure after a healthy state refresh: %s', async (failure, expected) => {
    const card = await startView()
    backend.start.mockRejectedValueOnce(new Error(failure))
    await card.startCalibration()
    await flushPromises()
    expect(card.error.value).toBe(expected)
    expect(wrapper!.get('.field__error').text()).toBe(expected)
  })

  it('traces a run\'s measured offsets, cleared by the next run and by another device', async () => {
    const card = await startView()
    const run = (status: CalibrationStatus, elapsed: number, percent = 0) =>
      emit(state({ status, running: true, elapsed, percent }))
    run(CalibrationStatus.CalibrationPreparing, 2)
    run(CalibrationStatus.CalibrationCalibrating, 8, 0.01)
    run(CalibrationStatus.CalibrationCalibrating, 9, 0.012)
    emit(state({ status: CalibrationStatus.CalibrationCalibrated, percent: 0.011 }))
    expect(card.trace.value).toEqual([{ at: 8, percent: 0.01 }, { at: 9, percent: 0.012 }])

    run(CalibrationStatus.CalibrationPreparing, 1)
    expect(card.trace.value).toEqual([])
    run(CalibrationStatus.CalibrationCalibrating, 8, 0.02)
    current.value.deviceId = 'b'
    await flushPromises()
    expect(card.trace.value).toEqual([])
  })
})
