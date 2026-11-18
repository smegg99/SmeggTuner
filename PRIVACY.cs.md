Poslední aktualizace: 12. listopadu 2026

## Kdo za zpracování odpovídá

SmeggTuner vyvíjí smegg99. Tyto zásady vysvětlují, jak s údaji nakládá aplikace SmeggTuner pro Android (balíček `me.smegg.smeggtuner`), její verze pro počítač a prohlížeč a web smeggtuner.com. Dostupné funkce se podle verze liší.

S dotazy nebo žádostmi týkajícími se soukromí se obracejte na smegg99 na adrese [support@smegg.me](mailto:support@smegg.me). Ke kontaktování vývojáře nepotřebujete účet v aplikaci.

## Mikrofon a zvukové soubory

SmeggTuner ze signálu mikrofonu přímo ve vašem zařízení vypočítává výšku tónu, frekvence jazýčků a zázněje. V systému Android žádá o oprávnění používat mikrofon. V nastavení aplikací v systému Android můžete toto oprávnění odmítnout nebo odebrat. Ladění naživo pomocí mikrofonu pak nebude dostupné. Otevření zvukového souboru přístup k mikrofonu nevyžaduje.

Analýza zvuku probíhá místně. SmeggTuner neodesílá zvuk z mikrofonu, importované zvukové soubory ani výsledky měření vývojáři, analytické službě ani reklamní službě. Živý zvuk se pro analýzu uchovává v paměti. Spuštěním ladičky se automaticky neukládá nahrávka z mikrofonu. V systému Android se snímání zvuku i přehrávání generovaných tónů zastaví, jakmile aplikace opustí popředí.

Když vyberete soubor WAV, SmeggTuner jej načte pro přehrávání a analýzu. Který soubor smí aplikace otevřít, určuje systémový dialog pro výběr dokumentu. Vybraný poskytovatel cloudového úložiště může soubor stáhnout podle vlastních pravidel. Dočasné místní kopie mohou zůstat v mezipaměti aplikace, dokud ji nevymažete vy nebo systém Android.

## Údaje uložené v zařízení

Nastavení jazyka, motivu, referenční frekvence a kalibrace zvuku se ukládají místně. Místní databáze může obsahovat také uložené výsledky práce z verzí, které ukládání ladění podporují.

Technické protokoly mohou obsahovat chyby aplikace, informace o zařízení nebo jeho zvukovém vybavení, nastavení a cesty k souborům. Protokoly včetně systémového protokolu Androidu zůstávají v zařízení. Aplikace je automaticky neodesílá ani neposílá vývojáři hlášení o pádech. Pokud se rozhodnete poslat podpoře protokol nebo snímek obrazovky, nejprve jej zkontrolujte a odstraňte vše, co nechcete sdílet.

Verze pro Android používá soukromé úložiště aplikace. Pro místní databázi ani soubory nezajišťuje další vrstvu šifrování. Jejich ochrana závisí také na nastavení zabezpečení zařízení. Zálohování v systému Android a přenos mezi zařízeními jsou povoleny. V závislosti na verzi Androidu, zařízení a nastavení zálohování může systém zkopírovat data aplikace k poskytovateli záloh nebo do jiného zařízení. Vývojář tyto zálohy nedostává. Spravovat je můžete v zařízení a u svého poskytovatele záloh.

## Síťová připojení, účty a třetí strany

Po instalaci funguje ladička pro Android bez účtu i bez připojení k internetu. Neobsahuje reklamní SDK, sledování využití aplikace pro analytické účely, registraci účtu ani automatické odesílání údajů o ladění. Nežádá o přístup ke kontaktům, poloze, fotoaparátu, knihovně fotografií ani reklamnímu identifikátoru. Výběrem souboru zpřístupníte pouze vybraný dokument, nikoli celé úložiště.

Verze pro Android na rozdíl od verze pro počítač nevyhledává aktualizace na GitHubu. Instalaci a aktualizace vydání z Google Play zajišťuje Google Play podle pravidel společnosti Google. Google nebo výrobce zařízení mohou podle vašich nastavení a svých zásad nezávisle zpracovávat informace o obchodě, operačním systému, zálohování nebo diagnostice.

Volitelná kontrola aktualizací ve verzi pro počítač kontaktuje GitHub přes HTTPS a předá název a verzi aplikace, aby vyhledala nová vydání. GitHub získá informace o připojení, například vaši IP adresu. Kontrolu můžete vypnout v Nastavení. Požadavek neobsahuje zvukové soubory ani soubory s údaji o ladění.

Odkazy na online zásady, web a GitHub můžete otevřít v externím prohlížeči. Samotné zásady lze číst i bez připojení přímo v aplikaci. Odkaz na e-mail podpory otevře vaši poštovní aplikaci. Zpráva se odešle pouze na váš pokyn. Cílový web obdrží obvyklé údaje o webovém požadavku včetně IP adresy a informací o prohlížeči a řídí se vlastními zásadami. Příspěvky zpracovává poskytovatel, na kterého odkaz vede. SmeggTuner nezískává údaje o vaší platební kartě. Pokud daná verze umožňuje exportovat nebo tisknout dokument, obdrží jej zvolené cílové umístění nebo tisková služba.

