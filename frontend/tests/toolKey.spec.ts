// tests/toolKey.spec.ts
import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'
import { createVuetify } from 'vuetify'
import { VBtn } from 'vuetify/components'
import ToolbarSource from '~/components/shell/ToolbarSource.vue'

const source = ref({ kind: 'mic', deviceId: 'mic-1', name: 'A long sound card name' })
vi.mock('~/composables/useAudioDevices', () => ({
  useAudioDevices: () => ({
    current: source, devices: ref([]), loading: ref(false),
    refresh: vi.fn(), selectMic: vi.fn(), pickFile: vi.fn(),
  }),
}))
vi.mock('~/composables/useTuner', () => ({ useTuner: () => ({ running: ref(false) }) }))
vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))

describe('source key width', () => {
  it('keeps a fixed width and a complete accessible name when the source changes', async () => {
    const w = mount(ToolbarSource, {
      global: {
        plugins: [createVuetify({ components: { VBtn } })],
        stubs: {
          VMenu: { template: '<div><slot name="activator" :props="{}" /></div>' },
          VList: true, VListItem: true, VListSubheader: true,
        },
      },
    })
    const button = w.get('button')
    expect(button.attributes('title')).toBe(source.value.name)
    expect(button.attributes('style')).toContain('width: 125px')
    source.value = { kind: 'file', deviceId: '', name: 'A much longer recording filename.wav' }
    await w.vm.$nextTick()
    expect(button.attributes('title')).toBe(source.value.name)
    expect(button.attributes('style')).toContain('width: 125px')
  })
})
