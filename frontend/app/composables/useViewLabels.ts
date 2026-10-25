// Whether the room keys carry their names: only when the window is wide enough for the tuner's bar with
// them, the widest bar, so every room shows them alike at one window size.
import { onBeforeUnmount, onMounted, ref, watch, type Ref, type ShallowRef } from 'vue'

// The names are shown only with this much room to spare, so a bar on the edge cannot flip every frame.
const MARGIN = 2

type Part = Readonly<ShallowRef<HTMLElement | null>>
type Keys = Readonly<ShallowRef<{ $el: HTMLElement } | null>>

const width = (el: Element | null | undefined) => el?.getBoundingClientRect().width ?? 0

/** useViewLabels measures the bar; the tuner's need is its content with the named keys, read from the probe while measuring() holds. */
export function useViewLabels(bar: Part, spacer: Part, views: Keys, probe: Keys, measuring: () => boolean, locale: Readonly<Ref<string>>) {
  const labelled = ref(true)
  let need = 0
  let frame = 0

  function fit() {
    frame = 0
    const box = bar.value
    const keys = views.value?.$el
    const named = probe.value?.$el
    if (!box || !keys || !named) return
    const room = width(box)
    const free = width(spacer.value)
    if (measuring() || need === 0) {
      // No free space means the bar is squeezing something: the tuner needs more than it has.
      need = free < 1 ? room + MARGIN : room - free - width(keys) + width(named)
    }
    labelled.value = room >= need + MARGIN
  }

  // The observer only asks for a frame: a label change inside its callback would resize what it watches.
  const queueFit = () => {
    if (!frame) frame = requestAnimationFrame(fit)
  }

  let observer: ResizeObserver | undefined
  onMounted(() => {
    observer = new ResizeObserver(queueFit)
    for (const el of [bar.value, spacer.value, probe.value?.$el]) {
      if (el) observer.observe(el)
    }
    fit()
  })
  onBeforeUnmount(() => {
    observer?.disconnect()
    cancelAnimationFrame(frame)
  })

  // The probe's names change width with the language; the tuner's need is measured again there.
  watch(locale, () => {
    need = 0
    queueFit()
  })

  return labelled
}
