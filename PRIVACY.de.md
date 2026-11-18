Stand: 12. November 2026

## Wer verantwortlich ist

SmeggTuner wird von smegg99 entwickelt. Diese Erklärung beschreibt, wie die Android-App SmeggTuner (Paket `me.smegg.smeggtuner`), die Desktop- und Browserversionen sowie smeggtuner.com mit Informationen umgehen. Die Funktionen unterscheiden sich je nach Version.

Bei Fragen oder Anliegen zum Datenschutz erreichen Sie smegg99 unter [support@smegg.me](mailto:support@smegg.me). Für die Kontaktaufnahme mit dem Entwickler benötigen Sie kein App-Konto.

## Mikrofon und Audiodateien

SmeggTuner berechnet aus dem Mikrofonsignal auf Ihrem Gerät die Tonhöhe, die Frequenzen der Stimmzungen und ihre Schwebungen. Unter Android bittet die App um Zugriff auf das Mikrofon. Sie können die Berechtigung in den Android-App-Einstellungen verweigern oder entziehen. Stimmen über das Mikrofon ist dann nicht möglich. Zum Öffnen einer Audiodatei ist kein Mikrofonzugriff erforderlich.

Die Audioanalyse erfolgt lokal. SmeggTuner überträgt weder Mikrofonsignale noch importierte Audiodateien oder Stimmmessungen an den Entwickler, einen Analysedienst oder einen Werbedienst. Zur Analyse wird das Live-Signal im Arbeitsspeicher gehalten. Der Start des Stimmgeräts speichert nicht automatisch eine Mikrofonaufnahme. Unter Android enden Aufnahme und Tonausgabe des Signalgenerators, sobald die App den Vordergrund verlässt.

Wenn Sie eine WAV-Datei auswählen, liest SmeggTuner sie zur Wiedergabe und Analyse. Die Android-Dokumentauswahl bestimmt, auf welche Datei die App zugreifen kann. Ein von Ihnen gewählter Cloudspeicherdienst kann die Datei nach seinen eigenen Verfahren herunterladen. Temporäre lokale Kopien können im App-Cache verbleiben, bis Android oder Sie ihn leeren.

## Auf dem Gerät gespeicherte Informationen

Die App speichert Einstellungen wie Sprache, Farbschema, Kammerton und Audiokalibrierung lokal. Ihre lokale Datenbank kann außerdem gespeicherte Stimmarbeiten aus Versionen enthalten, die diese Funktionen anbieten.

Technische Protokolle können App-Fehler, Angaben zum Gerät oder Audiogerät, Einstellungen und Dateipfade enthalten. Die Protokolle bleiben lokal. Dazu gehört auch das Systemprotokoll von Android. Die App lädt diese Protokolle nicht automatisch hoch und sendet dem Entwickler keine automatischen Absturzberichte. Wenn Sie für eine Supportanfrage ein Protokoll oder einen Screenshot senden möchten, prüfen Sie den Inhalt vorher und entfernen Sie alles, was Sie nicht weitergeben möchten.

Die Android-App nutzt ihren privaten App-Speicher. Sie verschlüsselt ihre lokale Datenbank und ihre Dateien nicht zusätzlich. Ihr Schutz hängt auch von den Sicherheitseinstellungen des Geräts ab. Android-Sicherung und Geräteübertragung sind aktiviert. Je nach Android-Version, Gerät und Sicherungseinstellungen kann das Betriebssystem App-Daten zu Ihrem Sicherungsanbieter oder auf ein anderes Gerät kopieren. Der Entwickler erhält diese Sicherungen nicht. Sie können sie über Ihr Gerät und Ihren Sicherungsanbieter verwalten.

## Netzwerkverbindungen, Konten und Dritte

