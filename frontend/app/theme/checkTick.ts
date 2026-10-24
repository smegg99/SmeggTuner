// app/theme/checkTick.ts
import { defineComponent, h, markRaw } from 'vue'

// The instrument's checkbox tick: an inline stroked path, so it lands on the same
// sub-pixel position the original drew it at. A CSS mask is snapped to whole pixels.
export const CheckTick = markRaw(defineComponent({
  name: 'CheckTick',
  setup: () => () => h('svg', { viewBox: '0 0 16 16', fill: 'none', class: 'check-tick' }, [
    h('path', {
      'd': 'M3.5 8.5 L6.5 11.5 L12.5 4.5',
      'stroke': 'currentColor',
      'stroke-width': '2.2',
      'stroke-linecap': 'round',
      'stroke-linejoin': 'round',
    }),
  ]),
}))
