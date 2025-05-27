import React, { useState, useRef, useEffect } from "react";
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
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { CreateWordlistDTO, Wordlist } from "@/api/wordlists";
import * as wordlistsApi from "@/api/wordlists";

const { height: SCREEN_HEIGHT } = Dimensions.get("window");

type Language = { code: string; name: string; flag: string };

// Top 10 languages for language learning
const LANGUAGES: Language[] = [
  { code: "en", name: "English", flag: "🇬🇧" },
  { code: "es", name: "Spanish", flag: "🇪🇸" },
  { code: "fr", name: "French", flag: "🇫🇷" },
  { code: "de", name: "German", flag: "🇩🇪" },
  { code: "it", name: "Italian", flag: "🇮🇹" },
  { code: "pt", name: "Portuguese", flag: "🇵🇹" },
  { code: "zh", name: "Chinese", flag: "🇨🇳" },
  { code: "ja", name: "Japanese", flag: "🇯🇵" },
  { code: "ko", name: "Korean", flag: "🇰🇷" },
  { code: "ar", name: "Arabic", flag: "🇸🇦" },
];

type CreateWordlistModalProps = {
  visible: boolean;
  onClose: () => void;
  onSuccess: (wordlistName: string) => void;
};

const CreateWordlistModal = ({
  visible,
  onClose,
  onSuccess,
}: CreateWordlistModalProps) => {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [selectedLanguage, setSelectedLanguage] = useState<Language | null>(
    null,
  );
  const [errors, setErrors] = useState<Record<string, string|null>>({});

  const slideAnim = useRef(new Animated.Value(SCREEN_HEIGHT)).current;
  const backdropAnim = useRef(new Animated.Value(0)).current;
  const queryClient = useQueryClient();

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
      ]).start();
    }
  }, [visible]);

  const { mutate: createWordlist, isPending: isCreateWordlistPending } = useMutation<
    void,
    Error,
    CreateWordlistDTO
  >({
    mutationFn: (dto) => wordlistsApi.addWordlist(dto),
    onError: (error) => {
        setErrors({ submit: "Failed to create wordlist. Please try again." });
        console.error(error);
    },
    onSuccess: (_, dto) => {
      queryClient.invalidateQueries({queryKey:["wordlists"]});
      onSuccess?.(dto.name);
      handleClose();
    },
  });

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    if (!name.trim()) {
      newErrors.name = "Wordlist name is required";
    } else if (name.trim().length < 3) {
      newErrors.name = "Name must be at least 3 characters";
    }

    if (!selectedLanguage) {
      newErrors.language = "Please select a language";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = () => {
    if (validateForm()) {
      createWordlist({
        name: name.trim(),
        description: description.trim(),
        language: selectedLanguage!.code,
      });
    }
  };

  const handleClose = () => {
    setName("");
    setDescription("");
    setSelectedLanguage(null);
    setErrors({});
    onClose();
  };

  if (!visible) return null;

  return (
    <Modal
      transparent
      visible={visible}
      animationType="none"
      onRequestClose={handleClose}
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
            <Text style={styles.title}>Create New Wordlist</Text>
            <TouchableOpacity style={styles.closeButton} onPress={handleClose}>
              <Ionicons name="close" size={24} color="#636E72" />
            </TouchableOpacity>
          </View>

          <ScrollView style={styles.form} showsVerticalScrollIndicator={false}>
            {/* Name Input */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>
                Wordlist Name <Text style={styles.required}>*</Text>
              </Text>
              <TextInput
                style={[styles.input, errors.name && styles.inputError]}
                placeholder="e.g., Travel Essentials"
                placeholderTextColor="#B2BEC3"
                value={name}
                onChangeText={(text) => {
                  setName(text);
                  if (errors.name) {
                    setErrors({ ...errors, name: null });
                  }
                }}
                autoCapitalize="words"
                maxLength={50}
              />
              {errors.name && (
                <Text style={styles.errorText}>{errors.name}</Text>
              )}
            </View>

            {/* Description Input */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>Description (Optional)</Text>
              <TextInput
                style={[styles.input, styles.textArea]}
                placeholder="Add a description for your wordlist"
                placeholderTextColor="#B2BEC3"
                value={description}
                onChangeText={setDescription}
                multiline
                numberOfLines={3}
                maxLength={200}
              />
            </View>

            {/* Language Selection */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>
                Language <Text style={styles.required}>*</Text>
              </Text>
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
                      selectedLanguage?.code === lang.code &&
                        styles.languageItemSelected,
                    ]}
                    onPress={() => {
                      setSelectedLanguage(lang);
                      if (errors.language) {
                        setErrors({ ...errors, language: null });
                      }
                    }}
                  >
                    <Text style={styles.languageFlag}>{lang.flag}</Text>
                    <Text
                      style={[
                        styles.languageName,
                        selectedLanguage?.code === lang.code &&
                          styles.languageNameSelected,
                      ]}
                    >
                      {lang.name}
                    </Text>
                  </TouchableOpacity>
                ))}
              </ScrollView>
              {errors.language && (
                <Text style={styles.errorText}>{errors.language}</Text>
              )}
            </View>

            {/* Submit Error */}
            {errors.submit && (
              <View style={styles.submitError}>
                <Text style={styles.errorText}>{errors.submit}</Text>
              </View>
            )}
          </ScrollView>

          {/* Action Buttons */}
          <View style={styles.actions}>
            <TouchableOpacity
              style={[styles.button, styles.cancelButton]}
              onPress={handleClose}
            >
              <Text style={styles.cancelButtonText}>Cancel</Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={[
                styles.button,
                styles.createButton,
                isCreateWordlistPending && styles.buttonDisabled,
              ]}
              onPress={handleSubmit}
              disabled={isCreateWordlistPending}
            >
              {isCreateWordlistPending ? (
                <ActivityIndicator color="#FFFFFF" size="small" />
              ) : (
                <>
                  <Text style={styles.createButtonText}>Create Wordlist</Text>
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

export default CreateWordlistModal;
