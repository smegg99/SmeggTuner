<template>
  <div class="sheet">
    <slot />
  </div>
</template>

<script setup lang="ts">
// Not a v-list: its Material idioms (ripple, dividers) fight the flat look.
</script>

<style scoped>
.sheet {
  background: rgb(var(--v-theme-chrome));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  display: flex;
  flex-direction: column;
  /* cap width so a long device name can't outgrow the window; rows ellipsise inside */
  max-width: 42cqw;
  min-width: 14cqw;
  overflow: hidden;
  padding-block: 0.4cqh;
}
</style>
