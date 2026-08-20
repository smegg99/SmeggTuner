<template>
  <thead>
    <tr>
      <th class="head head--note">
        {{ t('record.table.note') }}
      </th>
      <th
        v-for="head in reedHeads"
        :key="head.key"
        class="head head--num"
      >
        {{ head.label }}
      </th>
      <th
        v-for="head in beatHeads"
        :key="head.key"
        class="head head--num"
      >
        {{ head.label }}
      </th>

      <th class="head head--kill" />
    </tr>
  </thead>
</template>

<script setup lang="ts">
// Goal/Error columns are dropped from the view; the rows still carry them.
interface Head { key: string, label: string }

defineProps<{
  reedHeads: Head[]
  beatHeads: Head[]
}>()

const { t } = useI18n()
</script>

<style scoped>
.head {
  background: rgb(var(--v-theme-chrome));
  color: rgb(var(--v-theme-ink2));
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.04em;
  padding: 5.6px 7.7px;
  position: sticky;
  text-transform: uppercase;
  top: 0;
  z-index: 2;
}

.head--note {
  text-align: left;
}

.head--kill {
  width: 33px;
}

.head--num {
  border-left: 1px solid rgb(var(--v-theme-lineSoft));
  text-align: right;
}
</style>
