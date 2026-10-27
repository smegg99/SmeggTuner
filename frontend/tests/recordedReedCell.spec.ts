// tests/recordedReedCell.spec.ts
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import RecordedReedCell from '~/components/record/RecordedReedCell.vue'
import RecordedTableRow from '~/components/record/RecordedTableRow.vue'
import type { ReedCell } from '~/utils/recordedRows'

const cell: ReedCell = {
  key: 'reed-1', reed: 0, slot: 0, present: true, inTol: false,
  curr: '+1.2', raw: 1.2, goal: '0.0', error: '+1.2',
}

describe('recorded reed keyboard editing', () => {
  let wrapper: ReturnType<typeof mount<typeof RecordedReedCell>>
  beforeEach(() => {
    vi.stubGlobal('useI18n', () => ({
      t: (key: string, params?: { index: number }) => key === 'record.table.reed' ? `Reed ${params?.index}` : key,
      n: (value: number) => String(value),
    }))
    wrapper = mount(RecordedReedCell, { props: { cell, noteName: 'A4' }, attachTo: document.body })
  })
  afterEach(() => {
    wrapper.unmount()
    vi.unstubAllGlobals()
  })

  it.each(['Enter', ' '])('opens from %s and commits a comma decimal with focus restored', async (key) => {
    const button = wrapper.get('button')
    expect(button.attributes('type')).toBe('button')
    expect(button.attributes('aria-label')).toBe('A4: Reed 1: +1.2')
    button.element.focus()
    const event = new KeyboardEvent('keydown', { key, bubbles: true, cancelable: true })
    button.element.dispatchEvent(event)
    await nextTick()
    expect(event.defaultPrevented).toBe(true)
    const input = wrapper.get('input')
    expect(input.attributes('aria-label')).toBe('A4: Reed 1')
    expect(document.activeElement).toBe(input.element)
    expect(input.element.value).toBe('1.2')
    await input.setValue('2,5')
    await input.trigger('keydown', { key: 'Enter' })
    await nextTick()
    expect(wrapper.emitted('edit')).toEqual([[2.5]])
    expect(document.activeElement).toBe(wrapper.get('button').element)
  })

  it('cancels with Escape and restores focus without emitting', async () => {
    await wrapper.get('button').trigger('click')
    await wrapper.get('input').setValue('3')
    await wrapper.get('input').trigger('keydown', { key: 'Escape' })
    await nextTick()
    expect(wrapper.emitted('edit')).toBeUndefined()
    expect(document.activeElement).toBe(wrapper.get('button').element)
  })

  it('cancels on blur without stealing focus from the next control', async () => {
    await wrapper.get('button').trigger('click')
    const next = document.createElement('button')
    document.body.append(next)
    next.focus()
    await nextTick()
    expect(wrapper.find('input').exists()).toBe(false)
    expect(document.activeElement).toBe(next)
    expect(wrapper.emitted('edit')).toBeUndefined()
    next.remove()
  })

  it.each(['', 'invalid', 'Infinity'])('never commits invalid input %j', async (value) => {
    await wrapper.get('button').trigger('click')
    await wrapper.get('input').setValue(value)
    await wrapper.get('input').trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('edit')).toBeUndefined()
  })

  it.each([{ present: false }, { slot: -1 }])('keeps uneditable cells out of the tab order: %j', async (change) => {
    await wrapper.setProps({ cell: { ...cell, ...change } })
    expect(wrapper.find('button').exists()).toBe(false)
    expect(wrapper.attributes('tabindex')).toBeUndefined()
    await wrapper.trigger('dblclick')
    expect(wrapper.find('input').exists()).toBe(false)
  })

  it('preserves double-click entry without resetting an active draft', async () => {
    await wrapper.trigger('dblclick')
    await wrapper.get('input').setValue('4.5')
    await wrapper.get('input').trigger('dblclick')
    expect(wrapper.get('input').element.value).toBe('4.5')
  })

  it('passes note context and the edited take/slot through the real table row', async () => {
    const row = mount(RecordedTableRow, {
      props: {
        row: {
          key: 'row', note: 71, noteName: 'H4', take: 3, manual: false, latest: false,
          reedCells: [{ ...cell, reed: 2, slot: 1 }], beatCells: [],
        },
        removable: false, showRegister: false,
      },
      global: {
        components: { RecordRecordedReedCell: RecordedReedCell },
        stubs: { RecordRecordedRowMarks: true, VIcon: true },
      },
    })
    try {
      expect(row.get('button').attributes('aria-label')).toBe('H4: Reed 3: +1.2')
      await row.get('button').trigger('keydown', { key: 'Enter' })
      expect(row.get('input').attributes('aria-label')).toBe('H4: Reed 3')
      await row.get('input').setValue('-0.5')
      await row.get('input').trigger('keydown', { key: 'Enter' })
      expect(row.emitted('edit')).toEqual([[{ take: 3, slot: 1, value: -0.5 }]])
    }
    finally {
      row.unmount()
    }
  })
})
