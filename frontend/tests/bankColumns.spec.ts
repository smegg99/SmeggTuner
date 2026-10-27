import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import type { RowDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/record/models.js'
import RecordedTable from '~/components/record/RecordedTable.vue'
import RecordedTableHead from '~/components/record/RecordedTableHead.vue'
import RecordedTableRow from '~/components/record/RecordedTableRow.vue'
import RecordedRowMarks from '~/components/record/RecordedRowMarks.vue'
import type { Bank } from '~/types/session'

/*
 * THE COLUMNS ARE NAMED ONLY WHEN THEY CAN HONESTLY BE NAMED.
 *
 * A card whose columns say L | M1 | M2 is making a claim about which rank of reeds each
 * number came out of, and a number under the wrong rank looks exactly like a number under
 * the right one - the technician files a real reed to a pitch nobody asked for.
 *
 * So: named when the instrument has said what its ranks are and the reading fits them, and
 * NUMBERED - which says less and claims nothing - the moment anything does not line up.
 */

// The table reads the display unit off the config; the default is the reference unit.
vi.mock('~/composables/useConfigSync', () => ({
  useConfigSync: () => ({ config: { tuner: {} } }),
}))

vi.stubGlobal('useI18n', () => ({
  t: (key: string, args?: Record<string, unknown>) =>
    args?.index !== undefined ? `Reed ${args.index}` : key,
  // useRecordFormat writes every number through the locale's own formatter, so the stub owes
  // it one. What is under test here is the COLUMN HEADINGS, not the digits under them.
  n: (value: number, opts?: Intl.NumberFormatOptions) =>
    value.toFixed(opts?.minimumFractionDigits ?? 1),
}))

// The table was split into head / row / marks components, auto-imported in the app
// and registered here by their auto-import names, the same way the ValueBox test
// wires in TunerDeviationBar. The Vuetify chrome is stubbed; the column-naming and
// two-rows behaviour under test lives in the real head and row.
const global = {
  mocks: { t: (key: string) => key },
  components: {
    RecordRecordedTableHead: RecordedTableHead,
    RecordRecordedTableRow: RecordedTableRow,
    RecordRecordedRowMarks: RecordedRowMarks,
  },
  stubs: {
    'v-icon': true,
    'v-tooltip': true,
    'UiToolGroup': true,
  },
}

function row(note: number, register: string | undefined, devs: number[]): RowDTO {
  return {
    note,
    noteName: `n${note}`,
    register,
    take: 0,
    at: '',
    slots: devs.length,
    reeds: devs.map((c, i) => ({
      reed: i,
      curr: c,
      goal: 0,
      error: c,
      currHz: c,
      goalHz: 0,
      errorHz: c,
      inTol: true,
    })),
    beats: [],
    manual: false,
  }
}

/*
 * The reed columns come first, then the beat columns - they share a class, and only the reed
 * ones are named after a rank. The note column is not one of them and carries its own class.
 *
 * The hook was `.group-head` while the table had two header rows: rank groups across the top,
 * Note / Curr / Goal / Error under each. With the desired-beating curve out of the table each
 * group holds one number, so the second row went and the group heads became the column heads.
 * The assertions below are untouched by that - what a column is CALLED is the whole of what
 * they test, and it is the same question either way.
 */
const heads = (wrapper: ReturnType<typeof mount>, reeds: number) =>
  wrapper.findAll('.head--num').slice(0, reeds).map(th => th.text())

describe('the tuning table columns', () => {
  it('calls a column what is written on the switch', () => {
    const banks: Bank[] = ['L', 'M1', 'M2']
    const wrapper = mount(RecordedTable, {
      props: { rows: [row(60, 'LMM', [-4, 0, 4])], reedCount: 3, banks },
      global,
    })
    expect(heads(wrapper, 3)).toEqual(['L', 'M1', 'M2'])
  })

  // An instrument nobody has described the ranks of gets the numbered columns this table
  // always had. Less information, none of it wrong.
  it('numbers the columns of an instrument that has not said what its ranks are', () => {
    const wrapper = mount(RecordedTable, {
      props: { rows: [row(60, undefined, [-4, 0, 4])], reedCount: 3 },
      global,
    })
    expect(heads(wrapper, 3)).toEqual(['Reed 1', 'Reed 2', 'Reed 3'])
  })

  /*
   * A bank list that does not line up with the reeds being shown cannot name anything: which
   * of the three columns is the missing one? There is no honest answer, so it does not
   * invent one.
   */
  it('refuses to name columns when the ranks and the reeds disagree', () => {
    const banks: Bank[] = ['M1', 'M2']
    const wrapper = mount(RecordedTable, {
      props: { rows: [row(60, 'MMM', [-4, 0, 4])], reedCount: 3, banks },
      global,
    })
    expect(heads(wrapper, 3)).toEqual(['Reed 1', 'Reed 2', 'Reed 3'])
  })

  /*
   * ONE NOTE ON TWO REGISTERS IS TWO ROWS, and they must both survive to the DOM: a row keyed
   * on the note alone would collide them, and Vue would render one and drop the other -
   * silently, and looking perfectly normal.
   */
  it('draws a note played on two registers as two rows', () => {
    const banks: Bank[] = ['M1', 'M2', 'M3']
    const wrapper = mount(RecordedTable, {
      props: {
        rows: [row(60, 'MMM', [-4, 0, 4]), row(60, 'M2', [7])],
        reedCount: 3,
        banks,
      },
      global,
    })
    expect(wrapper.findAll('tbody tr.recorded-row')).toHaveLength(2)

    // And each says which switch it came off, because that is the only thing telling them
    // apart.
    const text = wrapper.find('tbody').text()
    expect(text).toContain('MMM')
    expect(text).toContain('M2')
  })
})
