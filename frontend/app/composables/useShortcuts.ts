import { onBeforeUnmount, onMounted } from 'vue'
import { useNoteSounds } from '~/composables/useNoteSounds'
import { useRecord } from '~/composables/useRecord'
import { useShell } from '~/composables/useShell'
import { AUTO_NOTE, NOTE_MAX, NOTE_MIN, useTuner } from '~/composables/useTuner'

// Global keys; must not collide with keys owned elsewhere: FileTransport (space/home/end/L/M/F/S/zoom, only while a file is open), RecordControls (Enter/Backspace), the curve canvas (arrows).

/** Default pinned note (A4) when the detector has nothing to hand over. */
const DEFAULT_PIN = 69

/** A key pressed into a field is text, not a command. */
function typing(target: EventTarget | null): boolean {
  const el = target as HTMLElement | null
  return Boolean(el && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA' || el.isContentEditable))
}

export function useShortcuts() {
  const { setView } = useShell()
  const { note, manualNote, setManualNote, frozen, setFrozen, canFreeze } = useTuner()
  const { enabled: sounds } = useNoteSounds()
  const { sessionId, armed, readings, toggleRecording, undo } = useRecord()

  // Needs a session, else the key would silently do nothing; the engine runs its own warm-up once armed.
  function toggleRecord() {
    if (!sessionId.value) return
    toggleRecording(!armed.value)
  }

  function undoTake() {
    if (!sessionId.value || !readings.value) return
    void undo()
  }

  function step(delta: number) {
    if (manualNote.value === AUTO_NOTE) return // auto: the detector owns the note, not the keys
    void setManualNote(Math.min(NOTE_MAX, Math.max(NOTE_MIN, manualNote.value + delta)))
  }

  // Switching to manual pins whatever the detector is currently on.
  function toggleManual() {
    const manual = manualNote.value !== AUTO_NOTE
    void setManualNote(manual ? AUTO_NOTE : (note.value || DEFAULT_PIN))
  }

  function onKey(event: KeyboardEvent) {
    // A modified key belongs to the desktop, not to us.
    if (event.ctrlKey || event.metaKey || event.altKey) return
    if (typing(event.target)) return

    const keys: Record<string, () => void> = {
      '1': () => setView('tune'),
      '2': () => setView('workshop'),
      '3': () => setView('settings'),
      'a': toggleManual,
      'n': () => { sounds.value = !sounds.value },
      // Freeze needs a reading on screen to hold (including one a stopped run left).
      'h': () => { if (canFreeze.value) void setFrozen(!frozen.value) },
      'r': toggleRecord,
      'z': undoTake,
      ',': () => step(-1),
      '.': () => step(1),
    }

    const run = keys[event.key.length === 1 ? event.key.toLowerCase() : event.key]
    if (!run) return

    event.preventDefault()
    run()
  }

  onMounted(() => window.addEventListener('keydown', onKey))
  onBeforeUnmount(() => window.removeEventListener('keydown', onKey))
}
