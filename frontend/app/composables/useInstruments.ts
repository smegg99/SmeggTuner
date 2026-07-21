import { computed, ref, shallowRef } from 'vue'
import * as SessionService from '~~bindings/smegg.me/smeggtuner/services/session/service.js'
import type { Instrument, InstrumentTemplate } from '~/types/session'

// The instrument library. Starts empty by design: no default instruments ship, so a session is never seeded with a plausible-but-wrong accordion.
const list = shallowRef<InstrumentTemplate[]>([])
const loading = ref(false)
const error = ref('')

let loaded = false

export function useInstruments() {
  const { t } = useI18n()

  async function refresh() {
    loading.value = true
    error.value = ''
    try {
      list.value = (await SessionService.Instruments()) as unknown as InstrumentTemplate[]
      loaded = true
    }
    catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    }
    finally {
      loading.value = false
    }
  }

  // load is refresh, once: the list only changes when this app changes it.
  async function load() {
    if (loaded) return
    await refresh()
  }

  function find(id: string): InstrumentTemplate | undefined {
    return list.value.find(i => i.id === id)
  }

  // Returns a URL, not image data: embedding base64 over the binding would be megabytes per open. Served by the asset server (services/session/images.go).
  function imageOf(i: InstrumentTemplate): string {
    // The revision in the URL busts the webview cache: a replaced photo at the same src would paint the stale cached copy. See services/session/images.go.
    return i.hasImage ? `/instruments/${i.id}/image?v=${i.imageRev ?? 0}` : ''
  }

  // apply overlays a template but keeps the bench instrument's serial number.
  function apply(i: InstrumentTemplate, onto: Instrument): Instrument {
    // Name comes from the shelf entry (i.name), not the instrument inside it, or the session shows "Instrument not named".
    return { ...i.instrument, name: i.name, serial: onto.serial }
  }

  // Empty id creates a new instrument; an existing id edits it in place.
  async function save(i: InstrumentTemplate) {
    const saved = await SessionService.SaveInstrumentSpec(i as never)
    await refresh()
    return saved as unknown as InstrumentTemplate | null
  }

  async function fromBench(name: string) {
    const saved = await SessionService.SaveInstrument(name)
    await refresh()
    return saved as unknown as InstrumentTemplate | null
  }

  async function remove(id: string) {
    await SessionService.DeleteInstrument(id)
    await refresh()
  }

  // Hands Go only a path; Go opens, decodes (so a non-image fails there) and caps it - not the webview.
  async function setImage(id: string) {
    const path = await SessionService.OpenImageDialog(
      t('instrument.file.imageTitle'),
      t('instrument.file.imageFilter'),
    )
    if (!path) return false

    await SessionService.SetInstrumentImage(id, path)
    await refresh()
    return true
  }

  async function clearImage(id: string) {
    await SessionService.SetInstrumentImage(id, '')
    await refresh()
  }

  // importFile reads a .stif (accordion and photograph) onto the shelf.
  async function importFile(): Promise<InstrumentTemplate | undefined> {
    const path = await SessionService.OpenInstrumentDialog(
      t('instrument.file.openTitle'),
      t('instrument.file.filterName'),
    )
    if (!path) return undefined

    const added = await SessionService.ImportInstrument(path)
    await refresh()
    return (added ?? undefined) as unknown as InstrumentTemplate | undefined
  }

  // exportFile writes a .stif, photograph and all.
  async function exportFile(i: InstrumentTemplate): Promise<boolean> {
    const suggested = await SessionService.SuggestInstrumentFileName(i.name)
    const path = await SessionService.SaveFileDialog(
      'instrument',
      suggested,
      t('instrument.file.exportTitle'),
      t('instrument.file.filterName'),
    )
    if (!path) return false

    await SessionService.ExportInstrument(i.id, path)
    return true
  }

  return {
    list,
    loading,
    error,
    // empty is the fresh-install state: the session dialog offers actions instead of a dead dropdown.
    empty: computed(() => !loading.value && list.value.length === 0),
    load,
    refresh,
    find,
    imageOf,
    apply,
    save,
    fromBench,
    remove,
    setImage,
    clearImage,
    importFile,
    exportFile,
  }
}
