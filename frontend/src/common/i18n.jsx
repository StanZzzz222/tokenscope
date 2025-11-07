import { createContext, useContext, useState } from "react";
import en from "@/assets/i18n/en-US.json";
import zh from "@/assets/i18n/zh-CN.json";
import { useEffect } from "react";

const I18nContext = createContext();

export const I18nProvider = ({ children }) => {
  const [locale, setLocale] = useState(en);

  useEffect(() => {
    const language = localStorage.getItem("locale")
    if (language) i18n.setLocale(language)
  }, [])

  const i18n = {
    setLocale: (language) => {
      switch (language) {
        case "en":
          setLocale(en);
          break;
        case "zh":
          setLocale(zh);
          break;
        default:
          setLocale(en);
      }
      localStorage.setItem("locale", language)
    },
    getLocale: () => {
      const language = localStorage.getItem("locale")
      return language
    },
    get: (key) => locale[key] || key,
  };

  return <I18nContext.Provider value={i18n}>{children}</I18nContext.Provider>;
};

export const useI18n = () => useContext(I18nContext);