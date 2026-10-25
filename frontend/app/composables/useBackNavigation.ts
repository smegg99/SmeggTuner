// app/composables/useBackNavigation.ts
import { onBeforeUnmount, onMounted, watch } from 'vue'
import { useShell } from './useShell'
import { useSessions } from './useSessions'
import { error, keyOf } from './sessionStore'

const steps: (() => Promise<boolean>)[] = []

declare global {
  interface Window {
    smeggTunerBack?: { setEnabled(enabled: boolean): void }
    __smeggTunerBack?: () => Promise<boolean>
  }
}

export function registerBackStep(step: () => Promise<boolean>): void {
  steps.push(step)
  onBeforeUnmount(() => {
    const index = steps.indexOf(step)
    if (index >= 0) steps.splice(index, 1)
  })
}

function overlayOpen(): boolean {
  return Boolean(document.querySelector('.v-overlay--active:not(.v-tooltip)'))
}

export function useBackNavigation(): void {
  const { view, openSession, closeSession, setView } = useShell()
  const sessions = useSessions()
  let running: Promise<boolean> | undefined
  let observer: MutationObserver | undefined
  let stopWatch: (() => void) | undefined
  let enabled: boolean | undefined

  function publish() {
    const next = view.value !== 'tune' || overlayOpen()
    if (next !== enabled) window.smeggTunerBack?.setEnabled(next)
    enabled = next
  }

  async function back(): Promise<boolean> {
    if (overlayOpen()) {
      window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
      return true
    }
    for (const step of [...steps].reverse()) {
      if (await step()) return true
    }
    if (view.value === 'workshop' && openSession.value) {
      if (await sessions.close()) closeSession()
      return true
    }
    if (view.value !== 'tune') {
      await setView(view.value === 'calibrate' ? 'workshop' : 'tune')
      return true
    }
    return false
  }

  onMounted(() => {
    if (!window.smeggTunerBack) return
    window.__smeggTunerBack = () => {
      if (!running) {
        running = back().catch((failure) => {
          error.value = keyOf(failure)
          return true
        }).finally(() => {
          running = undefined
          publish()
        })
      }
      return running
    }
    stopWatch = watch(view, publish, { flush: 'post', immediate: true })
    observer = new MutationObserver(publish)
    observer.observe(document.body, { subtree: true, childList: true, attributes: true, attributeFilter: ['class'] })
  })
  onBeforeUnmount(() => {
    stopWatch?.()
    observer?.disconnect()
    window.smeggTunerBack?.setEnabled(false)
    delete window.__smeggTunerBack
  })
}
