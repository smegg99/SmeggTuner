# Translating SmeggTuner

Edit the YAML files in [locales/pl](locales/pl) for Polish or
[locales/en](locales/en) for English. Both languages have the same keys and
section order. You only need a text editor to change translations.

1. Find the message in the English file, then find the same key in your language.
2. Translate the text after `value:` or inside `plural.forms`.
3. Keep keys, `{placeholders}`, `argument`, and placeholder types unchanged.
   `#` comments mark sections; they are not displayed or translated.

Send the edited YAML files if you cannot run the checks below. A maintainer can
generate the application files for you. Do not edit `frontend/i18n/generated`
or `backend/common/i18n`: those files are generated.

## Where to look

| File | Contents |
| --- | --- |
| `common.yaml` | Shared buttons, units, navigation, window titles, About and Help |
| `tuner.yaml` | Live readings, WAV playback, instrument calibration and browser messages |
| `settings.yaml` | Microphone, appearance, language and sound-card calibration |
| `session.yaml` | Sessions, accordion definitions and target curves |
| `record.yaml` | Recording controls, reading tables and the report dialog |
| `report.yaml` | Printed reports and export errors |
| `log.yaml` | Diagnostic messages |
| `shortcut.yaml` | Keyboard key labels |

The Workshop is disabled in 1.4, but its session, recording and report code
still uses these catalogs. Keep those translations.

## Short text and sentences

Most labels fit on one line:

```yaml
cancel: {value: "Cancel"}
```

Translate only `Cancel`. Keep the quotes so punctuation such as `:` and `#`
is treated as text. Inside double quotes, write a literal quote as `\"`.
Use spaces for indentation, not tabs.

Long messages use `>-`. The indented lines become one paragraph, with spaces
between lines and no trailing newline:

```yaml
emptyHint:
  value: >-
    Create a session to save measurements,
    or use the tuner without one.
```

Use natural phrasing. Buttons and table headings need short translations;
tooltips and explanations have more room. Keep product names, reed symbols
such as `L` and `M`, and physical units such as `Hz` unchanged.

## Placeholders

The app inserts a value where a name appears in braces. You may move it to
fit your sentence, but do not translate the name or remove the braces:

```yaml
card.updated:
  value: "Updated {when}"
  placeholders:
    when: {type: string}
```

Keep the `placeholders` block. The generator checks that the sentence and
its declarations agree. In this example, `{when}` is an already formatted
date and time.

## Plurals

Translate each form separately. `argument` names the number used to choose
the form; other numbers in the message do not select it.

English uses `one` for 1 and `other` for 0, 2, 3, and so on. Polish uses:

| Form | Examples |
| --- | --- |
| `one` | 1 |
| `few` | 2, 3, 4, 22, 23, 24 |
| `many` | 0, 5, 12, 14, 21, 25 |
| `other` | Fractions; keep this required form even for integer counts |

See [English readings](locales/en/session.yaml) and
[Polish readings](locales/pl/session.yaml), under `card.readings`, for a
complete example. Keep every form present in your language's catalog.
Do not copy English's two-form structure over Polish's four forms.

## Tuning terminology

| Term | Meaning here |
| --- | --- |
| Reed | One vibrating metal tongue that produces a tone |
| Reed rank / bank | A set of reeds, identified by symbols such as L or M |
| Register | A switch that selects which reed ranks sound together |
| Beat / beating | The pulsation caused by two nearby pitches, not musical rhythm |
| Reading | A measured pitch or beat value |
| Pass | A named set of readings within a session |
| Target / goal curve | Desired deviations or beatings across the note range |
| File toward zero | Remove metal with a file to adjust a reed, not save a computer file |
| Reference unit | A deviation of 0.06% of the note frequency; distinct from a cent |

## Generate and check

From the repository root, with the project's Go, Node.js, Python and `just`
prerequisites installed:

```sh
just locales
just check-locales
```

The first command regenerates translations only. The second checks missing
keys, placeholders, plural forms, stale generated files and unused messages.
It does not build the desktop app or regenerate icons.

When reviewing in the app, switch languages in Settings and check long text,
buttons, and counts such as 0, 1, 2, 5, 12 and 22. Translation changes need a
frontend rebuild or a running development server to appear.

## Adding a language

Copy the eight English YAML files into `locales/<language-code>` and translate
them. Adapt plural forms to that language rather than retaining English's
forms. Ask a maintainer to wire the language into the app before testing it:
adding a folder alone does not enable it.

The maintainer must update the locale list in [config.cue](config.cue), the
language setting in [the configuration schema](../backend/common/config/config.cue),
the locale registration in [Nuxt](../frontend/nuxt.config.ts), and the language
selection in [theme sync](../frontend/app/composables/useThemeSync.ts) and
[browser startup](../frontend/app/web/bootAppearance.ts). Also extend
[date formatting and shared messages](../frontend/i18n/i18n.config.ts) and
[report checks](../scripts/check-locales.mjs), then regenerate and test.
