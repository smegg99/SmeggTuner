<p align="center">
  <img src="assets/smeggtuner-readme.svg" alt="SmeggTuner" width="240">
</p>

<p align="center">
  <a href="README.md">English</a> | Polski | <a href="https://smeggtuner.com">smeggtuner.com</a>
</p>

# SmeggTuner

Tuner do strojenia akordeonów. Mierzy osobno każdy stroik nuty, do trzech naraz, razem z dudnieniem
między nimi. Słucha mikrofonu albo czyta nagranie WAV. Działa tak jak profesjonalny Accordion Tuner
Dirka, ale jest darmowy.

Działa na Linuksie i Windowsie. Wersja na Androida i wersja przeglądarkowa są eksperymentalne.

To wciąż beta, sprawdzona na niewielu akordeonach i kartach dźwiękowych. Zanim zaczniesz piłować
stroik, porównaj odczyt z tunerem, któremu ufasz.

<p align="center">
  <img src="assets/screenshots/smeggtuner-recording-analysis.png" alt="Trzy stroiki C4 zmierzone z nagrania" width="760">
</p>

## Co potrafi

Tuner mierzy dźwięki, oktawy albo akordy. W trybie Dźwięki pokazuje stroiki jednej nuty, w trybach
Oktawy i Akordy kilka nut naraz. Odchyłkę podaje w jednostkach Dirka (0,06 % częstotliwości nuty, czyli
mniej więcej cent) albo w centach, a dudnienie w hercach albo w jednostkach. Wykrywanie nut ma trzy
ustawienia: Szybkie, Zrównoważone i Stabilne. Stabilne wybiera i blokuje nutę tak samo jak tuner Dirka.

Każdy mikrofon ma własną kalibrację karty dźwiękowej. Trwa około minuty i nie potrzebuje żadnego
dźwięku. Szum tła tuner poznaje raz dla każdego mikrofonu, a przyciskiem na górnym pasku można go
nauczyć od nowa. Głośność wejścia może ustawiać się sama. Przycisk ze śnieżynką zatrzymuje każdy
zablokowany odczyt, dopóki dźwięk nie ucichnie.

Nagranie WAV mierzy się tak samo jak mikrofon. Można je odtwarzać, przewijać, zapętlać i powiększać.
Kliknięcie w pasek nut gra wybraną nutę. A4 ustawia się od 430 do 450 Hz; jest też transpozycja, filtr
przydźwięku sieciowego 50 albo 60 Hz i cztery sposoby nazywania nut, w tym polski (C Cis D ... H).

Kiedy aplikacja nie może wystartować, wysypie się, przestanie odpowiadać albo zostanie zamknięta siłą,
pokazuje okno z raportem, który można skopiować, zapisać albo wysłać mailem.

Warsztat z wersji 1.3 (biblioteka akordeonów, sesje, nagrane odczyty, krzywe docelowe i drukowane
raporty) jest w 1.4 wyłączony. Warsztat wróci w którymś z kolejnych wydań.

## Tuner Dirka

| | Profesjonalny Accordion Tuner Dirka | SmeggTuner 1.4 |
|---|---|---|
| Cena | 298 EUR i 148 EUR za moduł nagrywania i raportów | Za darmo |
| Kod | Zamknięty | MIT |
| Systemy | Windows, macOS | Linux, Windows; Android i przeglądarka eksperymentalnie |
| Stroiki naraz | 3 | 3 |
| Dźwięki, oktawy, akordy | Tak | Tak |
| Kalibracja karty dźwiękowej | Dla każdego urządzenia | Dla każdego mikrofonu |
| Wykrywanie nut | Jedno ustawienie | Szybkie, Zrównoważone, Stabilne |
| Pliki WAV | Brak w menu | Tak |
| Nagrywanie i raporty | Płatny moduł | Nie w 1.4 |

