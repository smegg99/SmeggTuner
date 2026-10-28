<!-- lang:en -->

## SmeggTuner 1.4.0

The measuring engine was rebuilt to work the way Dirk's professional Accordion Tuner does. Readings
are now shown in its units, 0.06 % of the note's frequency each (about one cent); cents are one
setting away. A new detection setting picks how quickly a new note is read and locked: Fast,
Balanced or Steady.

The sound card is now calibrated per microphone, on its own screen. A card that was never
calibrated still measures, and a key in the top bar says so. The app can set the microphone level
by itself while measuring, and it learns the room's background noise once per microphone, with a
key to learn it again.

The top bar was redone: the Tones, Octaves and Chords modes and the reed counts sit in it, A4 is
typed straight into the note card, and every control has a tooltip. Holding stable readings moved
from the settings to the snowflake key in the top bar (H, or Space on the tuner): it holds each
locked reading until the note stops, and the separate key that froze the reading at once is gone.
Clicking the note strip plays the note without crackling.

The Workshop (sessions, recording and printed cards) is hidden in this release, so a first start
shows a tuner and not a session manager. Nothing already stored is deleted.

When the app cannot start, crashes, stops responding or is killed, a window says so and offers a
report to copy, save or email, with the recent log in it. The email carries the report and the log
file as attachments where the mail client takes them. A settings file the app cannot read opens the
same window, naming the file and the setting, instead of the app closing without a word. A log file
is now kept by default, next to the settings file or in the app's data folder when installed.

New: Linux ARM64 packages, and experimental editions for Android and the browser.

### Breaking changes

- **Settings file.** 1.4 refuses settings that 1.3 wrote and names the one it found at startup.
  Delete these from `config.json`, or delete the whole file to get the defaults:
  - `audio.clock_ppm`
  - the whole `engine` section
  - `tuner.unit`
  - `report.logo_path`
  - `tuner.note_sounds`

  The sound card correction is now measured per microphone by the built-in calibration, so 1.3's
  `clock_ppm` is not carried over. The note strip plays the pressed note whenever the sound is not
  muted, so the `note_sounds` switch is gone.
- **Sessions from 1.3.** The new engine stores readings in a different shape, and 1.4 does not show
  sessions while the Workshop is hidden. Sessions saved by 1.3 are left as they are; nothing is
  converted or deleted.

### Supported platforms

**Linux x86_64:** AppImage, DEB, RPM, and Arch Linux packages.

**Linux ARM64:** DEB, RPM, and Arch Linux packages.

**Windows x86_64:** installer and portable executable.

**Android ARM64:** experimental APK.

**Web:** experimental browser edition.

iOS and macOS are not currently supported.

<!-- lang:pl -->

## SmeggTuner 1.4.0

Silnik pomiarowy powstał od nowa i mierzy tak jak profesjonalny Accordion Tuner Dirka. Odczyty są
teraz w jego jednostkach: jedna to 0,06 % częstotliwości nuty, czyli mniej więcej cent. Centy można
włączyć w ustawieniach. Nowe ustawienie wykrywania nut decyduje, jak szybko tuner łapie i blokuje nową
nutę: Szybkie, Zrównoważone albo Stabilne.

Karta dźwiękowa ma teraz osobną kalibrację dla każdego mikrofonu, na własnym ekranie. Nieskalibrowana
karta też mierzy, a przycisk na górnym pasku o tym przypomina. Podczas pomiaru aplikacja może sama
ustawiać głośność mikrofonu. Szum tła poznaje raz dla każdego mikrofonu, a osobny przycisk pozwala
nauczyć go od nowa.

Górny pasek jest przerobiony. Są na nim tryby Dźwięki, Oktawy i Akordy oraz liczba stroików, A4 wpisuje
się prosto w kartę nuty, a każdy przycisk ma podpowiedź. Zatrzymywanie stabilnych odczytów przeszło
z ustawień na przycisk ze śnieżynką na górnym pasku (H albo spacja w tunerze). Trzyma każdy
zablokowany odczyt, dopóki dźwięk nie ucichnie. Osobnego przycisku, który od razu zamrażał odczyt,
już nie ma. Kliknięcie w pasek nut gra nutę bez trzasków.

Warsztat (sesje, nagrywanie i drukowane karty) jest w tym wydaniu ukryty, więc po pierwszym
uruchomieniu widać tuner, a nie menedżer sesji. Nic z zapisanych danych nie jest usuwane.

Kiedy aplikacja nie może wystartować, wysypie się, przestanie odpowiadać albo zostanie zamknięta siłą,
pokazuje okno z raportem do skopiowania, zapisania albo wysłania mailem. Raport zawiera ostatnie wpisy
dziennika, a mail dostaje raport i plik dziennika jako załączniki, jeśli program pocztowy na to
pozwala. Plik ustawień, którego aplikacja nie potrafi odczytać, otwiera to samo okno z nazwą pliku
i ustawienia, zamiast po cichu zamykać aplikację. Plik dziennika jest teraz zapisywany domyślnie, obok
pliku ustawień albo, po instalacji, w folderze danych aplikacji.

Nowości: pakiety dla Linuksa ARM64 i eksperymentalne wersje na Androida i do przeglądarki.

### Zmiany niezgodne wstecz

- **Plik ustawień.** Wersja 1.4 nie przyjmuje ustawień zapisanych przez 1.3 i przy starcie pokazuje,
  które znalazła. Usuń z `config.json` te klucze albo cały plik, żeby wrócić do wartości domyślnych:
  - `audio.clock_ppm`
  - całą sekcję `engine`
  - `tuner.unit`
  - `report.logo_path`
  - `tuner.note_sounds`

  Poprawkę karty dźwiękowej mierzy teraz wbudowana kalibracja, osobno dla każdego mikrofonu, więc
  `clock_ppm` z 1.3 nie jest przenoszony. Pasek nut gra naciśniętą nutę zawsze, gdy dźwięk nie jest
  wyciszony, więc przełącznik `note_sounds` zniknął.
- **Sesje z 1.3.** Nowy silnik zapisuje odczyty inaczej, a 1.4 nie pokazuje sesji, dopóki Warsztat
  jest ukryty. Sesje zapisane przez 1.3 zostają tak, jak były; nic nie jest konwertowane ani usuwane.

### Obsługiwane platformy

**Linux x86_64:** AppImage, DEB, RPM i pakiet dla Arch Linux.

**Linux ARM64:** DEB, RPM i pakiet dla Arch Linux.

**Windows x86_64:** instalator i wersja przenośna.

**Android ARM64:** eksperymentalny plik APK.

**Przeglądarka:** eksperymentalna wersja webowa.

iOS i macOS nie są na razie obsługiwane.
