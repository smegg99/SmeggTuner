// vitest.config.ts
import { fileURLToPath } from 'node:url'
import path from 'node:path'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vitest/config'

const dir = path.dirname(fileURLToPath(import.meta.url))

/*
 * The frontend had no tests at all. It has some now, and they are pointed at the
 * two things worth testing here: the pure geometry, and the one rule the app must
 * never break - that a reed the engine could not measure is not printed.
 *
 * Component tests mount the presentational half over a known model. The Wails
 * bindings are not mocked and never should be: what they would be asserting is
 * that a mock returns what the mock was told to return.
 */
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '~': path.resolve(dir, 'app'),
      '~~bindings': path.resolve(dir, 'bindings'),
    },
  },
  test: {
    environment: 'happy-dom',
    include: ['tests/**/*.spec.ts'],
  },
})
