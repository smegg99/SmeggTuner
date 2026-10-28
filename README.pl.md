<p align="center">
  <img src="assets/smeggtuner-app-icon.svg" alt="SmeggTuner" width="140">
</p>

<p align="center">
  <a href="README.md">English</a> | <b>Polski</b>
  <br>
  <a href="https://smeggtuner.com">smeggtuner.com</a>
</p>

# SmeggTuner

Darmowy, otwartoźródłowy tuner do akordeonu. Zagraj dźwięk, a pokaże każdy głos rejestru osobno
i dudnienia między nimi, więc widać, który stroik piłować i o ile.

Działa tak jak profesjonalny Accordion Tuner Dirka, ale za darmo. Działa na Linuksie i Windowsie,
z mikrofonem albo z nagraniem WAV.

<p align="center">
  <img src="assets/screenshots/smeggtuner-recording-analysis.png" alt="Trzy głosy C4 zmierzone z nagrania" width="760">
</p>

> **Wersja beta.** Testy na prawdziwych akordeonach i kartach dźwiękowych są jeszcze ograniczone.
> Zanim zaczniesz piłować stroik, sprawdź odczyt tunerem, któremu ufasz.

## Funkcje

- Mierzy do trzech głosów jednego dźwięku naraz i dudnienie każdej pary.
- Trzy tryby: Dźwięki (głosy jednego dźwięku), Oktawy i Akordy.
- Odczyty w jednostkach Dirka (0,06 % częstotliwości nuty, mniej więcej jeden cent) albo w
  centach. Dudnienia w Hz albo w jednostkach.
- Wykrywanie nut: Szybkie, Zrównoważone albo Stabilne. Stabilne działa dokładnie jak tuner Dirka.
- Kalibracja karty dźwiękowej, osobno dla każdego mikrofonu. Trwa około minuty i nie wymaga
  dźwięku.
- Automatyczna głośność wejścia i szum tła uczony raz dla każdego mikrofonu. Przycisk na górnym
  pasku uczy go od nowa.
- Zatrzymuj stabilne odczyty: przycisk ze śnieżynką trzyma każdy zablokowany odczyt, aż dźwięk
  ucichnie.
- Otwiera nagrania WAV z odtwarzaniem, przewijaniem, pętlą i powiększeniem, więc można stroić
  bez akordeonu na stole.
- Kliknięcie paska nut odtwarza nutę.
- A4 od 430 do 450 Hz, transpozycja, przypięta nuta przesuwana strzałkami, tolerancje i cztery
  sposoby nazywania nut, w tym polski (C Cis D ... H).
- Filtr przydźwięku sieciowego 50 Hz lub 60 Hz.
- Polski i angielski, jasny i ciemny motyw, ikona w zasobniku, skróty klawiszowe.
- Gdy nie może się uruchomić, ulegnie awarii, przestanie odpowiadać albo zostanie zamknięta, mówi o
  tym w oknie z raportem, który można skopiować, zapisać albo wysłać e-mailem, razem z ostatnimi
  wpisami dziennika. Plik dziennika jest domyślnie zapisywany.

Warsztat z poprzednich wersji jest w 1.4 wyłączony. Były w nim biblioteka akordeonów, sesje,
nagrane odczyty, docelowe krzywe dudnień i drukowane raporty. Kod nadal tu jest i Warsztat wróci
w jednym z kolejnych wydań.

## SmeggTuner a tuner Dirka

| | Profesjonalny Accordion Tuner Dirka | SmeggTuner 1.4 |
|---|---|---|
| Cena | 298 EUR, plus 148 EUR za moduł nagrywania i raportów | Za darmo |
| Kod | Zamknięty | Otwarty, MIT |
| Systemy | Windows, macOS | Linux (x86_64, ARM64), Windows; eksperymentalna wersja w przeglądarce |
| Głosy mierzone naraz | Do 3 | Do 3 |
| Dźwięki, Oktawy, Akordy | Tak | Tak |
| Jednostki lub centy, dudnienia w Hz | Tak | Tak |
| Kalibracja karty dźwiękowej | Tak, dla każdego urządzenia | Tak, dla każdego mikrofonu |
| Automatyczna głośność wejścia | Tak | Tak |
| Szybkość wykrywania | Jedno zachowanie | Szybkie, Zrównoważone, Stabilne (Stabilne to Dirk) |
| Analiza pliku WAV | Nie ma jej w menu | Tak, z przewijaniem, pętlą i powiększeniem |
| Nagrywanie i raporty | Płatny moduł | Nie w 1.4, Warsztat wróci później |

