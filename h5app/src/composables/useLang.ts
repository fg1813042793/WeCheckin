import { useLocale } from 'uview-pro'

export function useLang() {
  const { setLocale } = useLocale()

  // 切换语言
  function switchLang(lang: string, locale?: string) {
    // 切换uniapp语言
    locale && uni.setLocale(locale)
    uni.setStorageSync('UNI_LOCALE', locale)
    // 切换uView Pro语言
    setLocale(lang)
  }

  return {
    switchLang,
  }
}