Nach der Installation funktioniert das Android-Stimmgerät ohne Konto und ohne Internetverbindung. Die App enthält kein Werbe-SDK und keinen integrierten Analysedienst zur Nutzerverfolgung. Es gibt keine Kontoregistrierung und keinen automatischen Upload von Stimmdaten. Die App fordert keinen Zugriff auf Kontakte, Standort, Kamera, Fotomediathek oder Werbe-ID an. Mit der Dateiauswahl erlauben Sie nur den Zugriff auf das gewählte Dokument.

Anders als die Desktopversion sucht die Android-Version nicht auf GitHub nach Updates. Google Play verwaltet Installation und Updates der Play-Version nach den Richtlinien von Google. Google oder der Gerätehersteller können unabhängig davon Informationen über den App-Store, das Betriebssystem, Sicherungen oder Diagnosen gemäß Ihren Einstellungen und ihren eigenen Richtlinien verarbeiten.

Auf dem Desktop fragt die optionale Updateprüfung über HTTPS bei GitHub nach neuen Versionen. Dabei übermittelt sie den Namen und die Version der App. GitHub erhält Verbindungsdaten wie Ihre IP-Adresse. Sie können die Suche in den Einstellungen ausschalten. Audio- und Stimmdateien sind nicht Teil dieser Anfrage.

Links zur Online-Datenschutzerklärung, zur Website und zu GitHub öffnen bei Auswahl einen externen Browser. Die Erklärung selbst lässt sich ohne Verbindung in der App lesen. E-Mail-Links zum Support öffnen Ihr E-Mail-Programm. Eine Nachricht wird erst versandt, wenn Sie sie selbst abschicken. Beim Aufruf eines Webziels erhält dieses die üblichen Anfragedaten, darunter IP-Adresse und Browserinformationen. Es gelten die Richtlinien des jeweiligen Anbieters. Spenden wickelt der verlinkte Anbieter ab. SmeggTuner erhält keine Zahlungskartendaten. Wenn eine Version das Exportieren oder Drucken von Dokumenten ermöglicht, erhält das von Ihnen gewählte Ziel oder der Druckdienst das jeweilige Dokument.

## Website und Browserversion

