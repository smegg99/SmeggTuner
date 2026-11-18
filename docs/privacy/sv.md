Senast uppdaterad den 12 november 2026

## Vem ansvarar?

SmeggTuner utvecklas av smegg99. Den här policyn beskriver hur Android-appen SmeggTuner (paketet `me.smegg.smeggtuner`), skrivbords- och webbläsarversionerna samt smeggtuner.com hanterar uppgifter. Funktionerna skiljer sig mellan versionerna.

Vid frågor eller begäranden som rör integritet kan du kontakta smegg99 på [support@smegg.me](mailto:support@smegg.me). Du behöver inget konto i appen för att kontakta utvecklaren.

## Mikrofon och ljudfiler

SmeggTuner använder mikrofonljudet för att beräkna tonhöjd, stämtungornas frekvenser och svävningar på din enhet. På Android begär appen mikrofonbehörighet. Du kan neka eller återkalla behörigheten i Androids appinställningar. Då går det inte att stämma direkt med mikrofonen. Du kan öppna en ljudfil utan mikrofonbehörighet.

Ljudet analyseras lokalt. SmeggTuner laddar inte upp mikrofonljud, importerade ljudfiler eller stämningsmätningar till utvecklaren, en analystjänst eller en annonstjänst. Ljudet från mikrofonen hålls i arbetsminnet för analys. Ingen mikrofoninspelning sparas automatiskt när du startar stämapparaten. På Android upphör upptagningen och uppspelningen av genererade toner när appen lämnar förgrunden.

När du väljer en WAV-fil läser SmeggTuner filen för uppspelning och analys. Androids dokumentväljare styr vilken fil appen får tillgång till. Om du väljer en molnlagringstjänst kan den ladda ned filen enligt sina egna rutiner. Tillfälliga lokala kopior kan ligga kvar i appens cache tills Android eller du själv tömmer den.

## Uppgifter som lagras på enheten

Appen sparar inställningar som språk, tema, referenston och ljudkalibrering lokalt. Den lokala databasen kan även innehålla sparat stämningsarbete från versioner med sådana funktioner.

Tekniska loggar kan innehålla appfel, uppgifter om enheten eller ljudenheten, inställningar och filsökvägar. Loggarna lagras lokalt, även Androids systemlogg. Appen laddar inte automatiskt upp dessa loggar eller skickar kraschrapporter till utvecklaren. Om du väljer att skicka en logg eller skärmbild för att få hjälp, granska den först och ta bort det du inte vill dela.

Android-appen använder sitt privata lagringsutrymme. Den lägger inte till något extra krypteringslager för sin lokala databas eller sina filer. Skyddet beror även på enhetens säkerhetsinställningar. Säkerhetskopiering och överföring mellan enheter är aktiverade i Android. Beroende på Android-version, enhet och inställningar för säkerhetskopiering kan operativsystemet kopiera appdata till din säkerhetskopieringstjänst eller en annan enhet. Utvecklaren får inte dessa säkerhetskopior. Du kan hantera dem via enheten och din säkerhetskopieringstjänst.

## Nätverksanslutningar, konton och tredje parter

Efter installationen fungerar Android-stämapparaten utan konto eller internetanslutning. Den innehåller inget annons-SDK, ingen spårning för analys i appen, ingen kontoregistrering och ingen automatisk uppladdning av stämningsdata. Den begär inte åtkomst till dina kontakter, din plats, kamera, ditt bildbibliotek eller annons-id. När du väljer en fil ger du bara tillgång till det valda dokumentet.

Android-versionen gör inte skrivbordsversionens uppdateringskontroller mot GitHub. Google Play hanterar installation och uppdatering av Play-versionen enligt Googles policyer. Google eller enhetstillverkaren kan självständigt behandla uppgifter om appbutiken, operativsystemet, säkerhetskopior eller diagnostik enligt dina inställningar och sina egna policyer.

På skrivbordet kan den valfria uppdateringskontrollen kontakta GitHub via HTTPS för att söka efter nya versioner. Begäran innehåller appens namn och version. GitHub får anslutningsuppgifter som din IP-adress. Du kan stänga av kontrollerna i inställningarna. Ljudfiler och stämningsfiler ingår inte i begäran.

Om du väljer en länk till policyn på nätet, webbplatsen eller GitHub öppnas en extern webbläsare. Själva policyn kan läsas i appen utan anslutning. E-postlänkar för support öppnar ditt e-postprogram. Ett meddelande skickas bara om du själv väljer att skicka det. Webbplatsen du besöker tar emot vanliga uppgifter från webbförfrågan, bland annat IP-adress och webbläsaruppgifter, och tillämpar sin egen policy. Donationer hanteras av den länkade leverantören. SmeggTuner får inte dina betalkortsuppgifter. Om en version låter dig exportera eller skriva ut ett dokument får den valda destinationen eller utskriftstjänsten dokumentet.

