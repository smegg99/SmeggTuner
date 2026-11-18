// Explicit imports keep Vite from walking the entire repository to match a root glob.
import type { LocaleCode } from '../types/locale'
import cs from '../../../PRIVACY.cs.md?raw'
import da from '../../../PRIVACY.da.md?raw'
import de from '../../../PRIVACY.de.md?raw'
import el from '../../../PRIVACY.el.md?raw'
import en from '../../../PRIVACY.md?raw'
import es from '../../../PRIVACY.es.md?raw'
import et from '../../../PRIVACY.et.md?raw'
import fi from '../../../PRIVACY.fi.md?raw'
import fr from '../../../PRIVACY.fr.md?raw'
import hu from '../../../PRIVACY.hu.md?raw'
import it from '../../../PRIVACY.it.md?raw'
import ja from '../../../PRIVACY.ja.md?raw'
import ko from '../../../PRIVACY.ko.md?raw'
import lt from '../../../PRIVACY.lt.md?raw'
import lv from '../../../PRIVACY.lv.md?raw'
import nb from '../../../PRIVACY.nb.md?raw'
import nl from '../../../PRIVACY.nl.md?raw'
import pl from '../../../PRIVACY.pl.md?raw'
import ptBR from '../../../PRIVACY.pt-BR.md?raw'
import ptPT from '../../../PRIVACY.pt-PT.md?raw'
import ro from '../../../PRIVACY.ro.md?raw'
import ru from '../../../PRIVACY.ru.md?raw'
import sk from '../../../PRIVACY.sk.md?raw'
import sv from '../../../PRIVACY.sv.md?raw'
import th from '../../../PRIVACY.th.md?raw'
import zhHans from '../../../PRIVACY.zh-Hans.md?raw'
import zhHant from '../../../PRIVACY.zh-Hant.md?raw'

export const policies: Record<LocaleCode, string> = {
  cs, da, de, el, en, es, et, fi, fr, hu, it, ja, ko, lt, lv, nb, nl, pl, ro, ru, sk, sv, th,
  'pt-BR': ptBR, 'pt-PT': ptPT, 'zh-Hans': zhHans, 'zh-Hant': zhHant,
}
