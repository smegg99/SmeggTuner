// app/composables/useShortcuts.ts
import { onBeforeUnmount, onMounted } from 'vue'
import { useAudioDevices } from '~/composables/useAudioDevices'
import { useTransport } from '~/composables/useTransport'
import { useFileRun } from '~/composables/useFileRun'
import { useOutputLevel } from '~/composables/useOutputLevel'
import { useRecord } from '~/composables/useRecord'
import { useShell } from '~/composables/useShell'
import { AUTO_NOTE, useTuner } from '~/composables/useTuner'
import { shortcutFor } from '~/utils/shortcuts'
import { WORKSHOP } from '~/utils/features'
import type { View } from '~/composables/useShell'

// The rooms 1, 2 and 3 open, in tab order; without the workshop, 2 is Settings.
const ROOMS: View[] = WORKSHOP ? ['tune', 'workshop', 'settings'] : ['tune', 'settings']

// The shell owns global keys once, including transport keys shared by both tuning rooms.

function controlOwnsKey(event: KeyboardEvent): boolean {
  const el = event.target
  if (!(el instanceof HTMLElement)) return false
  if (el.isContentEditable || el.closest('input, textarea, select, [contenteditable="true"]')) return true
  const control = el.closest('button, a[href], [role="button"], [role="slider"], [role="tab"]')
  return Boolean(control && ['Enter', ' ', 'Home', 'End', 'ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown'].includes(event.key))
}

export function useShortcuts() {
  const { view, setView } = useShell()
  const { manualNote, setAuto, stepNote, frozen, setFrozen, canFreeze, takeNow } = useTuner()
  const { playPause } = useFileRun()
  const { sessionId, armed, readings, recording, toggleRecording, undo } = useRecord()
  const { current } = useAudioDevices()

  const transport = useTransport()
  const { muted, setMuted } = useOutputLevel()

  const manual = () => manualNote.value !== AUTO_NOTE

  // Needs a session, else the key would silently do nothing; the engine runs its own warm-up once armed.
  function toggleRecord() {
    if (!sessionId.value) return
    toggleRecording(!armed.value)
  }

  // Removes the take captured last; with none left it leaves record mode (services/record.Undo).
  function undoTake() {
    if (!sessionId.value || (!readings.value && !armed.value)) return
    void undo()
  }

  function freeze() {
    if (canFreeze.value) void setFrozen(!frozen.value)
  }

  const actions: Record<string, () => void> = {
    rooms1: () => setView(ROOMS[0]!),
    rooms2: () => setView(ROOMS[1]!),
    rooms3: () => { if (ROOMS[2]) void setView(ROOMS[2]) },
    manual: () => void setAuto(manual()),
    freeze,
    record: toggleRecord,
    undo: undoTake,
    noteDown: () => { if (manual()) void stepNote(-1) },
    noteUp: () => { if (manual()) void stepNote(1) },
    takeNow: () => { if (recording.value) void takeNow() },
  }

  const centre = () => (transport.view.from + transport.view.to) / 2
  const fileActions: Record<string, () => void> = {
    ' ': () => void playPause(),
    'Home': () => void transport.seek(transport.state.from),
    'End': () => void transport.seek(transport.state.to),
    'l': () => void transport.setLoop(!transport.state.loop),
    'm': () => void setMuted(!muted.value),
    'f': () => transport.fit(),
    's': () => transport.fitSelection(),
    '+': () => transport.zoom(0.5, centre()),
    '=': () => transport.zoom(0.5, centre()),
    '-': () => transport.zoom(2, centre()),
  }

  function onKey(event: KeyboardEvent) {
    // A modified key belongs to the desktop, not to us.
    if (event.defaultPrevented || event.isComposing || event.ctrlKey || event.metaKey || event.altKey) return
    if (controlOwnsKey(event)) return
    if (document.querySelector('.v-overlay--active:not(.v-tooltip)')) return

    const file = current.value.kind === 'file'
    if (file && transport.state.available && (view.value === 'tune' || view.value === 'calibrate')) {
      const key = event.key.length === 1 ? event.key.toLowerCase() : event.key
      const run = fileActions[key]
      if (run) {
        if (event.repeat && !['+', '=', '-'].includes(key)) return
        event.preventDefault()
        run()
        return
      }
    }

    const action = shortcutFor(event.key, { tune: view.value === 'tune', file })
    const run = action && actions[action]
    if (!run) return
    if (event.repeat && action !== 'noteDown' && action !== 'noteUp') return

    event.preventDefault()
    run()
  }

  onMounted(() => window.addEventListener('keydown', onKey))
  onBeforeUnmount(() => window.removeEventListener('keydown', onKey))
}
