Sist oppdatert 12. november 2026

## Hvem er ansvarlig?

SmeggTuner utvikles av smegg99. Denne erklæringen beskriver hvordan Android-appen SmeggTuner (pakken `me.smegg.smeggtuner`), skrivebords- og nettleserutgavene og smeggtuner.com håndterer opplysninger. Funksjonene varierer mellom utgavene.

Har du spørsmål eller forespørsler om personvern, kan du kontakte smegg99 på [support@smegg.me](mailto:support@smegg.me). Du trenger ingen konto i appen for å kontakte utvikleren.

## Mikrofon og lydfiler

SmeggTuner bruker mikrofonlyden til å beregne tonehøyde, stemtungenes frekvenser og slag på enheten din. På Android ber appen om tilgang til mikrofonen. Du kan avslå eller trekke tilbake tillatelsen i appinnstillingene i Android. Da kan du ikke stemme direkte med mikrofonen. Du kan åpne en lydfil uten mikrofontilgang.

Lyden analyseres lokalt. SmeggTuner laster ikke opp mikrofonlyd, importerte lydfiler eller stemmemålinger til utvikleren, en analysetjeneste eller en annonsetjeneste. Mikrofonlyden holdes i arbeidsminnet mens den analyseres. Det lagres ikke automatisk et mikrofonopptak når stemmeapparatet starter. På Android stopper opptaket og avspillingen av genererte toner når appen forlater forgrunnen.

Når du velger en WAV-fil, leser SmeggTuner den for avspilling og analyse. Dokumentvelgeren i Android styrer hvilken fil appen får tilgang til. Hvis du velger en skylagringstjeneste, kan den laste ned filen i tråd med sine egne rutiner. Midlertidige lokale kopier kan bli liggende i appens hurtigbuffer til Android eller du selv tømmer den.

## Opplysninger som lagres på enheten

Appen lagrer innstillinger som språk, tema, kammertone og lydkalibrering lokalt. Den lokale databasen kan også inneholde lagret stemmearbeid fra utgaver som tilbyr disse funksjonene.

Tekniske logger kan inneholde appfeil, opplysninger om enheten eller lydenheten, innstillinger og filstier. Loggene lagres lokalt, også systemloggen i Android. Appen laster ikke automatisk opp disse loggene eller sender krasjrapporter til utvikleren. Hvis du velger å sende en logg eller et skjermbilde for å få hjelp, bør du først se gjennom innholdet og fjerne det du ikke vil dele.

Android-appen bruker sitt private lagringsområde. Den legger ikke til et ekstra krypteringslag for den lokale databasen eller filene. Beskyttelsen avhenger også av enhetens sikkerhetsinnstillinger. Sikkerhetskopiering og enhetsoverføring er aktivert i Android. Avhengig av Android-versjon, enhet og innstillingene for sikkerhetskopiering kan operativsystemet kopiere appdata til leverandøren av sikkerhetskopier eller til en annen enhet. Utvikleren mottar ikke disse sikkerhetskopiene. Du kan administrere dem via enheten og leverandøren.

## Nettverksforbindelser, kontoer og tredjeparter

Etter installasjonen fungerer Android-stemmeapparatet uten konto eller internettforbindelse. Det har verken annonse-SDK, analyseverktøy som sporer bruk i appen, kontoregistrering eller automatisk opplasting av stemmedata. Det ber ikke om tilgang til kontaktene dine, posisjonen din, kameraet, bildebiblioteket eller annonse-ID-en. Når du velger en fil, gir du bare tilgang til det valgte dokumentet.

Android-utgaven utfører ikke skrivebordsutgavens oppdateringssjekker mot GitHub. Google Play håndterer installasjon og oppdatering av Play-utgaven etter Googles retningslinjer. Google eller enhetsprodusenten kan uavhengig av utvikleren behandle opplysninger om appbutikken, operativsystemet, sikkerhetskopiering eller diagnostikk i tråd med innstillingene dine og sine egne retningslinjer.

På skrivebordet kan den valgfrie oppdateringskontrollen kontakte GitHub over HTTPS for å se etter nye utgaver. Forespørselen inneholder appens navn og versjon. GitHub mottar tilkoblingsopplysninger som IP-adressen din. Du kan slå av kontrollen i innstillingene. Lydfiler og stemmefiler inngår ikke i forespørselen.

Hvis du velger en lenke til personvernerklæringen på nett, nettstedet eller GitHub, åpnes en ekstern nettleser. Selve erklæringen kan leses i appen uten nett. E-postlenker til brukerstøtte åpner e-postprogrammet ditt. En melding sendes bare hvis du selv velger å sende den. Nettstedet du besøker, mottar vanlige opplysninger om nettforespørselen, blant annet IP-adresse og nettleseropplysninger, og bruker sine egne retningslinjer. Donasjoner håndteres av leverandøren lenken viser til. SmeggTuner mottar ikke betalingskortopplysningene dine. Hvis en utgave lar deg eksportere eller skrive ut et dokument, mottar målet eller utskriftstjenesten du velger, det aktuelle dokumentet.

