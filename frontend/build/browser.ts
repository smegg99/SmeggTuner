// build/browser.ts
import type { Plugin } from 'postcss'
import { flattenRoot } from '../app/plugins/browser/layers'

export default function browserCss(): Plugin {
  return {
    postcssPlugin: 'smeggtuner-browser',
    OnceExit(root) {
      flattenRoot(root)
    },
  }
}
