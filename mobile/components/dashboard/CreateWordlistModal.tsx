import * as wordlistsApi from "@/api/wordlists";
import { CreateWordlistDTO } from "@/api/wordlists";
import { useTheme } from "@/contexts/ThemeContext";
import { useUserSession } from "@/hooks/useUserSession";
import { Ionicons } from "@expo/vector-icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import React, { useEffect, useRef, useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import {
  ActivityIndicator,
  Animated,
  Dimensions,
  Keyboard,
  Modal,
  PanResponder,
  Platform,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  TouchableWithoutFeedback,
  View,
} from "react-native";
import * as Sentry from "@sentry/react-native";
import { usePostHog } from "posthog-react-native";
import {
  ACTIVATION_EVENT_NAMES,
  captureActivationEvent,
} from "@/utils/activationEvents";

const { height: SCREEN_HEIGHT } = Dimensions.get("window");

type Language = {
  code: string;
  name: string;
  flag: string;
};

export const LANGUAGES: Language[] = [
  { code: "en", name: "English", flag: "🇬🇧" },
  { code: "es", name: "Spanish", flag: "🇪🇸" },
  { code: "fr", name: "French", flag: "🇫🇷" },
  { code: "de", name: "German", flag: "🇩🇪" },
  { code: "it", name: "Italian", flag: "🇮🇹" },
  { code: "pt", name: "Portuguese", flag: "🇵🇹" },
  { code: "ja", name: "Japanese", flag: "🇯🇵" },
];

interface CreateWordlistModalProps {
  visible: boolean;
  onClose: () => void;
  onSuccess?: (wordlist: wordlistsApi.Wordlist) => void;
  onError?: (error: Error) => void;
}

