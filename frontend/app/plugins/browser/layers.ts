// app/plugins/browser/layers.ts
import postcss from 'postcss'
import type { Root } from 'postcss'
import cascadeLayers from '@csstools/postcss-cascade-layers'
import selectorParser from 'postcss-selector-parser'
import { compare, selectorSpecificity } from '@csstools/selector-specificity'

export const THEME_STYLESHEET_ID = 'vuetify-theme-stylesheet'

// Static files and the live theme must use the same layer order and specificity stride.
const order = `
@layer vuetify-core, vuetify-components, vuetify-overrides, vuetify-utilities, vuetify-final;
@layer vuetify-core.reset, vuetify-core.base;
@layer vuetify-utilities.theme-base, vuetify-utilities.typography, vuetify-utilities.helpers, vuetify-utilities.theme-background, vuetify-utilities.theme-foreground;
@layer vuetify-final.transitions, vuetify-final.trumps;
`
const names = new Set(order.match(/vuetify-[\w.-]+/g))
const bound = { a: 2, b: 0, c: 0 }

export function flattenLayers(css: string): string {
  const root = postcss.parse(css, { from: undefined })
  flattenRoot(root)
  return root.toString()
}

export function flattenRoot(root: Root): void {
  root.walkAtRules('layer', (rule) => {
    const parents: string[] = []
    for (let parent = rule.parent; parent && parent.type !== 'root'; parent = parent.parent) {
      if (parent.type === 'atrule' && parent.name === 'layer') parents.unshift(parent.params)
    }
    for (const name of rule.params.split(',')) {
      const full = [...parents, name.trim()].join('.')
      if (!names.has(full)) throw rule.error(`Unregistered CSS layer: ${full}`)
    }
  })
  root.walkRules((rule) => {
    if (rule.parent?.type === 'atrule' && /keyframes$/i.test(rule.parent.name)) return
    const selectors = selectorParser().astSync(rule.selector)
    for (const selector of selectors.nodes) {
      if (compare(selectorSpecificity(selector), bound) > 0) {
        throw rule.error(`CSS selector exceeds the shared layer specificity bound: ${rule.selector}`)
      }
    }
  })

  const sentinel = postcss.rule({ selector: '#smegg-layer-bound#smegg-layer-stride' })
  sentinel.append({ prop: '--smegg-layer-bound', value: '0' })
  root.prepend(sentinel)
  root.prepend(postcss.parse(order, { from: undefined }).nodes)
  const result = postcss([cascadeLayers()]).process(root, { from: undefined, map: false })
  result.sync()
  const warnings = result.warnings()
  if (warnings.length) throw new Error(warnings.map(warning => warning.text).join('\n'))
  sentinel.remove()
}
