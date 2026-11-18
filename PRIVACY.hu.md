Utolsó frissítés: 2026. november 12.

## Ki felel az adatkezelésért?

A SmeggTuner fejlesztője smegg99. Ez a tájékoztató ismerteti, hogyan kezeli az adatokat a SmeggTuner androidos alkalmazása (`me.smegg.smeggtuner` csomag), az asztali és a böngészős kiadás, valamint a smeggtuner.com webhely. Az elérhető funkciók kiadásonként eltérnek.

Adatvédelmi kérdéssel vagy kéréssel a [support@smegg.me](mailto:support@smegg.me) címen fordulhat smegg99-hez. A fejlesztő megkereséséhez nincs szüksége alkalmazásfiókra.

## Mikrofon és hangfájlok

A SmeggTuner a mikrofon jeléből az Ön eszközén számítja ki a hangmagasságot, a sípok frekvenciáját és a lebegést. Androidon engedélyt kér a mikrofon használatára. Az Android alkalmazásbeállításaiban megtagadhatja vagy visszavonhatja ezt az engedélyt. Ilyenkor az élő mikrofonos hangolás nem érhető el. Hangfájl megnyitásához nincs szükség mikrofonengedélyre.

A hangelemzés helyben történik. A SmeggTuner nem küld mikrofonhangot, importált hangfájlokat vagy hangolási méréseket a fejlesztőnek, analitikai vagy hirdetési szolgáltatásnak. Az elemzéshez az élő hang a memóriában marad. A hangoló elindítása nem ment automatikusan mikrofonfelvételt. Androidon a hangrögzítés és a generált hangok lejátszása leáll, amikor az alkalmazás a háttérbe kerül.

WAV-fájl kiválasztásakor a SmeggTuner beolvassa a fájlt lejátszáshoz és elemzéshez. Az Android dokumentumválasztója határozza meg, melyik fájlhoz férhet hozzá az alkalmazás. Ha felhőtárhely-szolgáltatót választ, az a saját eljárásai szerint letöltheti a fájlt. Ideiglenes helyi másolatok maradhatnak az alkalmazás gyorsítótárában mindaddig, amíg az Android vagy Ön ki nem üríti azt.

## Az eszközön tárolt adatok

Az alkalmazás helyben tárolja az olyan beállításokat, mint a nyelv, a téma, a referenciahang és a hangkalibrálás. A helyi adatbázis olyan kiadásokból származó mentett hangolási munkákat is tartalmazhat, amelyek támogatják ezt a funkciót.

A műszaki naplók alkalmazáshibákat, az eszköz vagy a hangeszköz adatait, beállításokat és fájlútvonalakat tartalmazhatnak. A naplók, köztük az Android rendszernaplója, helyben maradnak. Az alkalmazás nem tölti fel automatikusan a naplókat, és nem küld összeomlási jelentést a fejlesztőnek. Ha segítségkéréshez naplót vagy képernyőképet küld, előbb nézze át, és távolítson el belőle mindent, amit nem szeretne megosztani.

Az androidos alkalmazás a saját, elkülönített tárhelyét használja. Helyi adatbázisához és fájljaihoz nem biztosít további titkosítási réteget. A védelem az eszköz biztonsági beállításaitól is függ. Az androidos biztonsági mentés és az eszközök közötti átvitel engedélyezett. Az Android verziójától, az eszköztől és a mentési beállításoktól függően az operációs rendszer az alkalmazás adatait a biztonsági mentést kezelő szolgáltatóhoz vagy másik eszközre másolhatja. Ezeket a mentéseket a fejlesztő nem kapja meg. Az eszközön és a mentési szolgáltatónál kezelheti őket.

## Hálózati kapcsolatok, fiókok és külső szolgáltatók

