// Split from useSessions; its return is spread back into that composable's API unchanged.
import { computed } from 'vue'
import * as SessionService from '~~bindings/smegg.me/smeggtuner/services/session/service.js'
import { active, run } from '~/composables/sessionStore'
import type { FitDTO } from '~~bindings/smegg.me/smeggtuner/services/session/models.js'

export function useSessionCurve() {
  const { t } = useI18n()

  // `value` is in the unit the user is authoring in.
  async function setAnchor(note: number, reed: number, value: number, unit: string) {
    return run(() => SessionService.SetAnchor(note, reed, value, unit))
  }

  // One beating for a note; the backend derives every reed from RefReed and Asymmetry.
  async function setBeating(note: number, value: number, unit: string) {
    return run(() => SessionService.SetBeating(note, value, unit))
  }

  async function clearAnchor(note: number) {
    return run(() => SessionService.ClearAnchor(note))
  }

  // Clear the list, keep the curve settings (RefReed, Asymmetry, flags). DropCurve takes them.
  async function clearAnchors() {
    const notes = (active.value?.curve?.anchors ?? []).map(anchor => anchor.note)
    return run(async () => {
      for (const note of notes) await SessionService.ClearAnchor(note)
    })
  }

  // Where the reference reed sits in the tremolo, in percent; read on the NEXT beating.
  async function setAsymmetry(percent: number) {
    return run(() => SessionService.SetAsymmetry(percent))
  }

  async function setInterpolate(on: boolean) {
    return run(() => SessionService.SetInterpolate(on))
  }

  async function setExtrapolateLeft(on: boolean) {
    return run(() => SessionService.SetExtrapolateLeft(on))
  }

  async function setExtrapolateRight(on: boolean) {
    return run(() => SessionService.SetExtrapolateRight(on))
  }

  // Which reed the curve calls "at pitch", and the unit the next anchors are typed in.
  async function setRefReed(reed: number) {
    return run(() => SessionService.SetRefReed(reed))
  }

  async function setCurveUnit(unit: string) {
    return run(() => SessionService.SetCurveUnit(unit))
  }

  async function dropCurve() {
    return run(() => SessionService.DropCurve())
  }

  async function fitCurve(): Promise<FitDTO | undefined> {
    const fit = await run(() => SessionService.FitCurve())
    return fit ?? undefined
  }

  async function importCurve(fromID: string) {
    return run(() => SessionService.ImportCurve(fromID))
  }

  async function importCurveFile() {
    const path = await run(() => SessionService.OpenFileDialog(
      t('session.file.openTitle'),
      t('session.file.openFilter'),
    ))
    if (!path) return
    return run(() => SessionService.ImportCurveFile(path))
  }

  return {
    hasCurve: computed(() => (active.value?.curve?.anchors?.length ?? 0) > 0),
    setAnchor,
    setBeating,
    clearAnchor,
    clearAnchors,
    setAsymmetry,
    setInterpolate,
    setExtrapolateLeft,
    setExtrapolateRight,
    setRefReed,
    setCurveUnit,
    dropCurve,
    fitCurve,
    importCurve,
    importCurveFile,
  }
}
