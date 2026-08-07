/**
 * Maps API error messages to i18n translation keys
 */

interface ErrorMapping {
  pattern: RegExp | string;
  i18nKey: string;
  isFieldError?: boolean;
  field?: string;
}

const errorMappings: ErrorMapping[] = [
  // Validation errors for signup fields - exact matches from API
  {
    pattern: "The email field is required.",
    i18nKey: "common.required",
    isFieldError: true,
    field: "email",
  },
  {
    pattern: "The firstname field is required.",
    i18nKey: "common.required",
    isFieldError: true,
    field: "firstName",
  },
  {
    pattern: "The lastname field is required.",
    i18nKey: "common.required",
    isFieldError: true,
    field: "lastName",
  },
  {
    pattern: "The password field is required.",
    i18nKey: "common.required",
    isFieldError: true,
    field: "password",
  },
  {
    pattern: "The email field must be a valid email address.",
    i18nKey: "errors.invalidEmail",
    isFieldError: true,
    field: "email",
  },
  {
    pattern: "password must contain at least 8 characters",
    i18nKey: "errors.shortPassword",
    isFieldError: true,
    field: "password",
  },
  {
    pattern: "password must contain at most 72 UTF-8 bytes",
    i18nKey: "errors.longPassword",
    isFieldError: true,
    field: "password",
  },
];

export interface MappedError {
  i18nKey: string;
  isFieldError: boolean;
  field?: string;
  originalMessage: string;
}

/**
 * Maps an API error message to an i18n translation key
 * @param errorMessage The error message from the API
 * @returns Mapped error with i18n key and field information
 */
export function mapErrorToI18n(errorMessage: string): MappedError {
  const trimmedMessage = errorMessage.trim();

  // Find matching error mapping using exact string matching
  for (const mapping of errorMappings) {
    if (
      typeof mapping.pattern === "string" &&
      trimmedMessage === mapping.pattern
    ) {
      return {
        i18nKey: mapping.i18nKey,
        isFieldError: mapping.isFieldError || false,
        field: mapping.field,
        originalMessage: errorMessage,
      };
    }
  }

  // Default fallback for non-signup errors
  return {
    i18nKey: "errors.general",
    isFieldError: false,
    originalMessage: errorMessage,
  };
}

/**
 * Checks if an error should be displayed as a field error
 * @param errorMessage The error message from the API
 * @returns Object with field name and whether it's a field error
 */
export function getFieldError(errorMessage: string): {
  isFieldError: boolean;
  field?: string;
} {
  const mapped = mapErrorToI18n(errorMessage);
  return {
    isFieldError: mapped.isFieldError,
    field: mapped.field,
  };
}
