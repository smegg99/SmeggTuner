// nuxt.config.ts
import { execSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import path from 'node:path'
import vuetify, { transformAssetUrls } from 'vite-plugin-vuetify'
import wailsPlugin from '@wailsio/runtime/plugins/vite'
import svgLoader from 'vite-svg-loader'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const bindingsDir = path.resolve(__dirname, 'bindings')

const localeFiles = ['about.json', 'common.json', 'curve.json', 'file.json', 'record.json', 'report.json', 'routes.json', 'session.json', 'settings.json', 'title.json', 'tuner.json']

/*
 * The build stamps its own version and commit.
 *
 * A version typed into a source file is a version that is wrong the moment somebody
 * forgets to change it, and "which build is this" is the first question of every bug
 * report. If git is not there (a source tarball, a clean-room CI image) it says so
 * rather than inventing a plausible hash.
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

/*
 * The version is the nearest TAG, and nothing else.
 *
 * `git describe --always` falls back to the commit hash when there are no tags, so
 * the status bar read "v0920a8f 0920a8f": the same hash, printed twice, one of them
 * pretending to be a version. Untagged is a real state and it should say so.
 *
 * It says so by being EMPTY, not by the word "dev". The status bar shows the version beside
 * its keys only when there IS one (v-if), so an untagged build simply does not carry a badge
 * saying "dev" around forever - and the About dialog, which has room to be exact, spells the
 * empty case out as a development build. One build, said once, in the place with room for it.
 */
const VERSION = process.env.SMEGGTUNER_VERSION
  ?? stamp('git describe --tags --abbrev=0', '')
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
