// tests/updateStartup.spec.ts
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { nextTick, onBeforeUnmount, reactive, ref, watchEffect } from 'vue'
import { useUpdatePrompts } from '@s99ab/composables/releases/useUpdatePrompts'
import { closeAllDialogs, dialogEntries, resetDialogs } from '@s99ab/internal/dialog'
import AppRoot from '~/components/shell/AppRoot.vue'

const config = reactive({ preferences: { language: 'en' } })
const ready = ref(false)
const locale = ref('en')
const platform = ref({ android: false })
let native = true
let wrapper: VueWrapper | undefined
const unsubscribe = vi.fn()
const source = {
  state: vi.fn(async (lang: string) => ({
    current: '1.4.3', notes: { version: '1.4.3', text: lang + ' notes', lang },
  })),
  notesShown: vi.fn(async () => {}),
  ignore: vi.fn(async () => {}),
  download: vi.fn(async () => {}),
  onAvailable: vi.fn(() => unsubscribe),
}
const startPrompts = vi.fn(useUpdatePrompts)

vi.mock('~/composables/useConfigSync', () => ({ useConfigSync: () => ({ config, ready }) }))
vi.mock('~/composables/useThemeSync', () => ({ initializeThemeSync: vi.fn() }))
vi.mock('~/composables/s99wails', () => ({
  isWails: () => native,
  updateActions: {
    state: (lang: string) => source.state(lang),
    notesShown: () => source.notesShown(),
    ignore: () => source.ignore(),
    download: () => source.download(),
    onAvailable: () => source.onAvailable(),
  },
  installCrashHeartbeat: () => () => {},
  installFrontendLogBridge: () => () => {},
}))
vi.mock('~~bindings/github.com/smegg99/s99wails/v2/services/logger/service.js', () => ({}))
vi.mock('vuetify', () => ({ useDisplay: () => ({ platform }) }))
vi.mock('@s99ab/composables/useTranslate', () => ({
  useLocale: () => locale,
  useTranslate: () => (_key: string, fallback: string) => fallback,
}))

function render() {
  wrapper = mount(AppRoot, {
    global: { stubs: { VApp: { template: '<div><slot /></div>' }, ShellAppWindow: true, ShellBootScreen: true, UiDialogHost: true } },
  })
  return wrapper
}

beforeEach(() => {
  ready.value = false
  config.preferences.language = 'en'
  locale.value = 'en'
  native = true
  platform.value.android = false
  vi.clearAllMocks()
  Object.defineProperty(document, 'visibilityState', { configurable: true, value: 'visible' })
  vi.stubGlobal('ref', ref)
  vi.stubGlobal('watchEffect', watchEffect)
  vi.stubGlobal('onBeforeUnmount', onBeforeUnmount)
  vi.stubGlobal('useI18n', () => ({ locale }))
  vi.stubGlobal('useHead', vi.fn())
  vi.stubGlobal('useLogger', () => ({ debug: vi.fn() }))
  vi.stubGlobal('useUpdatePrompts', startPrompts)
})

afterEach(async () => {
  wrapper?.unmount()
  await flushPromises()
  closeAllDialogs()
  resetDialogs()
  vi.unstubAllGlobals()
})

describe('update prompts during app startup', () => {
  it('waits for config even when the initial language already matches', async () => {
    render()
    await flushPromises()
    expect(source.state).not.toHaveBeenCalled()
    ready.value = true
    await flushPromises()
    expect(source.state).toHaveBeenCalledExactlyOnceWith('en')
  })

  it('waits for the saved Polish locale before opening release notes', async () => {
    render()
    await flushPromises()
    config.preferences.language = 'pl'
    ready.value = true
    await flushPromises()
    expect(source.state).not.toHaveBeenCalled()
    expect(dialogEntries).toHaveLength(0)
    locale.value = 'pl'
    await flushPromises()
    expect(source.state).toHaveBeenCalledExactlyOnceWith('pl')
    expect(dialogEntries[0]?.props).toMatchObject({ notes: 'pl notes', notesLocale: 'pl' })
  })

  it('does not remount prompts when the language changes later', async () => {
    ready.value = true
    render()
    await flushPromises()
    config.preferences.language = 'pl'
    await nextTick()
    locale.value = 'pl'
    await flushPromises()
    expect(startPrompts).toHaveBeenCalledOnce()
    expect(source.onAvailable).toHaveBeenCalledOnce()
    expect(unsubscribe).not.toHaveBeenCalled()
  })

  it.each(['browser', 'android'])('does not start prompts on %s', async (target) => {
    ready.value = true
    native = target !== 'browser'
    platform.value.android = target === 'android'
    render()
    await flushPromises()
    expect(source.state).not.toHaveBeenCalled()
    expect(source.onAvailable).not.toHaveBeenCalled()
  })

  it('disposes the prompt subscription with the root', async () => {
    ready.value = true
    const view = render()
    await flushPromises()
    view.unmount()
    wrapper = undefined
    await flushPromises()
    expect(unsubscribe).toHaveBeenCalledOnce()
    expect(dialogEntries.filter(entry => entry.open)).toHaveLength(0)
  })
})