export const CreateWordlistModal: React.FC<CreateWordlistModalProps> = ({
  visible,
  onClose,
  onSuccess,
  onError,
}) => {
  const slideAnim = useRef(new Animated.Value(SCREEN_HEIGHT)).current;
  const backdropAnim = useRef(new Animated.Value(0)).current;
  const keyboardOffsetAnim = useRef(new Animated.Value(0)).current;
  const queryClient = useQueryClient();
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { hasOptimisticSubscription } = useUserSession();
  const posthog = usePostHog();
  const [currentStep, setCurrentStep] = useState(1);
  const totalSteps = 4;
  const styles = createStyles(theme);

  // Handle modal close with animation
  const handleClose = () => {
    Animated.parallel([
      Animated.timing(slideAnim, {
        toValue: SCREEN_HEIGHT,
        duration: 300,
        useNativeDriver: true,
      }),
      Animated.timing(backdropAnim, {
        toValue: 0,
        duration: 300,
        useNativeDriver: true,
      }),
    ]).start(() => {
      onClose();
    });
  };

  // Pan responder for swipe to dismiss
  const panResponder = useRef(
    PanResponder.create({
      onStartShouldSetPanResponder: () => false,
      onMoveShouldSetPanResponder: (_, gestureState) => {
        // Only respond to downward swipes with sufficient movement
        return (
          gestureState.dy > 10 &&
          Math.abs(gestureState.dx) < Math.abs(gestureState.dy)
        );
      },
      onPanResponderMove: (_, gestureState) => {
        // Only allow downward movement
        if (gestureState.dy > 0) {
          slideAnim.setValue(gestureState.dy);
        }
      },
      onPanResponderRelease: (_, gestureState) => {
        // If swiped down more than 100 pixels or with velocity, close the modal
        if (gestureState.dy > 100 || gestureState.vy > 0.5) {
          handleClose();
        } else {
          // Bounce back to original position
          Animated.spring(slideAnim, {
            toValue: 0,
            useNativeDriver: true,
            tension: 40,
            friction: 8,
          }).start();
        }
      },
    }),
  ).current;

  const mutation = useMutation<
    wordlistsApi.Wordlist,
    Error,
    wordlistsApi.CreateWordlistDTO
  >({
    mutationFn: (dto) => wordlistsApi.addWordlist(dto),
    onSuccess: (data) => {
      Sentry.addBreadcrumb({
        message: "Wordlist created successfully",
        type: "user.action",
        data: {
          wordlistName: data.name,
          language: data.languageCode,
        },
      });
      captureActivationEvent(posthog, ACTIVATION_EVENT_NAMES.WORDLIST_CREATED, {
        wordlistId: data.id,
        language: data.languageCode,
        source: "create_wordlist_modal",
      });
      queryClient.invalidateQueries({ queryKey: ["wordlists"] });

      // Call success callback and close modal immediately
      onSuccess?.(data);
      onClose();
    },
    onError: (error) => {
      Sentry.captureException(error, {
        data: {
          action: "create_wordlist",
          error_message: error.message,
        },
      });
      console.log(error);
      onError?.(error);
    },
  });

  const {
    control,
    handleSubmit,
    formState: { errors },
    reset,
    watch,
    setValue,
  } = useForm<CreateWordlistDTO>({
    defaultValues: {
      name: "",
      description: "",
      languageCode: undefined,
      pronunciationSystem: undefined,
    },
  });

  // Watch selected language to fetch pronunciation systems
  const selectedLanguage = watch("languageCode");

  // Fetch pronunciation systems for selected language
  const {
    data: pronunciationData,
    isLoading: isLoadingPronunciation,
    error: pronunciationError,
  } = useQuery({
    queryKey: ["pronunciationSystems", selectedLanguage],
    queryFn: () => wordlistsApi.getPronunciationSystems(selectedLanguage!),
    enabled: !!selectedLanguage,
    staleTime: 5 * 60 * 1000, // Cache for 5 minutes
    retry: 2, // Retry on failure
  });

  // Auto-select default pronunciation system when language changes
  useEffect(() => {
    if (pronunciationData?.defaultSystem) {
      // Only set if no pronunciation system is currently selected or if language changed
      const currentPronunciation = watch("pronunciationSystem");
      if (
        !currentPronunciation ||
        !pronunciationData.supportedSystems.includes(currentPronunciation)
      ) {
        setValue("pronunciationSystem", pronunciationData.defaultSystem);
      }
    }
  }, [pronunciationData, setValue, watch]);

  // Wizard navigation functions
  const nextStep = () => {
    if (currentStep < totalSteps) {
      setCurrentStep((prev) => prev + 1);
    }
  };

  const prevStep = () => {
    if (currentStep > 1) {
      // Handle backwards navigation through skipped steps
      if (currentStep === 4 && !pronunciationData?.canChange) {
        // Skip step 3 (pronunciation) and go directly to step 2 (language)
        setCurrentStep(2);
      } else {
        setCurrentStep((prev) => prev - 1);
      }
    }
  };

  const handleLanguageSelection = (
    languageCode: string,
    onChange: (value: string) => void,
  ) => {
    // Set the language via react-hook-form
    onChange(languageCode);

    // Auto-advance after a delay to allow for visual feedback
    setTimeout(() => {
      nextStep(); // Always go to next step, the step content will handle what to show
    }, 300);
  };

  // Animation effect
  useEffect(() => {
    if (visible) {
      Animated.parallel([
        Animated.timing(slideAnim, {
          toValue: 0,
          duration: 300,
          useNativeDriver: true,
        }),
        Animated.timing(backdropAnim, {
          toValue: 1,
          duration: 300,
          useNativeDriver: true,
        }),
      ]).start();
    } else {
      Animated.parallel([
        Animated.timing(slideAnim, {
          toValue: SCREEN_HEIGHT,
          duration: 300,
          useNativeDriver: true,
        }),
        Animated.timing(backdropAnim, {
          toValue: 0,
          duration: 300,
          useNativeDriver: true,
        }),
      ]).start(() => {
        reset(); // Reset form when modal is fully closed
        setCurrentStep(1); // Reset to first step
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [visible]);

  // Keyboard event listeners for iOS only (Android handles this automatically)
  useEffect(() => {
    if (!visible || Platform.OS !== "ios") return;

    const keyboardWillShow = (event: any) => {
      const keyboardHeight = event.endCoordinates.height;
      Animated.timing(keyboardOffsetAnim, {
        toValue: -keyboardHeight, // Move modal up by full keyboard height
        duration: event.duration || 250,
        useNativeDriver: true,
      }).start();
    };

    const keyboardWillHide = (event: any) => {
      Animated.timing(keyboardOffsetAnim, {
        toValue: 0,
        duration: event.duration || 250,
        useNativeDriver: true,
      }).start();
    };

    const showListener = Keyboard.addListener(
      "keyboardWillShow",
      keyboardWillShow,
    );
    const hideListener = Keyboard.addListener(
      "keyboardWillHide",
      keyboardWillHide,
    );

    return () => {
      showListener.remove();
      hideListener.remove();
    };
  }, [visible, keyboardOffsetAnim]);

  const handleFormSubmit = (data: CreateWordlistDTO) => {
    const payload = {
      ...data,
      hasOptimisticSubscription,
    };
    return mutation.mutateAsync(payload);
  };

  // Step validation functions
  const canProceedFromStep = (step: number): boolean => {
    switch (step) {
      case 1:
        const name = watch("name");
        return !!name && name.length >= 3;
      case 2:
        const language = watch("languageCode");
        return !!language;
      case 3:
        // If no pronunciation options, this is the description step (optional)
        if (!pronunciationData?.canChange) {
          return true; // Description is optional
        }
        // Otherwise this is the pronunciation step
        const pronunciation = watch("pronunciationSystem");
        return !!pronunciation;
      case 4:
        return true; // Description is optional
      default:
        return false;
    }
  };

  // Step components
  const renderStepContent = () => {
    switch (currentStep) {
      case 1:
        return renderStep1(); // Name
      case 2:
        return renderStep3(); // Language selection
      case 3:
        // If no pronunciation options, show description at step 3
        if (!pronunciationData?.canChange) {
          return renderStep2(); // Description
        }
        return renderStep4(); // Pronunciation (after language)
      case 4:
        return renderStep2(); // Description moved to step 4
      default:
        return null;
    }
  };

  // Step 1: Name
  const renderStep1 = () => (
    <View style={styles.stepContainer}>
      <Text style={styles.stepTitle}>
        {t("createWordlist.step1.title", {
          defaultValue: "What's your wordlist called?",
        })}
      </Text>

      <View style={styles.stepInputContainer}>
        <Controller
          control={control}
          name="name"
          rules={{
            required: t("createWordlist.nameRequired"),
            minLength: {
              value: 3,
              message: t("createWordlist.nameMinLength"),
            },
            maxLength: {
              value: 50,
              message: t("createWordlist.nameMaxLength"),
            },
          }}
          render={({ field: { onChange, onBlur, value } }) => (
            <TextInput
              style={[styles.stepInput, errors.name && styles.inputError]}
              placeholder={t("createWordlist.namePlaceholder")}
              placeholderTextColor={theme.colors.text.placeholder}
              value={value}
              onChangeText={onChange}
              onBlur={onBlur}
              autoCapitalize="words"
              maxLength={50}
              autoFocus
              accessibilityLabel={t("createWordlist.nameLabel")}
              accessibilityHint="Enter a name for your new wordlist"
            />
          )}
        />
        {errors.name && (
          <Text style={styles.errorText}>{errors.name.message}</Text>
        )}
      </View>
    </View>
  );

  // Step 2: Description (now step 4)
  const renderStep2 = () => (
    <View style={styles.stepContainer}>
      <Text style={styles.stepTitle}>
        {t("createWordlist.step2.title", {
          defaultValue: "Add a description?",
        })}
      </Text>
      <Text style={styles.stepOptionalText}>
        {t("createWordlist.step2.optional", {
          defaultValue: "Optional",
        })}
      </Text>

      <View style={styles.step2InputContainer}>
        <Controller
          control={control}
          name="description"
          rules={{
            maxLength: {
              value: 200,
              message: t("createWordlist.descriptionMaxLength"),
            },
          }}
          render={({ field: { onChange, onBlur, value } }) => (
            <TextInput
              style={[styles.stepInput, styles.stepTextArea]}
              placeholder={t("createWordlist.descriptionPlaceholder")}
              placeholderTextColor={theme.colors.text.placeholder}
              value={value}
              onChangeText={onChange}
              onBlur={onBlur}
              multiline
              numberOfLines={4}
              maxLength={200}
              autoFocus
              blurOnSubmit={false}
              textAlignVertical="top"
              accessibilityLabel={t("createWordlist.descriptionLabel")}
              accessibilityHint="Optionally describe what this wordlist is for"
            />
          )}
        />
        {errors.description && (
          <Text style={styles.errorText}>{errors.description.message}</Text>
        )}
      </View>
    </View>
  );

  // Step 3: Language (now step 2)
  const renderStep3 = () => (
    <View style={styles.stepContainer}>
      <Text style={styles.stepTitle}>
        {t("createWordlist.step3.title", {
          defaultValue: "Which language are you learning?",
        })}
      </Text>

      <View style={styles.stepInputContainer}>
        <Controller
          control={control}
          name="languageCode"
          rules={{
            required: t("createWordlist.languageRequired"),
          }}
          render={({ field: { onChange, value } }) => (
            <ScrollView
              horizontal
              showsHorizontalScrollIndicator={false}
              style={styles.languageScroll}
              contentContainerStyle={styles.languageScrollContent}
              accessibilityLabel="Select language for wordlist"
            >
              {LANGUAGES.map((lang) => (
                <TouchableOpacity
                  key={lang.code}
                  style={[
                    styles.wizardLanguageItem,
                    value === lang.code && styles.wizardLanguageItemSelected,
                  ]}
                  onPress={() => handleLanguageSelection(lang.code, onChange)}
                  accessibilityRole="radio"
                  accessibilityLabel={`Select ${t(`dashboard.languages.${lang.name.toLowerCase()}`)} language`}
                  accessibilityState={{ selected: value === lang.code }}
                >
                  <Text style={styles.wizardLanguageFlag}>{lang.flag}</Text>
                  <Text
                    style={[
                      styles.wizardLanguageName,
                      value === lang.code && styles.wizardLanguageNameSelected,
                    ]}
                  >
                    {t(`dashboard.languages.${lang.name.toLowerCase()}`)}
                  </Text>
                </TouchableOpacity>
              ))}
            </ScrollView>
          )}
        />
        {errors.languageCode && (
          <Text style={styles.errorText}>{errors.languageCode.message}</Text>
        )}
      </View>
    </View>
  );

  // Step 4: Pronunciation (conditional)
  const renderStep4 = () => {
    if (!pronunciationData?.canChange) {
      // Skip this step if no pronunciation options
      return null;
    }

    return (
      <View style={styles.stepContainer}>
        <Text style={styles.stepTitle}>
          {t("createWordlist.step4.title", {
            defaultValue: "Choose pronunciation system",
          })}
        </Text>

        <View style={styles.stepInputContainer}>
          <Controller
            control={control}
            name="pronunciationSystem"
            rules={{
              validate: (value) => {
                if (pronunciationData?.canChange && !value) {
                  return t("createWordlist.pronunciationRequired");
                }
                return true;
              },
            }}
            render={({ field: { onChange, value } }) => (
              <View>
                {isLoadingPronunciation ? (
                  <View style={styles.loadingContainer}>
                    <ActivityIndicator
                      size="small"
                      color={theme.colors.primary}
                    />
                  </View>
                ) : pronunciationError ? (
                  <View style={styles.errorContainer}>
                    <Text style={styles.errorText}>
                      {t("createWordlist.pronunciationLoadError")}
                    </Text>
                  </View>
                ) : (
                  <ScrollView
                    horizontal
                    showsHorizontalScrollIndicator={false}
                    style={styles.languageScroll}
                    contentContainerStyle={styles.languageScrollContent}
                  >
                    {pronunciationData?.supportedSystems.map((system) => (
                      <TouchableOpacity
                        key={system}
                        style={[
                          styles.wizardPronunciationItem,
                          value === system &&
                            styles.wizardPronunciationItemSelected,
                        ]}
                        onPress={() => {
                          onChange(system);
                          // Auto-advance to next step after pronunciation selection
                          setTimeout(() => {
                            nextStep();
                          }, 300); // Small delay for visual feedback
                        }}
                        accessibilityRole="radio"
                        accessibilityLabel={`Select ${t(`pronunciationSystems.${system}`)} pronunciation system`}
                        accessibilityState={{ selected: value === system }}
                      >
                        <Text
                          style={[
                            styles.wizardPronunciationName,
                            value === system &&
                              styles.wizardPronunciationNameSelected,
                          ]}
                        >
                          {t(`pronunciationSystems.${system}`)}
                        </Text>
                      </TouchableOpacity>
                    ))}
                  </ScrollView>
                )}
              </View>
            )}
          />
          {errors.pronunciationSystem && (
            <Text style={styles.errorText}>
              {errors.pronunciationSystem.message}
            </Text>
          )}
        </View>
      </View>
    );
  };

  if (!visible) return null;

  return (
    <Modal
      transparent
      visible={visible}
      animationType="none"
      onRequestClose={onClose}
      accessibilityViewIsModal={true}
    >
      <TouchableWithoutFeedback onPress={handleClose}>
        <Animated.View
          style={[
            styles.backdrop,
            {
              opacity: backdropAnim,
            },
          ]}
        />
      </TouchableWithoutFeedback>

      <Animated.View
        style={[
          styles.modalContent,
          {
            transform: [
              { translateY: slideAnim },
              // Only apply keyboard offset on iOS, Android handles this automatically
              ...(Platform.OS === "ios"
                ? [{ translateY: keyboardOffsetAnim }]
                : []),
            ],
          },
        ]}
        {...panResponder.panHandlers}
      >
        {/* Header with Progress */}
        <View style={styles.wizardHeader}>
          <View style={styles.handle} />
          <View style={styles.headerContent}>
            <View style={styles.titleContainer}>
              <Text style={styles.title}>{t("createWordlist.title")}</Text>
            </View>
            <TouchableOpacity
              style={styles.closeButton}
              onPress={handleClose}
              accessibilityRole="button"
              accessibilityLabel={t("common.close")}
              accessibilityHint="Close the create wordlist dialog"
            >
              <Ionicons name="close" size={24} color="#636E72" />
            </TouchableOpacity>
          </View>

          {/* Progress Indicator */}
          <View style={styles.progressContainer}>
            <View style={styles.progressDots}>
              {[1, 2, 3, 4].map((step) => {
                // Skip step 3 (pronunciation) if no pronunciation options
                if (step === 3 && !pronunciationData?.canChange) {
                  return null;
                }

                // Calculate the visual step and actual step for proper coloring
                let isActive = false;
                let isCompleted = false;

                if (pronunciationData?.canChange) {
                  // Normal 4-step flow: 1=Name, 2=Language, 3=Pronunciation, 4=Description
                  isActive = currentStep === step;
                  isCompleted = currentStep > step;
                } else {
                  // 3-step flow: 1=Name, 2=Language, 3=Description (step 4 shows as step 3)
                  if (step === 4) {
                    // Step 4 (description) should behave as visual step 3
                    isActive = currentStep === 4;
                    isCompleted = currentStep > 4; // Never completed in this flow
                  } else {
                    // Steps 1 and 2 work normally
                    isActive = currentStep === step;
                    isCompleted = currentStep > step && currentStep !== 4; // Handle transition to description
                    // When on step 4 (description), steps 1 and 2 should be completed
                    if (currentStep === 4 && step < 3) {
                      isCompleted = true;
                      isActive = false;
                    }
                  }
                }

                return (
                  <View
                    key={step}
                    style={[
                      styles.progressDot,
                      isActive && styles.progressDotActive,
                      isCompleted && styles.progressDotCompleted,
                    ]}
                  />
                );
              })}
            </View>
            <Text style={styles.progressText}>
              {pronunciationData?.canChange
                ? `${currentStep} of ${totalSteps}`
                : `${currentStep === 4 ? 3 : currentStep} of ${totalSteps - 1}`}
            </Text>
          </View>
        </View>

        {/* Step Content */}
        <View style={styles.stepContentContainer}>{renderStepContent()}</View>

        {/* Navigation Buttons */}
        <View style={styles.wizardActions}>
          {/* Navigation Buttons */}
          <View style={styles.navigationButtons}>
            {currentStep > 1 && (
              <TouchableOpacity
                style={[styles.button, styles.backButton]}
                onPress={prevStep}
                accessibilityRole="button"
                accessibilityLabel={t("common.back")}
              >
                <Ionicons
                  name="chevron-back"
                  size={20}
                  color={theme.colors.text.secondary}
                />
                <Text style={styles.backButtonText}>{t("common.back")}</Text>
              </TouchableOpacity>
            )}

            {/* Next/Create Button */}
            {currentStep === 4 ||
            (currentStep === 3 && !pronunciationData?.canChange) ? (
              // Description step - Only show Create button (no skip needed since optional)
              <TouchableOpacity
                style={[
                  styles.button,
                  styles.createButton,
                  mutation.isPending && styles.buttonDisabled,
                ]}
                onPress={handleSubmit(handleFormSubmit)}
                disabled={mutation.isPending}
                accessibilityRole="button"
                accessibilityLabel={
                  mutation.isPending
                    ? "Creating wordlist..."
                    : t("createWordlist.createButton")
                }
              >
                {mutation.isPending ? (
                  <ActivityIndicator color="#FFFFFF" size="small" />
                ) : (
                  <>
                    <Text style={styles.createButtonText}>
                      {t("createWordlist.createButton")}
                    </Text>
                    <Ionicons name="add-circle" size={20} color="#FFFFFF" />
                  </>
                )}
              </TouchableOpacity>
            ) : (
              // Regular Next Button
              <TouchableOpacity
                style={[
                  styles.button,
                  styles.nextButton,
                  !canProceedFromStep(currentStep) && styles.buttonDisabled,
                ]}
                onPress={() => {
                  if (currentStep === 3 && !pronunciationData?.canChange) {
                    // Skip to step 4 (description) if no pronunciation options
                    setCurrentStep(4);
                  } else {
                    nextStep();
                  }
                }}
                disabled={!canProceedFromStep(currentStep)}
                accessibilityRole="button"
                accessibilityLabel="Next"
              >
                <Text style={styles.nextButtonText}>
                  {t("common.next", { defaultValue: "Next" })}
                </Text>
                <Ionicons name="chevron-forward" size={20} color="#FFFFFF" />
              </TouchableOpacity>
            )}
          </View>

          {/* Submit Error */}
          {mutation.error && (
            <View style={styles.submitError}>
              <Text style={styles.errorText}>
                {t("createWordlist.errorMessage")}
              </Text>
            </View>
          )}
        </View>
      </Animated.View>
    </Modal>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    backdrop: {
      position: "absolute",
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor:
        theme.mode === "light" ? "rgba(0, 0, 0, 0.4)" : "rgba(0, 0, 0, 0.6)",
    },
    modalContent: {
      backgroundColor: theme.colors.background.surface,
      borderTopLeftRadius: 24,
      borderTopRightRadius: 24,
      maxHeight: SCREEN_HEIGHT * 0.85,
      shadowColor: theme.colors.text.primary,
      shadowOffset: { width: 0, height: -4 },
      shadowOpacity: 0.1,
      shadowRadius: 12,
      elevation: 8,
      position: "absolute",
      bottom: 0,
      left: 0,
      right: 0,
    },
    header: {
      paddingTop: 10,
      paddingBottom: 16,
      paddingHorizontal: 20,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.ui.border,
    },
    handle: {
      width: 40,
      height: 4,
      backgroundColor: theme.colors.ui.disabled,
      borderRadius: 2,
      alignSelf: "center",
      marginBottom: 16,
    },
    titleContainer: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      flex: 1,
    },
    title: {
      fontSize: 20,
      fontWeight: "600",
      color: theme.colors.text.primary,
      textAlign: "center",
    },
    closeButton: {
      position: "absolute",
      right: 20,
      top: 0,
      width: 32,
      height: 32,
      borderRadius: 16,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
    },
    form: {
      flex: 1,
      paddingHorizontal: 20,
      paddingTop: 16,
    },
    inputGroup: {
      marginBottom: 18,
    },
    firstInputGroup: {
      marginTop: 6,
    },
    reducedMarginGroup: {
      marginBottom: 12,
    },
    addDescriptionButton: {
      flexDirection: "row",
      alignItems: "center",
      paddingVertical: 12,
      paddingHorizontal: 16,
      borderRadius: 12,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      borderStyle: "dashed",
      backgroundColor: theme.colors.background.default,
    },
    addDescriptionIcon: {
      marginRight: 8,
    },
    addDescriptionText: {
      fontSize: 16,
      color: theme.colors.primary,
      fontWeight: "500",
    },
    descriptionHeader: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      marginBottom: 8,
    },
    removeDescriptionButton: {
      padding: 4,
    },
    compactGuidelinesButton: {
      flexDirection: "row",
      alignItems: "center",
      paddingVertical: 10,
      paddingHorizontal: 14,
      borderRadius: 8,
      backgroundColor: theme.colors.background.subtle,
      marginBottom: 6,
    },
    compactGuidelinesIcon: {
      marginRight: 8,
    },
    compactGuidelinesText: {
      flex: 1,
      fontSize: 14,
      color: theme.colors.text.secondary,
      fontWeight: "500",
    },
    // Wizard-specific styles
    wizardHeader: {
      paddingTop: 8,
      paddingBottom: 12,
      paddingHorizontal: 20,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.ui.border,
    },
    headerContent: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      marginBottom: 12,
      position: "relative",
      minHeight: 32,
    },
    progressContainer: {
      alignItems: "center",
    },
    progressDots: {
      flexDirection: "row",
      justifyContent: "center",
      marginBottom: 6,
    },
    progressDot: {
      width: 8,
      height: 8,
      borderRadius: 4,
      backgroundColor: theme.colors.ui.disabled,
      marginHorizontal: 4,
    },
    progressDotActive: {
      backgroundColor: theme.colors.primary,
    },
    progressDotCompleted: {
      backgroundColor: theme.colors.success,
    },
    progressText: {
      fontSize: 12,
      color: theme.colors.text.secondary,
      fontWeight: "500",
    },
    stepContentContainer: {
      flex: 1,
      paddingHorizontal: 20,
    },
    stepContainer: {
      flex: 1,
      justifyContent: "center",
      paddingVertical: 16,
    },
    stepTitle: {
      fontSize: 22,
      fontWeight: "600",
      color: theme.colors.text.primary,
      textAlign: "center",
      marginBottom: 8,
    },
    stepOptionalText: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      textAlign: "center",
      marginBottom: 20,
      fontStyle: "italic",
    },
    stepSubtitle: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      textAlign: "center",
      marginBottom: 24,
      lineHeight: 22,
    },
    stepInputContainer: {
      marginTop: 20,
    },
    step2InputContainer: {
      marginTop: 12,
    },
    stepInput: {
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      borderRadius: 16,
      paddingHorizontal: 20,
      paddingVertical: 18,
      fontSize: 18,
      color: theme.colors.text.primary,
      backgroundColor: theme.colors.ui.inputBackground,
      textAlign: "center",
    },
    stepTextArea: {
      height: 100,
      textAlignVertical: "top",
      textAlign: "left",
      paddingTop: 18,
    },
    languageScrollContent: {
      paddingHorizontal: 10,
    },
    wizardLanguageItem: {
      alignItems: "center",
      paddingHorizontal: 20,
      paddingVertical: 16,
      borderRadius: 16,
      borderWidth: 2,
      borderColor: theme.colors.ui.border,
      marginHorizontal: 8,
      backgroundColor: theme.colors.background.default,
      minWidth: 100,
    },
    wizardLanguageItemSelected: {
      borderColor: theme.colors.primary,
      backgroundColor:
        theme.mode === "light" ? "#FFF5F0" : theme.colors.background.secondary,
    },
    wizardLanguageFlag: {
      fontSize: 32,
      marginBottom: 8,
    },
    wizardLanguageName: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      fontWeight: "500",
      textAlign: "center",
    },
    wizardLanguageNameSelected: {
      color: theme.colors.primary,
      fontWeight: "600",
    },
    wizardPronunciationItem: {
      paddingHorizontal: 20,
      paddingVertical: 16,
      borderRadius: 16,
      borderWidth: 2,
      borderColor: theme.colors.ui.border,
      marginHorizontal: 8,
      backgroundColor: theme.colors.background.default,
      minWidth: 120,
    },
    wizardPronunciationItemSelected: {
      borderColor: theme.colors.primary,
      backgroundColor:
        theme.mode === "light" ? "#FFF5F0" : theme.colors.background.secondary,
    },
    wizardPronunciationName: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      fontWeight: "500",
      textAlign: "center",
    },
    wizardPronunciationNameSelected: {
      color: theme.colors.primary,
      fontWeight: "600",
    },
    wizardActions: {
      paddingHorizontal: 20,
      paddingTop: 16,
      paddingBottom: Platform.OS === "ios" ? 34 : 24,
      backgroundColor: theme.colors.background.surface,
    },
    navigationButtons: {
      flexDirection: "row",
      gap: 12,
    },
    backButton: {
      flex: 1,
      backgroundColor: theme.colors.background.surface,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
    },
    backButtonText: {
      fontSize: 16,
      fontWeight: "600",
      color: theme.colors.text.secondary,
      marginLeft: 4,
    },
    nextButton: {
      flex: 2,
      backgroundColor: theme.colors.primary,
    },
    nextButtonText: {
      fontSize: 16,
      fontWeight: "600",
      color: theme.colors.text.inverse,
      marginRight: 4,
    },
    step2Actions: {
      flexDirection: "row",
      gap: 12,
      flex: 1,
    },
    skipButton: {
      flex: 1,
      backgroundColor: theme.colors.background.surface,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
    },
    skipButtonText: {
      fontSize: 16,
      fontWeight: "600",
      color: theme.colors.text.secondary,
    },
    label: {
      fontSize: 16,
      fontWeight: "500",
      color: theme.colors.text.primary,
      marginBottom: 8,
    },
    required: {
      color: theme.colors.primary,
    },
    input: {
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      borderRadius: 12,
      paddingHorizontal: 16,
      paddingVertical: 14,
      fontSize: 16,
      color: theme.colors.text.primary,
      backgroundColor: theme.colors.ui.inputBackground,
    },
    inputError: {
      borderColor: theme.colors.error,
    },
    textArea: {
      height: 80,
      textAlignVertical: "top",
      paddingTop: 14,
    },
    errorText: {
      color: theme.colors.error,
      fontSize: 14,
      marginTop: 6,
    },
    languageScroll: {
      marginTop: 8,
    },
    languageItem: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: 16,
      paddingVertical: 12,
      borderRadius: 12,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      marginRight: 10,
      backgroundColor: theme.colors.background.default,
    },
    languageItemSelected: {
      borderColor: theme.colors.primary,
      backgroundColor:
        theme.mode === "light" ? "#FFF5F0" : theme.colors.background.secondary,
    },
    languageFlag: {
      fontSize: 24,
      marginRight: 8,
    },
    languageName: {
      fontSize: 16,
      color: theme.colors.text.secondary,
    },
    languageNameSelected: {
      color: theme.colors.primary,
      fontWeight: "500",
    },
    submitError: {
      marginTop: -8,
      marginBottom: 16,
    },
    actions: {
      flexDirection: "row",
      paddingHorizontal: 20,
      paddingTop: 16,
      paddingBottom: Platform.OS === "ios" ? 34 : 24,
      borderTopWidth: 1,
      borderTopColor: theme.colors.ui.border,
      gap: 12,
      backgroundColor: theme.colors.background.surface,
    },
    button: {
      flex: 1,
      paddingVertical: 16,
      borderRadius: 12,
      flexDirection: "row",
      justifyContent: "center",
      alignItems: "center",
    },
    cancelButton: {
      backgroundColor: theme.colors.background.surface,
    },
    cancelButtonText: {
      fontSize: 16,
      fontWeight: "600",
      color: theme.colors.text.secondary,
    },
    createButton: {
      backgroundColor: theme.colors.primary,
      shadowColor: theme.colors.primary,
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.2,
      shadowRadius: 8,
      elevation: 5,
    },
    createButtonText: {
      fontSize: 16,
      fontWeight: "600",
      color: theme.colors.text.inverse,
    },
    buttonIcon: {
      marginLeft: 6,
    },
    buttonDisabled: {
      opacity: 0.7,
    },
    helpText: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      marginBottom: 12,
      fontStyle: "italic",
    },
    loadingContainer: {
      paddingVertical: 20,
      alignItems: "center",
    },
    pronunciationItem: {
      paddingHorizontal: 16,
      paddingVertical: 12,
      borderRadius: 12,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      marginRight: 10,
      backgroundColor: theme.colors.background.default,
    },
    pronunciationItemSelected: {
      borderColor: theme.colors.primary,
      backgroundColor:
        theme.mode === "light" ? "#FFF5F0" : theme.colors.background.secondary,
    },
    pronunciationName: {
      fontSize: 16,
      color: theme.colors.text.secondary,
    },
    pronunciationNameSelected: {
      color: theme.colors.primary,
      fontWeight: "500",
    },
    errorContainer: {
      paddingVertical: 16,
      paddingHorizontal: 12,
      backgroundColor:
        theme.mode === "light" ? "#FFF5F5" : theme.colors.background.secondary,
      borderRadius: 8,
      borderWidth: 1,
      borderColor: theme.mode === "light" ? "#FFE6E6" : theme.colors.ui.border,
    },
  });
