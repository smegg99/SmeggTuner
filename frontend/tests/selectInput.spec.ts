// tests/selectInput.spec.ts
import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { reactive, ref } from 'vue'
import { createVuetify } from 'vuetify'
import { VSelect } from 'vuetify/components'
import LanguageSelect from '~/components/settings/appearance/LanguageSelect.vue'

vi.mock('~/composables/useThemeSync', () => ({
  useThemeSync: () => ({
    prefs: ref(reactive({ language: 'pl' })),
    localeItems: ref([{ code: 'en', name: 'English' }, { code: 'pl', name: 'Polski' }]),
    setLanguage: vi.fn(),
  }),
}))
vi.stubGlobal('useI18n', () => ({ t: (key: string) => key, locale: ref('pl') }))

describe('language selection', () => {
  it('renders the label belonging to the stored locale', () => {
    const w = mount(LanguageSelect, {
      global: {
        plugins: [createVuetify({ components: { VSelect } })],
        stubs: { UiFormField: { template: '<div><slot /></div>' } },
      },
    })
    expect(w.text()).toContain('Polski')
    expect(w.findComponent(VSelect).props('items')).toEqual([
      { value: 'en', label: 'English' }, { value: 'pl', label: 'Polski' },
    ])
  })
})
