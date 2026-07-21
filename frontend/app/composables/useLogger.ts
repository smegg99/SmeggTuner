import * as LoggerService from '~~bindings/github.com/smegg99/s99wails/services/logger/service.js'
import type { LogLevel, Fields } from '~/types/logger'

const levelToBackend: Record<LogLevel, (msg: string, fields: Fields) => Promise<void>> = {
  trace: LoggerService.LogDebug,
  debug: LoggerService.LogDebug,
  info: LoggerService.LogInfo,
  warn: LoggerService.LogWarn,
  error: LoggerService.LogError,
}

const levelToConsole: Record<LogLevel, (...args: unknown[]) => void> = {
  trace: console.debug,
  debug: console.debug,
  info: console.info,
  warn: console.warn,
  error: console.error,
}

function log(level: LogLevel, msg: string, fields: Fields = {}) {
  const hasFields = Object.keys(fields).length > 0
  levelToConsole[level](
    `[frontend] ${msg}`,
    ...(hasFields ? [fields] : []),
  )
  levelToBackend[level](msg, fields).catch(() => {})
}

export function useLogger() {
  return {
    trace: (msg: string, fields?: Fields) => log('trace', msg, fields ?? {}),
    debug: (msg: string, fields?: Fields) => log('debug', msg, fields ?? {}),
    info: (msg: string, fields?: Fields) => log('info', msg, fields ?? {}),
    warn: (msg: string, fields?: Fields) => log('warn', msg, fields ?? {}),
    error: (msg: string, fields?: Fields) => log('error', msg, fields ?? {}),
  }
}
