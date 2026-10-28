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
Clicking the note strip plays the note without crackling. A settings file the app cannot read now
opens a window naming the file and the setting, instead of the app closing without a word.

The Workshop (sessions, recording and printed cards) is hidden in this release, so a first start
shows a tuner and not a session manager. Nothing already stored is deleted.

When the app cannot start, crashes, stops responding or is killed, a window says so and offers a
report to copy, save or email, with the recent log in it; the email carries the report and the log
file as attachments where the mail client takes them. A log file is now kept by default, next to the
settings file or in the app's data folder when installed.

New: Linux ARM64 packages, and an experimental edition that runs in the browser.

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

**Web:** experimental browser edition.

Android, iOS, and macOS are not currently supported.

<!-- lang:pl -->

## SmeggTuner 1.4.0

Silnik pomiarowy został przebudowany tak, żeby działał jak profesjonalny Tuner Dirka.
Odczyty są teraz pokazywane w jego jednostkach, każda to 0,06 % częstotliwości nuty (mniej więcej
jeden cent); centy można przywrócić w ustawieniach. Nowe ustawienie detekcji decyduje, jak szybko
nowa nuta jest odczytywana i blokowana: Szybkie, Zrównoważone albo Stabilne.

Karta dźwiękowa jest teraz kalibrowana osobno dla każdego mikrofonu, na własnym ekranie. Karta,
której nigdy nie skalibrowano, nadal mierzy, a przycisk na górnym pasku o tym przypomina. Aplikacja
może sama ustawiać poziom mikrofonu podczas pomiaru i uczy się szumu tła raz dla każdego mikrofonu,
z przyciskiem, żeby nauczyć się go od nowa.

Górny pasek został przerobiony: są na nim tryby Dźwięki, Oktawy i Akordy oraz liczba głosów, A4
wpisuje się wprost w karcie nuty, a każdy przycisk ma podpowiedź. Zatrzymywanie stabilnych odczytów
przeniosło się z ustawień na przycisk ze śnieżynką na górnym pasku (H albo spacja w stroiku): trzyma
każdy zablokowany odczyt, aż dźwięk ucichnie, a osobny przycisk, który od razu zamrażał odczyt,
zniknął. Kliknięcie paska nut odtwarza nutę bez trzasków. Plik ustawień, którego aplikacja nie
potrafi odczytać, otwiera teraz okno z nazwą pliku i ustawienia, zamiast zamykać aplikację bez
słowa.

Warsztat (sesje, nagrywanie i drukowane karty) jest w tym wydaniu ukryty, żeby pierwsze
uruchomienie pokazywało stroik, a nie menedżera sesji. Nic z tego, co już zapisano, nie jest
usuwane.

Gdy aplikacja nie może się uruchomić, ulegnie awarii, przestanie odpowiadać albo zostanie zamknięta,
okno o tym mówi i proponuje raport do skopiowania, zapisania albo wysłania e-mailem, razem z
ostatnimi wpisami dziennika; e-mail dostaje raport i plik dziennika jako załączniki, jeśli program
pocztowy je przyjmuje. Plik dziennika jest teraz domyślnie zapisywany, obok pliku ustawień albo, po
instalacji, w folderze danych aplikacji.

Nowe: pakiety dla Linuksa ARM64 i eksperymentalna wersja działająca w przeglądarce.

### Zmiany niekompatybilne

- **Plik ustawień.** Wersja 1.4 odrzuca ustawienia zapisane przez 1.3 i przy starcie pokazuje,
  które znalazła. Usuń z `config.json` te klucze albo usuń cały plik, żeby wrócić do ustawień
  domyślnych:
  - `audio.clock_ppm`
  - całą sekcję `engine`
  - `tuner.unit`
  - `report.logo_path`
  - `tuner.note_sounds`

  Korekta karty dźwiękowej jest teraz mierzona osobno dla każdego mikrofonu przez wbudowaną
  kalibrację, więc wartość `clock_ppm` z 1.3 nie jest przenoszona. Pasek nut gra naciśniętą nutę
  zawsze, gdy dźwięk nie jest wyciszony, więc przełącznik `note_sounds` nie jest już potrzebny.
- **Sesje z 1.3.** Nowy silnik zapisuje odczyty w innej postaci, a 1.4 nie pokazuje sesji, dopóki
  Warsztat jest ukryty. Sesje zapisane przez 1.3 zostają bez zmian; nic nie jest konwertowane ani
  usuwane.

### Obsługiwane platformy

**Linux x86_64:** AppImage, pakiety DEB i RPM oraz pakiet dla Arch Linux.

**Linux ARM64:** pakiety DEB i RPM oraz pakiet dla Arch Linux.

**Windows x86_64:** instalator oraz wersja przenośna.

**Przeglądarka:** eksperymentalna wersja webowa.

Systemy Android, iOS i macOS nie są obecnie obsługiwane.
