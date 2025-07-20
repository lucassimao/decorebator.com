import { useEffect } from "react";
import * as userApi from "@/api/users";
import i18n from "@/i18n";
import { useUserSession } from "@/hooks/useUserSession";

export function useI18n() {
  // Get user data from centralized session
  const { user } = useUserSession();

  useEffect(() => {
    if (user?.preferredLanguage) {
      // Map backend language codes to i18n language codes if needed
      const languageMap: Record<string, string> = {
        en: "en",
        pt_BR: "pt-BR",
        pt_PT: "pt-PT",
        de: "de",
        it: "it",
        fr: "fr",
        es: "es",
        ja: "ja",
      };

      const i18nLanguage =
        languageMap[user.preferredLanguage] || user.preferredLanguage;

      if (i18n.language !== i18nLanguage) {
        i18n.changeLanguage(i18nLanguage);
      }
    }
  }, [user?.preferredLanguage]);

  const changeLanguage = async (language: string) => {
    // Update i18n immediately for responsive UI
    await i18n.changeLanguage(language);

    // Map i18n language codes to backend language codes
    const backendLanguageMap: Record<string, string> = {
      en: "en",
      "pt-BR": "pt_BR",
      "pt-PT": "pt_PT",
      de: "de",
      it: "it",
      fr: "fr",
      es: "es",
      ja: "ja",
    };

    const backendLanguage = backendLanguageMap[language] || language;

    // Update user profile with new language preference
    try {
      await userApi.update({ preferredLanguage: backendLanguage });
    } catch (error) {
      console.error("Failed to update language preference:", error);
      // Optionally revert language change if API call fails
    }
  };

  return {
    currentLanguage: i18n.language,
    changeLanguage,
    t: i18n.t,
  };
}
