// nuxt.config.ts
import { execSync } from 'node:child_process'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import path from 'node:path'
import vuetify, { transformAssetUrls } from 'vite-plugin-vuetify'
import wailsPlugin from '@wailsio/runtime/plugins/vite'
import svgLoader from 'vite-svg-loader'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const bindingsDir = path.resolve(__dirname, 'bindings')

const localeFiles = ['about.json', 'common.json', 'curve.json', 'file.json', 'record.json', 'report.json', 'routes.json', 'session.json', 'settings.json', 'title.json', 'tuner.json']

/*
 * VERSION is the manually controlled release source of truth. The commit stays automatic;
 * if git is unavailable (for example in a source tarball), leave it empty instead of
 * inventing a plausible hash.
 */
function stamp(command: string, fallback: string): string {
  try {
    const out = execSync(command, { stdio: ['ignore', 'pipe', 'ignore'] }).toString().trim()
    return out || fallback
  }
  catch {
    return fallback
  }
}

const releaseVersion = readFileSync(path.resolve(__dirname, '../VERSION'), 'utf8').trim()
if (!/^\d+\.\d+\.\d+$/.test(releaseVersion)) {
  throw new Error(`invalid VERSION "${releaseVersion}": expected MAJOR.MINOR.PATCH`)
}

const VERSION = `v${releaseVersion}`
const COMMIT = process.env.SMEGGTUNER_COMMIT
  ?? stamp('git rev-parse --short HEAD', '')

export default defineNuxtConfig({
  modules: [
    '@unocss/nuxt',
    '@nuxtjs/i18n',
    '@nuxt/eslint',
  ],
  ssr: false,
  devtools: { enabled: true },
  app: {
    pageTransition: false,
    layoutTransition: false,
    head: {
      style: [
        {
          // Prevent flash before Vuetify/JS loads. The actual theme is applied
          // from the backend config once the app mounts.
          innerHTML: `html, body { background-color: #000000; }`,
        },
      ],
    },
  },
  css: ['~/assets/css/fonts.css', '~/assets/css/scrollbar.css', '~/assets/css/vuetify.scss', '~/assets/css/overlays.css'],
  runtimeConfig: {
    public: {
      version: VERSION,
      commit: COMMIT,
    },
  },
  alias: {
    '~~bindings': bindingsDir,
  },
  build: {
    transpile: ['vuetify'],
  },
  // Wails' dev proxy dials tcp4 127.0.0.1. Left to itself Nuxt binds the IPv6
  // loopback only ([::1]), the proxy is refused, and the app shows a white
  // window with one "connection refused" line to explain it.
  devServer: { host: '127.0.0.1', port: 9245 },
  features: {
    inlineStyles: false,
  },
  experimental: {
    payloadExtraction: true,
  },
  compatibilityDate: '2025-07-15',
  nitro: {
    compressPublicAssets: true,
  },
  vite: {
    plugins: [
      wailsPlugin(bindingsDir) as any,
      vuetify({
        autoImport: true,
        styles: { configFile: 'assets/css/vuetify.scss' },
      }),
      svgLoader({
        defaultImport: 'component',
      }),
    ],
    vue: {
      template: {
        transformAssetUrls,
      },
    },
    optimizeDeps: {
      include: [
        '@wailsio/runtime',
      ],
    },
  },
  eslint: {
    config: {
      stylistic: true,
    },
  },
  i18n: {
    strategy: 'no_prefix',
    defaultLocale: 'en',
    langDir: 'locales',
    locales: [
      {
        code: 'en',
        name: 'English',
        language: 'en-GB',
        files: localeFiles.map(f => `en/${f}`),
      },
      {
        code: 'pl',
        name: 'Polski',
        language: 'pl-PL',
        files: localeFiles.map(f => `pl/${f}`),
      },
    ],
    compilation: {
      strictMessage: false,
      escapeHtml: false,
    },
    detectBrowserLanguage: false,
    vueI18n: 'i18n.config.ts',
  },
})