Beim Besuch von smeggtuner.com und beim Laden der Browserversion werden Anfragen an die Hosting-Infrastruktur gestellt. Verbindungs- und Sicherheitsdaten können die IP-Adresse, die angeforderte URL, den Zeitpunkt, Browserinformationen und Fehler umfassen. Hosting- und Sicherheitsanbieter verarbeiten diese Daten, um den Dienst bereitzustellen und Missbrauch zu verhindern. Cloudflare schützt den Webverkehr. Die Verarbeitung beschreibt die [Datenschutzerklärung von Cloudflare](https://www.cloudflare.com/privacypolicy/).

Bei aktivierter Analyse nutzt die Website die cookiefreie Besuchsstatistik von Umami. Erfasst werden können besuchte Seiten, verweisende Websites, Browser- und Gerätekategorien, der ungefähre Standort, Klicks auf gekennzeichnete Links und die Seitenleistung. Umami verarbeitet Verbindungsdaten, um Besuche zu unterscheiden. Auch ohne Cookies werden also Informationen verarbeitet. Der Analyse-Endpunkt ist über diese Website erreichbar. Diese Website-Statistiken sind von der Android-App getrennt und enthalten weder Mikrofonaufnahmen noch Stimmmessungen.

Die Website speichert Ihre Farbschema-Einstellung im Browser und kann ein Cookie für die Sprachwahl verwenden. Die Kontaktseite begrenzt automatisierte Anfragen mit Cloudflare Turnstile. Dabei werden Browser- und Verbindungssignale verarbeitet, wie in den [Datenschutzhinweisen zu Turnstile](https://www.cloudflare.com/turnstile-privacy-policy/) beschrieben. Zum Lesen dieser Erklärung müssen Sie die Prüfung der Kontaktseite nicht durchlaufen.

Das Browser-Stimmgerät analysiert Mikrofonsignale nach Ihrer Freigabe lokal im Browser. Es lädt diese Audiodaten nicht zur Analyse hoch. Über die Berechtigungen und die Einstellungen für Websitedaten im Browser können Sie den Zugriff entziehen und lokale Einstellungen löschen.

## Support, Aufbewahrung und Löschung

Wenn Sie den Entwickler kontaktieren, werden Ihre freiwilligen Angaben, Ihre E-Mail-Adresse und die Korrespondenz zur Beantwortung der Anfrage und zur Untersuchung gemeldeter Probleme verwendet. Senden Sie Aufnahmen, personenbezogene Daten oder vertrauliche Dokumente nur, wenn sie für Ihre Anfrage erforderlich sind. Informationen werden weder verkauft noch für gezielte Werbung genutzt.

Lokale Einstellungen und gespeicherte Arbeiten bleiben erhalten, bis Sie sie entfernen, den App-Speicher löschen oder die App deinstallieren. Unter Android setzen Sie die App-Daten über Einstellungen > Apps > SmeggTuner > Speicher > Speicherinhalt löschen zurück. Die Bezeichnungen können je nach Gerät variieren. Das Leeren des Caches entfernt temporäre Dateien, setzt jedoch nicht alle Einstellungen zurück. Weder die Deinstallation noch das Löschen der App-Daten entfernt anderweitig exportierte Dateien, an andere Personen gesandte Kopien oder bei einem anderen Dienst gespeicherte Sicherungen. Verwalten Sie diese Kopien separat.

Supportkorrespondenz wird so lange aufbewahrt, wie es zur Beantwortung der Anfrage, zur weiteren Bearbeitung des Problems oder zur Erfüllung gesetzlicher Pflichten erforderlich ist. Betriebs- und Sicherheitsprotokolle der Website dienen dem Betrieb, der Missbrauchsprävention und der Untersuchung von Vorfällen. Besuchsstatistiken dienen dazu, die Nutzung der Website zu verstehen und sie zu verbessern. Die Aufbewahrungsdauer hängt von diesen Zwecken und den jeweiligen Diensteinstellungen ab. Über den genannten Datenschutzkontakt können Sie Auskunft über die beim Entwickler gespeicherten Informationen oder deren Löschung verlangen. Auf Dateien, die nur auf Ihrem Gerät liegen, kann der Entwickler nicht aus der Ferne zugreifen. Er kann sie auch nicht löschen.

## Ihre Möglichkeiten und Rechte

Soweit Datenschutzrecht anwendbar ist, können Sie unter Umständen Auskunft, Berichtigung oder Löschung Ihrer personenbezogenen Daten verlangen, die Verarbeitung einschränken oder ihr widersprechen und eine Kopie in einem übertragbaren Format erhalten. Sie können sich auch bei Ihrer örtlichen Datenschutzaufsichtsbehörde beschweren. Zur Ausübung dieser Rechte wenden Sie sich an den Entwickler. Um eine Offenlegung gegenüber anderen Personen zu verhindern, kann eine Identitätsprüfung erforderlich sein.

Die für Bereitstellung und Schutz der Website sowie die Beantwortung von Anfragen erforderliche Verarbeitung beruht auf dem jeweiligen Dienstzweck und, soweit gesetzlich zulässig, auf berechtigten Interessen. Informationen können auch zur Erfüllung gesetzlicher Pflichten aufbewahrt werden. Dienstleister können Informationen nach ihren jeweiligen Bedingungen und Schutzvorkehrungen in anderen Ländern als Ihrem verarbeiten. Da kein App-Konto angelegt wird, gibt es kein SmeggTuner-Konto zu löschen.

## Änderungen

Diese Seite wird aktualisiert, wenn sich relevante Funktionen oder der Umgang mit Daten ändern. Das Datum oben kennzeichnet die aktuelle Fassung. Eine mit der App ausgelieferte Kopie lässt sich im Info-Dialog offline lesen. Die aktuelle Webfassung finden Sie unter [smeggtuner.com/de/privacy](https://smeggtuner.com/de/privacy).
