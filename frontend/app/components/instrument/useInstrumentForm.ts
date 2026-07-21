import { computed, reactive, ref, watch } from 'vue'
import type { Ref } from 'vue'
import { banksOfRegisters } from '~/utils/banks'
import { A4_DEFAULT, A4_MAX, A4_MIN } from '~/types/session'
import { DEFAULT_BEAT_TOLERANCE, DEFAULT_TOLERANCE } from '~/types/config'
import type { Bank, InstrumentTemplate, Register } from '~/types/session'

const TOL_MAX = 50

type Props = { instrument?: InstrumentTemplate | null }
type Emit = (e: 'submit', i: InstrumentTemplate) => void

// Instrument-editor form state, validation, and emitted payload.
export function useInstrumentForm(props: Props, emit: Emit, open: Ref<boolean>) {
  const editing = ref(false)
  const nameError = ref(false)

  const form = reactive({
    id: '',
    name: '',
    make: '',
    model: '',
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

  const banks = computed<Bank[]>(() => banksOfRegisters(registers.value))

  // The reed count is the widest switch's bank count.
  const reedCount = computed(() =>
    Math.max(1, ...registers.value.map(r => r.banks.length)),
  )

  // Only a backwards range when both ends are set.
  const rangeError = computed(() => form.lo !== 0 && form.hi !== 0 && form.lo > form.hi)

  watch(open, (isOpen) => {
    if (!isOpen) return

    nameError.value = false
    editing.value = Boolean(props.instrument)

    const i = props.instrument
    Object.assign(form, {
      id: i?.id ?? '',
      name: i?.name ?? '',
      make: i?.instrument.make ?? '',
      model: i?.instrument.model ?? '',
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
  })

  function commitA4() {
    const n = Math.round(Number(a4Draft.value))
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

  function submit() {
    commitA4()
    commitTol()
    commitBeat()

    const name = form.name.trim()
    nameError.value = name === ''
    if (nameError.value || rangeError.value) return

    emit('submit', {
      id: form.id,
      name,
      hasImage: props.instrument?.hasImage ?? false,
      instrument: {
        // The name on the instrument itself, so a session copy and a .stif export are self-describing.
        name,
        make: form.make.trim(),
        model: form.model.trim(),
        serial: '', // serial belongs to a physical accordion, not the model
        reedCount: reedCount.value,
        banks: banks.value,
        registers: registers.value,
        lo: form.lo || undefined,
        hi: form.hi || undefined,
        a4: form.a4,
        tolerance: form.tolerance,
        beatTolerance: form.beatTolerance,
      },
    })
  }

  return {
    form,
    editing,
    nameError,
    a4Draft,
    tolDraft,
    beatDraft,
    registers,
    banks,
    reedCount,
    rangeError,
    commitA4,
    commitTol,
    commitBeat,
    submit,
  }
}
