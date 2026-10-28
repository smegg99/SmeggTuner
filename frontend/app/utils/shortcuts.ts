// The global key map, pure so it can be tested; useShortcuts runs the actions.

export type ShortcutAction
  = | 'rooms1' | 'rooms2' | 'rooms3'
    | 'manual' | 'hold' | 'record' | 'undo'
    | 'noteDown' | 'noteUp' | 'takeNow'

export interface ShortcutContext {
  /** the tune room is showing */
  tune: boolean
  /** the input is a recording: Space is FileTransport's play/pause */
  file: boolean
}

const EVERYWHERE: Record<string, ShortcutAction> = {
  '1': 'rooms1',
  '2': 'rooms2',
  '3': 'rooms3',
  'a': 'manual',
  'h': 'hold',
  'r': 'record',
  'z': 'undo',
  ',': 'noteDown',
  '.': 'noteUp',
}

// The reference tuner's keys; only in the tune room, where the arrows and Enter mean nothing else.
const TUNE_ROOM: Record<string, ShortcutAction> = {
  ArrowLeft: 'noteDown',
  ArrowRight: 'noteUp',
  Enter: 'takeNow',
  Backspace: 'undo',
}

/** shortcutFor names the action a key press asks for, or undefined. */
export function shortcutFor(key: string, ctx: ShortcutContext): ShortcutAction | undefined {
  const k = key.length === 1 ? key.toLowerCase() : key
  if (k === ' ') return ctx.tune && !ctx.file ? 'hold' : undefined
  return EVERYWHERE[k] ?? (ctx.tune ? TUNE_ROOM[k] : undefined)
}
