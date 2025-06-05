import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import * as userApi from "@/api/users";
import i18n from "@/i18n";

export function useI18n() {
  // Fetch user profile to get preferredLanguage
  const { data: profile } = useQuery({
    queryKey: ["userProfile"],
    queryFn: userApi.getProfile,
    retry: false, // Don't retry if user is not authenticated
  });

  useEffect(() => {
    if (profile?.preferredLanguage) {
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
        languageMap[profile.preferredLanguage] || profile.preferredLanguage;

      if (i18n.language !== i18nLanguage) {
        i18n.changeLanguage(i18nLanguage);
      }
    }
  }, [profile?.preferredLanguage]);

  const changeLanguage = async (language: string) => {
    // Update i18n immediately for responsive UI
    await i18n.changeLanguage(language);

    // Update user profile with new language preference
    try {
      await userApi.update({ preferredLanguage: language });
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
