import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick, ref, shallowRef } from 'vue'
// The real bus: nothing but a counter and a last-request, with no backend behind it, so the
// test drives the same singleton the component watches rather than a lookalike.
import { useWorkshop } from '~/composables/useWorkshop'
import SessionDetail from '~/components/session/SessionDetail.vue'

/*
 * THE PRINT KEY HAS TO REACH THE REPORT.
 *
 * It did not. The workshop's toolbar raised `print` through useWorkshop and the switch that
 * answers those requests handled new / import / export / edit and simply had no case for it,
 * so the key sat there looking enabled and did nothing at all. A dead control is worse than
 * a missing one: the technician concludes the export is broken, not that it was never wired.
 *
 * The request is answered HERE and not in the workshop because a card is printed of the
 * recording on screen, and this is the component that has one. So what this pins is the whole
 * chain - key -> bus -> dialog -> backend - and that the pass it exports is the pass being
 * looked at rather than whatever the backend would have picked by default.
 */

vi.stubGlobal('useI18n', () => ({
  t: (key: string) => key,
  d: () => 'when',
}))

const exportSession = vi.fn(async () => true)

vi.mock('~/composables/useReport', () => ({
  useReport: () => ({
    busy: ref(false),
    error: ref(''),
    exportSession,
    clearError: vi.fn(),
  }),
}))

vi.mock('~/composables/useRecord', () => ({
  useRecord: () => ({
    table: shallowRef({ rows: [{ note: 69 }], reedCount: 3, tolerance: 1, banks: [] }),
    show: vi.fn(),
    editReed: vi.fn(),
    fetch: vi.fn(),
  }),
}))

vi.mock('~/composables/useSessions', () => ({
  useSessions: () => ({
    active: shallowRef({
      id: 's1',
      name: 'box',
      a4: 442,
      curve: null,
      instrument: { name: 'Hohner Morino', banks: ['M1', 'M2'], reedCount: 2 },
      readings: 4,
    }),
    busy: ref(false),
  }),
}))

vi.mock('~/composables/useConfigSync', () => ({
  useConfigSync: () => ({ config: { tuner: { scale_naming: 'cdefgab' } } }),
}))

const global = {
  stubs: {
    'UiToolGroup': {
      name: 'UiToolGroup',
      props: ['modelValue', 'items'],
      template: '<div class="views" />',
    },
    /*
     * The name the APP renders, which is not the name the file has. Nuxt auto-imports
     * components/record/RecordedTable.vue as RecordRecordedTable, and this stub used to be
     * registered as `RecordedTable` - the same typo the component had.
     *
     * So the stub answered to the mistake and the suite stayed green while the real session
     * screen rendered a blank space: Vue resolved the unknown tag to a native element, which
     * draws nothing and reports nothing. A stub registered under a name the app does not use
     * tests the stub.
     */
    'RecordRecordedTable': { name: 'RecordRecordedTable', props: ['removable'], template: '<div class="recorded-table" />' },
    'CurveEditor': { name: 'CurveEditor', template: '<div class="curve-editor" />' },
    'RecordReportDialog': {
      name: 'RecordReportDialog',
      props: ['modelValue', 'passLabel', 'busy', 'error'],
      template: '<div class="report-dialog" />',
    },
    'v-icon': true,
  },
}

function detail() {
  return mount(SessionDetail, { props: { id: 's1', src: '' }, global })
}

describe('the print key', () => {
  it('opens the report when the toolbar asks, and not before', async () => {
    const w = detail()
    expect(w.findComponent({ name: 'RecordReportDialog' }).props('modelValue')).toBe(false)

    useWorkshop().ask('print')
    await nextTick()

    expect(w.findComponent({ name: 'RecordReportDialog' }).props('modelValue')).toBe(true)
  })

  it('ignores requests that are not its own', async () => {
    const w = detail()

    useWorkshop().ask('import')
    await nextTick()

    expect(w.findComponent({ name: 'RecordReportDialog' }).props('modelValue')).toBe(false)
  })

  it('exports the pass on screen', async () => {
    const w = detail()
    useWorkshop().ask('print')
    await nextTick()

    const options = { format: 'pdf', date: '2026-07-16' }
    await w.findComponent({ name: 'RecordReportDialog' }).vm.$emit('export', options)
    await nextTick()

    /*
     * The session's own readings are what a card is printed of, and the FORMAT reaches the
     * backend untouched. All three - pdf, html, csv - are one export: the PDF is the HTML
     * sheet laid out by a headless browser on the Go side, so nothing about it is decided
     * here, and a frontend that special-cased it would be a second place to get it wrong.
     */
    expect(exportSession).toHaveBeenCalledWith(options)
  })

  it('closes on a written file and stays open on a cancel', async () => {
    const w = detail()
    const dialog = () => w.findComponent({ name: 'RecordReportDialog' })

    useWorkshop().ask('print')
    await nextTick()
    await dialog().vm.$emit('export', { format: 'html' })
    await nextTick()
    expect(dialog().props('modelValue')).toBe(false)

    // A cancelled save dialog comes back false. He is mid-sentence: the form stays up.
    exportSession.mockResolvedValueOnce(false)
    useWorkshop().ask('print')
    await nextTick()
    await dialog().vm.$emit('export', { format: 'html' })
    await nextTick()
    expect(dialog().props('modelValue')).toBe(true)
  })
})

/*
 * A SESSION OPENS ON WHAT IT MEASURED, and swaps to what it should be.
 *
 * The desired-beating curve was briefly out of this screen and is back, so the swap is a
 * behaviour again. It opens on the readings either way: a technician who has just recorded is
 * looking at what he caught, not at the goal he caught it against.
 */
describe('the session views', () => {
  it('starts on the recording and swaps to the curve', async () => {
    const w = detail()
    expect(w.find('.read__recorded').exists()).toBe(true)
    expect(w.find('.curve-editor').exists()).toBe(false)

    await w.findComponent({ name: 'UiToolGroup' }).vm.$emit('update:modelValue', 'curve')
    await flushPromises()
    expect(w.find('.curve-editor').exists()).toBe(true)
    expect(w.find('.read__recorded').exists()).toBe(false)
  })
})

/*
 * READINGS STAY REMOVABLE. A number typed over a measurement is a different claim about the
 * same reed, and the honest way to change what a session says is to play the note again - so
 * what the table offers is REMOVAL, and it offers it on every reading there is.
 */
describe('the recorded table', () => {
  it('lets a reading be taken out', () => {
    const w = detail()
    expect(w.findComponent({ name: 'RecordRecordedTable' }).props('removable')).toBe(true)
  })
})
