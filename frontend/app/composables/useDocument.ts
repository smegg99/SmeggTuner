// What a dialog or a room needs to show one bundled document: its title, and the language it is read in,
// which follows the interface until the document's own language links change it.
import { computed, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'
import type { DocumentKind, DocumentLocale } from '~/utils/documents'

export function useDocument(kind: MaybeRefOrGetter<DocumentKind>) {
  const { t, locale } = useI18n()
  const ofInterface = (): DocumentLocale => locale.value === 'pl' ? 'pl' : 'en'
  const language = ref<DocumentLocale>(ofInterface())
  const reset = () => {
    language.value = ofInterface()
  }
  watch([locale, () => toValue(kind)], reset)
  const title = computed(() => ({
    guide: t('common.about.guide'),
    privacy: t('common.about.privacy'),
    license: t('common.about.license'),
  })[toValue(kind)])
  return { language, title, reset }
}
