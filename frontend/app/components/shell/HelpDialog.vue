<!-- frontend/app/components/shell/HelpDialog.vue -->
<template>
  <v-dialog
    v-model="open"
    :aria-label="t('about.helpTitle')"
    width="var(--dlg-lg)"
  >
    <UiDialogCard
      :title="t('about.helpTitle')"
      @close="open = false"
    >
      <div class="help">
        <p class="help__lead">
          {{ t(WORKSHOP ? 'about.help.intro' : 'about.help.introTuner') }}
        </p>

        <h3 class="help__h">
          {{ t('about.help.rooms.title') }}
        </h3>
        <ul class="help__rooms">
          <li><b>{{ t('about.help.rooms.tuneName') }}</b> - {{ t('about.help.rooms.tune') }}</li>
          <li v-if="WORKSHOP">
            <b>{{ t('about.help.rooms.workshopName') }}</b> - {{ t('about.help.rooms.workshop') }}
          </li>
          <li><b>{{ t('about.help.rooms.settingsName') }}</b> - {{ t('about.help.rooms.settings') }}</li>
        </ul>

        <h3 class="help__h">
          {{ t('about.help.everywhere.title') }}
        </h3>
        <dl class="help__keys">
          <div
            v-for="k in GLOBAL_KEYS"
            :key="k.desc"
            class="help__key"
          >
            <dt>
              <template
                v-for="(key, i) in k.keys"
                :key="key"
              >
                <span
                  v-if="i"
                  class="help__join"
                >{{ k.or ? t('shortcut.or') : '/' }}</span>
                <UiKbd :keys="key" />
              </template>
            </dt>
            <dd>{{ t(k.desc) }}</dd>
          </div>
        </dl>

        <h3 class="help__h">
          {{ t('about.help.shortcuts.title') }}
        </h3>
        <dl class="help__keys">
          <div
            v-for="k in KEYS"
            :key="k.desc"
            class="help__key"
          >
            <dt>
              <template
                v-for="(key, i) in k.keys"
                :key="key"
              >
                <span
                  v-if="i"
                  class="help__join"
                >{{ k.or ? t('shortcut.or') : '/' }}</span>
                <UiKbd :keys="key" />
              </template>
            </dt>
            <dd>{{ t(k.desc) }}</dd>
          </div>
        </dl>
      </div>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { WORKSHOP } from '~/utils/features'

const { t } = useI18n()

const open = defineModel<boolean>({ required: true })

// One cap per key in UiKbd's registry form; `or` joins keys that do the same, a slash a pair of opposites.
interface KeyRow { keys: string[], or?: boolean, desc: string }

// Kept in sync with useShortcuts by eye; this only describes the bindings, it does not create them.
// Recording belongs to the workshop's sessions, so its keys are described only with it.
const RECORDING = ['about.help.everywhere.record', 'about.help.everywhere.takeNow', 'about.help.everywhere.undo']
const GLOBAL_KEYS = computed<KeyRow[]>(() => [
  { keys: WORKSHOP ? ['1', '2', '3'] : ['1', '2'], desc: WORKSHOP ? 'about.help.everywhere.rooms' : 'about.help.everywhere.roomsTuner' },
  { keys: ['a'], desc: 'about.help.everywhere.manual' },
  { keys: [',', '.'], desc: 'about.help.everywhere.step' },
  { keys: ['ArrowLeft', 'ArrowRight'], desc: 'about.help.everywhere.arrows' },
  { keys: ['h', 'space'], or: true, desc: 'about.help.everywhere.hold' },
  { keys: ['r'], desc: 'about.help.everywhere.record' },
  { keys: ['enter'], desc: 'about.help.everywhere.takeNow' },
  { keys: ['z', 'backspace'], or: true, desc: 'about.help.everywhere.undo' },
].filter(k => WORKSHOP || !RECORDING.includes(k.desc)))

// Kept in sync with useShortcuts' file keys by eye; this only describes those keys.
const KEYS: KeyRow[] = [
  { keys: ['space'], desc: 'about.help.shortcuts.space' },
  { keys: ['home', 'end'], desc: 'about.help.shortcuts.homeEnd' },
  { keys: ['l'], desc: 'about.help.shortcuts.loop' },
  { keys: ['m'], desc: 'about.help.shortcuts.mute' },
  { keys: ['f', 's'], desc: 'about.help.shortcuts.fit' },
  { keys: ['plus', 'minus'], desc: 'about.help.shortcuts.zoom' },
]
</script>

<style scoped>
.help {
  display: flex;
  flex-direction: column;
  gap: var(--gap);
}

.help__lead {
  color: rgb(var(--v-theme-ink));
  font-size: 13px;
  line-height: 1.5;
  margin: 0;
}

.help__h {
  color: rgb(var(--v-theme-ink));
  font-size: 14px;
  font-weight: 700;
  margin: var(--gap) 0 0;
}

.help__rooms {
  color: rgb(var(--v-theme-ink2));
  display: flex;
  flex-direction: column;
  font-size: 12px;
  gap: var(--gap);
  line-height: 1.5;
  margin: 0;
  padding-inline-start: var(--gap-wide);
}

.help__rooms b {
  color: rgb(var(--v-theme-ink));
}

/* Two columns: one-per-row ran the dialog past the 640px minimum window. */
.help__keys {
  display: grid;
  gap: var(--gap) var(--gap);
  grid-template-columns: auto 1fr auto 1fr;
  margin: 3px 0 0;
}

@media (max-width: 720px) {
  .help__keys { grid-template-columns: auto 1fr; }
}

.help__key {
  display: contents;
}

.help__key dd {
  align-self: center;
  color: rgb(var(--v-theme-ink2));
  font-size: 12px;
  margin: 0;
}

.help__key dt {
  align-items: center;
  display: flex;
  gap: 4px;
  justify-content: flex-end;
}

.help__join {
  color: rgb(var(--v-theme-ink3));
  font-size: 11px;
}
</style>
