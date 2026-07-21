import { ref, watch } from 'vue'

// A string draft absorbs partial input ("-", "", "4.") so it never reaches the config; only a committed value (blur/enter) is clamped and written back.
interface DraftNumberOptions {
  min: number
  max: number
  /** rounding grain: 1 rounds to whole numbers, 0.1 keeps one decimal. */
  step?: number
}

export function useDraftNumber(
  read: () => number,
  write: (value: number) => void,
  options: DraftNumberOptions,
) {
  const draft = ref(String(read()))

  // A change from elsewhere replaces the draft, but only when it disagrees with
  // what is typed, so a keystroke is never pulled out from under the cursor.
  watch(read, (value) => {
    if (value !== Number(draft.value)) draft.value = String(value)
  })

  function commit() {
    const current = read()
    const value = Number(draft.value)
    if (draft.value.trim() === '' || !Number.isFinite(value)) {
      draft.value = String(current)
      return
    }

    const step = options.step ?? 1
    // toFixed(6) sweeps up the floating-point crumbs a divide leaves behind, so
    // a 0.1 step does not commit 3.0000000004.
    const clean = Number((Math.min(options.max, Math.max(options.min, Math.round(value / step) * step)).toFixed(6)))
    draft.value = String(clean)
    if (clean !== current) write(clean)
  }

  return { draft, commit }
}
