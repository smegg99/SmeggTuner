// The browser's "ResizeObserver loop completed with undelivered notifications" is not a failure: it says
// a size change was handed to the next frame. It arrives as a window error, so the log bridge would file
// it as one; this listener, installed before the bridge, keeps it out and reports it once.
const NOTICE = 'ResizeObserver loop'

export function filterResizeNotice(once: () => void): () => void {
  let said = false
  const onError = (event: ErrorEvent) => {
    if (!String(event.message).includes(NOTICE)) return
    event.stopImmediatePropagation()
    event.preventDefault()
    if (!said) {
      said = true
      once()
    }
  }
  window.addEventListener('error', onError)
  return () => window.removeEventListener('error', onError)
}
