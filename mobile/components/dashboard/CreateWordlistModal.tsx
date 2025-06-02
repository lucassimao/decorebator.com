import React, { useRef, useEffect } from "react";
import {
  View,
  Text,
  StyleSheet,
  Modal,
  TouchableOpacity,
  TouchableWithoutFeedback,
  TextInput,
  ScrollView,
  Animated,
  Dimensions,
  KeyboardAvoidingView,
  Platform,
  ActivityIndicator,
  Alert,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useForm, Controller } from "react-hook-form";
import { CreateWordlistDTO } from "@/api/wordlists";
import * as wordlistsApi from "@/api/wordlists";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
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

  const mutation = useMutation<
    wordlistsApi.Wordlist,
    Error,
    wordlistsApi.CreateWordlistDTO
  >({
    mutationFn: (dto) => wordlistsApi.addWordlist(dto),
    onSuccess: (data) => {
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
      console.log(error);
      onError?.(error);
    },
  });

  const {
    control,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<CreateWordlistDTO>({
    defaultValues: {
      name: "",
      description: "",
      languageCode: undefined,
    },
  });

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
    >
      <TouchableWithoutFeedback onPress={onClose}>
        <Animated.View
          style={[
            styles.backdrop,
            {
              opacity: backdropAnim,
            },
          ]}
        />
      </TouchableWithoutFeedback>

      <KeyboardAvoidingView
        behavior={Platform.OS === "ios" ? "padding" : "height"}
        style={styles.container}
      >
        <Animated.View
          style={[
            styles.modalContent,
            {
              transform: [{ translateY: slideAnim }],
            },
          ]}
        >
          {/* Header */}
          <View style={styles.header}>
            <View style={styles.handle} />
            <Text style={styles.title}>{t("createWordlist.title")}</Text>
            <TouchableOpacity style={styles.closeButton} onPress={onClose}>
              <Ionicons name="close" size={24} color="#636E72" />
            </TouchableOpacity>
          </View>

          <ScrollView style={styles.form} showsVerticalScrollIndicator={false}>
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
                    autoFocus
                    style={[styles.input, errors.name && styles.inputError]}
                    placeholder={t("createWordlist.namePlaceholder")}
                    placeholderTextColor="#B2BEC3"
                    value={value}
                    onChangeText={onChange}
                    onBlur={onBlur}
                    autoCapitalize="words"
                    maxLength={50}
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
                  />
                )}
              />
              {errors.description && (
                <Text style={styles.errorText}>
                  {errors.description.message}
                </Text>
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
              onPress={onClose}
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
      </KeyboardAvoidingView>
    </Modal>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: "flex-end",
  },
  backdrop: {
    position: "absolute",
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: "rgba(0, 0, 0, 0.4)",
  },
  modalContent: {
    backgroundColor: "#FFFFFF",
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    maxHeight: SCREEN_HEIGHT * 0.9,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: -4 },
    shadowOpacity: 0.1,
    shadowRadius: 12,
    elevation: 8,
  },
  header: {
    paddingTop: 12,
    paddingBottom: 20,
    paddingHorizontal: 20,
    borderBottomWidth: 1,
    borderBottomColor: "#F0F0F0",
  },
  handle: {
    width: 40,
    height: 4,
    backgroundColor: "#DFE6E9",
    borderRadius: 2,
    alignSelf: "center",
    marginBottom: 16,
  },
  title: {
    fontSize: 24,
    fontWeight: "600",
    color: "#2D3436",
    textAlign: "center",
  },
  closeButton: {
    position: "absolute",
    right: 20,
    top: 32,
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: "#F5F5F5",
    justifyContent: "center",
    alignItems: "center",
  },
  form: {
    padding: 20,
  },
  inputGroup: {
    marginBottom: 24,
  },
  label: {
    fontSize: 16,
    fontWeight: "500",
    color: "#2D3436",
    marginBottom: 8,
  },
  required: {
    color: "#FF7B54",
  },
  input: {
    borderWidth: 1,
    borderColor: "#E0E0E0",
    borderRadius: 12,
    paddingHorizontal: 16,
    paddingVertical: 14,
    fontSize: 16,
    color: "#2D3436",
    backgroundColor: "#FAFAFA",
  },
  inputError: {
    borderColor: "#FF6B6B",
  },
  textArea: {
    height: 80,
    textAlignVertical: "top",
    paddingTop: 14,
  },
  errorText: {
    color: "#FF6B6B",
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
    borderColor: "#E0E0E0",
    marginRight: 10,
    backgroundColor: "#FAFAFA",
  },
  languageItemSelected: {
    borderColor: "#FF7B54",
    backgroundColor: "#FFF5F0",
  },
  languageFlag: {
    fontSize: 24,
    marginRight: 8,
  },
  languageName: {
    fontSize: 16,
    color: "#636E72",
  },
  languageNameSelected: {
    color: "#FF7B54",
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
    borderTopColor: "#F0F0F0",
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
    backgroundColor: "#F5F5F5",
  },
  cancelButtonText: {
    fontSize: 16,
    fontWeight: "600",
    color: "#636E72",
  },
  createButton: {
    backgroundColor: "#FF7B54",
    shadowColor: "#FF7B54",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.2,
    shadowRadius: 8,
    elevation: 5,
  },
  createButtonText: {
    fontSize: 16,
    fontWeight: "600",
    color: "#FFFFFF",
  },
  buttonIcon: {
    marginLeft: 6,
  },
  buttonDisabled: {
    opacity: 0.7,
  },
});
