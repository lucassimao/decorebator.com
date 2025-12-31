import { formatInTimeZone } from "date-fns-tz";
import * as Localization from "expo-localization";
import i18n from "@/i18n";

/**
 * Get the device's current timezone
 * @returns IANA timezone identifier (e.g., 'America/New_York') or 'UTC' as fallback
 */
export const getDeviceTimezone = (): string => {
  try {
    const calendarTimezone = Localization.getCalendars()?.[0]?.timeZone || "";
    const intlTimezone =
      typeof Intl !== "undefined"
        ? Intl.DateTimeFormat().resolvedOptions().timeZone || ""
        : "";
    const candidate = calendarTimezone || intlTimezone;
    const normalized = normalizeTimezone(candidate);
    if (normalized) {
      if (normalized !== candidate) {
        console.warn(
          "Normalized device timezone:",
          candidate,
          "->",
          normalized,
        );
      }
      return normalized;
    }
    if (candidate) {
      console.warn(
        "Invalid timezone from device, falling back to UTC:",
        candidate,
      );
    }
    return "UTC";
  } catch (error) {
    console.warn("Failed to get device timezone, falling back to UTC:", error);
    return "UTC";
  }
};

const normalizeTimezone = (value: string): string | null => {
  if (!value) {
    return null;
  }
  if (looksLikeIanaTimezone(value)) {
    return value;
  }
  const upper = value.toUpperCase();
  if (upper === "UTC" || upper === "GMT") {
    return "UTC";
  }
  const match = upper.match(/^(?:UTC|GMT)([+-])(\d{1,2})(?::?(\d{2}))?$/);
  if (!match) {
    return null;
  }
  const sign = match[1];
  const hours = Number(match[2]);
  const minutes = match[3] ? Number(match[3]) : 0;
  if (!Number.isFinite(hours) || !Number.isFinite(minutes)) {
    return null;
  }
  if (minutes !== 0) {
    return null;
  }
  if (hours > 14) {
    return null;
  }
  const etcSign = sign === "+" ? "-" : "+";
  return `Etc/GMT${etcSign}${hours}`;
};

const looksLikeIanaTimezone = (value: string): boolean => {
  if (!value) {
    return false;
  }
  if (value.startsWith("Etc/")) {
    return true;
  }
  return value.includes("/");
};

/**
 * Format a UTC date string with proper timezone conversion
 * @param utcDateString - ISO 8601 date string in UTC (from API)
 * @returns Formatted date string in the user's local timezone and locale
 */
export const formatDate = (utcDateString: string): string => {
  try {
    const userTimezone = getDeviceTimezone();

    // Convert UTC date to user's timezone and format using their locale
    return formatInTimeZone(
      new Date(utcDateString),
      userTimezone,
      "PPP", // Long localized date format: "January 1st, 2025"
    );
  } catch (error) {
    console.warn(
      "Failed to format subscription date with timezone, using fallback:",
      error,
    );

    // Fallback to current implementation if date-fns-tz fails
    try {
      const date = new Date(utcDateString);
      const locale = i18n.language.startsWith("en") ? "en-US" : i18n.language;

      return date.toLocaleDateString(locale, {
        year: "numeric",
        month: "long",
        day: "numeric",
      });
    } catch (fallbackError) {
      console.error("Even fallback date formatting failed:", fallbackError);
      return utcDateString; // Return original string as last resort
    }
  }
};