Ceny ze strony [dirksprojects.com](https://www.dirksprojects.com/), stan z października 2026.

## Plany

- [ ] Powrót Warsztatu
- [ ] Wersja przeglądarkowa pod adresem app.smeggtuner.com
- [ ] Testy na większej liczbie akordeonów, mikrofonów i kart dźwiękowych
- [ ] macOS i iOS, kiedy będę miał na czym je testować

## Pobieranie

Najnowsza wersja jest na stronie
[wydań](https://github.com/smegg99/SmeggTuner/releases) i na
[smeggtuner.com](https://smeggtuner.com):

- Linux x86_64: AppImage, DEB, RPM, pakiet dla Arch Linux
- Linux ARM64: DEB, RPM, pakiet dla Arch Linux
- Windows x86_64: instalator i wersja przenośna
- Android ARM64: APK, wersja eksperymentalna

Co się zmieniło, opisuje [CHANGELOG.md](CHANGELOG.md).

<table>
  <tr>
    <td><img src="assets/screenshots/smeggtuner-tuning.png" alt="Dwa stroiki i ich dudnienie" width="400"></td>
    <td><img src="assets/screenshots/smeggtuner-tuner-idle.png" alt="Tuner czeka na dźwięk" width="400"></td>
  </tr>
  <tr>
    <td><img src="assets/screenshots/smeggtuner-input-devices.png" alt="Wybór wejścia" width="400"></td>
    <td><img src="assets/screenshots/smeggtuner-settings.png" alt="Ustawienia" width="400"></td>
  </tr>
  <tr>
    <td><img src="assets/screenshots/smeggtuner-calibration-warmup.png" alt="Rozgrzewanie przed kalibracją karty dźwiękowej" width="400"></td>
    <td><img src="assets/screenshots/smeggtuner-calibration-complete.png" alt="Kalibracja karty dźwiękowej zakończona" width="400"></td>
  </tr>
</table>

## Budowanie

Budować można tylko na Linuksie. Potrzebujesz Go 1.27 z cgo, nagłówków GTK 3 i WebKit2GTK 4.1
(`libgtk-3-dev libwebkit2gtk-4.1-dev` w Debianie i Ubuntu, `gtk3-devel webkit2gtk4.1-devel`
w Fedorze), Node.js `^22.22.3 || ^24.15.0 || >=26.0.0`, pnpm 11.3.0, [just](https://github.com/casey/just),
Pythona 3 i `rsvg-convert`. Obok tego repozytorium muszą leżeć klony
[S99AB](https://github.com/smegg99/S99AB), [S99Wails](https://github.com/smegg99/S99Wails)
i [S99Logger](https://github.com/smegg99/S99Logger). Część tych bibliotek może jeszcze nie być
opublikowana.

```sh
pnpm --dir ../S99AB install
just build
```

Gotowy plik to `bin/smeggtuner`. `just run` buduje aplikację i od razu ją uruchamia.

## Praca nad kodem

`just dev` przebudowuje wersję debug i uruchamia ją ponownie po każdej zmianie w źródłach. `just test`
uruchamia testy Go, frontendu i skryptów budujących. `just check` sprawdza do tego pliki generowane,
formatowanie, vet, tłumaczenia, typy i lint, niczego nie zmieniając.

Tłumaczenia leżą w `i18n/locales/{en,pl}`, a schemat ustawień w `backend/common/config/config.cue`.
`just generate` odtwarza wszystko, co z nich powstaje, a do tego bindingi, ikony i pliki pomocnicze
S99Wails.

`just web` buduje wersję przeglądarkową do `dist/web`. `just release` przepuszcza wszystkie
sprawdzenia i buduje każdy cel do `dist/release/<wersja>`; APK na Androida potrzebuje klucza do
podpisu, opisanego w [build/Android.md](build/Android.md).

`deploy.json` wdraża wersję przeglądarkową przez [S99Deploy](https://github.com/smegg99/S99Deploy):
`deploy/app/fetch.sh` pobiera archiwum webowe wersji z `VERSION` z jej wydania na GitHubie i sprawdza je
z plikiem `SHA256SUMS` tego wydania, a `deploy/app` to mały serwer w Go, który je udostępnia. Na serwerze
`sudo depl up smeggtuner-app` wdraża wersję, którą wskazuje `VERSION` w repozytorium.

## Ustawienia

Przy pierwszym uruchomieniu aplikacja zapisuje `config.json` z domyślnymi wartościami: w katalogu,
z którego ją uruchomiono, a po instalacji z pakietu w `~/.config/smeggtuner`. Baza danych i katalog
`logs/` trafiają obok niego, a po instalacji do `~/.local/share/smeggtuner`. Jeśli w pliku jest
ustawienie, którego schemat nie zna, aplikacja nie wystartuje i pokaże okno z nazwą pliku i tego
ustawienia.

## Licencja

MIT, szczegóły w [LICENSE](LICENSE).
