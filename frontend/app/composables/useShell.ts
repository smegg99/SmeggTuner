import { computed, ref } from 'vue'
import { flushEdits } from './pendingEdits'

// The top-level views; module scope so the toolbar and window agree without prop drilling.
export type View = 'tune' | 'workshop' | 'settings' | 'calibrate' | 'soundcard'

const view = ref<View>('tune')
// The view left last, for a room's Return key.
const previous = ref<View>('tune')
// What the open room says beside its title in the main bar (the device, the instrument); '' for nothing.
const pageMeta = ref('')
// Id of the open session, or null for the list.
const openSession = ref<string | null>(null)
let viewRequest = 0

async function setView(next: View): Promise<boolean> {
  const request = ++viewRequest
  if (view.value === next) return true
  if (!await flushEdits() || request !== viewRequest) return false
  previous.value = view.value
  view.value = next
  return true
}

export function useShell() {
  return {
    view,
    previous,
    pageMeta,
    openSession,

    // Does not stop the engine (would be a cycle to useTuner); AppWindow watches view and does that.
    setView,
    openSessionAt: (id: string) => { openSession.value = id },
    closeSession: () => { openSession.value = null },

    reading: computed(() => view.value === 'workshop' && openSession.value !== null),

  }
}
