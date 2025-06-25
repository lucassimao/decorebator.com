import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import * as Localization from "expo-localization";

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

// Function to get the best matching language based on device locale
export function getDeviceLanguage(): string {
  // Get the device locale
  const deviceLocale = Localization.getLocales()[0];

  if (!deviceLocale) {
    return "en"; // Fallback to English if no locale found
  }

  const deviceLanguageCode = deviceLocale.languageCode;
  const deviceRegion = deviceLocale.regionCode;

  // First, try to find exact match (e.g., pt-BR)
  const exactMatch = `${deviceLanguageCode}-${deviceRegion}`;
  if (exactMatch in resources) {
    return exactMatch;
  }

  // Special handling for Portuguese
  if (deviceLanguageCode === "pt") {
    // If the device is set to Portuguese from Brazil, use pt-BR
    if (deviceRegion === "BR") {
      return "pt-BR";
    }
    // For Portugal or any other Portuguese variant, use pt-PT
    // This includes when just "pt" is set without a region
    return "pt-PT";
  }

  // For other languages, try to find a match by language code only
  const supportedLanguageCodes = Object.keys(resources);
  const languageMatch = supportedLanguageCodes.find(
    (code) => code.split("-")[0] === deviceLanguageCode,
  );

  if (languageMatch) {
    return languageMatch;
  }

  // If no match found, fallback to English
  return "en";
}

// Get the initial language based on device settings
const initialLanguage = getDeviceLanguage();

i18n.use(initReactI18next).init({
  resources,
  lng: initialLanguage,
  fallbackLng: "en",
  interpolation: {
    escapeValue: false, // React already escapes values
  },
  compatibilityJSON: "v4", // For React Native
});

export default i18n;
