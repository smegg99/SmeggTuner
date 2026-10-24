// app/plugins/inputKind.ts

// A select's inner input matches :focus-visible even after a click, so the field
// reads the last input kind from <html data-input> to ring keyboard focus only.
export default defineNuxtPlugin((app) => {
  const root = document.documentElement
  const pointer = () => {
    root.dataset.input = 'pointer'
  }
  const keyboard = () => {
    root.dataset.input = 'keyboard'
  }
  window.addEventListener('pointerdown', pointer, true)
  window.addEventListener('keydown', keyboard, true)

  const dispose = () => {
    window.removeEventListener('pointerdown', pointer, true)
    window.removeEventListener('keydown', keyboard, true)
  }
  app.vueApp.onUnmount(dispose)
  if (import.meta.hot) import.meta.hot.dispose(dispose)
})
