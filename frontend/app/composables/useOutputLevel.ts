import { useTransport } from '~/composables/useTransport'
import * as TunerService from '~~bindings/smegg.me/smeggtuner/services/tuner/service.js'

// One slider over two backend devices: file playback (services/audio) and the reference tone (services/tuner). Wraps useTransport so the two cannot drift.
export function useOutputLevel() {
  const { volume, muted, setVolume: setSpeakerVolume, setMuted: setSpeakerMuted } = useTransport()

  const setVolume = (next: number) => {
    setSpeakerVolume(next)
    void TunerService.SetToneVolume(muted.value ? 0 : next)
  }

  const setMuted = async (next: boolean) => {
    await setSpeakerMuted(next)
    void TunerService.SetToneVolume(next ? 0 : volume.value)
  }

  return { volume, muted, setVolume, setMuted }
}
