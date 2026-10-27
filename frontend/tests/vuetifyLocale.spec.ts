// tests/vuetifyLocale.spec.ts
import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h, nextTick, reactive, ref } from 'vue'
import type { App } from 'vue'
import { useLocale } from 'vuetify'
import { en, pl } from 'vuetify/locale'
import { initializeThemeSync } from '~/composables/useThemeSync'

vi.mock('@smegg99/s99ab/appearance', () => ({
  registerAppearanceSource: vi.fn(),
  resetAppearanceSource: vi.fn(),
  useAppearance: () => ({ setMode: vi.fn() }),
}))
vi.mock('~/composables/s99wails', () => ({
  useSystemTheme: () => ({ theme: ref(null) }),
  resolveDesktopTheme: () => ({ accentFor: () => '#2563eb' }),
}))
vi.mock('~~bindings/github.com/smegg99/s99wails/v2/services/theme/service.js', () => ({
  GetTheme: vi.fn(),
}))

afterEach(() => vi.unstubAllGlobals())

describe('Vuetify locale integration', () => {
  it('loads Polish widget messages and follows both persisted and active language changes', async () => {
    vi.stubGlobal('defineNuxtPlugin', (plugin: unknown) => plugin)
    const { default: plugin } = await import('~/plugins/vuetify')
    const locale = ref('en')
    const config = reactive({ preferences: { language: 'pl', theme: 'auto', accent_mode: 'auto', accent_color: '' } })
    vi.stubGlobal('useConfigSync', () => ({ config }))
    vi.stubGlobal('useI18n', () => ({
      locale,
      setLocale: async (code: string) => { locale.value = code },
    }))
    vi.stubGlobal('useLogger', () => ({ error: vi.fn() }))
    const probe = defineComponent({
      setup() {
        initializeThemeSync()
        const widgetLocale = useLocale()
        return () => h('span', [
          widgetLocale.t('$vuetify.open'),
          widgetLocale.t('$vuetify.noDataText'),
          widgetLocale.t('$vuetify.pagination.ariaLabel.next'),
        ].join('|'))
      },
    })
    const installer = {
      install(vueApp: App) {
        const setup = plugin as unknown as { setup: (app: { vueApp: App }) => void }
        setup.setup({ vueApp })
      },
    }
    const wrapper = mount(probe, { global: { plugins: [installer] } })
    try {
      await flushPromises()
      expect(wrapper.text()).toBe([pl.open, pl.noDataText, pl.pagination.ariaLabel.next].join('|'))
      config.preferences.language = 'en'
      await flushPromises()
      expect(wrapper.text()).toBe([en.open, en.noDataText, en.pagination.ariaLabel.next].join('|'))
      locale.value = 'pl'
      await nextTick()
      expect(wrapper.text()).toBe([pl.open, pl.noDataText, pl.pagination.ariaLabel.next].join('|'))
    }
    finally {
      wrapper.unmount()
    }
  })
})
