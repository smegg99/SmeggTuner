<template>
  <div class="ident">
    <span class="ident__pic">
      <img
        v-if="src"
        class="ident__img"
        :src="src"
        :alt="instrument"
      >
      <v-icon
        v-else
        icon="mdi-folder-music-outline"
        class="ident__none"
      />
    </span>

    <span class="ident__m">
      <span class="ident__n">{{ session?.name }}</span>
      <span class="ident__s">{{ instrument || t('session.card.unnamedInstrument') }}</span>
    </span>

    <span class="ident__sp" />

    <UiToolGroup
      v-model="view"
      class="ident__views"
      :items="VIEWS"
      :label="t('session.title')"
    />

    <UiChip class="ident__chip">
      {{ t('session.card.a4', { hz: (session?.a4 ?? 0).toFixed(1) }) }}
    </UiChip>

    <UiChip
      v-for="r in registers"
      :key="r.name"
      mono
      class="ident__chip"
    >
      {{ r.name }}
    </UiChip>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ToolItem } from '~/components/ui/ToolGroup.vue'
import type { SessionDTO } from '~~bindings/smegg.me/smeggtuner/services/session/models.js'

const props = defineProps<{
  session: SessionDTO | null
  src: string
  busy: boolean
}>()

const view = defineModel<'recorded' | 'curve'>('view', { required: true })

const { t } = useI18n()

const instrument = computed(() => {
  const i = props.session?.instrument
  return i?.name || [i?.make, i?.model].filter(Boolean).join(' ')
})

const registers = computed(() => props.session?.instrument?.registers ?? [])

const VIEWS = computed<ToolItem[]>(() => [
  { value: 'recorded', label: t('sessionPage.tabs.recorded'), icon: 'mdi-table' },
  { value: 'curve', label: t('sessionPage.tabs.curve'), icon: 'mdi-chart-bell-curve-cumulative' },
])
</script>

<style scoped>
.ident {
  align-items: center;
  border-bottom: 1px solid rgb(var(--v-theme-lineSoft));
  display: flex;
  gap: 0.7cqw;
  padding: 0.8cqh 0.9cqw;
}

.ident__pic {
  align-items: center;
  aspect-ratio: 4 / 3;
  background: rgb(var(--v-theme-sunk));
  border: 1px solid rgb(var(--v-theme-lineSoft));
  border-radius: 2px;
  display: flex;
  flex: 0 0 auto;
  height: 5cqh;
  justify-content: center;
  overflow: hidden;
}

.ident__img {
  height: 100%;
  object-fit: cover;
  width: 100%;
}

.ident__none {
  color: rgb(var(--v-theme-ink3));
  font-size: 2.4cqh;
}

.ident__m {
  display: flex;
  flex-direction: column;
  gap: 0.1cqh;
  min-width: 0;
}

.ident__n {
  font-size: 1.8cqh;
  font-weight: 600;
}

.ident__s {
  color: rgb(var(--v-theme-ink3));
  font-size: 1.45cqh;
}

.ident__sp {
  flex: 1 1 auto;
}

.ident__views {
  flex: 0 0 auto;
}

/* Metrics copied from UiToolKey .key so these static labels match the view keys' height. */
.ident__chip {
  align-items: center;
  display: inline-flex;
  font-size: 1.75cqh;
  font-weight: 500;
  height: 3.9cqh;
  justify-content: center;
  line-height: 1;
  padding: 0 0.85cqw;
}
</style>
