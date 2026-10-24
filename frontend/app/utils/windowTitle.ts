import type { SessionProgress } from '~/utils/sessionProgress'
import type { View } from '~/composables/useShell'

// The window title without the app name (the Go side prepends it); t is injected so the rules are testable.
export type Translate = (key: string, params?: Record<string, unknown>) => string

// An uncalibrated mic says so in every room, so the title keeps asking for the calibration.
export function windowTitle(view: View, progress: SessionProgress, t: Translate, uncalibrated = false): string {
  const title = roomTitle(view, progress, t)
  return uncalibrated ? t('title.uncalibrated', { title }) : title
}

function roomTitle(view: View, progress: SessionProgress, t: Translate): string {
  const room = t(`title.room.${view}`)

  if (progress.kind === 'none' || !progress.name) return room

  if (progress.kind === 'idle') {
    return t('title.session', { room, name: progress.name })
  }

  if (progress.total === null) {
    return t('title.count', { room, name: progress.name, done: progress.done })
  }

  return t('title.progress', {
    room,
    name: progress.name,
    done: progress.done,
    total: progress.total,
  })
}