Az androidos hangoló telepítés után fiók és internetkapcsolat nélkül is működik. Nincs benne hirdetési SDK, az alkalmazáshasználatot elemző követő, fiókregisztráció vagy a hangolási adatok automatikus feltöltésére szolgáló funkció. Nem kér hozzáférést a névjegyekhez, a helyadatokhoz, a kamerához, a fotótárhoz vagy a hirdetési azonosítóhoz. Fájl kiválasztásával csak a kiválasztott dokumentumhoz ad hozzáférést, a teljes tárhelyhez nem.

Az androidos kiadás nem végez az asztali kiadáséhoz hasonló frissítésellenőrzést a GitHubon. A Play-kiadás telepítését és frissítéseit a Google Play kezeli a Google szabályzatai szerint. A Google vagy az eszköz gyártója az Ön beállításai és saját szabályzata alapján önállóan kezelhet áruházi, operációs rendszerrel kapcsolatos, biztonsági mentési vagy diagnosztikai adatokat.

Asztali gépen a választható frissítésellenőrző HTTPS-kapcsolaton keresztül az alkalmazás nevével és verziószámával kérdezi le az elérhető kiadásokat a GitHubtól. A GitHub kapcsolati adatokat kap, például az Ön IP-címét. Az ellenőrzéseket a Beállításokban kikapcsolhatja. Hang- és hangolási fájlok nem részei a kérésnek.

Az online tájékoztatóhoz, a webhelyhez és a GitHubhoz tartozó hivatkozások külső böngészőt nyitnak meg, ha rájuk kattint. Maga a tájékoztató az alkalmazásban internetkapcsolat nélkül is olvasható. A támogatás e-mail-hivatkozása megnyitja a levelezőprogramját. Az üzenet csak akkor megy el, ha Ön elküldi. A felkeresett webhely megkapja a szokásos webes kérésadatokat, köztük az IP-címet és a böngésző adatait, és saját szabályzata szerint jár el. Az adományokat a hivatkozott szolgáltató kezeli. A SmeggTuner nem kapja meg a bankkártyaadatait. Ha egy kiadásban dokumentumot exportál vagy nyomtat, a dokumentumot az Ön által választott célhely vagy nyomtatási szolgáltatás kapja meg.

## A webhely és a böngészős kiadás

