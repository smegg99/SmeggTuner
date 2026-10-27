// tests/instrumentReference.spec.ts
import { describe, expect, it } from 'vitest'
import { effectScope, ref } from 'vue'
import { Instrument, Template } from '~~bindings/smegg.me/smeggtuner/backend/core/session/models.js'
import { useInstrumentForm } from '~/components/instrument/useInstrumentForm'

describe('instrument reference pitch drafts', () => {
  it('leaves a recorded fractional reference unchanged when navigation commits drafts', () => {
    const scope = effectScope()
    try {
      scope.run(() => {
        const source = ref(new Template({
          id: 'instrument', name: 'Accordion', instrument: new Instrument({ a4: 440.5 }),
        }))
        const editor = useInstrumentForm(source)
        editor.commitA4()
        expect(editor.form.a4).toBe(440.5)
        expect(editor.a4Draft.value).toBe('440.5')
        expect(editor.collect()?.instrument.a4).toBe(440.5)
      })
    }
    finally {
      scope.stop()
    }
  })

  it.each([
    ['441.2', 441.2], ['429.9', 430], ['450.1', 450], ['invalid', 440],
  ])('commits %s as %s without integer rounding', (draft, expected) => {
    const scope = effectScope()
    try {
      scope.run(() => {
        const editor = useInstrumentForm(ref(null))
        editor.a4Draft.value = draft
        editor.commitA4()
        expect(editor.form.a4).toBe(expected)
      })
    }
    finally {
      scope.stop()
    }
  })
})
