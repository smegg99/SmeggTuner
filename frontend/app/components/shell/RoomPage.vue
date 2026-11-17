<!-- app/components/shell/RoomPage.vue -->
<!-- A room that is one scrolling page: the guide and About, which a compact window shows as rooms where a
     wide one opens dialogs. A page's own key sits at its head, beside the first lines; links go under it. -->
<template>
  <UiPanel
    class="room tuner-panel"
    flush
  >
    <div
      class="room__body"
      :class="{ 'room__body--centred': centred }"
    >
      <div
        v-if="$slots.action"
        class="room__action"
      >
        <slot name="action" />
      </div>
      <slot />
    </div>
    <template
      v-if="$slots.footer"
      #footer
    >
      <slot name="footer" />
    </template>
  </UiPanel>
</template>

<script setup lang="ts">
defineProps<{
  /** a page shorter than the room sits in its middle */
  centred?: boolean
}>()
</script>

<style scoped>
.room__body {
  overflow-y: auto;
  overscroll-behavior: contain;
  scroll-padding-block: 24px;
  padding: 16px;
  font-size: var(--s99ab-font-size);
  line-height: 1.5;
}
.room__body--centred { display: grid; align-items: center; padding-block: 8px; }
/* The page's text runs beside the key and on under it. */
.room__action { float: right; margin: 0 0 var(--gap) 12px; }
/* The foot's links are worked with a finger: the control height, not a strip's dense one. */
.room :deep(.panel__foot) {
  --s99ab-control-h: inherit;
  padding: var(--gap);
}
</style>
