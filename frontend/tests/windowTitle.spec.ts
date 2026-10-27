import { describe, expect, it } from 'vitest'
import { windowTitle } from '~/utils/windowTitle'
import type { Translate } from '~/utils/windowTitle'
import type { SessionProgress } from '~/utils/sessionProgress'

// A fake translator echoing key and params: this tests which case fires, not
// the English wording.
const t: Translate = (key, params) => {
  if (!params) return key
  const filled = Object.entries(params).map(([k, v]) => `${k}=${String(v)}`).join(',')
  return `${key}(${filled})`
}

const none: SessionProgress = { kind: 'none' }

describe('the window title phrase', () => {
  it('is the room alone when no session is open', () => {
    expect(windowTitle('tune', none, t)).toBe('title.room.tune')
    expect(windowTitle('workshop', none, t)).toBe('title.room.workshop')
    expect(windowTitle('settings', none, t)).toBe('title.room.settings')
    expect(windowTitle('calibrate', none, t)).toBe('title.room.calibrate')
  })

  // The name shows in every room, as the footer label does.
  it('names the session whenever one is open, in any room', () => {
    const idle: SessionProgress = { kind: 'idle', name: 'Bugari Armando' }
    expect(windowTitle('tune', idle, t)).toBe('title.session(room=title.room.tune,name=Bugari Armando)')
    expect(windowTitle('settings', idle, t)).toBe('title.session(room=title.room.settings,name=Bugari Armando)')
  })

  it('carries the figures while a pass is being recorded', () => {
    const active: SessionProgress = { kind: 'active', name: 'Bugari', done: 12, total: 41, inTune: 8 }
    expect(windowTitle('tune', active, t))
      .toBe('title.progress(room=title.room.tune,name=Bugari,done=12,total=41)')
  })

  // Keyboard unknown, so there is nothing to be 12 out of.
  it('drops the denominator when the instrument never said its keyboard', () => {
    const active: SessionProgress = { kind: 'active', name: 'Bugari', done: 12, total: null, inTune: null }
    expect(windowTitle('tune', active, t))
      .toBe('title.count(room=title.room.tune,name=Bugari,done=12)')
  })

  it('does not put an unnamed session in the title', () => {
    const idle: SessionProgress = { kind: 'idle', name: '' }
    expect(windowTitle('workshop', idle, t)).toBe('title.room.workshop')
  })

  // An uncalibrated mic is named in every room, around whatever the room's title is.
  it('asks for the sound card calibration while the mic is uncalibrated', () => {
    expect(windowTitle('tune', none, t, true)).toBe('title.uncalibrated(title=title.room.tune)')
    const idle: SessionProgress = { kind: 'idle', name: 'Bugari' }
    expect(windowTitle('settings', idle, t, true))
      .toBe('title.uncalibrated(title=title.session(room=title.room.settings,name=Bugari))')
    expect(windowTitle('soundcard', none, t)).toBe('title.room.soundcard')
  })
})
