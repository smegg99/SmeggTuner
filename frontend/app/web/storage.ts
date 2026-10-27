// app/web/storage.ts
// The engine's config lives in this browser only; storage can be blocked (private mode), so both directions fail quietly.
const KEY = 'smeggtuner.web.config'

export function loadConfig(): string {
  try {
    return localStorage.getItem(KEY) ?? ''
  }
  catch {
    return ''
  }
}

export function saveConfig(json: string): void {
  try {
    localStorage.setItem(KEY, json)
  }
  catch {
    // Settings then last for this visit only.
  }
}

// storedDeviceId is the microphone the settings saved (config audio.device_id), "" for the system default.
export function storedDeviceId(json: string): string {
  try {
    const id = (JSON.parse(json) as { audio?: { device_id?: unknown } })?.audio?.device_id
    return typeof id === 'string' ? id : ''
  }
  catch {
    return ''
  }
}

// storedPreferences is what the boot screen needs before the engine exists: the saved theme and language.
export function storedPreferences(json: string): { theme: string, language: string } {
  try {
    const p = (JSON.parse(json) as { preferences?: { theme?: unknown, language?: unknown } })?.preferences
    return {
      theme: typeof p?.theme === 'string' ? p.theme : 'auto',
      language: typeof p?.language === 'string' ? p.language : '',
    }
  }
  catch {
    return { theme: 'auto', language: '' }
  }
}