A smeggtuner.com felkeresése és a böngészős kiadás betöltése kéréseket küld a kiszolgáló infrastruktúrának. A kapcsolati és biztonsági adatok közé tartozhat az IP-cím, a kért URL, az időpont, a böngésző adatai és az esetleges hibák. A tárhely- és biztonsági szolgáltatók ezeket az adatokat a szolgáltatás biztosítására és a visszaélések megelőzésére dolgozzák fel. A webforgalom védelmét a Cloudflare biztosítja. Adatkezeléséről a [Cloudflare adatvédelmi tájékoztatójában](https://www.cloudflare.com/privacypolicy/) olvashat.

Ha az analitika engedélyezve van, a webhely az Umami sütik nélküli látogatási statisztikáit használja. Ezek tartalmazhatják a megtekintett oldalakat, a hivatkozó webhelyeket, a böngésző és az eszköz kategóriáját, a hozzávetőleges helyet, a megjelölt hivatkozásokra történő kattintásokat és az oldalak teljesítményadatait. Az Umami kapcsolati adatokat dolgoz fel a látogatások megkülönböztetéséhez. A sütik mellőzése tehát nem jelenti azt, hogy nincs adatkezelés. Az analitikai végpont ezen a webhelyen keresztül érhető el. Ezek a webes statisztikák elkülönülnek az androidos alkalmazástól, és nem tartalmaznak mikrofonhangot vagy hangolási méréseket.

A webhely a böngésző tárhelyén tárolja a választott témát, és sütit használhat a nyelvi beállítás megjegyzésére. A kapcsolatfelvételi oldal a Cloudflare Turnstile szolgáltatással korlátozza az automatizált kéréseket. A szolgáltatás a böngészőből és a kapcsolatból származó jeleket a [Turnstile adatvédelmi tájékoztatójában](https://www.cloudflare.com/turnstile-privacy-policy/) leírtak szerint dolgozza fel. E tájékoztató olvasásához nem kell elvégeznie a kapcsolatfelvételi oldalon megjelenő ellenőrzést.

A böngészős hangoló az engedély megadása után helyben, a böngészőben dolgozza fel a mikrofon hangját. A hangot nem tölti fel elemzésre. A böngésző engedélybeállításaiban visszavonhatja a hozzáférést, a webhelytárhely beállításaiban pedig törölheti a helyi beállításokat.

## Támogatás, megőrzés és törlés

Ha kapcsolatba lép a fejlesztővel, az Ön által elküldött adatokat, e-mail-címét és levelezését a kérés megválaszolására és a jelzett problémák kivizsgálására használja fel. Csak akkor küldjön felvételt, személyes adatot vagy érzékeny dokumentumot, ha az szükséges a kérése teljesítéséhez. Az adatokat nem értékesítik, és nem használják célzott hirdetéshez.

A helyi beállítások és a mentett munkák addig maradnak meg, amíg nem törli őket, nem üríti ki az alkalmazás tárhelyét, vagy nem távolítja el az alkalmazást. Androidon az alkalmazás adatainak alaphelyzetbe állításához válassza a Beállítások > Alkalmazások > SmeggTuner > Tárhely > Tárhely törlése menüpontot. Az elnevezések eszközönként eltérhetnek. A gyorsítótár ürítése eltávolítja az ideiglenes fájlokat, de nem állít alaphelyzetbe minden beállítást. Az eltávolítás vagy az alkalmazás adatainak törlése nem törli a máshová exportált fájlokat, másoknak elküldött másolatokat vagy más szolgáltatónál tárolt mentéseket. Ezeket külön kezelje.

A támogatással folytatott levelezést a kérés megválaszolásához, a probléma utókövetéséhez vagy a vonatkozó jogi kötelezettségek teljesítéséhez szükséges ideig őrzik meg. A webhely üzemeltetési és biztonsági naplóit a szolgáltatás működtetése, valamint a visszaélések megelőzése és kivizsgálása érdekében őrzik. A látogatási statisztikák megőrzése a webhely használatának megértését és a webhely fejlesztését szolgálja. A megőrzés időtartama ezektől a céloktól és az adott szolgáltatás beállításaitól függ. A fenti adatvédelmi elérhetőségen érdeklődhet a fejlesztőnél tárolt adatokról, illetve kérheti törlésüket. A fejlesztő nem férhet hozzá távolról a kizárólag az Ön eszközén lévő fájlokhoz, és nem törölheti őket.

## Az Ön lehetőségei és jogai

Ha az adatvédelmi jogszabályok alkalmazandók, Ön jogosult lehet személyes adataihoz hozzáférni, azok helyesbítését, törlését vagy kezelésük korlátozását kérni, az adatkezelés ellen tiltakozni, valamint az adatokról hordozható másolatot kapni. Panaszt tehet a helyi adatvédelmi hatóságnál is. Jogai gyakorlásához forduljon a fejlesztőhöz. Személyazonossága ellenőrzésére szükség lehet annak elkerüléséhez, hogy az adatokat másnak adják ki.

A webhely működtetéséhez és védelméhez, valamint a kérések megválaszolásához szükséges adatkezelés az adott szolgáltatási célon és jogos érdekeken alapul, amennyiben ezt a jog megengedi. Jogi kötelezettségek teljesítéséhez is szükség lehet adatok megőrzésére. A szolgáltatók az Ön országától eltérő országokban is kezelhetnek adatokat a rájuk vonatkozó feltételek és garanciák mellett. Az alkalmazásban nem jön létre fiók, így nincs törölhető SmeggTuner-fiók.

## Változások

Ez az oldal az érintett funkciók vagy adatkezelési gyakorlatok változásakor frissül. A fenti dátum jelöli az aktuális változatot. A tájékoztató másolata az alkalmazás része, a Névjegy ablakból internetkapcsolat nélkül is olvasható. Az aktuális webes változat a [smeggtuner.com/hu/privacy](https://smeggtuner.com/hu/privacy) oldalon érhető el.
