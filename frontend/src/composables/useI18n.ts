import { ref, watch } from 'vue'
import arLocale from '../locales/ar.json'
import enLocale from '../locales/en.json'

export type SupportedLocale = 'ar' | 'en'

const locales: Record<SupportedLocale, Record<string, string>> = {
  ar: arLocale,
  en: enLocale,
}

const currentLang = ref<SupportedLocale>('ar')

// Synchronize language and text direction to html document element in browser
if (typeof document !== 'undefined') {
  watch(
    currentLang,
    (lang) => {
      document.documentElement.dir = lang === 'ar' ? 'rtl' : 'ltr'
      document.documentElement.lang = lang
    },
    { immediate: true },
  )
}

export function useI18n() {
  function t(
    key: string,
    params?: Record<string, string | number | undefined>,
  ): string {
    let str = locales[currentLang.value]?.[key] || key
    if (params) {
      for (const [k, v] of Object.entries(params)) {
        str = str.replace(new RegExp(`\\{${k}\\}`, 'g'), String(v ?? ''))
      }
    }
    return str
  }

  function renderUIMessage(msg: any): string {
    if (!msg) return ''
    if (typeof msg === 'string') return msg
    if (msg.key) {
      return t(msg.key, msg.params)
    }
    return ''
  }

  function toggleLang() {
    currentLang.value = currentLang.value === 'ar' ? 'en' : 'ar'
  }

  function setLang(lang: SupportedLocale) {
    currentLang.value = lang
  }

  return {
    currentLang,
    t,
    toggleLang,
    setLang,
    renderUIMessage,
  }
}
