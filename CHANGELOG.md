<!-- lang:en -->

## SmeggTuner 1.4.3

SmeggTuner now uses AGPL-3.0-only instead of MIT. Distributed derivatives must provide their
corresponding source code, and modified versions offered over a network must offer it to their users.
Commercial use is still allowed. Copies previously released under MIT retain their MIT permissions.

The About dialog now links to smegg.me and includes an offline license viewer with a scrollable text
area. Its links are arranged in two rows.

`just android-release` builds the signed APK and app bundle with one password prompt. Frontend
dependencies now match the shared layer, and build commands select the pinned pnpm version correctly.
Go was updated to 1.27.4 and Wails to 3.1.0.

The Workshop remains hidden. Android and the web version are still experimental.

### Supported platforms

**Linux x86_64:** AppImage, DEB, RPM, Arch Linux package, and per-user archive.

**Linux ARM64:** DEB, RPM, Arch Linux package, and per-user archive.

**Windows x86_64:** installer and portable executable.

**Android ARM64:** experimental signed APK and app bundle (AAB).

**Web:** experimental, at app.smeggtuner.com.

iOS and macOS are not currently supported.

<!-- lang:pl -->

## SmeggTuner 1.4.3

SmeggTuner korzysta teraz z licencji AGPL-3.0-only zamiast MIT. Rozpowszechniane wersje pochodne muszą
udostępniać odpowiadający im kod źródłowy, a zmodyfikowane wersje oferowane przez sieć muszą oferować
go swoim użytkownikom. Użytek komercyjny nadal jest dozwolony. Kopie wcześniej wydane na MIT zachowują
uprawnienia wynikające z MIT.

Okno "O programie" zawiera teraz odnośnik do smegg.me i podgląd licencji dostępny bez internetu,
z przewijanym polem tekstowym. Odnośniki zostały ułożone w dwóch rzędach.

`just android-release` buduje podpisany plik APK i pakiet aplikacji po jednokrotnym podaniu hasła.
Zależności frontendu są zgodne ze wspólną warstwą, a polecenia budujące wybierają właściwą wersję pnpm.
Zaktualizowano Go do wersji 1.27.4, a Wails do 3.1.0.

Warsztat pozostaje ukryty. Wersje na Androida i do przeglądarki nadal są eksperymentalne.

### Obsługiwane platformy

**Linux x86_64:** AppImage, pakiety DEB, RPM i Arch Linux oraz archiwum do instalacji bez uprawnień administratora.

**Linux ARM64:** pakiety DEB, RPM i Arch Linux oraz archiwum do instalacji bez uprawnień administratora.

**Windows x86_64:** instalator oraz wersja przenośna.

**Android ARM64:** eksperymentalny podpisany plik APK i pakiet aplikacji AAB.

**Przeglądarka:** wersja eksperymentalna pod app.smeggtuner.com.

Systemy iOS i macOS nie są obecnie obsługiwane.
