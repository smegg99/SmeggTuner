// 1.4 ships without the workshop (sessions, recording, reports): too much for a first run. Its code stays;
// every way into it, and every label naming it, is gated here. Set true to bring it back.
export const WORKSHOP = false

// The browser edition (SMEGGTUNER_EDITION=web): tuner and settings only, the engine in a worker. nuxt.config.ts defines it.
export const WEB: boolean = typeof __SMEGGTUNER_WEB__ !== 'undefined' && __SMEGGTUNER_WEB__

// Where the web edition sends people for the installed app: the downloads section of the site.
export const DOWNLOADS_URL = 'https://smeggtuner.com/#downloads'

// Browsers expose no system accent, so the web edition's default accent is the app icon's colour.
export const WEB_ACCENT = '#ff617c'
