// tests/reportNotes.spec.ts
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createVuetify } from 'vuetify'
import { VBtn, VTextarea, VTextField } from 'vuetify/components'
import ReportDialog from '~/components/record/ReportDialog.vue'
import { active, busy, error } from '~/composables/sessionStore'
import { useSessions } from '~/composables/useSessions'
import { SessionDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/session/models.js'
import { MAX_NOTES } from '~/utils/accordion'

/*
 * THIS MUST AGREE WITH services/session maxNotes, EXACTLY.
 *
 * The note is typed here and printed on the card by Go. The field caps what can be
 * typed; the service refuses what is too long. If the field's cap were the larger of
 * the two, a technician would write a note, watch it save, and find it missing.
 */

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..')

const backend = vi.hoisted(() => ({ setNotes: vi.fn() }))
vi.mock('~~bindings/smegg.me/smeggtuner/backend/services/session/service.js', () => ({ SetNotes: backend.setNotes }))
vi.mock('@wailsio/runtime', async original => ({
  ...await original<typeof import('@wailsio/runtime')>(),
  Events: { On: vi.fn() },
}))

let wrapper: ReturnType<typeof mount> | undefined

beforeEach(() => {
  vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
  backend.setNotes.mockReset().mockResolvedValue(undefined)
  active.value = new SessionDTO({ notes: 'saved note' })
  busy.value = false
  error.value = ''
})

afterEach(() => {
  wrapper?.unmount()
  vi.unstubAllGlobals()
})

function dialog() {
  wrapper = mount(ReportDialog, {
    props: { modelValue: true },
    global: {
      plugins: [createVuetify({ components: { VBtn, VTextarea, VTextField } })],
      stubs: {
        VDialog: { template: '<div><slot /></div>' },
        UiDialogCard: { template: '<div><slot /><slot name="actions" /></div>' },
        UiFormField: { template: '<div><slot id="field" /></div>' },
        UiToolGroup: true,
      },
    },
  })
  return wrapper
}

function exportButton() {
  return wrapper!.findAllComponents(VBtn).find(button => button.text() === 'record.report.export')!
}

describe('the printed note', () => {
  it('is capped at what the service accepts', () => {
    const go = readFileSync(path.join(root, 'backend/services/session/sessions.go'), 'utf8')
    const declared = /const maxNotes = (\d+)/.exec(go)

    expect(declared, 'services/session no longer declares maxNotes').not.toBeNull()
    expect(MAX_NOTES).toBe(Number(declared![1]))
  })

  it('distinguishes a successful void SetNotes from a failed SetNotes', async () => {
    const sessions = useSessions()
    expect(await sessions.setNotes('saved')).toBe(true)
    backend.setNotes.mockRejectedValueOnce(new Error('session.error.saveFailed'))
    expect(await sessions.setNotes('failed')).toBe(false)
    expect(error.value).toBe('session.error.saveFailed')
  })

  it('keeps the draft and shows a failed save without exporting, then allows retry', async () => {
    const w = dialog()
    await w.get('textarea').setValue('edited note')
    backend.setNotes.mockRejectedValueOnce(new Error('session.error.saveFailed'))
    await exportButton().trigger('click')
    await flushPromises()
    expect(w.emitted('export')).toBeUndefined()
    expect(w.find('.form__error').text()).toBe('session.error.saveFailed')
    expect((w.get('textarea').element as HTMLTextAreaElement).value).toBe('edited note')
    expect(w.emitted('update:modelValue')).toBeUndefined()
    await exportButton().trigger('click')
    await flushPromises()
    expect(backend.setNotes).toHaveBeenLastCalledWith('edited note')
    expect(w.emitted('export')).toHaveLength(1)
  })

  it('waits for note saving and blocks repeated export requests while it is pending', async () => {
    let finish!: () => void
    backend.setNotes.mockReturnValueOnce(new Promise<void>((resolve) => {
      finish = resolve
    }))
    const w = dialog()
    await exportButton().trigger('click')
    await exportButton().trigger('click')
    expect(w.emitted('export')).toBeUndefined()
    expect(backend.setNotes).toHaveBeenCalledTimes(1)
    finish()
    await flushPromises()
    expect(w.emitted('export')).toHaveLength(1)
  })

  it.each(['x', '\u0105', '\u{1f600}'])('accepts exactly 500 code points: %s', async (letter) => {
    const w = dialog()
    const text = letter.repeat(MAX_NOTES)
    expect(w.get('textarea').attributes('maxlength')).toBeUndefined()
    await w.get('textarea').setValue(text)
    await exportButton().trigger('click')
    await flushPromises()
    expect(backend.setNotes).toHaveBeenCalledExactlyOnceWith(text)
    expect(w.emitted('export')).toHaveLength(1)
  })

  it.each(['x', '\u0105', '\u{1f600}'])('refuses 501 code points without truncating the draft: %s', async (letter) => {
    const w = dialog()
    const text = letter.repeat(MAX_NOTES + 1)
    await w.get('textarea').setValue(text)
    await exportButton().trigger('click')
    await flushPromises()
    expect(backend.setNotes).not.toHaveBeenCalled()
    expect(w.emitted('export')).toBeUndefined()
    expect(w.find('.form__error').text()).toBe('session.error.notesTooLong')
    expect((w.get('textarea').element as HTMLTextAreaElement).value).toBe(text)
  })
})
