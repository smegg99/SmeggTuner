<p align="center">
  <img src="assets/smeggtuner-app-icon.svg" alt="SmeggTuner" width="140">
</p>

<p align="center">
  <b>English</b> | <a href="README.pl.md">Polski</a>
  <br>
  <a href="https://smeggtuner.com">smeggtuner.com</a>
</p>

# SmeggTuner

A free, open source tuner for accordions. Play a note and it shows every reed of the register on
its own, plus the beats between them, so you can see which reed to file and by how much.

It works the way Dirk's professional Accordion Tuner does, and it is free. It runs on Linux and
Windows, from a microphone or from a WAV recording.

<p align="center">
  <img src="assets/screenshots/smeggtuner-recording-analysis.png" alt="Three reeds of C4 measured from a recording" width="760">
</p>

> **Beta.** Testing on real accordions and sound cards is still limited. Check the readings against
> a tuner you trust before you file a reed.

## Features

- Measures up to three reeds of one note at the same time, and the beat between each pair.
- Three modes: Tones (the reeds of one note), Octaves, and Chords.
- Readings in Dirk's units (0.06 % of the note, about one cent) or in cents. Beats in Hz or units.
- Note detection can be Fast, Balanced or Steady. Steady behaves exactly like Dirk's tuner.
- Sound card calibration, kept per microphone. It takes about a minute and needs no sound.
- Automatic input volume, and background noise learned once per microphone. A key in the
  top bar relearns it.
- Hold stable readings: the snowflake key holds each locked reading until the sound stops.
- Opens WAV recordings with play, seek, loop and zoom, so you can tune without the accordion.
- Click the note strip to hear the note.
- A4 from 430 to 450 Hz, transpose, a pinned note you step with the arrow keys, tolerances, and
  four note namings, including Polish (C Cis D ... H).
- Mains hum filter for 50 Hz or 60 Hz.
- English and Polish, light and dark theme, tray icon, keyboard shortcuts.
- If it fails to start, crashes, stops responding or is killed, it says so in a window with a report
  you can copy, save or email, including the recent log. A log file is kept by default.

The Workshop from earlier versions is turned off in 1.4. It held the accordion library, sessions,
recorded takes, goal beat curves and printed reports. The code is still here, and it will come
back in a later release.

## SmeggTuner and Dirk's Tuner

| | Dirk's professional Accordion Tuner | SmeggTuner 1.4 |
|---|---|---|
| Price | 298 EUR, plus 148 EUR for the record and report module | Free |
| Source | Closed | Open, MIT |
| Systems | Windows, macOS | Linux (x86_64, ARM64), Windows; experimental browser edition |
| Reeds measured together | Up to 3 | Up to 3 |
| Tones, Octaves, Chords | Yes | Yes |
| Units or cents, beats in Hz | Yes | Yes |
| Sound card calibration | Yes, per device | Yes, per microphone |
| Automatic input volume | Yes | Yes |
| Detection speed | One behaviour | Fast, Balanced, Steady (Steady is Dirk's) |
| Analyse a WAV file | Not in the menus | Yes, with seek, loop and zoom |
| Record and report | Paid module | Not in 1.4, the Workshop returns later |

Dirk's prices are from [dirksprojects.com](https://www.dirksprojects.com/) in October 2026.

## Roadmap

These are plans, not promises.

- [ ] Bring the Workshop back: accordion library, sessions, recording, goal curves and reports.
- [ ] Release the Android build. It builds and runs in the emulator, but is not released yet.
- [ ] macOS and iOS. I have no hardware to test them on.
- [ ] Host the browser edition at app.smeggtuner.com.
- [ ] More testing on real accordions, microphones and sound cards.

## Download

Get the latest build from [smeggtuner.com](https://smeggtuner.com) or from the
[releases page](https://github.com/smegg99/Smegg-s-professional-Accordion-Tuner/releases).

- **Linux x86_64:** AppImage, DEB, RPM, Arch package
- **Linux ARM64:** DEB, RPM, Arch package
- **Windows x86_64:** installer and portable exe

What changed in this version is in [CHANGELOG.md](CHANGELOG.md).

## Screenshots

<table>
  <tr>
    <td><img src="assets/screenshots/smeggtuner-tuning.png" alt="Tuning a recording" width="400"><br>Two reeds and their beat, from a recording.</td>
    <td><img src="assets/screenshots/smeggtuner-tuner-idle.png" alt="The tuner waiting for a note" width="400"><br>The tuner waiting for a note.</td>
  </tr>
  <tr>
    <td><img src="assets/screenshots/smeggtuner-input-devices.png" alt="Choosing the input" width="400"><br>Pick a microphone or open a recording.</td>
    <td><img src="assets/screenshots/smeggtuner-settings.png" alt="Settings" width="400"><br>Settings.</td>
  </tr>
  <tr>
    <td><img src="assets/screenshots/smeggtuner-calibration-warmup.png" alt="Sound card calibration warming up" width="400"><br>Sound card calibration, warming up.</td>
    <td><img src="assets/screenshots/smeggtuner-calibration-complete.png" alt="Sound card calibration done" width="400"><br>Calibration done.</td>
  </tr>
</table>

## Building

You need Linux to develop. The app is Go and [Wails 3](https://v3.wails.io) with a Nuxt frontend.

- Go 1.27 with cgo, a C compiler and pkg-config
- Node.js `^22.22.3 || ^24.15.0 || >=26.0.0` and pnpm 11.3.0
- GTK 3 and WebKit2GTK 4.1 headers (`libgtk-3-dev libwebkit2gtk-4.1-dev` on Debian and Ubuntu,
  `gtk3-devel webkit2gtk4.1-devel` on Fedora)
- [just](https://github.com/casey/just), Python 3, and `rsvg-convert` from librsvg
- [S99AB](https://github.com/smegg99/S99AB) and [S99Wails](https://github.com/smegg99/S99Wails)
  checked out next to this repository

```sh
git clone https://github.com/smegg99/S99AB.git
git clone https://github.com/smegg99/S99Wails.git
git clone https://github.com/smegg99/Smegg-s-professional-Accordion-Tuner.git SmeggTuner
pnpm --dir S99AB install
cd SmeggTuner
pnpm --dir frontend install
just generate
just dev
```

| Command | What it does |
|---|---|
| `just dev` | Rebuild and restart the debug app when sources change |
| `just build` | Build `bin/smeggtuner` (`just build debug` for a debug build) |
| `just run` | Build and open the app |
| `just test` | Go, frontend and build script tests |
| `just check` | Generated files, formatting, vet, locales, types, lint and tests |
| `just web` | Build the browser edition into `dist/web` |
| `just install` | Install a per-user Linux build with a menu entry |
| `just release` | Run the checks, then build every release target into `dist/release` |

Android builds are covered in [build/Android.md](build/Android.md). Translations live in
`i18n/locales/{en,pl}`; run `just generate` after editing them.

The settings file, `config.json`, is written with defaults on first start. The schema is
[`backend/common/config/config.cue`](backend/common/config/config.cue). If the app finds a setting
it does not know, it shows a window naming the file and the setting.

## License

MIT, see [LICENSE](LICENSE).
