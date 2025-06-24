import * as wordlistsApi from "@/api/wordlists";
import { CreateWordlistDTO } from "@/api/wordlists";
import { useTheme } from "@/contexts/ThemeContext";
import { Ionicons } from "@expo/vector-icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import React, { useEffect, useRef, useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import {
  ActivityIndicator,
  Alert,
  Animated,
  Dimensions,
  Modal,
  PanResponder,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  TouchableWithoutFeedback,
  View,
} from "react-native";
import { ContentGuidelinesModal } from "./ContentGuidelinesModal";
import { captureException, addBreadcrumb } from "@/utils/sentry";
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
  const queryClient = useQueryClient();
  const { t } = useTranslation();
  const { theme } = useTheme();
  const [showContentGuidelines, setShowContentGuidelines] = useState(false);
  const [guidelinesExpanded, setGuidelinesExpanded] = useState(false);
  const guidelinesHeight = useRef(new Animated.Value(0)).current;
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
      addBreadcrumb("Wordlist created successfully", "user.action", {
        wordlistName: data.name,
        language: data.languageCode,
      });
      queryClient.invalidateQueries({ queryKey: ["wordlists"] });
      queryClient.invalidateQueries({ queryKey: ["dashboardStats"] });

      Alert.alert(
        t("common.success"),
        t("createWordlist.successMessage", { name: data.name }),
        [
          {
            text: t("common.ok"),
            onPress: () => onSuccess?.(data),
          },
        ],
      );
    },
    onError: (error) => {
      captureException(error, {
        user_action: {
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
      });
    }
  }, [visible]);

  const handleFormSubmit = (data: CreateWordlistDTO) =>
    mutation.mutateAsync(data);

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
            transform: [{ translateY: slideAnim }],
          },
        ]}
        {...panResponder.panHandlers}
      >
        {/* Header - Outside KeyboardAvoidingView */}
        <View style={styles.header}>
          <View style={styles.handle} />
          <Text
            style={styles.title}
            accessibilityRole="header"
            // accessibilityLevel={1} // Removed deprecated prop
          >
            {t("createWordlist.title")}
          </Text>
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

        <ScrollView
          style={styles.form}
          showsVerticalScrollIndicator={false}
          keyboardShouldPersistTaps="handled"
          contentContainerStyle={{ paddingBottom: 100 }}
        >
          {/* Name Input */}
          <View style={styles.inputGroup}>
            <Text style={styles.label}>
              {t("createWordlist.nameLabel")}{" "}
              <Text style={styles.required}>*</Text>
            </Text>
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
                  style={[styles.input, errors.name && styles.inputError]}
                  placeholder={t("createWordlist.namePlaceholder")}
                  placeholderTextColor="#B2BEC3"
                  value={value}
                  onChangeText={onChange}
                  onBlur={onBlur}
                  autoCapitalize="words"
                  maxLength={50}
                  accessibilityLabel={t("createWordlist.nameLabel")}
                  accessibilityHint="Enter a name for your new wordlist"
                  // accessibilityRequired={true} // Not a valid prop
                />
              )}
            />
            {errors.name && (
              <Text style={styles.errorText}>{errors.name.message}</Text>
            )}
          </View>

          {/* Description Input */}
          <View style={styles.inputGroup}>
            <Text style={styles.label}>
              {t("createWordlist.descriptionLabel")}
            </Text>
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
                  style={[styles.input, styles.textArea]}
                  placeholder={t("createWordlist.descriptionPlaceholder")}
                  placeholderTextColor="#B2BEC3"
                  value={value}
                  onChangeText={onChange}
                  onBlur={onBlur}
                  multiline
                  numberOfLines={3}
                  maxLength={200}
                  accessibilityLabel={t("createWordlist.descriptionLabel")}
                  accessibilityHint="Optionally describe what this wordlist is for"
                />
              )}
            />
            {errors.description && (
              <Text style={styles.errorText}>{errors.description.message}</Text>
            )}
          </View>

          {/* Language Selection */}
          <View style={styles.inputGroup}>
            <Text style={styles.label}>
              {t("createWordlist.languageLabel")}{" "}
              <Text style={styles.required}>*</Text>
            </Text>
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
                  accessibilityLabel="Select language for wordlist"
                >
                  {LANGUAGES.map((lang) => (
                    <TouchableOpacity
                      key={lang.code}
                      style={[
                        styles.languageItem,
                        value === lang.code && styles.languageItemSelected,
                      ]}
                      onPress={() => {
                        onChange(lang.code);
                      }}
                      accessibilityRole="radio"
                      accessibilityLabel={`Select ${t(`dashboard.languages.${lang.name.toLowerCase()}`)} language`}
                      accessibilityState={{ selected: value === lang.code }}
                    >
                      <Text style={styles.languageFlag}>{lang.flag}</Text>
                      <Text
                        style={[
                          styles.languageName,
                          value === lang.code && styles.languageNameSelected,
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
              <Text style={styles.errorText}>
                {errors.languageCode.message}
              </Text>
            )}
          </View>

          {/* Pronunciation System Selection */}
          {pronunciationData?.canChange && (
            <View style={styles.inputGroup}>
              <Text style={styles.label}>
                {t("createWordlist.pronunciationLabel")}
              </Text>
              <Text style={styles.helpText}>
                {t("createWordlist.pronunciationHelp")}
              </Text>
              <Controller
                control={control}
                name="pronunciationSystem"
                rules={{
                  validate: (value) => {
                    // If pronunciation system selection is available and required
                    if (pronunciationData?.canChange && !value) {
                      return t("createWordlist.pronunciationRequired");
                    }
                    // If value is set, ensure it's supported
                    if (
                      value &&
                      pronunciationData?.supportedSystems &&
                      !pronunciationData.supportedSystems.includes(value)
                    ) {
                      return t("createWordlist.pronunciationNotSupported");
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
                        <Text style={styles.helpText}>
                          {t("createWordlist.pronunciationDefaultWillBeUsed")}
                        </Text>
                      </View>
                    ) : (
                      <ScrollView
                        horizontal
                        showsHorizontalScrollIndicator={false}
                        style={styles.languageScroll}
                      >
                        {pronunciationData?.supportedSystems.map((system) => (
                          <TouchableOpacity
                            key={system}
                            style={[
                              styles.pronunciationItem,
                              value === system &&
                                styles.pronunciationItemSelected,
                            ]}
                            onPress={() => {
                              onChange(system);
                            }}
                            accessibilityRole="radio"
                            accessibilityLabel={`Select ${t(`pronunciationSystems.${system}`)} pronunciation system`}
                            accessibilityState={{
                              selected: value === system,
                            }}
                          >
                            <Text
                              style={[
                                styles.pronunciationName,
                                value === system &&
                                  styles.pronunciationNameSelected,
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
          )}

          {/* Content Guidelines Notice - Expandable */}
          <TouchableOpacity
            style={styles.guidelinesNotice}
            onPress={() => {
              const toValue = guidelinesExpanded ? 0 : 1;
              setGuidelinesExpanded(!guidelinesExpanded);
              Animated.timing(guidelinesHeight, {
                toValue,
                duration: 300,
                useNativeDriver: false,
              }).start();
            }}
            activeOpacity={0.8}
          >
            <View style={styles.guidelinesHeader}>
              <Ionicons
                name="information-circle"
                size={20}
                color={theme.colors.primary}
              />
              <Text style={styles.guidelinesTitle}>
                {t("createWordlist.contentGuidelines")}
              </Text>
              <View style={styles.expandIcon}>
                <Ionicons
                  name={guidelinesExpanded ? "chevron-up" : "chevron-down"}
                  size={20}
                  color={theme.colors.text.secondary}
                />
              </View>
            </View>

            <Animated.View
              style={{
                maxHeight: guidelinesHeight.interpolate({
                  inputRange: [0, 1],
                  outputRange: [0, 200],
                }),
                overflow: "hidden",
              }}
            >
              <Text style={styles.guidelinesText}>
                {t("createWordlist.contentGuidelinesText")}
              </Text>
              <TouchableOpacity
                style={styles.guidelinesLink}
                onPress={(e) => {
                  e.stopPropagation();
                  setShowContentGuidelines(true);
                }}
                accessibilityRole="button"
                accessibilityLabel={t("createWordlist.viewGuidelines")}
                accessibilityHint="View detailed content guidelines"
              >
                <Text style={styles.guidelinesLinkText}>
                  {t("createWordlist.viewGuidelines")}
                </Text>
                <Ionicons
                  name="arrow-forward"
                  size={16}
                  color={theme.colors.primary}
                />
              </TouchableOpacity>
            </Animated.View>
          </TouchableOpacity>

          {/* Submit Error */}
          {mutation.error && (
            <View style={styles.submitError}>
              <Text style={styles.errorText}>
                {t("createWordlist.errorMessage")}
              </Text>
            </View>
          )}
        </ScrollView>

        {/* Action Buttons */}
        <View style={styles.actions}>
          <TouchableOpacity
            style={[styles.button, styles.cancelButton]}
            onPress={handleClose}
            accessibilityRole="button"
            accessibilityLabel={t("common.cancel")}
            accessibilityHint="Cancel wordlist creation and close dialog"
          >
            <Text style={styles.cancelButtonText}>{t("common.cancel")}</Text>
          </TouchableOpacity>

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
            accessibilityHint="Create the new wordlist with entered details"
            accessibilityState={{ disabled: mutation.isPending }}
          >
            {mutation.isPending ? (
              <ActivityIndicator color="#FFFFFF" size="small" />
            ) : (
              <>
                <Text style={styles.createButtonText}>
                  {t("createWordlist.createButton")}
                </Text>
                <Ionicons
                  name="add-circle"
                  size={20}
                  color="#FFFFFF"
                  style={styles.buttonIcon}
                />
              </>
            )}
          </TouchableOpacity>
        </View>
      </Animated.View>

      <ContentGuidelinesModal
        visible={showContentGuidelines}
        onClose={() => setShowContentGuidelines(false)}
      />
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
      height: SCREEN_HEIGHT * 0.9,
      maxHeight: SCREEN_HEIGHT * 0.9,
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
      paddingTop: 12,
      paddingBottom: 20,
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
    title: {
      fontSize: 24,
      fontWeight: "600",
      color: theme.colors.text.primary,
      textAlign: "center",
    },
    closeButton: {
      position: "absolute",
      right: 20,
      top: 32,
      width: 32,
      height: 32,
      borderRadius: 16,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
    },
    form: {
      padding: 20,
    },
    guidelinesNotice: {
      backgroundColor:
        theme.mode === "light" ? "#FFF9F0" : theme.colors.background.secondary,
      borderWidth: 1,
      borderColor: theme.mode === "light" ? "#FFE6CC" : theme.colors.ui.border,
      borderRadius: 12,
      padding: 16,
      marginBottom: 24,
    },
    guidelinesHeader: {
      flexDirection: "row",
      alignItems: "center",
    },
    guidelinesTitle: {
      fontSize: 16,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginLeft: 8,
      flex: 1,
    },
    expandIcon: {
      marginLeft: 8,
    },
    guidelinesText: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      lineHeight: 20,
      marginTop: 12,
      marginBottom: 12,
    },
    guidelinesLink: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      paddingVertical: 8,
    },
    guidelinesLinkText: {
      fontSize: 14,
      color: theme.colors.primary,
      fontWeight: "500",
    },
    inputGroup: {
      marginBottom: 24,
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
      paddingBottom: 32,
      borderTopWidth: 1,
      borderTopColor: theme.colors.ui.border,
      gap: 12,
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
