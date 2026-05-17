import { createI18n, type MessageResolver, type PathValue } from 'vue-i18n'
import en from './locales/en'
import zhCN from './locales/zh-CN'

export type AppLocale = 'zh-CN' | 'en'

export function normalizeLocale(locale?: string): AppLocale {
  return locale === 'en' ? 'en' : 'zh-CN'
}

const flatMessageResolver: MessageResolver = (messages: unknown, path: string) => {
  if (!messages || typeof messages !== 'object') return null
  const record = messages as Record<string, unknown>
  return Object.prototype.hasOwnProperty.call(record, path) ? record[path] as PathValue : null
}

export const i18n = createI18n({
  legacy: false as const,
  locale: 'zh-CN',
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    en,
  },
  messageResolver: flatMessageResolver,
  missingWarn: false,
  fallbackWarn: false,
})

export function setI18nLocale(locale: string) {
  const normalized = normalizeLocale(locale)
  i18n.global.locale.value = normalized
  document.documentElement.lang = normalized
  return normalized
}
