Päivitetty viimeksi 12. marraskuuta 2026

## Kuka vastaa tietojen käsittelystä

SmeggTunerin kehittäjä on smegg99. Tässä tietosuojakäytännössä kuvataan, miten SmeggTunerin Android-sovellus (paketti `me.smegg.smeggtuner`), työpöytä- ja selainversiot sekä smeggtuner.com käsittelevät tietoja. Ominaisuudet vaihtelevat versioittain.

Tietosuojaa koskevat kysymykset ja pyynnöt voi lähettää kehittäjälle smegg99 osoitteeseen [support@smegg.me](mailto:support@smegg.me). Yhteydenotto ei edellytä sovelluksen käyttäjätiliä.

## Mikrofoni ja äänitiedostot

SmeggTuner laskee mikrofonin äänestä sävelkorkeuden, kielten taajuudet ja huojunnan omalla laitteellasi. Androidissa sovellus pyytää mikrofonin käyttöoikeutta. Voit evätä oikeuden tai perua sen Androidin sovellusasetuksissa. Tällöin reaaliaikainen viritys mikrofonilla ei ole käytettävissä. Äänitiedoston avaaminen ei vaadi mikrofonin käyttöoikeutta.

Ääni analysoidaan paikallisesti. SmeggTuner ei lähetä mikrofonin ääntä, tuotuja äänitiedostoja eikä viritysmittauksia kehittäjälle, analytiikkapalveluun tai mainospalveluun. Reaaliaikaista ääntä pidetään työmuistissa analysointia varten. Virittimen käynnistäminen ei tallenna mikrofonista automaattisesti äänitettä. Androidissa äänen kaappaus ja tuotettujen äänien toisto lakkaavat, kun sovellus siirtyy pois etualalta.

Kun valitset WAV-tiedoston, SmeggTuner lukee sen toistoa ja analysointia varten. Androidin tiedostovalitsin määrää, mihin tiedostoon sovellus saa pääsyn. Valitsemasi pilvitallennuspalvelu voi ladata tiedoston laitteelle omien käytäntöjensä mukaisesti. Tilapäiset paikalliset kopiot voivat jäädä sovelluksen välimuistiin, kunnes Android tai sinä tyhjennät sen.

## Laitteelle tallennetut tiedot

Sovellus säilyttää paikallisesti asetuksia, kuten kielen, teeman, vertailutaajuuden ja äänen kalibroinnin. Sen paikallisessa tietokannassa voi lisäksi olla tallennettuja viritystöitä sellaisista versioista, joissa tallennus on käytettävissä.

Tekniset lokit voivat sisältää sovellusvirheitä, laitteen tai äänilaitteen tietoja, asetuksia ja tiedostopolkuja. Lokit, myös Androidin järjestelmäloki, säilyvät paikallisesti. Sovellus ei automaattisesti lähetä lokeja verkkoon eikä lähetä kaatumisraportteja kehittäjälle. Jos päätät lähettää lokin tai kuvakaappauksen tukeen, tarkista se ensin ja poista kaikki tiedot, joita et halua jakaa.

Android-sovellus käyttää sovelluksen yksityistä tallennustilaa. Se ei lisää paikalliselle tietokannalle tai tiedostoille erillistä salauskerrosta. Suojaus riippuu myös laitteesi tietoturva-asetuksista. Androidin varmuuskopiointi ja siirto toiseen laitteeseen ovat käytössä. Android-versiosta, laitteesta ja varmuuskopiointiasetuksista riippuen käyttöjärjestelmä voi kopioida sovelluksen tietoja varmuuskopiointipalveluusi tai toiseen laitteeseen. Kehittäjä ei saa näitä varmuuskopioita. Voit hallita niitä laitteesi ja varmuuskopiointipalvelusi kautta.

## Verkkoyhteydet, tilit ja kolmannet osapuolet

