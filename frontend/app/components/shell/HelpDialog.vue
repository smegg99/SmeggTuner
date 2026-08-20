<template>
  <v-dialog
    v-model="open"
    width="var(--dlg-lg)"
  >
    <UiDialogCard
      :title="t('about.helpTitle')"
      @close="open = false"
    >
      <div class="help">
        <p class="help__lead">
          {{ t('about.help.intro') }}
        </p>

        <h3 class="help__h">
          {{ t('about.help.rooms.title') }}
        </h3>
        <ul class="help__rooms">
          <li><b>{{ t('about.help.rooms.tuneName') }}</b> - {{ t('about.help.rooms.tune') }}</li>
          <li><b>{{ t('about.help.rooms.workshopName') }}</b> - {{ t('about.help.rooms.workshop') }}</li>
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
            <dt><kbd>{{ k.keys }}</kbd></dt>
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
            <dt><kbd>{{ k.keys }}</kbd></dt>
            <dd>{{ t(k.desc) }}</dd>
          </div>
        </dl>
      </div>
    </UiDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
const { t } = useI18n()

const open = defineModel<boolean>({ required: true })

// Kept in sync with useShortcuts by eye; this only describes the bindings, it does not create them.
const GLOBAL_KEYS = [
  { keys: '1 / 2 / 3', desc: 'about.help.everywhere.rooms' },
  { keys: 'A', desc: 'about.help.everywhere.manual' },
  { keys: ', / .', desc: 'about.help.everywhere.step' },
  { keys: 'H', desc: 'about.help.everywhere.hold' },
  { keys: 'N', desc: 'about.help.everywhere.sounds' },
  { keys: 'R', desc: 'about.help.everywhere.record' },
  { keys: 'Z', desc: 'about.help.everywhere.undo' },
]

// Kept in sync with FileTransport by eye; this only describes those keys.
const KEYS = [
  { keys: 'Space', desc: 'about.help.shortcuts.space' },
  { keys: 'Home / End', desc: 'about.help.shortcuts.homeEnd' },
  { keys: 'L', desc: 'about.help.shortcuts.loop' },
  { keys: 'M', desc: 'about.help.shortcuts.mute' },
  { keys: 'F / S', desc: 'about.help.shortcuts.fit' },
  { keys: '+ / −', desc: 'about.help.shortcuts.zoom' },
]
</script>

<style scoped>
.help {
  display: flex;
  flex-direction: column;
  gap: 11.2px;
  padding: 4.8px 0;
}

.help__lead {
  color: rgb(var(--v-theme-ink));
  font-size: 16.8px;
  line-height: 1.5;
  margin: 0;
}

.help__h {
  color: rgb(var(--v-theme-ink));
  font-size: 16px;
  font-weight: 700;
  margin: 8px 0 0;
}

.help__rooms {
  color: rgb(var(--v-theme-ink2));
  display: flex;
  flex-direction: column;
  font-size: 15.6px;
  gap: 5.6px;
  line-height: 1.5;
  margin: 0;
  padding-inline-start: 19.8px;
}

.help__rooms b {
  color: rgb(var(--v-theme-ink));
}

/* Two columns: one-per-row ran the dialog past the 640px minimum window. */
.help__keys {
  display: grid;
  gap: 7.2px 13.2px;
  grid-template-columns: auto 1fr auto 1fr;
  margin: 3.2px 0 0;
}

.help__key {
  display: contents;
}

.help__key dt {
  text-align: end;
}

.help__key dd {
  align-self: center;
  color: rgb(var(--v-theme-ink2));
  font-size: 15.6px;
  margin: 0;
}

kbd {
  background: rgb(var(--v-theme-chrome2));
  border: 1px solid rgb(var(--v-theme-line));
  border-radius: 2px;
  color: rgb(var(--v-theme-ink));
  display: inline-block;
  font-family: var(--font-mono);
  font-size: 13.6px;
  line-height: 1;
  padding: 5.6px 7.7px;
  white-space: nowrap;
}
</style>
