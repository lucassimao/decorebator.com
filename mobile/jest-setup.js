// jest-setup.js
// Import and configure i18n for tests
import i18n from "@/i18n";

process.env.EXPO_PUBLIC_API_URL ??= "http://localhost:3000";

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

jest.mock("expo-notifications", () => ({
  addNotificationResponseReceivedListener: jest.fn(() => ({
    remove: jest.fn(),
  })),
  clearLastNotificationResponse: jest.fn(),
  getExpoPushTokenAsync: jest.fn(() => Promise.resolve({ data: "test-token" })),
  getLastNotificationResponseAsync: jest.fn(() => Promise.resolve(null)),
  getPermissionsAsync: jest.fn(() =>
    Promise.resolve({ status: "granted", canAskAgain: true }),
  ),
  requestPermissionsAsync: jest.fn(() =>
    Promise.resolve({ status: "granted", canAskAgain: true }),
  ),
  setNotificationHandler: jest.fn(),
}));

jest.mock("@/utils/offlineManager", () => ({
  __esModule: true,
  default: {
    clearCache: jest.fn(() => Promise.resolve()),
    getNetworkStatus: jest.fn(() => true),
    setUserPremiumStatus: jest.fn(),
    subscribeToNetworkChanges: jest.fn(() => jest.fn()),
  },
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

jest.mock("expo-constants", () => ({
  default: {
    expoConfig: {
      version: "1.0.0",
      runtimeVersion: "1.0.0",
    },
    appOwnership: "expo",
  },
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

// Reanimated's native worklets module is unavailable in Jest. Individual
// animation tests can override this deterministic static implementation.
jest.mock("react-native-reanimated", () => ({
  __esModule: true,
  default: { View: "Animated.View", Text: "Animated.Text" },
  Easing: { cubic: "cubic", out: (value) => value },
  cancelAnimation: jest.fn(),
  interpolate: (value, input, output) =>
    output[0] +
    ((value - input[0]) / (input[1] - input[0])) * (output[1] - output[0]),
  runOnJS: (fn) => fn,
  useAnimatedStyle: (fn) => fn(),
  useReducedMotion: () => false,
  useSharedValue: (value) => ({ value }),
  withDelay: (_delay, value) => value,
  withSpring: (value) => value,
  withTiming: (value, _config, callback) => {
    callback?.(true);
    return value;
  },
}));

jest.mock("expo-iap", () => ({
  ErrorCode: {
    UserCancelled: "user-cancelled",
    Pending: "pending",
    DeferredPayment: "deferred-payment",
    Unknown: "unknown",
    ItemUnavailable: "item-unavailable",
    SkuNotFound: "sku-not-found",
    FeatureNotSupported: "feature-not-supported",
  },
  getAvailablePurchases: jest.fn(() => Promise.resolve([])),
  isEligibleForIntroOfferIOS: jest.fn(() => Promise.resolve(false)),
  useIAP: () => ({
    connected: false,
    subscriptions: [],
    fetchProducts: jest.fn(() => Promise.resolve()),
    requestPurchase: jest.fn(() => Promise.resolve()),
    finishTransaction: jest.fn(() => Promise.resolve()),
    reconnect: jest.fn(() => Promise.resolve(false)),
  }),
}));

jest.mock("@react-native-async-storage/async-storage", () => {
  const storage = new Map();
  return {
    getAllKeys: jest.fn(() => Promise.resolve([...storage.keys()])),
    multiGet: jest.fn((keys) =>
      Promise.resolve(keys.map((key) => [key, storage.get(key) ?? null])),
    ),
    multiSet: jest.fn((entries) => {
      entries.forEach(([key, value]) => storage.set(key, value));
      return Promise.resolve();
    }),
    multiRemove: jest.fn((keys) => {
      keys.forEach((key) => storage.delete(key));
      return Promise.resolve();
    }),
    getItem: jest.fn((key) => Promise.resolve(storage.get(key) ?? null)),
    setItem: jest.fn((key, value) => {
      storage.set(key, value);
      return Promise.resolve();
    }),
    removeItem: jest.fn((key) => {
      storage.delete(key);
      return Promise.resolve();
    }),
    clear: jest.fn(() => {
      storage.clear();
      return Promise.resolve();
    }),
  };
});

// Configure i18n for tests
beforeAll(async () => {
  await i18n.changeLanguage("en");
  i18n.options.lng = "en";
  i18n.options.fallbackLng = "en";
});

// Mock React Navigation
jest.mock("@react-navigation/native", () => ({
  useNavigation: jest.fn(() => ({
    goBack: jest.fn(),
    navigate: jest.fn(),
    replace: jest.fn(),
    push: jest.fn(),
  })),
  useFocusEffect: jest.fn(),
  useRoute: jest.fn(() => ({
    params: {},
  })),
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
  AppState: {
    addEventListener: jest.fn(() => ({ remove: jest.fn() })),
  },
  Linking: {
    addEventListener: jest.fn(() => ({ remove: jest.fn() })),
    canOpenURL: jest.fn(() => Promise.resolve(true)),
    openURL: jest.fn(() => Promise.resolve()),
  },
  SafeAreaView: "SafeAreaView",
  ScrollView: "ScrollView",
  View: "View",
  Text: "Text",
  TextInput: "TextInput",
  TouchableOpacity: "TouchableOpacity",
  TouchableWithoutFeedback: "TouchableWithoutFeedback",
  Pressable: "Pressable",
  Switch: "Switch",
  KeyboardAvoidingView: "KeyboardAvoidingView",
  Image: "Image",
  ImageBackground: "ImageBackground",
  Keyboard: {
    dismiss: jest.fn(),
    addListener: jest.fn(() => ({
      remove: jest.fn(),
    })),
  },
  StyleSheet: {
    create: (styles) => styles,
    flatten: (styles) => styles,
  },
  ActivityIndicator: "ActivityIndicator",
  Modal: "Modal",
  AccessibilityInfo: {
    isScreenReaderEnabled: jest.fn(() => Promise.resolve(false)),
    setAccessibilityFocus: jest.fn(),
    announceForAccessibility: jest.fn(),
  },
  findNodeHandle: jest.fn(() => 1),
  Dimensions: {
    get: () => ({ width: 375, height: 667 }),
  },
  Animated: {
    View: "Animated.View",
    Text: "Animated.Text",
    Value: class AnimatedValue {
      constructor(value) {
        this._value = value;
      }
      setValue(value) {
        this._value = value;
      }
    },
    timing: (value, config) => ({
      start: (callback) => {
        value.setValue(config.toValue);
        if (callback) callback({ finished: true });
      },
      stop: jest.fn(),
    }),
    spring: (value, config) => ({
      start: (callback) => {
        value.setValue(config.toValue);
        if (callback) callback({ finished: true });
      },
      stop: jest.fn(),
    }),
    sequence: (animations) => ({
      start: (callback) => {
        animations.forEach((anim) => anim.start());
        if (callback) callback({ finished: true });
      },
      stop: jest.fn(),
    }),
    parallel: (animations) => ({
      start: (callback) => {
        animations.forEach((anim) => anim.start());
        if (callback) callback({ finished: true });
      },
      stop: jest.fn(),
    }),
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
    useTheme: jest.fn(() => ({
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
    })),
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

// Mock react-native-safe-area-context
jest.mock("react-native-safe-area-context", () => ({
  useSafeAreaInsets: jest.fn(() => ({
    top: 0,
    right: 0,
    bottom: 0,
    left: 0,
  })),
  SafeAreaProvider: ({ children }) => children,
  SafeAreaView: ({ children }) => children,
}));

// Mock expo-linear-gradient
jest.mock("expo-linear-gradient", () => ({
  LinearGradient: ({ children }) => children,
}));

// Mock Sentry
jest.mock("@sentry/react-native", () => ({
  init: jest.fn(),
  captureException: jest.fn(),
  captureMessage: jest.fn(),
  addBreadcrumb: jest.fn(),
  setContext: jest.fn(),
  setUser: jest.fn(),
}));

// Mock API modules
jest.mock("@/api/users", () => ({
  login: jest.fn(),
  signup: jest.fn(),
  setAuth: jest.fn(),
  requestResetEmailPassword: jest.fn(),
}));

// Mock utility functions
jest.mock("@/utils/countryDetection", () => ({
  getDetectedCountry: jest.fn(() => "US"),
}));

// Mock hooks
jest.mock("@/hooks/useSnackbar", () => ({
  useSnackbar: () => ({
    show: jest.fn(),
  }),
}));

// Mock expo-web-browser
jest.mock("expo-web-browser", () => ({
  openBrowserAsync: jest.fn(),
}));

// Mock expo-router
jest.mock("expo-router", () => ({
  router: {
    replace: jest.fn(),
    push: jest.fn(),
  },
  Link: ({ children }) => children,
}));

// Mock posthog
jest.mock("posthog-react-native", () => ({
  usePostHog: () => ({
    capture: jest.fn(),
    identify: jest.fn(),
    reset: jest.fn(),
  }),
}));

// Duplicate navigation mock removed (already defined above)

// Remove duplicate React Native mock - it's already defined above
