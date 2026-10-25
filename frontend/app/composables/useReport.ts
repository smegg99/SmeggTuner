import { ref } from 'vue'
import * as ReportService from '~~bindings/smegg.me/smeggtuner/backend/services/report/service.js'
import type { OptionsDTO } from '~~bindings/smegg.me/smeggtuner/backend/services/report/models.js'
import type { ReportOptions } from '~/types/record'

// Android PDF uses the system print UI; file exports return a path after writing.

const ERROR_KEY_PATTERN = /(?:session|report)\.error\.[A-Za-z0-9]+/
const ERROR_UNEXPECTED = 'report.error.renderFailed'

const busy = ref(false)
const error = ref('')

function keyOf(err: unknown): string {
  const text = err instanceof Error ? err.message : String(err)
  return ERROR_KEY_PATTERN.exec(text)?.[0] ?? ERROR_UNEXPECTED
}

export function useReport() {
  /** Close the report dialog after a file is written or Android accepts the print job. */
  async function exportSession(opts: ReportOptions): Promise<boolean> {
    busy.value = true
    error.value = ''
    try {
      const res = await ReportService.Export(opts as unknown as OptionsDTO)
      return Boolean(res?.path || res?.printing)
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
