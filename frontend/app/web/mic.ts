// app/web/mic.ts

export type MicReason = 'denied' | 'none' | 'unsupported' | 'failed'

export class MicError extends Error {
  override name = 'MicError'
  constructor(readonly reason: MicReason, message?: string) {
    super(message ?? reason)
  }
}

// micConstraints asks the browser for the raw signal; any of the three processing stages changes what the engine reads.
export function micConstraints(deviceId: string): MediaStreamConstraints {
  return {
    audio: {
      echoCancellation: false,
      noiseSuppression: false,
      autoGainControl: false,
      channelCount: 1,
      ...(deviceId ? { deviceId: { exact: deviceId } } : {}),
    },
  }
}

export async function requestMic(deviceId: string): Promise<MediaStream> {
  if (!navigator.mediaDevices?.getUserMedia) throw new MicError('unsupported')
  try {
    return await navigator.mediaDevices.getUserMedia(micConstraints(deviceId))
  }
  catch (e) {
    const name = (e as DOMException)?.name
    if (name === 'NotAllowedError' || name === 'SecurityError') throw new MicError('denied', String(e))
    if (name === 'NotFoundError' || name === 'OverconstrainedError') throw new MicError('none', String(e))
    throw new MicError('failed', String(e))
  }
}

// appliedProcessing is what the browser kept switched on although the request turned it off.
export function appliedProcessing(stream: MediaStream): string[] {
  const s = stream.getAudioTracks()[0]?.getSettings() ?? {}
  return (['echoCancellation', 'noiseSuppression', 'autoGainControl'] as const).filter(k => s[k] === true)
}
