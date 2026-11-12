// tests/updateToggle.spec.ts
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { defineComponent, reactive, ref } from 'vue'
import UpdateToggle from '~/components/settings/appearance/UpdateToggle.vue'

const config = reactive({ preferences: { check_updates: undefined as boolean | undefined } })
const platform = ref({ android: false })
const save = vi.fn<() => Promise<void>>()
const check = vi.fn<() => Promise<void>>()

vi.mock('~/composables/useConfigSync', () => ({ useConfigSync: () => ({ config, saveConfigNow: save }) }))
vi.mock('~/composables/s99wails', () => ({ updateActions: { check: () => check() } }))
vi.mock('vuetify', () => ({ useDisplay: () => ({ platform }) }))

const Checkbox = defineComponent({
  props: { modelValue: Boolean, label: String, hint: String },
  emits: ['update:modelValue'],
  template: '<input type="checkbox" :checked="modelValue" @change="$emit(\'update:modelValue\', $event.target.checked)">',
})
let wrapper: VueWrapper | undefined

function render() {
  wrapper = mount(UpdateToggle, { global: { stubs: { VCheckbox: Checkbox } } })
  return wrapper
}

beforeEach(() => {
  config.preferences.check_updates = undefined
  platform.value.android = false
  save.mockReset().mockResolvedValue()
  check.mockReset().mockResolvedValue()
  vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
})

afterEach(() => {
  wrapper?.unmount()
  vi.unstubAllGlobals()
})

describe('update preference', () => {
  it('defaults to enabled and uses localized label and hint keys', () => {
    const checkbox = render().findComponent(Checkbox)
    expect(checkbox.props()).toMatchObject({
      modelValue: true,
      label: 'settings.checkUpdates',
      hint: 'settings.checkUpdatesHint',
    })
  })

  it('persists disabling immediately without checking', async () => {
    await render().get('input').setValue(false)
    await flushPromises()
    expect(config.preferences.check_updates).toBe(false)
    expect(save).toHaveBeenCalledOnce()
    expect(check).not.toHaveBeenCalled()
  })

  it('waits for persistence before checking when enabled', async () => {
    config.preferences.check_updates = false
    let saved!: () => void
    save.mockReturnValueOnce(new Promise<void>((resolve) => {
      saved = resolve
    }))
    await render().get('input').setValue(true)
    expect(config.preferences.check_updates).toBe(true)
    expect(save).toHaveBeenCalledOnce()
    expect(check).not.toHaveBeenCalled()
    saved()
    await flushPromises()
    expect(check).toHaveBeenCalledOnce()
  })

  it('does not check if disabled while the save is pending', async () => {
    config.preferences.check_updates = false
    let saved!: () => void
    save.mockReturnValueOnce(new Promise<void>((resolve) => {
      saved = resolve
    }))
    const view = render()
    await view.get('input').setValue(true)
    await view.get('input').setValue(false)
    saved()
    await flushPromises()
    expect(check).not.toHaveBeenCalled()
  })

  it('keeps a failed check quiet', async () => {
    config.preferences.check_updates = false
    check.mockRejectedValueOnce(new Error('offline'))
    await render().get('input').setValue(true)
    await flushPromises()
    expect(check).toHaveBeenCalledOnce()
    expect(config.preferences.check_updates).toBe(true)
  })

  it('is absent on Android', () => {
    platform.value.android = true
    expect(render().find('input').exists()).toBe(false)
    expect(check).not.toHaveBeenCalled()
  })
})