## Web a verze pro prohlížeč

Při návštěvě smeggtuner.com nebo načtení verze pro prohlížeč se odesílají požadavky na hostingovou infrastrukturu. Údaje o připojení a zabezpečení mohou zahrnovat IP adresu, požadovanou URL, čas, informace o prohlížeči a chyby. Poskytovatelé hostingu a zabezpečení tyto údaje zpracovávají za účelem provozu služby a předcházení zneužití. Webový provoz chrání společnost Cloudflare. Zpracování údajů popisují její [zásady ochrany soukromí](https://www.cloudflare.com/privacypolicy/).

Je-li zapnuta analytika, používá web statistiky návštěvnosti Umami bez souborů cookie. Mohou zahrnovat navštívené stránky, odkazující weby, kategorie prohlížečů a zařízení, přibližnou polohu, kliknutí na označené odkazy a výkon stránek. Umami rozlišuje návštěvy pomocí údajů o připojení, které zpracovává i bez souborů cookie. Analytické požadavky se obsluhují prostřednictvím tohoto webu. Tyto statistiky jsou oddělené od aplikace pro Android a neobsahují zvuk z mikrofonu ani měření ladění.

Web ukládá volbu motivu do úložiště prohlížeče a pro volbu jazyka může použít soubor cookie. Kontaktní stránka používá Cloudflare Turnstile k omezení automatizovaných požadavků. Zpracovává přitom údaje o prohlížeči a připojení, jak vysvětluje [oznámení o ochraně soukromí služby Turnstile](https://www.cloudflare.com/turnstile-privacy-policy/). Pro přečtení těchto zásad nemusíte ověření na kontaktní stránce absolvovat.

Po udělení oprávnění v prohlížeči zpracovává webová ladička zvuk z mikrofonu místně v prohlížeči. Zvuk neodesílá k analýze. V nastavení oprávnění prohlížeče a úložiště webu můžete přístup odebrat a smazat místní nastavení.

## Podpora, doba uchování a výmaz

Pokud kontaktujete vývojáře, vámi zaslané údaje, vaše e-mailová adresa a korespondence se použijí k odpovědi na žádost a k prošetření nahlášených problémů. Nahrávky, osobní údaje ani citlivé dokumenty neposílejte, pokud nejsou pro vyřízení žádosti nutné. Údaje se neprodávají ani nepoužívají k cílené reklamě.

Místní nastavení a uložené výsledky práce zůstávají v zařízení, dokud je neodstraníte, nevymažete úložiště aplikace nebo aplikaci neodinstalujete. V systému Android můžete data aplikace vymazat v nabídce Nastavení > Aplikace > SmeggTuner > Úložiště > Vymazat úložiště. Názvy položek se podle zařízení liší. Vymazáním mezipaměti se odstraní dočasné soubory, ale neobnoví se všechna nastavení. Odinstalováním aplikace ani vymazáním jejích dat neodstraníte soubory exportované jinam, kopie zaslané dalším osobám ani zálohy u jiné služby. Tyto kopie spravujte samostatně.

Korespondence s podporou se uchovává po dobu potřebnou k vyřízení žádosti, dalšímu řešení problému nebo splnění příslušných zákonných povinností. Provozní a bezpečnostní záznamy webu se uchovávají pro provoz služby, prevenci a šetření zneužití. Statistiky návštěvnosti slouží k pochopení způsobu používání webu a jeho zlepšování. Doba uchování závisí na těchto účelech a nastavení příslušné služby. Na výše uvedené kontaktní adrese se můžete vývojáře zeptat na údaje, které uchovává, nebo požádat o jejich výmaz. Vývojář nemůže vzdáleně přistupovat k souborům uloženým pouze ve vašem zařízení ani je odstranit.

## Vaše možnosti a práva

Pokud se na zpracování vztahují předpisy o ochraně osobních údajů, můžete mít právo na přístup ke svým osobním údajům, jejich opravu či výmaz, omezení zpracování, vznesení námitky proti zpracování a získání přenositelné kopie. Můžete také podat stížnost u příslušného místního dozorového úřadu. Pro uplatnění těchto práv kontaktujte vývojáře. Aby údaje nebyly zpřístupněny někomu jinému, může být nutné ověření totožnosti.

Zpracování potřebné k provozu a zabezpečení webu či k vyřízení žádostí vychází z účelu příslušné služby a z oprávněných zájmů v rozsahu povoleném zákonem. Údaje mohou být uchovávány také pro plnění zákonných povinností. Poskytovatelé služeb mohou údaje zpracovávat i mimo vaši zemi za příslušných podmínek a s příslušnými ochrannými opatřeními. Aplikace nevytváří uživatelský účet, takže u SmeggTuner není žádný účet k odstranění.

## Změny

Při změnách příslušných funkcí nebo způsobu nakládání s údaji bude tato stránka aktualizována. Datum výše označuje aktuální verzi. Kopie zásad je součástí aplikace a lze ji číst offline v dialogu O aplikaci. Aktuální webová verze je k dispozici na [smeggtuner.com/cs/privacy](https://smeggtuner.com/cs/privacy).