## Nettstedet og nettleserutgaven

Når du besøker smeggtuner.com eller laster inn nettleserutgaven, sendes forespørsler til driftsinfrastrukturen. Tilkoblings- og sikkerhetsopplysninger kan omfatte IP-adresse, forespurt URL, tidspunkt, nettleseropplysninger og feil. Drifts- og sikkerhetsleverandørene behandler disse opplysningene for å levere tjenesten og forebygge misbruk. Cloudflare beskytter nettrafikken og beskriver behandlingen i [Cloudflares personvernerklæring](https://www.cloudflare.com/privacypolicy/).

Når besøksanalyse er aktivert, bruker nettstedet statistikk fra Umami uten informasjonskapsler. Den kan omfatte besøkte sider, henvisende nettsteder, nettleser- og enhetskategorier, omtrentlig posisjon, klikk på merkede lenker og sideytelse. Umami behandler tilkoblingsopplysninger for å skille besøk fra hverandre. Det behandles altså opplysninger selv om det ikke brukes informasjonskapsler. Analyseendepunktet ligger på dette nettstedet. Nettstedets statistikk er atskilt fra Android-appen og inneholder verken mikrofonlyd eller stemmemålinger.

Nettstedet lagrer temavalget ditt i nettleseren og kan bruke en informasjonskapsel for språkvalget. Kontaktsiden bruker Cloudflare Turnstile for å begrense automatiserte forespørsler. Turnstile behandler signaler fra nettleseren og tilkoblingen slik det er beskrevet i [personvernerklæringen for Turnstile](https://www.cloudflare.com/turnstile-privacy-policy/). Du trenger ikke gjennomføre kontrollen på kontaktsiden for å lese denne erklæringen.

Etter at du har gitt tillatelse i nettleseren, behandler stemmeapparatet mikrofonlyden lokalt i nettleseren. Det laster ikke opp lyden for analyse. Via nettleserens tillatelser og innstillinger for nettstedsdata kan du trekke tilbake tilgangen og slette lokale innstillinger.

## Brukerstøtte, lagring og sletting

Hvis du kontakter utvikleren, brukes opplysningene du velger å sende, e-postadressen din og korrespondansen til å besvare henvendelsen og undersøke meldte problemer. Send bare opptak, personopplysninger eller sensitive dokumenter hvis de er nødvendige for henvendelsen. Opplysninger selges ikke og brukes ikke til målrettet reklame.

Lokale innstillinger og lagret arbeid blir værende til du fjerner dem, tømmer appens lagring eller avinstallerer appen. På Android kan du nullstille appdata under Innstillinger > Apper > SmeggTuner > Lagring > Tøm lagring. Navnene varierer fra enhet til enhet. Tømming av hurtigbufferen fjerner midlertidige filer, men nullstiller ikke alle innstillinger. Avinstallering eller sletting av appdata sletter ikke filer du har eksportert andre steder, kopier du har sendt til andre, eller sikkerhetskopier hos en annen tjeneste. Disse kopiene må håndteres separat.

Korrespondanse med brukerstøtte lagres så lenge det er nødvendig for å besvare henvendelsen, følge opp saken eller oppfylle gjeldende rettslige forpliktelser. Nettstedets drifts- og sikkerhetslogger lagres av hensyn til drift, forebygging av misbruk og undersøkelse av hendelser. Besøksstatistikk beholdes for å forstå hvordan nettstedet brukes og forbedre det. Lagringstiden avhenger av formålene og de aktuelle tjenesteinnstillingene. Via kontaktadressen ovenfor kan du spørre hvilke opplysninger utvikleren har, eller be om at de slettes. Utvikleren kan ikke få fjerntilgang til eller slette filer som bare finnes på enheten din.

## Dine valg og rettigheter

Der personvernlovgivningen gjelder, kan du ha rett til innsyn i, retting eller sletting av egne personopplysninger, begrensning av eller innsigelse mot behandlingen og utlevering av en kopi i et overførbart format. Du kan også klage til din lokale personvernmyndighet. Kontakt utvikleren for å utøve rettighetene. Det kan være nødvendig å bekrefte identiteten din for å unngå at opplysninger gis til andre.

Behandling som er nødvendig for å levere og sikre nettstedet eller besvare henvendelser, bygger på tjenestens aktuelle formål og berettigede interesser der loven tillater det. Opplysninger kan også lagres for å oppfylle rettslige forpliktelser. Tjenesteleverandører kan behandle opplysninger i andre land enn ditt etter sine gjeldende vilkår og sikkerhetstiltak. Det opprettes ingen appkonto, så det finnes ingen SmeggTuner-konto å slette.

## Endringer

Denne siden oppdateres når relevante funksjoner eller behandlingen av opplysninger endres. Datoen øverst viser gjeldende versjon. En kopi av erklæringen følger med appen og kan leses uten nett i dialogen Om. Gjeldende nettversjon finnes på [smeggtuner.com/nb/privacy](https://smeggtuner.com/nb/privacy).
