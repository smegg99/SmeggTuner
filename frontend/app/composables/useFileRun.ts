import { computed } from 'vue'
import { useTransport } from '~/composables/useTransport'
import { useTuner } from '~/composables/useTuner'

// A recording's run: starting the engine and playing the file are one action, so Play is Start.
export function useFileRun() {
  const { running, starting, start, stop } = useTuner()
  const { state, setPaused, seek } = useTransport()

  const playing = computed(() => running.value && !state.paused)

  async function playPause() {
    if (starting.value) return
    if (!running.value) {
      // Clear a leftover pause first, or Start would run and play nothing.
      if (state.paused) await setPaused(false)
      await start()
      return
    }
    await setPaused(!state.paused)
  }

  async function halt() {
    await stop()
    await seek(state.from)
  }

  return { playing, running, playPause, halt }
}
