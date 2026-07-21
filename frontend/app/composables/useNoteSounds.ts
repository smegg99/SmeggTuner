import { computed } from 'vue'
import { useConfigSync } from '~/composables/useConfigSync'
import { useTuner } from '~/composables/useTuner'

// The mic hears this tone, so the engine would measure it; the record guard lives in Go (services/tuner.observe), not here.
export function useNoteSounds() {
  const { config } = useConfigSync()
  const { playTone, stopTone } = useTuner()

  const enabled = computed<boolean>({
    get: () => config.tuner?.note_sounds ?? false,
    set: (on: boolean) => {
      config.tuner.note_sounds = on

      // Turning it off with a finger still down would leave the tone droning.
      if (!on) void stopTone()
    },
  })

  const press = (note: number) => {
    if (!enabled.value) return
    void playTone(note)
  }

  const release = () => {
    if (!enabled.value) return
    void stopTone()
  }

  return { enabled, press, release }
}
