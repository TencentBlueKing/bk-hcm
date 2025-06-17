import { createI18n, Locale } from 'vue-i18n';
import Cookies from 'js-cookie';

import langMap from './lang';

interface ILANG_PKG {
  [propName: string]: string;
}

const en: ILANG_PKG = {};
const zh: ILANG_PKG = {};
Object.keys(langMap).forEach((key) => {
  en[key] = langMap[key][0] || key;
  zh[key] = langMap[key][1] || key;
});

// const language = (navigator.language || 'en').toLocaleLowerCase();
const localLanguage = (Cookies.get('blueking_language') || 'zh-cn') as Locale;

const i18n = createI18n({
  silentTranslationWarn: true,
  legacy: false,
  locale: localLanguage,
  fallbackLocale: 'zh-cn',
  messages: {
    // 'zh-cn': Object.assign(lang.zhCN, zh),
    'zh-cn': zh,
    // en: Object.assign(lang.enUS, en),
    en,
  },
});

export const isChinese = localLanguage === 'zh-cn';

export default i18n;
