<template>
  <v-menu>
    <template #activator="{ props: menu }">
      <v-btn
        v-bind="menu"
        :disabled="busy"
        :title="t('curve.menu.label')"
        :aria-label="t('curve.menu.label')"
        append-icon="mdi-menu-down"
      >
        {{ t('curve.menu.label') }}
      </v-btn>
    </template>

    <v-list class="tuner-menu">
      <div>
        <v-list-subheader>{{ t('curve.menu.ends') }}</v-list-subheader>
        <v-list-item
          :active="curve?.extrapolateLeft ?? false"
          :append-icon="(curve?.extrapolateLeft ?? false) ? 'mdi-check' : undefined"
          :title="t('curve.menu.extendLow')"
          @click="setExtrapolateLeft(!curve?.extrapolateLeft)"
        />
        <v-list-item
          :active="curve?.extrapolateRight ?? false"
          :append-icon="(curve?.extrapolateRight ?? false) ? 'mdi-check' : undefined"
          :title="t('curve.menu.extendHigh')"
          @click="setExtrapolateRight(!curve?.extrapolateRight)"
        />
      </div>

      <div>
        <v-list-subheader>{{ t('curve.menu.replace') }}</v-list-subheader>
        <v-list-item
          :disabled="!readings"
          :title="t('curve.menu.fit')"
          @click="ask(t('curve.menu.fitBody', { n: readings }, readings), fitCurve)"
        />
        <v-list-item
          :title="t('curve.menu.fromFile')"
          @click="ask(t('curve.menu.replaceBody'), importCurveFile)"
        />
      </div>

      <!-- Every session holding a beat list; a long shelf scrolls here rather than off the window. -->
      <div v-if="donors.length">
        <v-list-subheader>{{ t('curve.menu.fromSession') }}</v-list-subheader>
        <div class="donors">
          <v-list-item
            v-for="s in donors"
            :key="s.id"
            :title="s.name"
            @click="ask(t('curve.menu.replaceBody'), () => importCurve(s.id))"
          />
        </div>
      </div>
    </v-list>
  </v-menu>

  <v-dialog
    v-model="confirming"
    :aria-label="t('curve.menu.replaceTitle')"
    width="var(--dlg-sm)"
  >
    <UiDialogCard
      :title="t('curve.menu.replaceTitle')"
      @close="confirming = false"
    >
      <p class="dialog-warn">
        {{ pendingBody }}
      </p>
      <template #actions>
        <v-btn
          @click="confirming = false"
        >
          {{ t('common.cancel') }}
        </v-btn><v-btn
          color="error"
          @click="confirm"
        >
          {{ t('curve.menu.replaceConfirm') }}
        </v-btn>
      </template>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useSessions } from '~/composables/useSessions'

// Settings of the whole beat list, and the ways to replace it: a fit to this pass, or another session's list.
const { t } = useI18n()
const {
  active, list, busy, readings,
  setExtrapolateLeft, setExtrapolateRight, fitCurve, importCurve, importCurveFile,
} = useSessions()

const curve = computed(() => active.value?.curve ?? null)

// Summary.hasCurve is exactly what ImportCurve accepts.
const donors = computed(() => list.value.filter(s => s.hasCurve && s.id !== active.value?.id))

const confirming = ref(false)
const pendingBody = ref('')
let pending: (() => Promise<unknown>) | null = null

function ask(body: string, run: () => Promise<unknown>) {
  pendingBody.value = body
  pending = run
  confirming.value = true
}

async function confirm() {
  confirming.value = false
  const run = pending
  pending = null
  await run?.()
}
</script>

<style scoped>
.donors {
  max-height: 192.5px;
  overflow-y: auto;
}
</style>
