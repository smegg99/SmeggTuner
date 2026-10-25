import { onMounted, onScopeDispose, ref, watch, toValue, type MaybeRefOrGetter, type Ref } from 'vue'

/** Number of full rows that fit in a fixed panel. Pagination owns overflow. */
export function usePageSize(host: Ref<HTMLElement | null>, rowHeight: MaybeRefOrGetter<number>) {
  const count = ref(1)
  let observer: ResizeObserver | undefined
  const measure = () => {
    const element = host.value
    const style = element ? getComputedStyle(element) : undefined
    const padding = style ? parseFloat(style.paddingTop) + parseFloat(style.paddingBottom) : 0
    count.value = Math.max(1, Math.floor(((element?.clientHeight ?? 0) - padding) / toValue(rowHeight)))
  }
  // The count changes the rows inside host; measured on the next frame, outside the observer's callback.
  let frame = 0
  const queue = () => {
    if (!frame) frame = requestAnimationFrame(() => {
      frame = 0
      measure()
    })
  }
  onMounted(() => {
    observer = new ResizeObserver(queue)
    if (host.value) observer.observe(host.value)
    measure()
  })
  watch(host, (next, previous) => {
    if (previous) observer?.unobserve(previous)
    if (next) observer?.observe(next)
    measure()
  })
  watch(() => toValue(rowHeight), measure)
  onScopeDispose(() => {
    observer?.disconnect()
    cancelAnimationFrame(frame)
  })
  return count
}
