import * as usersApi from "@/api/users";
import { useSnackbar } from "@/hooks/useSnackbar";
import { ProgressiveSignupForm } from "@/components/auth/ProgressiveSignupForm";
import { TraditionalSignupForm } from "@/components/auth/TraditionalSignupForm";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { router } from "expo-router";
import { usePostHog } from "posthog-react-native";
import * as React from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { backendLanguageMap } from "@/hooks/useI18n";
import { getDetectedCountry } from "@/utils/countryDetection";
import { mapErrorToI18n } from "@/utils/errorMapping";
import AsyncStorage from "@react-native-async-storage/async-storage";
import {
  ImageBackground,
  Keyboard,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  StyleSheet,
  TouchableWithoutFeedback,
  View,
} from "react-native";
import z from "zod";
import { authLightTheme } from "@/theme/authTheme";
import type { Theme } from "@/contexts/ThemeContext";
import { useTheme } from "@/contexts/ThemeContext";
import type { ResponsiveValues } from "@/contexts/ThemeContext";
import { SafeAreaView } from "react-native-safe-area-context";
import {
  ACTIVATION_EVENT_NAMES,
  captureActivationEvent,
} from "@/utils/activationEvents";

const schema = z
  .object({
    fullName: z
      .string()
      .min(2, "Required")
      .regex(/\s/, "Please enter your full name"),
    email: z.string().email().min(2, "Required"),
    password: z.string().min(5, "Required"),
  })
  .required();

