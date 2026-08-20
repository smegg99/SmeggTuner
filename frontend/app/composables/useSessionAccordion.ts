import { computed } from 'vue'
import * as SessionService from '~~bindings/smegg.me/smeggtuner/services/session/service.js'
import { useInstruments } from '~/composables/useInstruments'
import { useSessions } from '~/composables/useSessions'
import type { InstrumentTemplate } from '~/types/session'

// The accordion the open session is working on, and the three things you can do to it that are
// not editing a field. Shared: the session header owns the controls, the Accordion tab owns the
// form, and both have to agree on which accordion is in hand.
export function useSessionAccordion() {
  const { active, refresh } = useSessions()
  const library = useInstruments()

  // The remembered accordion when the session links to one, otherwise the session's own copy -
  // which an imported .stsf always carries, and which is what "Remember it" puts on the shelf.
  const source = computed<InstrumentTemplate | null>(() => {
    const s = active.value
    if (!s) return null

    const kept = s.instrumentId ? library.find(s.instrumentId) : undefined
    if (kept) return kept

    return {
      id: '',
      name: s.instrument?.name ?? '',
      hasImage: false,
      instrument: (s.instrument ?? { serial: '', reedCount: 1 }) as InstrumentTemplate['instrument'],
    }
  })

  // A photograph and a .stif file belong to the shelf entry, so both need one to exist.
  const remembered = computed(() => Boolean(source.value?.id))
  const imageSrc = computed(() =>
    (source.value?.id && source.value.hasImage ? library.imageOf(source.value) : ''))

  async function setPhoto() {
    if (source.value?.id) await library.setImage(source.value.id)
  }

  async function exportIt() {
    if (source.value?.id) await library.exportFile(source.value)
  }

  async function remember() {
    const s = active.value
    if (!s) return

    await SessionService.AdoptInstrument(s.id, s.instrument?.name?.trim() || s.name)
    await library.refresh()
    await refresh()
  }

  return { source, remembered, imageSrc, setPhoto, exportIt, remember }
}