## Webbplatsen och webbläsarversionen

När du besöker smeggtuner.com eller läser in webbläsarversionen skickas förfrågningar till webbhotellets infrastruktur. Anslutnings- och säkerhetsuppgifter kan omfatta IP-adress, begärd URL, tidpunkt, webbläsaruppgifter och fel. Drift- och säkerhetsleverantörerna behandlar uppgifterna för att tillhandahålla tjänsten och förebygga missbruk. Cloudflare skyddar webbtrafiken och beskriver sin behandling i [Cloudflares integritetspolicy](https://www.cloudflare.com/privacypolicy/).

När besöksanalys är aktiverad använder webbplatsen statistik från Umami utan kakor. Den kan omfatta besökta sidor, hänvisande webbplatser, kategorier av webbläsare och enheter, ungefärlig plats, klick på märkta länkar och sidornas prestanda. Umami behandlar anslutningsuppgifter för att skilja besök åt. Uppgifter behandlas alltså även utan kakor. Analysens mottagningspunkt finns på den här webbplatsen. Webbplatsens statistik är skild från Android-appen och innehåller varken mikrofonljud eller stämningsmätningar.

Webbplatsen sparar ditt temaval i webbläsarens lagring och kan använda en kaka för ditt språkval. Kontaktsidan använder Cloudflare Turnstile för att begränsa automatiserade förfrågningar. Turnstile behandlar signaler från webbläsaren och anslutningen enligt [Turnstiles integritetsmeddelande](https://www.cloudflare.com/turnstile-privacy-policy/). Du behöver inte genomföra kontaktsidans kontroll för att läsa denna policy.

När du har gett ditt tillstånd i webbläsaren behandlar stämapparaten mikrofonljudet lokalt i webbläsaren. Den laddar inte upp ljudet för analys. Med webbläsarens behörigheter och inställningar för webbplatsdata kan du återkalla åtkomsten och rensa lokala inställningar.

## Support, lagring och radering

Om du kontaktar utvecklaren används de uppgifter du väljer att skicka, din e-postadress och korrespondensen för att besvara din fråga och undersöka rapporterade problem. Skicka bara inspelningar, personuppgifter eller känsliga dokument om de behövs för ditt ärende. Uppgifter säljs inte och används inte för riktad reklam.

Lokala inställningar och sparat arbete finns kvar tills du tar bort dem, rensar appens lagring eller avinstallerar appen. På Android återställer du appdata via Inställningar > Appar > SmeggTuner > Lagring > Rensa lagringsutrymme. Namnen varierar mellan enheter. När du rensar cacheminnet tas tillfälliga filer bort utan att alla inställningar återställs. Om du avinstallerar appen eller rensar appdata tas inte filer som exporterats till andra platser, kopior som du skickat till andra personer eller säkerhetskopior hos andra tjänster bort. Hantera dessa kopior separat.

Supportkorrespondens sparas så länge den behövs för att besvara förfrågan, följa upp problemet eller uppfylla tillämpliga rättsliga skyldigheter. Webbplatsens drift- och säkerhetsloggar sparas för drift, förebyggande av missbruk och utredning av incidenter. Besöksstatistik sparas för att förstå hur webbplatsen används och förbättra den. Lagringstiden beror på dessa ändamål och respektive tjänsts inställningar. Via kontaktadressen ovan kan du fråga vilka uppgifter utvecklaren har eller begära att de raderas. Utvecklaren kan inte på distans komma åt eller radera filer som bara finns på din enhet.

## Dina val och rättigheter

Där dataskyddslagstiftning gäller kan du ha rätt att få tillgång till, rätta eller radera dina personuppgifter, begränsa eller invända mot behandlingen samt få en kopia i ett portabelt format. Du kan också lämna klagomål till din lokala dataskyddsmyndighet. Kontakta utvecklaren för att utöva dessa rättigheter. Din identitet kan behöva kontrolleras för att uppgifter inte ska lämnas ut till någon annan.

Behandling som behövs för att tillhandahålla och skydda webbplatsen eller besvara förfrågningar grundas på tjänstens aktuella ändamål och berättigade intressen där lagen tillåter det. Uppgifter kan även sparas för att uppfylla rättsliga skyldigheter. Tjänsteleverantörer kan behandla uppgifter i andra länder än ditt enligt sina tillämpliga villkor och skyddsåtgärder. Det skapas inget appkonto, så det finns inget SmeggTuner-konto att radera.

## Ändringar

Sidan uppdateras när relevanta funktioner eller hanteringen av uppgifter ändras. Datumet ovan anger aktuell version. En kopia av policyn följer med appen och kan läsas offline i dialogrutan Om. Den aktuella webbversionen finns på [smeggtuner.com/sv/privacy](https://smeggtuner.com/sv/privacy).
