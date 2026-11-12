<!-- frontend/app/components/shell/HelpDialog.vue -->
<template>
  <ShellDocumentDialog
    v-model="open"
    kind="guide"
  />
</template>

<script setup lang="ts">
const open = defineModel<boolean>({ required: true })
</script>
