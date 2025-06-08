import i18n from "i18next";
import { initReactI18next } from "react-i18next";

// Import all language files
import en from "./locales/en.json";
import ptBR from "./locales/pt-BR.json";
import ptPT from "./locales/pt-PT.json";
import de from "./locales/de.json";
import it from "./locales/it.json";
import fr from "./locales/fr.json";
import es from "./locales/es.json";
import ja from "./locales/ja.json";

export const resources = {
  en: { translation: en },
  "pt-BR": { translation: ptBR },
  "pt-PT": { translation: ptPT },
  de: { translation: de },
  it: { translation: it },
  fr: { translation: fr },
  es: { translation: es },
  ja: { translation: ja },
};

export const supportedLanguages = [
  { code: "en", name: "English", nativeName: "English" },
  {
    code: "pt-BR",
    name: "Brazilian Portuguese",
    nativeName: "Português (Brasil)",
  },
  {
    code: "pt-PT",
    name: "European Portuguese",
    nativeName: "Português (Portugal)",
  },
  { code: "de", name: "German", nativeName: "Deutsch" },
  { code: "it", name: "Italian", nativeName: "Italiano" },
  { code: "fr", name: "French", nativeName: "Français" },
  { code: "es", name: "Spanish", nativeName: "Español" },
  { code: "ja", name: "Japanese", nativeName: "日本語" },
];

i18n.use(initReactI18next).init({
  resources,
  lng: "en", // Default language
  fallbackLng: "en",
  interpolation: {
    escapeValue: false, // React already escapes values
  },
  compatibilityJSON: "v4", // For React Native
});

export default i18n;