Asennuksen jälkeen Android-viritin toimii ilman tiliä ja internetyhteyttä. Sovelluksessa ei ole mainonnan kehityspakettia (SDK), käyttäjätoimintaa seuraavaa analytiikkaa, tilin rekisteröintiä eikä viritystietojen automaattista lähetystä verkkoon. Se ei pyydä pääsyä yhteystietoihin, sijaintiin, kameraan, kuvakirjastoon eikä mainostunnisteeseen. Tiedoston valinta antaa pääsyn vain kyseiseen asiakirjaan, ei koko tallennustilaan.

Android-versio ei tee työpöytäversion GitHub-päivitystarkistuksia. Google Play hallinnoi Play-version asennusta ja päivityksiä Googlen käytäntöjen mukaisesti. Google tai laitevalmistaja voi itsenäisesti käsitellä sovelluskauppaan, käyttöjärjestelmään, varmuuskopiointiin tai diagnostiikkaan liittyviä tietoja asetustesi ja omien käytäntöjensä mukaisesti.

Työpöydällä valinnainen päivitystarkistus ottaa yhteyden GitHubiin HTTPS-yhteydellä ja välittää sovelluksen nimen ja version etsiessään julkaisuja. GitHub saa tietoja verkkoyhteydestäsi, kuten IP-osoitteesi. Tarkistukset voi poistaa käytöstä asetuksissa. Pyyntö ei sisällä ääni- eikä viritystiedostoja.

Valinnaiset linkit tietosuojakäytännön verkkoversioon, verkkosivustolle ja GitHubiin avautuvat ulkoiseen selaimeen. Itse käytännön voi lukea sovelluksessa ilman yhteyttä. Tuen sähköpostilinkit avaavat sähköpostisovelluksesi. Viesti lähtee vain, jos päätät lähettää sen. Verkkolinkin kohde saa tavalliset verkkopyynnön tiedot, kuten IP-osoitteesi ja selaimen tiedot, ja soveltaa omaa tietosuojakäytäntöään. Lahjoitukset käsittelee linkitetty palveluntarjoaja. SmeggTuner ei saa maksukorttisi tietoja. Jos versio tukee asiakirjan vientiä tai tulostusta, valitsemasi kohde tai tulostuspalvelu saa asiakirjan.

## Verkkosivusto ja selainversio

