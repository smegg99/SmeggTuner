// app/composables/useLogger.ts
import * as LoggerService from '~~bindings/github.com/smegg99/s99wails/v2/services/logger/service.js'
import type { LogLevel, Fields } from '~/types/logger'
import type messages from '../../i18n/generated/en/log.json'

type LogKey = `log.${keyof typeof messages.log}`

const levelToBackend: Record<LogLevel, (msg: string, fields: Fields) => Promise<void>> = {
  trace: LoggerService.LogDebug,
  debug: LoggerService.LogDebug,
  info: LoggerService.LogInfo,
  warn: LoggerService.LogWarn,
  error: LoggerService.LogError,
}

export function useLogger() {
  function log(level: LogLevel, key: LogKey, fields: Fields = {}) {
    levelToBackend[level](key, fields).catch(() => {})
  }
  return {
    trace: (key: LogKey, fields?: Fields) => log('trace', key, fields),
    debug: (key: LogKey, fields?: Fields) => log('debug', key, fields),
    info: (key: LogKey, fields?: Fields) => log('info', key, fields),
    warn: (key: LogKey, fields?: Fields) => log('warn', key, fields),
    error: (key: LogKey, fields?: Fields) => log('error', key, fields),
  }
}
