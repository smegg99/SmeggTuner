// app/web/wakeLock.ts
import { onScopeDispose, watch, type Ref } from 'vue'

interface Sentinel {
  released: boolean
  release(): Promise<void>
}

// useWakeLock keeps the screen on while running; the browser drops the lock when the tab hides, so it is taken again on return.
export function useWakeLock(running: Ref<boolean>) {
  const api = (navigator as Navigator & { wakeLock?: { request(type: 'screen'): Promise<Sentinel> } }).wakeLock
  if (!api) return
  let lock: Sentinel | null = null

  async function acquire() {
    if (lock && !lock.released) return
    try {
      lock = await api!.request('screen')
    }
    catch {
      lock = null
    }
  }
  async function release() {
    const l = lock
    lock = null
    await l?.release().catch(() => {})
  }
  const onVisible = () => {
    if (document.visibilityState === 'visible' && running.value) void acquire()
  }

  watch(running, on => void (on ? acquire() : release()), { immediate: true })
  document.addEventListener('visibilitychange', onVisible)
  onScopeDispose(() => {
    document.removeEventListener('visibilitychange', onVisible)
    void release()
  })
}
