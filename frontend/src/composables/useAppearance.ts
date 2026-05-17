import { computed, onBeforeUnmount, onMounted, ref, watch, type Ref } from 'vue'
import { normalizeLocale, setI18nLocale, type AppLocale } from '../i18n'
import type { Preferences } from '../domain/types'

export function useAppearance(preferences: Ref<Preferences>) {
  const systemDark = refSystemDark()
  let offSystemAppearance: (() => void) | undefined

  const currentLanguage = computed<AppLocale>(() => normalizeLocale(preferences.value.language))
  const resolvedAppearance = computed<'light' | 'dark'>(() => {
    if (preferences.value.appearance === 'dark') return 'dark'
    if (preferences.value.appearance === 'light') return 'light'
    return systemDark.value ? 'dark' : 'light'
  })

  function bindSystemAppearance() {
    if (typeof window === 'undefined' || !window.matchMedia) return
    const media = window.matchMedia('(prefers-color-scheme: dark)')
    const update = () => { systemDark.value = media.matches }
    update()
    media.addEventListener('change', update)
    offSystemAppearance = () => media.removeEventListener('change', update)
  }

  function applyDocumentAppearance() {
    if (typeof document === 'undefined') return
    const appearance = preferences.value.appearance || 'system'
    const resolved = resolvedAppearance.value
    const targets = [document.documentElement, document.body].filter(Boolean) as HTMLElement[]
    for (const target of targets) {
      target.dataset.appearance = appearance
      target.dataset.resolvedAppearance = resolved
      target.style.colorScheme = resolved
    }
  }

  onMounted(bindSystemAppearance)
  onBeforeUnmount(() => offSystemAppearance?.())
  watch(currentLanguage, (language) => {
    setI18nLocale(language)
  }, { immediate: true })
  watch([() => preferences.value.appearance, resolvedAppearance], applyDocumentAppearance, { immediate: true })

  return { systemDark, currentLanguage, resolvedAppearance }
}

function refSystemDark() {
  const initial = typeof window !== 'undefined' && window.matchMedia?.('(prefers-color-scheme: dark)').matches
  return ref(Boolean(initial))
}
