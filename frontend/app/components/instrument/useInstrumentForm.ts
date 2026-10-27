import { computed, reactive, ref, watch } from 'vue'
import type { Ref } from 'vue'
import { banksOfRegisters } from '~/utils/banks'
import { A4_DEFAULT, A4_MAX, A4_MIN } from '~/utils/reference'
import { DEFAULT_BEAT_TOLERANCE, DEFAULT_TOLERANCE } from '~/types/config'
import type { Bank, BassRegister, InstrumentTemplate, Register } from '~/types/session'

const TOL_MAX = 50

// The accordion editor's state, validation, and the template it hands back. Driven by
// the accordion it is editing: a new one arrives, the form reloads from it.
export function useInstrumentForm(source: Ref<InstrumentTemplate | null>) {
  const nameError = ref(false)

  const form = reactive({
    id: '',
    name: '',
    lo: 0, // MIDI note; 0 means unset
    hi: 0,
    a4: A4_DEFAULT,
    tolerance: DEFAULT_TOLERANCE,
    beatTolerance: DEFAULT_BEAT_TOLERANCE,
  })

  // Text drafts, clamped on commit (not mid-keystroke) so a decimal or 442 can be typed through.
  const a4Draft = ref(String(A4_DEFAULT))
  const tolDraft = ref(String(DEFAULT_TOLERANCE))
  const beatDraft = ref(String(DEFAULT_BEAT_TOLERANCE))

  const registers = ref<Register[]>([])

  // The bass machine: 0 is no bass section; its registers only exist inside a declared machine.
  const bassReeds = ref(0)
  const bassRegisters = ref<BassRegister[]>([])

  const banks = computed<Bank[]>(() => banksOfRegisters(registers.value))

  // The reed count is the widest switch's bank count.
  const reedCount = computed(() =>
    Math.max(1, ...registers.value.map(r => r.banks.length)),
  )

  // Only a backwards range when both ends are set.
  const rangeError = computed(() => form.lo !== 0 && form.hi !== 0 && form.lo > form.hi)

  function reload() {
    nameError.value = false

    const i = source.value
    Object.assign(form, {
      id: i?.id ?? '',
      name: i?.name ?? '',
      lo: i?.instrument.lo ?? 0,
      hi: i?.instrument.hi ?? 0,
      a4: i?.instrument.a4 || A4_DEFAULT,
      tolerance: i?.instrument.tolerance || DEFAULT_TOLERANCE,
      beatTolerance: i?.instrument.beatTolerance || DEFAULT_BEAT_TOLERANCE,
    })
    a4Draft.value = String(form.a4)
    tolDraft.value = String(form.tolerance)
    beatDraft.value = String(form.beatTolerance)
    registers.value = (i?.instrument.registers ?? []).map(r => ({ name: r.name, banks: [...r.banks] }))
    bassReeds.value = i?.instrument.bassReeds ?? 0
    bassRegisters.value = (i?.instrument.bassRegisters ?? []).map(r => ({ name: r.name, feet: [...r.feet] }))
  }

  // Identity, not deep equality: a re-fetch of the same accordion must not throw away
  // edits in progress, and switching to another one must reload.
  watch(() => source.value?.id, reload, { immediate: true })

  function commitA4() {
    const n = Number(a4Draft.value)
    form.a4 = Number.isFinite(n) ? Math.min(A4_MAX, Math.max(A4_MIN, n)) : A4_DEFAULT
    a4Draft.value = String(form.a4)
  }

  function clampTol(draft: string, fallback: number): number {
    const n = Number(draft)
    return Number.isFinite(n) && n > 0 ? Math.min(TOL_MAX, n) : fallback
  }

  function commitTol() {
    form.tolerance = clampTol(tolDraft.value, DEFAULT_TOLERANCE)
    tolDraft.value = String(form.tolerance)
  }

  function commitBeat() {
    form.beatTolerance = clampTol(beatDraft.value, DEFAULT_BEAT_TOLERANCE)
    beatDraft.value = String(form.beatTolerance)
  }

  // The edited accordion, or null when it does not add up; the caller decides where to put it.
  // It does NOT commit the number drafts: a draft commits when its own field is left, and forcing
  // one mid-edit would clamp a half-typed 44 to the 430 floor under the cursor.
  function collect(): InstrumentTemplate | null {
    const name = form.name.trim()
    nameError.value = name === ''
    if (nameError.value || rangeError.value) return null

    return {
      id: form.id,
      name,
      hasImage: source.value?.hasImage ?? false,
      instrument: {
        // The name on the instrument itself, so a session copy and a .stif export are self-describing.
        name,
        serial: source.value?.instrument.serial ?? '',
        reedCount: reedCount.value,
        banks: banks.value,
        registers: registers.value,
        bassReeds: bassReeds.value || undefined,
        bassRegisters: bassReeds.value > 0 && bassRegisters.value.length ? bassRegisters.value : undefined,
        lo: form.lo || undefined,
        hi: form.hi || undefined,
        a4: form.a4,
        tolerance: form.tolerance,
        beatTolerance: form.beatTolerance,
      },
    }
  }

  return {
    form,
    nameError,
    a4Draft,
    tolDraft,
    beatDraft,
    registers,
    bassReeds,
    bassRegisters,
    banks,
    reedCount,
    rangeError,
    reload,
    commitA4,
    commitTol,
    commitBeat,
    collect,
  }
}
