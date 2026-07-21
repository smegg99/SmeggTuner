import { ref } from 'vue'
import * as ReportService from '~~bindings/smegg.me/smeggtuner/services/report/service.js'
import type { OptionsDTO } from '~~bindings/smegg.me/smeggtuner/services/report/models.js'
import type { ReportOptions } from '~/types/record'

// Thin wrapper: the backend builds the card, shows the save dialog and writes the file. A cancelled save returns an empty path (not an error), so exportSession reports success only when a file was written.

const ERROR_KEY_PATTERN = /(?:session|report)\.error\.[A-Za-z0-9]+/
const ERROR_UNEXPECTED = 'report.error.renderFailed'

const busy = ref(false)
const error = ref('')

function keyOf(err: unknown): string {
  const text = err instanceof Error ? err.message : String(err)
  return ERROR_KEY_PATTERN.exec(text)?.[0] ?? ERROR_UNEXPECTED
}

export function useReport() {
  /** Export the card (pdf/html/csv). True only when a file was written; false on cancel or failure, with `error` set. */
  async function exportSession(opts: ReportOptions): Promise<boolean> {
    busy.value = true
    error.value = ''
    try {
      const res = await ReportService.Export(opts as unknown as OptionsDTO)
      return Boolean(res?.path)
    }
    catch (err) {
      error.value = keyOf(err)
      return false
    }
    finally {
      busy.value = false
    }
  }

  function clearError() {
    error.value = ''
  }

  return { busy, error, exportSession, clearError }
}
