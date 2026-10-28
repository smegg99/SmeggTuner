// The main bar's height, published as --app-bar-h on the document, so dialog heads match it at every
// window size; the bar's own container queries set it, which no media query can follow.
import { onBeforeUnmount, onMounted, type ShallowRef } from 'vue'

export function useBarHeight(bar: Readonly<ShallowRef<HTMLElement | null>>) {
  let observer: ResizeObserver | undefined
  const publish = () => {
    const h = bar.value?.getBoundingClientRect().height
    if (h) document.documentElement.style.setProperty('--app-bar-h', `${h}px`)
  }
  onMounted(() => {
    observer = new ResizeObserver(publish)
    if (bar.value) observer.observe(bar.value)
    publish()
  })
  onBeforeUnmount(() => observer?.disconnect())
}