Ceny tunera Dirka pochodzą z [dirksprojects.com](https://www.dirksprojects.com/), stan na
październik 2026.

## Plany

To plany, nie obietnice.

- [ ] Przywrócić Warsztat: bibliotekę akordeonów, sesje, nagrywanie, krzywe docelowe i raporty.
- [ ] Wydać wersję na Androida. Buduje się i działa w emulatorze, ale nie jest jeszcze wydana.
- [ ] macOS i iOS. Nie mam sprzętu, żeby je przetestować.
- [ ] Udostępnić wersję przeglądarkową pod app.smeggtuner.com.
- [ ] Więcej testów na prawdziwych akordeonach, mikrofonach i kartach dźwiękowych.

## Pobieranie

Najnowsza wersja jest na [smeggtuner.com](https://smeggtuner.com) i na
[stronie wydań](https://github.com/smegg99/Smegg-s-professional-Accordion-Tuner/releases).

- **Linux x86_64:** AppImage, DEB, RPM, pakiet dla Arch Linux
- **Linux ARM64:** DEB, RPM, pakiet dla Arch Linux
- **Windows x86_64:** instalator i wersja przenośna

Zmiany w tej wersji opisuje [CHANGELOG.md](CHANGELOG.md).

## Zrzuty ekranu

<table>
  <tr>
    <td><img src="assets/screenshots/smeggtuner-tuning.png" alt="Strojenie z nagrania" width="400"><br>Dwa głosy i ich dudnienie, z nagrania.</td>
    <td><img src="assets/screenshots/smeggtuner-tuner-idle.png" alt="Tuner czeka na dźwięk" width="400"><br>Tuner czeka na dźwięk.</td>
  </tr>
  <tr>
    <td><img src="assets/screenshots/smeggtuner-input-devices.png" alt="Wybór wejścia" width="400"><br>Wybierz mikrofon albo otwórz nagranie.</td>
    <td><img src="assets/screenshots/smeggtuner-settings.png" alt="Ustawienia" width="400"><br>Ustawienia.</td>
  </tr>
  <tr>
    <td><img src="assets/screenshots/smeggtuner-calibration-warmup.png" alt="Rozgrzewanie przed kalibracją karty dźwiękowej" width="400"><br>Kalibracja karty dźwiękowej, rozgrzewanie.</td>
    <td><img src="assets/screenshots/smeggtuner-calibration-complete.png" alt="Kalibracja karty dźwiękowej zakończona" width="400"><br>Kalibracja zakończona.</td>
  </tr>
</table>

## Budowanie

Do pracy nad aplikacją potrzebny jest Linux. Aplikacja jest napisana w Go i
[Wails 3](https://v3.wails.io), z frontendem w Nuxt.

- Go 1.27 z cgo, kompilator C i pkg-config
- Node.js `^22.22.3 || ^24.15.0 || >=26.0.0` i pnpm 11.3.0
- Nagłówki GTK 3 i WebKit2GTK 4.1 (`libgtk-3-dev libwebkit2gtk-4.1-dev` w Debianie i Ubuntu,
  `gtk3-devel webkit2gtk4.1-devel` w Fedorze)
- [just](https://github.com/casey/just), Python 3 i `rsvg-convert` z librsvg
- [S99AB](https://github.com/smegg99/S99AB) i [S99Wails](https://github.com/smegg99/S99Wails)
  sklonowane obok tego repozytorium

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

| Polecenie | Co robi |
|---|---|
| `just dev` | Przebudowuje i restartuje wersję debug po każdej zmianie źródeł |
| `just build` | Buduje `bin/smeggtuner` (`just build debug` dla wersji debug) |
| `just run` | Buduje i otwiera aplikację |
| `just test` | Testy Go, frontendu i skryptów budowania |
| `just check` | Pliki generowane, formatowanie, vet, tłumaczenia, typy, lint i testy |
| `just web` | Buduje wersję przeglądarkową do `dist/web` |
| `just install` | Instaluje wersję dla jednego użytkownika z wpisem w menu |
| `just release` | Uruchamia sprawdzenia, potem buduje wszystkie cele wydania do `dist/release` |

Budowanie na Androida opisuje [build/Android.md](build/Android.md). Tłumaczenia są w
`i18n/locales/{en,pl}`; po ich zmianie uruchom `just generate`.

Plik ustawień, `config.json`, powstaje z wartościami domyślnymi przy pierwszym uruchomieniu.
Schemat jest w [`backend/common/config/config.cue`](backend/common/config/config.cue). Jeśli
aplikacja znajdzie nieznane ustawienie, pokaże okno z nazwą pliku i tego ustawienia.

## Licencja

MIT, zobacz [LICENSE](LICENSE).
