import { ref } from 'vue'

// A request bus from the workshop toolbar to the on-screen view (siblings under AppWindow). A counter, not a boolean: a flag reset after reading gets read twice or missed.
export type Request = 'new' | 'import' | 'print' | 'export'

const seq = ref(0)
const last = ref<Request | null>(null)

export function useWorkshop() {
  return {
    seq,
    last,

    ask: (what: Request) => {
      last.value = what
      seq.value++
    },
  }
}