export default function SignUpScreen() {
  const [signUpError, setSignUpError] = React.useState<Error | null>(null);
  const [detectedCountry, setDetectedCountry] = React.useState<string>("");
  const [keyboardVisible, setKeyboardVisible] = React.useState(false);
  const snackbar = useSnackbar();
  const { t, i18n } = useTranslation();
  const posthog = usePostHog();

  // Always use light theme for auth screens
  const theme = authLightTheme;

  // Get responsive values from the unified theme context
  const { responsive } = useTheme();

  // Determine if we should use progressive disclosure based on screen size
  const useProgressiveDisclosure = responsive.category === "small";

  // Memoize styles to prevent recreation on every render
  const styles = React.useMemo(
    () => createStyles(theme, responsive, keyboardVisible),
    [theme, responsive, keyboardVisible],
  );

  const scrollViewRef = React.useRef<ScrollView>(null);

  React.useEffect(() => {
    if (signUpError) {
      // Only show snackbar for non-email errors (email errors are shown inline)
      snackbar.show(signUpError.message, "error", 2000);
    }
  }, [signUpError, snackbar]);

  React.useEffect(() => {
    posthog.capture("signup_started", { source: "signup_screen" });
  }, [posthog]);

  // Keyboard listeners
  React.useEffect(() => {
    const keyboardWillShow = Keyboard.addListener(
      Platform.OS === "ios" ? "keyboardWillShow" : "keyboardDidShow",
      () => setKeyboardVisible(true),
    );

    const keyboardWillHide = Keyboard.addListener(
      Platform.OS === "ios" ? "keyboardWillHide" : "keyboardDidHide",
      () => setKeyboardVisible(false),
    );

    return () => {
      keyboardWillShow.remove();
      keyboardWillHide.remove();
    };
  }, []);

  // Detect user's country on component mount
  React.useEffect(() => {
    try {
      const country = getDetectedCountry();
      setDetectedCountry(country);
      if (__DEV__) {
        console.log("Detected country:", country);
      }
    } catch (error) {
      if (__DEV__) {
        console.warn("Failed to detect country:", error);
      }
      setDetectedCountry("US"); // Fallback to US
    }
  }, []);

  // Auto-focus email field on component mount for better UX
  // React.useEffect(() => {
  //   const timer = setTimeout(() => {
  //     emailInputRef.current?.focus();
  //   }, 300); // Small delay to ensure component is fully mounted

  //   return () => clearTimeout(timer);
  // }, []);

  const { mutate: signup } = useMutation<void, Error, usersApi.UserSignup>({
    mutationFn: (userData) => usersApi.signup(userData),
    onError: (error) => {
      const mappedError = mapErrorToI18n(error.message);

      if (mappedError.isFieldError && mappedError.field) {
        // Set error on specific field to display translated message below the input
        setError(mappedError.field as keyof typeof schema._type, {
          type: "manual",
          message: t(mappedError.i18nKey),
        });
      } else {
        // For other errors, use snackbar with translated message
        setSignUpError(new Error(t(mappedError.i18nKey)));
      }
    },
    onSuccess: async () => {
      captureActivationEvent(posthog, ACTIVATION_EVENT_NAMES.USER_SIGNED_UP, {
        source: "signup_screen",
      });

      // Set flag that user just signed up for welcome flow
      try {
        await AsyncStorage.setItem("justSignedUp", "true");
        if (__DEV__) {
          console.log("Set justSignedUp flag to true");
        }
      } catch (error) {
        console.warn("Failed to set signup flag:", error);
      }

      router.replace("/dashboard");
    },
  });

  const form = useForm({
    resolver: zodResolver(schema),
    mode: "onChange",
    defaultValues: {
      fullName: "",
      email: __DEV__ ? process.env.EXPO_PUBLIC_TEST_USER_EMAIL || "" : "",
      password: __DEV__ ? process.env.EXPO_PUBLIC_TEST_USER_PASSWORD || "" : "",
    },
  });

  const { setError } = form;

  const onSubmit = React.useCallback(
    (data: z.infer<typeof schema>) => {
      // Split full name into first and last names
      const nameParts = data.fullName.trim().split(/\s+/);
      const firstName = nameParts[0];
      const lastName = nameParts.slice(1).join(" ") || nameParts[0]; // Use first name as last if only one name

      // Get current UI language and convert to backend format
      const currentLanguage = i18n.language;
      const backendLanguage =
        backendLanguageMap[currentLanguage] || currentLanguage;

      // Add detected country and language to signup data
      const signupData = {
        firstName,
        lastName,
        email: data.email,
        password: data.password,
        country: detectedCountry || "US", // Include detected country or fallback to US
        preferredLanguage: backendLanguage, // Include detected UI language
      };
      signup(signupData);
    },
    [detectedCountry, signup, i18n.language],
  );

  return (
    <ImageBackground
      source={require("../assets/images/signup-bg3.png")}
      style={styles.backgroundImage}
      imageStyle={styles.backgroundImageStyle}
      resizeMode="cover"
    >
      <SafeAreaView style={styles.safeArea}>
        <KeyboardAvoidingView
          style={styles.container}
          behavior={responsive.keyboardBehavior}
          keyboardVerticalOffset={responsive.keyboardOffset}
        >
          <TouchableWithoutFeedback onPress={Keyboard.dismiss}>
            <ScrollView
              ref={scrollViewRef}
              contentContainerStyle={styles.scrollContent}
              showsVerticalScrollIndicator={false}
              keyboardShouldPersistTaps="handled"
            >
              {/* Add some space at the top for the background to show */}
              <View style={styles.topSpacer} />

              {/* Form Card */}
              <View style={styles.formCard}>
                {useProgressiveDisclosure ? (
                  <ProgressiveSignupForm
                    form={form}
                    onSubmit={onSubmit}
                    theme={theme}
                    responsive={responsive}
                  />
                ) : (
                  <TraditionalSignupForm
                    form={form}
                    onSubmit={onSubmit}
                    theme={theme}
                    responsive={responsive}
                  />
                )}
              </View>
            </ScrollView>
          </TouchableWithoutFeedback>
        </KeyboardAvoidingView>
      </SafeAreaView>
    </ImageBackground>
  );
}

const createStyles = (
  theme: Theme,
  responsive: ResponsiveValues,
  keyboardVisible: boolean,
) =>
  StyleSheet.create({
    backgroundImage: {
      flex: 1,
      width: "100%",
      height: "100%",
      backgroundColor: "#FFF9F0", // Fallback warm background color
    },
    backgroundImageStyle: {
      // Force full coverage regardless of aspect ratio
      width: "100%",
      height: "100%",
    },
    safeArea: {
      flex: 1,
      backgroundColor: "transparent", // Let image show through
    },
    container: {
      flex: 1,
    },
    scrollContent: {
      flexGrow: 1,
      paddingHorizontal: responsive.spacing.horizontal,
    },
    topSpacer: {
      height: keyboardVisible
        ? responsive.spacing.vertical
        : responsive.screenHeight * 0.12, // Balanced spacing
    },
    bottomSpacer: {
      height: responsive.spacing.vertical * 4, // Increased to account for fixed footer
    },
    formCard: {
      backgroundColor: "rgba(255, 255, 255, 0.95)",
      borderRadius: theme.borderRadius.xl,
      padding: responsive.spacing.formPadding,
      ...theme.shadows.lg,
      // Add subtle border for definition
      borderWidth: 1,
      borderColor: "rgba(255, 255, 255, 0.3)",
    },
  });