Kun käyt sivustolla smeggtuner.com tai lataat selainversion, laite lähettää pyyntöjä sivuston ylläpitoinfrastruktuurille. Yhteys- ja turvallisuustietoihin voi sisältyä IP-osoite, pyydetty URL-osoite, ajankohta, selaimen tietoja ja virheitä. Ylläpito- ja tietoturvapalvelujen tarjoajat käsittelevät tietoja palvelun toimittamiseksi ja väärinkäytösten estämiseksi. Cloudflare suojaa verkkoliikennettä. Sen tietojen käsittely kuvataan [Cloudflaren tietosuojakäytännössä](https://www.cloudflare.com/privacypolicy/).

Kun analytiikka on käytössä, sivusto kerää evästeettömiä Umami-tilastoja. Niihin voi sisältyä vieraillut sivut, viittaavat sivustot, selain- ja laiteluokat, likimääräinen sijainti, merkittyjen linkkien napsautukset ja sivujen suorituskyky. Umami käsittelee verkkoyhteyden tietoja erottaakseen käynnit toisistaan. Evästeettömyys ei tarkoita, ettei tietoja käsitellä. Analytiikan päätepiste toimii tämän sivuston kautta. Sivuston tilastot ovat erillään Android-sovelluksesta eivätkä sisällä mikrofonin ääntä tai viritysmittauksia.

Sivusto tallentaa teemavalintasi selaimen tallennustilaan ja voi käyttää kielivalintaevästettä. Yhteystietosivulla Cloudflare Turnstile rajoittaa automaattisia pyyntöjä. Se käsittelee selaimen ja yhteyden signaaleja [Turnstilen tietosuojailmoituksessa](https://www.cloudflare.com/turnstile-privacy-policy/) kuvatulla tavalla. Tämän tietosuojakäytännön lukeminen ei edellytä yhteystietosivun tarkistuksen suorittamista.

Selainviritin käsittelee mikrofonin äänen paikallisesti selaimessasi, kun olet antanut selaimelle luvan. Ääntä ei lähetetä verkkoon analysoitavaksi. Selaimen käyttöoikeuksien ja sivuston tallennustilan asetuksilla voit perua käyttöoikeuden ja tyhjentää paikalliset valinnat.

## Tuki, säilytys ja poistaminen

Kun otat yhteyttä kehittäjään, lähettämiäsi tietoja, sähköpostiosoitettasi ja kirjeenvaihtoasi käytetään pyyntöösi vastaamiseen ja ilmoittamiesi ongelmien selvittämiseen. Älä lähetä äänitteitä, henkilötietoja tai arkaluonteisia asiakirjoja, elleivät ne ole pyyntösi käsittelylle tarpeen. Tietoja ei myydä eikä käytetä kohdennettuun mainontaan.

Paikalliset asetukset ja tallennetut työt säilyvät, kunnes poistat ne, tyhjennät sovelluksen tallennustilan tai poistat sovelluksen. Androidissa voit nollata sovelluksen tiedot valitsemalla Asetukset > Sovellukset > SmeggTuner > Tallennustila > Tyhjennä tallennustila. Nimet vaihtelevat laitteittain. Välimuistin tyhjennys poistaa tilapäistiedostot, mutta ei nollaa kaikkia asetuksia. Sovelluksen tai sen tietojen poistaminen ei poista muualle vietyjä tiedostoja, toiselle henkilölle lähetettyjä kopioita eikä muiden palvelujen säilyttämiä varmuuskopioita. Hallitse niitä erikseen.

Tuen kirjeenvaihtoa säilytetään niin kauan kuin on tarpeen pyyntöön vastaamiseksi, asian jatkotoimia varten tai sovellettavien lakisääteisten velvoitteiden täyttämiseksi. Sivuston toiminta- ja tietoturvalokeja säilytetään palvelun tuottamista, väärinkäytösten estämistä ja niiden selvittämistä varten. Kävijätilastoja säilytetään sivuston käytön ymmärtämiseksi ja sivuston kehittämiseksi. Säilytysaika riippuu näistä tarkoituksista ja palvelujen asetuksista. Voit kysyä kehittäjän hallussa olevista tiedoista tai pyytää niiden poistamista ottamalla yhteyttä yllä mainittuun tietosuojaosoitteeseen. Kehittäjä ei voi etäyhteydellä käyttää tai poistaa vain laitteellasi olevia tiedostoja.

## Valintasi ja oikeutesi

Kun tietosuojalainsäädäntöä sovelletaan, sinulla voi olla oikeus tutustua henkilötietoihisi, oikaista tai poistaa niitä, rajoittaa tai vastustaa niiden käsittelyä ja saada tiedoista siirrettävä kopio. Voit myös tehdä valituksen paikalliselle tietosuojaviranomaiselle. Ota yhteys kehittäjään käyttääksesi oikeuksiasi. Henkilöllisyys voidaan joutua varmistamaan, jotta tietoja ei luovuteta sivulliselle.

Sivuston tarjoamiseen ja suojaamiseen tai pyyntöihin vastaamiseen tarvittava käsittely perustuu kyseisen palvelun tarkoitukseen ja oikeutettuihin etuihin silloin, kun laki sen sallii. Tietoja voidaan säilyttää myös lakisääteisten velvoitteiden täyttämiseksi. Palveluntarjoajat voivat käsitellä tietoja asuinmaasi ulkopuolella sovellettavien ehtojensa ja suojatoimiensa mukaisesti. Sovellukseen ei luoda käyttäjätiliä, joten poistettavaa SmeggTuner-tiliä ei ole.

## Muutokset

Tätä sivua päivitetään, kun asiaan liittyvät ominaisuudet tai tietojen käsittelykäytännöt muuttuvat. Yllä oleva päivämäärä kertoo nykyisen version. Sovelluksen mukana toimitettu kopio on luettavissa ilman verkkoyhteyttä Tietoja-ikkunassa. Nykyinen verkkoversio on osoitteessa [smeggtuner.com/fi/privacy](https://smeggtuner.com/fi/privacy).
