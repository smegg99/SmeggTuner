<template>
  <!-- One tooltip for the whole window, styled as the layer's tooltips are. It follows the pointer
       from control to control, disabled ones included, so no control needs its own. -->
  <v-tooltip
    v-if="host"
    :model-value="open"
    :activator="host"
    :text="text"
    :open-on-hover="false"
    :open-on-focus="false"
    :open-on-click="false"
    :transition="false"
    location="bottom"
  />
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, shallowRef } from 'vue'
import { tipHost, tipText } from '~/utils/tooltip'

// How long the pointer rests on a control before its tip shows, as a native tooltip waits.
const DELAY_MS = 450

const host = shallowRef<HTMLElement | null>(null)
const text = ref('')
const open = ref(false)
let timer: ReturnType<typeof setTimeout> | undefined
let watcher: MutationObserver | undefined

// A title the app changes while its tip shows (Start becoming Stop) moves into the tip at once.
function follow(el: HTMLElement) {
  watcher?.disconnect()
  watcher = new MutationObserver(() => {
    const next = tipText(el)
    if (next) text.value = next
  })
  watcher.observe(el, { attributes: true, attributeFilter: ['title', 'aria-label'] })
}

function hide() {
  clearTimeout(timer)
  open.value = false
}

function onMove(event: PointerEvent) {
  const el = tipHost(event.target, event.clientX, event.clientY)
  if (el === host.value) return
  hide()
  watcher?.disconnect()
  host.value = el
  if (!el) return
  const tip = tipText(el)
  if (!tip) {
    host.value = null
    return
  }
  text.value = tip
  follow(el)
  timer = setTimeout(() => {
    open.value = true
  }, DELAY_MS)
}

function onLeave(event: PointerEvent) {
  if (!event.relatedTarget) {
    hide()
    host.value = null
  }
}

onMounted(() => {
  document.addEventListener('pointermove', onMove, { passive: true, capture: true })
  document.addEventListener('pointerout', onLeave, { passive: true, capture: true })
  for (const name of ['pointerdown', 'wheel', 'keydown'] as const) {
    document.addEventListener(name, hide, { passive: true, capture: true })
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('pointermove', onMove, { capture: true })
  document.removeEventListener('pointerout', onLeave, { capture: true })
  for (const name of ['pointerdown', 'wheel', 'keydown'] as const) {
    document.removeEventListener(name, hide, { capture: true })
  }
  watcher?.disconnect()
  clearTimeout(timer)
})
</script>
