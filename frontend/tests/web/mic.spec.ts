// tests/web/mic.spec.ts
import { afterEach, describe, expect, it, vi } from 'vitest'
import { MicError, micConstraints, requestMic } from '~/web/mic'

function stubGetUserMedia(impl: (c: MediaStreamConstraints) => Promise<MediaStream>) {
  const getUserMedia = vi.fn(impl)
  vi.stubGlobal('navigator', { ...navigator, mediaDevices: { getUserMedia } })
  return getUserMedia
}

afterEach(() => vi.unstubAllGlobals())

describe('microphone request', () => {
  it('requests all processing off', () => {
    expect(micConstraints('dev')).toEqual({
      audio: { echoCancellation: false, noiseSuppression: false, autoGainControl: false, channelCount: 1, deviceId: { exact: 'dev' } },
    })
    expect(micConstraints('').audio).not.toHaveProperty('deviceId')
  })

  it('denied permission yields MicError denied and start can retry', async () => {
    const gum = stubGetUserMedia(() => Promise.reject(new DOMException('no', 'NotAllowedError')))
    await expect(requestMic('')).rejects.toMatchObject({ reason: 'denied' })
    await expect(requestMic('')).rejects.toBeInstanceOf(MicError)
    expect(gum).toHaveBeenCalledTimes(2)
  })

  it('missing device yields MicError none', async () => {
    stubGetUserMedia(() => Promise.reject(new DOMException('gone', 'NotFoundError')))
    await expect(requestMic('x')).rejects.toMatchObject({ reason: 'none' })
  })

  it('no mediaDevices at all yields MicError unsupported', async () => {
    vi.stubGlobal('navigator', { ...navigator, mediaDevices: undefined })
    await expect(requestMic('')).rejects.toMatchObject({ reason: 'unsupported' })
  })
})
