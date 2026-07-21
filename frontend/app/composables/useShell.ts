import { computed, ref } from 'vue'

// The top-level views; module scope so the toolbar and window agree without prop drilling.
export type View = 'tune' | 'workshop' | 'settings' | 'calibrate'

export type Section = 'sessions' | 'instruments'

const view = ref<View>('tune')
const section = ref<Section>('sessions')

// Id of the open session, or null for the list.
const openSession = ref<string | null>(null)

export function useShell() {
  return {
    view,
    section,
    openSession,

    // Does not stop the engine (would be a cycle to useTuner); AppWindow watches view and does that.
    setView: (next: View) => { view.value = next },
    setSection: (next: Section) => {
      section.value = next
      openSession.value = null
    },

    openSessionAt: (id: string) => {
      section.value = 'sessions'
      openSession.value = id
    },
    closeSession: () => { openSession.value = null },

    reading: computed(() => view.value === 'workshop' && openSession.value !== null),

  }
}
