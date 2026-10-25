// app/composables/pendingEdits.ts
const editors = new Set<() => Promise<boolean>>()
let flushing: Promise<boolean> | undefined

export function registerPendingEdit(flush: () => Promise<boolean>): () => void {
  editors.add(flush)
  return () => {
    editors.delete(flush)
  }
}

export function flushEdits(): Promise<boolean> {
  if (flushing) return flushing
  flushing = (async () => {
    for (const flush of [...editors]) {
      if (!await flush()) return false
    }
    return true
  })().finally(() => { flushing = undefined })
  return flushing
}
