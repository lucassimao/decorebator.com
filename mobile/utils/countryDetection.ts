import * as Localization from "expo-localization";

/**
 * Detects the user's country based on device locale settings
 * @returns ISO 3166-1 alpha-2 country code (e.g., "US", "GB", "DE")
 */
export function detectUserCountry(): string {
  try {
    // Get device locale information
    const locales = Localization.getLocales();

    if (locales && locales.length > 0) {
      const primaryLocale = locales[0];

      // Extract region code (country) from locale
      if (primaryLocale.regionCode) {
        return primaryLocale.regionCode.toUpperCase();
      }
    }

    // Fallback to US if no region detected
    return "US";
  } catch (error) {
    console.warn("Failed to detect country from device locale:", error);
    // Fallback to US on any error
    return "US";
  }
}

/**
 * Validates if a country code is a valid ISO 3166-1 alpha-2 code
 * @param countryCode - The country code to validate
 * @returns boolean indicating if the code is valid
 */
export function isValidCountryCode(countryCode: string): boolean {
  if (!countryCode || countryCode.length !== 2) {
    return false;
  }

  // Basic validation - should be two uppercase letters
  return /^[A-Z]{2}$/.test(countryCode.toUpperCase());
}

/**
 * Gets the detected country with validation
 * @returns Validated country code or "US" as fallback
 */
export function getDetectedCountry(): string {
  const detected = detectUserCountry();

  if (isValidCountryCode(detected)) {
    return detected;
  }

  return "US";
}
