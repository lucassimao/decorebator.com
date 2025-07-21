// jest-setup.js
import "@testing-library/jest-native/extend-expect";

// Mock all Expo modules that cause issues
jest.mock("expo-secure-store", () => ({
  getItemAsync: jest.fn(),
  setItemAsync: jest.fn(),
  deleteItemAsync: jest.fn(),
}));

jest.mock("expo-localization", () => ({
  getLocales: jest.fn(() => [{ languageCode: "en", regionCode: "US" }]),
  getCalendars: jest.fn(() => [
    { calendar: "gregorian", timeZone: "America/New_York" },
  ]),
}));

jest.mock("@expo/vector-icons", () => ({
  Ionicons: "Ionicons",
  MaterialIcons: "MaterialIcons",
}));

jest.mock("expo-router", () => ({
  router: { replace: jest.fn(), push: jest.fn() },
}));

jest.mock("expo-web-browser", () => ({
  openAuthSessionAsync: jest.fn(),
  openBrowserAsync: jest.fn(),
}));

jest.mock("expo-mail-composer", () => ({
  isAvailableAsync: jest.fn(() => Promise.resolve(true)),
  composeAsync: jest.fn(),
}));

jest.mock("expo-auth-session", () => ({
  makeRedirectUri: jest.fn(() => "mock://redirect"),
}));

jest.mock("expo-updates", () => ({
  isEmbeddedLaunch: true,
  isEnabled: true,
  checkForUpdateAsync: jest.fn(() => Promise.resolve({ isAvailable: false })),
  fetchUpdateAsync: jest.fn(() => Promise.resolve({ isNew: false })),
  reloadAsync: jest.fn(() => Promise.resolve()),
}));

jest.mock("expo-device", () => ({
  DeviceType: { PHONE: 1 },
  deviceType: 1,
}));

jest.mock("expo-application", () => ({
  applicationName: "Test App",
  applicationVersion: "1.0.0",
  nativeApplicationVersion: "1.0.0",
  applicationId: "com.test.app",
}));

jest.mock("react-native-purchases", () => ({
  PurchasesPackage: {},
  PURCHASES_ERROR_CODE: {},
  PACKAGE_TYPE: {},
  configure: jest.fn(),
  getOfferings: jest.fn(),
  purchasePackage: jest.fn(),
  restorePurchases: jest.fn(),
  getCustomerInfo: jest.fn(),
}));

jest.mock("react-native-purchases-ui", () => ({
  presentPaywallIfNeeded: jest.fn(),
  RevenueCatUI: {
    presentPaywall: jest.fn(),
  },
}));

jest.mock("@react-native-async-storage/async-storage", () => ({
  getAllKeys: jest.fn(() => Promise.resolve([])),
  multiRemove: jest.fn(() => Promise.resolve()),
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
}));

// Mock i18n
jest.mock("@/i18n", () => ({
  language: "en",
}));

// Mock react-i18next
jest.mock("react-i18next", () => ({
  useTranslation: () => ({
    t: (key, params) => {
      // Simple translation mock - return the key with params interpolated
      if (params && key.includes("{{count}}")) {
        return key.replace("{{count}}", params.count);
      }
      return key;
    },
  }),
}));

// Mock React Navigation
jest.mock("@react-navigation/native", () => ({
  useNavigation: () => ({
    goBack: jest.fn(),
    navigate: jest.fn(),
    replace: jest.fn(),
    push: jest.fn(),
  }),
  useFocusEffect: jest.fn(),
  useRoute: () => ({
    params: {},
  }),
}));

// Mock React Query
jest.mock("@tanstack/react-query", () => ({
  useQuery: jest.fn(),
  useQueryClient: () => ({
    clear: jest.fn(),
    getQueryCache: () => ({ clear: jest.fn() }),
    getMutationCache: () => ({ clear: jest.fn() }),
    invalidateQueries: jest.fn(),
  }),
  useMutation: () => ({
    mutateAsync: jest.fn(),
    isPending: false,
    mutate: jest.fn(),
  }),
}));

// Mock React Native components
jest.mock("react-native", () => ({
  Platform: { OS: "ios" },
  Alert: { alert: jest.fn() },
  SafeAreaView: "SafeAreaView",
  ScrollView: "ScrollView",
  View: "View",
  Text: "Text",
  TouchableOpacity: "TouchableOpacity",
  StyleSheet: {
    create: (styles) => styles,
    flatten: (styles) => styles,
  },
  ActivityIndicator: "ActivityIndicator",
  Modal: "Modal",
  Dimensions: {
    get: () => ({ width: 375, height: 667 }),
  },
}));

// Mock useRevenueCat hook
jest.mock("@/hooks/useRevenueCat", () => ({
  usePaymentProvider: () => ({ data: { provider: "stripe" } }),
  useRevenueCat: () => ({
    isLoading: false,
    isRestoring: false,
    getCurrentOffering: () => null,
    purchasePackage: jest.fn(),
    restorePurchases: jest.fn(),
  }),
}));

// Mock ThemeContext with actual authLightTheme
jest.mock("@/contexts/ThemeContext", () => {
  const { authLightTheme } = jest.requireActual("@/theme/authTheme");

  return {
    useTheme: () => ({
      theme: authLightTheme,
      themeMode: "light",
      setThemeMode: jest.fn(),
      responsive: {
        spacing: {
          horizontal: 16,
          vertical: 16,
          formPadding: 20,
          elementSpacing: 12,
          minTouchTarget: 44,
        },
        fontSizes: { title: 24, headline: 18, body: 16, label: 14 },
        getValueForSize: (small, medium, large, xlarge) => medium, // Default to medium for tests
        getScaledFont: (fontName) => 16, // Default font size for tests
      },
    }),
  };
});

// Mock date utilities for consistent testing
jest.mock("@/utils/dateUtils", () => ({
  getDeviceTimezone: jest.fn(() => "America/New_York"), // Consistent timezone for tests
  formatDate: jest.fn((dateString) => {
    // Simple mock that returns a predictable format for testing
    return `Formatted: ${dateString}`;
  }),
}));

// Global console.warn suppression for test environment
const originalWarn = console.warn;
console.warn = (...args) => {
  // Suppress known Expo/React Native warnings in test environment
  const message = args[0];
  if (
    typeof message === "string" &&
    (message.includes("EXNativeModulesProxy") ||
      message.includes("ProgressBarAndroid") ||
      message.includes("Clipboard"))
  ) {
    return;
  }
  originalWarn(...args);
};
