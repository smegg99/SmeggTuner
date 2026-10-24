// nuxt.config.ts
import { execSync } from 'node:child_process'
import { readFileSync, readdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import path from 'node:path'
import vuetify, { transformAssetUrls } from 'vite-plugin-vuetify'
import wailsPlugin from '@wailsio/runtime/plugins/vite'
import svgLoader from 'vite-svg-loader'
import browserCss from './build/browser'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const bindingsDir = path.resolve(__dirname, 'bindings')

const localeFiles = readdirSync(path.resolve(__dirname, 'i18n/generated/en')).filter(file => file.endsWith('.json')).sort()

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
const androidConfig = JSON.parse(readFileSync(path.resolve(__dirname, '../build/android-support/config.json'), 'utf8'))
const COMMIT = process.env.SMEGGTUNER_COMMIT
  ?? stamp('git rev-parse --short HEAD', '')

// SMEGGTUNER_EDITION=web builds the browser edition: its own root and browsers, the engine served from /engine/.
const web = process.env.SMEGGTUNER_EDITION === 'web'
const browserTarget: string | string[] = web
  ? ['safari16.4', 'chrome111', 'firefox115', 'edge111']
  : `chrome${androidConfig.webViewMinMajor}`
// The web edition lays out at the device's own width: its compact layout is the phone one, not a scaled desktop window.
const desktopOnlyPlugins = ['/plugins/editKeys.ts', '/plugins/01.viewport.client.ts']

export default defineNuxtConfig({
  extends: ['@smegg99/s99ab'],
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
  css: ['~/assets/styles/app.scss', '~/assets/styles/instrument.scss'],
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
    entryImportMap: false,
  },
  compatibilityDate: '2025-07-15',
  nitro: {
    compressPublicAssets: true,
    ...(web && {
      output: { dir: path.resolve(__dirname, '.output-web') },
      // web-public is the manifest, service worker and runtime hook; scripts/web.sh builds the engine and icons into .web-assets.
      publicAssets: [
        { dir: path.resolve(__dirname, 'web-public'), baseURL: '/' },
        { dir: path.resolve(__dirname, '.web-assets'), baseURL: '/' },
      ],
    }),
  },
  vite: {
    css: {
      postcss: { plugins: [browserCss()] },
    },
    build: {
      target: browserTarget,
      cssTarget: browserTarget,
    },
    worker: { format: 'es' },
    define: { __SMEGGTUNER_WEB__: JSON.stringify(web) },
    // The repository's assets/ holds the artwork masters: mascot and key caps.
    server: { fs: { allow: [path.resolve(__dirname, '../assets')] } },
    resolve: { dedupe: ['vue', 'vuetify', 'vue-i18n', 'pinia'] },
    plugins: [
      wailsPlugin(bindingsDir),
      vuetify({
        autoImport: true,
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
  typescript: {
    tsConfig: { exclude: [path.resolve(__dirname, '../../S99AB/web/tests/**'), path.resolve(__dirname, '../../S99AB/node_modules/**')] },
  },
  hooks: {
    'app:resolve'(app) {
      if (!web) return
      app.mainComponent = path.resolve(__dirname, 'app/web/WebApp.vue')
      app.plugins = app.plugins.filter(p => !desktopOnlyPlugins.some(name => p.src.endsWith(name)))
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
    langDir: 'generated',
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
