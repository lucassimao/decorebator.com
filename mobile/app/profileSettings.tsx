import * as userApi from "@/api/users";
import { ChangePasswordModal } from "@/components/profileSettings/ChangePasswordModal";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import DateTimePicker from "@react-native-community/datetimepicker";
import { useNavigation } from "@react-navigation/native";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import * as ImagePicker from "expo-image-picker";
import * as MailComposer from "expo-mail-composer";
import { router } from "expo-router";
import * as WebBrowser from "expo-web-browser";
import React, { useEffect, useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import i18n, { supportedLanguages } from "@/i18n";
import { useTheme } from "@/contexts/ThemeContext";
import {
  ActivityIndicator,
  Alert,
  Dimensions,
  Image,
  KeyboardAvoidingView,
  Modal,
  Platform,
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from "react-native";

const { width: SCREEN_WIDTH } = Dimensions.get("window");

// Country list (partial)
const COUNTRIES = [
  { code: "US", name: "United States", flag: "🇺🇸" },
  { code: "GB", name: "United Kingdom", flag: "🇬🇧" },
  { code: "CA", name: "Canada", flag: "🇨🇦" },
  { code: "AU", name: "Australia", flag: "🇦🇺" },
  { code: "DE", name: "Germany", flag: "🇩🇪" },
  { code: "FR", name: "France", flag: "🇫🇷" },
  { code: "ES", name: "Spain", flag: "🇪🇸" },
  { code: "IT", name: "Italy", flag: "🇮🇹" },
  { code: "BR", name: "Brazil", flag: "🇧🇷" },
  { code: "JP", name: "Japan", flag: "🇯🇵" },
  // Add more countries as needed
];

async function blobToBase64(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => {
      reader.abort();
      reject(new Error("Problem reading blob as base64"));
    };
    reader.onload = () => {
      // reader.result is something like "data:image/png;base64,iVBORw0KGgoAAAANS…"
      // If you only need the raw base64 (no data: prefix), split it out:
      const dataUrl = reader.result as string;
      const base64 = dataUrl.split(",")[1];
      resolve(base64);
    };
    reader.readAsDataURL(blob);
  });
}

const uploadProfilePicture = async (uri: string): Promise<string> => {
  const response = await fetch(uri);
  const blob = await response.blob();

  await new Promise((resolve) => setTimeout(resolve, 1500));

  const base64Data = await blobToBase64(blob);

  const parts = uri.split(".");
  const extension = parts[parts.length - 1].toLowerCase();
  const res = await userApi.update({
    updateProfilePicture: {
      base64Data,
      extension,
    },
  });

  return res.profilePictureUrl;
};

const ProfileSettingsScreen: React.FC = () => {
  const navigation = useNavigation();
  const queryClient = useQueryClient();
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [showCountryPicker, setShowCountryPicker] = useState(false);
  const [tempProfilePicture, setTempProfilePicture] = useState<string | null>(
    null,
  );
  const [showChangePasswordModal, setShowChangePasswordModal] = useState(false);
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);

  const [showLanguagePicker, setShowLanguagePicker] = useState(false);

  // Fetch user profile
  const { data: profile, isLoading } = useQuery({
    queryKey: ["userProfile"],
    queryFn: userApi.getProfile,
  });

  // Form setup
  const {
    control,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors, isDirty },
  } = useForm<userApi.UpdateInput>({
    defaultValues: {
      firstName: "",
      lastName: "",
      country: "",
      dateOfBirth: "",
    },
  });

  // Set form values when profile loads
  useEffect(() => {
    if (profile) {
      reset({
        firstName: profile.firstName,
        lastName: profile.lastName,
        country: profile.country,
        dateOfBirth: profile.dateOfBirth,
        preferredLanguage: profile.preferredLanguage || "en",
      });
    }
  }, [profile, reset]);

  // Update profile mutation
  const updateMutation = useMutation({
    mutationFn: userApi.update,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["userProfile"] });
      Alert.alert(t("common.success"), t("profile.changesSaved"));
    },
    onError: () => {
      Alert.alert(t("common.error"), t("errors.general"));
    },
  });

  // Upload profile picture mutation
  const uploadPictureMutation = useMutation({
    mutationFn: uploadProfilePicture,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["userProfile"] });
      setTempProfilePicture(null);
    },
    onError: () => {
      Alert.alert(t("common.error"), t("errors.general"));
      setTempProfilePicture(null);
    },
  });

  // Delete account mutation
  const deleteAccountMutation = useMutation({
    mutationFn: userApi.deleteProfile,
    onSuccess: () => {
      Alert.alert(
        t("profile.accountDeleted"),
        t("profile.accountDeletedMessage"),
      );
      userApi.sigout();
      router.dismissAll();
      router.replace("/signup");
    },
    onError: () => {
      Alert.alert(t("common.error"), t("profile.deleteAccountError"));
    },
  });

  const handlePickImage = async () => {
    const permissionResult =
      await ImagePicker.requestMediaLibraryPermissionsAsync();

    if (!permissionResult.granted) {
      Alert.alert(
        t("profile.permissionRequired"),
        t("profile.photoLibraryPermission"),
      );
      return;
    }

    const result = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ["images"],
      allowsEditing: true,
      aspect: [1, 1],
      quality: 0.8,
    });

    if (!result.canceled && result.assets[0]) {
      setTempProfilePicture(result.assets[0].uri);
      uploadPictureMutation.mutate(result.assets[0].uri);
    }
  };

  const handleTakePhoto = async () => {
    const permissionResult = await ImagePicker.requestCameraPermissionsAsync();

    if (!permissionResult.granted) {
      Alert.alert(
        t("profile.permissionRequired"),
        t("profile.cameraPermission"),
      );
      return;
    }

    const result = await ImagePicker.launchCameraAsync({
      allowsEditing: true,
      aspect: [1, 1],
      quality: 0.8,
    });

    if (!result.canceled && result.assets[0]) {
      setTempProfilePicture(result.assets[0].uri);
      uploadPictureMutation.mutate(result.assets[0].uri);
    }
  };

  const handleProfilePicturePress = () => {
    Alert.alert(t("profile.changePhoto"), t("profile.chooseOption"), [
      { text: t("profile.takePhoto"), onPress: handleTakePhoto },
      { text: t("profile.chooseFromLibrary"), onPress: handlePickImage },
      { text: t("common.cancel"), style: "cancel" },
    ]);
  };

  const handleSubmitProfile = (data: userApi.UpdateInput) => {
    updateMutation.mutate(data);
  };

  const handleChangePassword = () => {
    setShowChangePasswordModal(true);
  };

  const handleDeleteAccount = () => {
    Alert.alert(
      t("settings.account.deleteAccount"),
      t("settings.account.deleteWarning"),
      [
        { text: t("common.cancel"), style: "cancel" },
        {
          text: t("settings.account.deleteAccount"),
          style: "destructive",
          onPress: () => {
            Alert.alert(
              t("profile.confirmDeletion"),
              t("profile.typeDeleteToConfirm"),
              [
                { text: t("common.cancel"), style: "cancel" },
                {
                  text: t("profile.confirm"),
                  style: "destructive",
                  onPress: () => deleteAccountMutation.mutate(),
                },
              ],
            );
          },
        },
      ],
    );
  };

  const getCountryName = (code: string) => {
    const country = COUNTRIES.find((c) => c.code === code);
    return country ? `${country.flag} ${country.name}` : code;
  };

  const formatDate = (dateString: string) => {
    if (!dateString) return "";
    const date = new Date(dateString);
    return date.toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color={theme.colors.primary} />
      </View>
    );
  }

  const watchedDateOfBirth = watch("dateOfBirth");
  const profilePictureUri = tempProfilePicture || profile?.profilePictureUrl;

  return (
    <SafeAreaView style={styles.container}>
      <KeyboardAvoidingView
        behavior={Platform.OS === "ios" ? "padding" : "height"}
        style={{ flex: 1 }}
      >
        <ScrollView
          contentContainerStyle={styles.scrollContent}
          showsVerticalScrollIndicator={false}
        >
          {/* Header */}
          <View style={styles.header}>
            <TouchableOpacity
              style={styles.backButton}
              onPress={() => navigation.goBack()}
            >
              <Ionicons
                name="arrow-back"
                size={24}
                color={theme.colors.text.primary}
              />
            </TouchableOpacity>
            <Text style={styles.headerTitle}>{t("profile.title")}</Text>
            <View style={{ width: 40 }} />
          </View>

          {/* Profile Picture */}
          <View style={styles.profilePictureSection}>
            <TouchableOpacity
              style={styles.profilePictureContainer}
              onPress={handleProfilePicturePress}
              disabled={uploadPictureMutation.isPending}
            >
              {uploadPictureMutation.isPending ? (
                <View style={styles.uploadingOverlay}>
                  <ActivityIndicator
                    size="large"
                    color={theme.colors.text.inverse}
                  />
                </View>
              ) : (
                <>
                  <Image
                    source={{
                      uri:
                        profilePictureUri || "https://via.placeholder.com/150",
                    }}
                    style={styles.profilePicture}
                  />
                  <View style={styles.editBadge}>
                    <MaterialIcons
                      name="camera-alt"
                      size={20}
                      color={theme.colors.text.inverse}
                    />
                  </View>
                </>
              )}
            </TouchableOpacity>
            <Text style={styles.emailText}>{profile?.email}</Text>
          </View>

          {/* Form Section */}
          <View style={styles.formSection}>
            {/* First Name */}
            <View style={styles.inputGroup}>
              <Text style={styles.inputLabel}>{t("profile.firstName")}</Text>
              <Controller
                control={control}
                name="firstName"
                rules={{
                  required: t("errors.fieldRequired"),
                  minLength: { value: 2, message: t("errors.fieldRequired") },
                }}
                render={({ field: { onChange, onBlur, value } }) => (
                  <TextInput
                    style={[
                      styles.input,
                      errors.firstName && styles.inputError,
                    ]}
                    placeholder={t("profile.firstNamePlaceholder")}
                    placeholderTextColor={theme.colors.text.placeholder}
                    value={value}
                    onChangeText={onChange}
                    onBlur={onBlur}
                    autoCapitalize="words"
                  />
                )}
              />
              {errors.firstName && (
                <Text style={styles.errorText}>{errors.firstName.message}</Text>
              )}
            </View>

            {/* Last Name */}
            <View style={styles.inputGroup}>
              <Text style={styles.inputLabel}>{t("profile.lastName")}</Text>
              <Controller
                control={control}
                name="lastName"
                rules={{
                  required: t("errors.fieldRequired"),
                  minLength: { value: 2, message: t("errors.fieldRequired") },
                }}
                render={({ field: { onChange, onBlur, value } }) => (
                  <TextInput
                    style={[styles.input, errors.lastName && styles.inputError]}
                    placeholder={t("profile.lastNamePlaceholder")}
                    placeholderTextColor={theme.colors.text.placeholder}
                    value={value}
                    onChangeText={onChange}
                    onBlur={onBlur}
                    autoCapitalize="words"
                  />
                )}
              />
              {errors.lastName && (
                <Text style={styles.errorText}>{errors.lastName.message}</Text>
              )}
            </View>

            {/* Country */}
            <View style={styles.inputGroup}>
              <Text style={styles.inputLabel}>
                {t("profile.country")} ({t("profile.optional")})
              </Text>
              <Controller
                control={control}
                name="country"
                render={({ field: { value } }) => (
                  <TouchableOpacity
                    style={styles.input}
                    onPress={() => setShowCountryPicker(true)}
                  >
                    <Text
                      style={[
                        styles.inputText,
                        !value && styles.placeholderText,
                      ]}
                    >
                      {value
                        ? getCountryName(value)
                        : t("profile.selectCountry")}
                    </Text>
                    <Ionicons
                      name="chevron-down"
                      size={20}
                      color={theme.colors.text.secondary}
                    />
                  </TouchableOpacity>
                )}
              />
            </View>

            {/* Language Preference */}
            <View style={styles.inputGroup}>
              <Text style={styles.inputLabel}>
                {t("settings.preferences.language")}
              </Text>
              <Controller
                control={control}
                name="preferredLanguage"
                render={({ field: { value } }) => (
                  <TouchableOpacity
                    style={styles.input}
                    onPress={() => setShowLanguagePicker(true)}
                  >
                    <Text
                      style={[
                        styles.inputText,
                        !value && styles.placeholderText,
                      ]}
                    >
                      {value
                        ? supportedLanguages.find((lang) => lang.code === value)
                            ?.nativeName || value
                        : t("dashboard.wordlists.selectLanguage")}
                    </Text>
                    <Ionicons
                      name="chevron-down"
                      size={20}
                      color={theme.colors.text.secondary}
                    />
                  </TouchableOpacity>
                )}
              />
            </View>

            {/* Date of Birth */}
            <View style={styles.inputGroup}>
              <Text style={styles.inputLabel}>{t("profile.dateOfBirth")}</Text>
              <Controller
                control={control}
                name="dateOfBirth"
                render={({ field: { value } }) => (
                  <TouchableOpacity
                    style={styles.input}
                    onPress={() => setShowDatePicker(true)}
                  >
                    <Text
                      style={[
                        styles.inputText,
                        !value && styles.placeholderText,
                      ]}
                    >
                      {value ? formatDate(value) : t("profile.dateOfBirth")}
                    </Text>
                    <Ionicons
                      name="calendar-outline"
                      size={20}
                      color={theme.colors.text.secondary}
                    />
                  </TouchableOpacity>
                )}
              />
            </View>

            {/* Save Button */}
            {isDirty && (
              <TouchableOpacity
                style={[
                  styles.saveButton,
                  updateMutation.isPending && styles.buttonDisabled,
                ]}
                onPress={handleSubmit(handleSubmitProfile)}
                disabled={updateMutation.isPending}
              >
                {updateMutation.isPending ? (
                  <ActivityIndicator
                    size="small"
                    color={theme.colors.text.inverse}
                  />
                ) : (
                  <Text style={styles.saveButtonText}>
                    {t("profile.saveChanges")}
                  </Text>
                )}
              </TouchableOpacity>
            )}
          </View>

          {/* Account Actions */}
          <View style={styles.actionsSection}>
            <TouchableOpacity
              style={styles.actionButton}
              onPress={handleChangePassword}
            >
              <MaterialIcons
                name="lock-outline"
                size={24}
                color={theme.colors.text.secondary}
              />
              <Text style={styles.actionText}>
                {t("profile.changePassword.title")}
              </Text>
              <Ionicons
                name="chevron-forward"
                size={20}
                color={theme.colors.text.secondary}
              />
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.actionButton}
              onPress={async () => {
                await WebBrowser.openBrowserAsync(
                  `https://decorebator.com/${i18n.language}/privacy`,
                );
              }}
            >
              <MaterialIcons
                name="privacy-tip"
                size={24}
                color={theme.colors.text.secondary}
              />
              <Text style={styles.actionText}>
                {t("settings.privacyPolicy")}
              </Text>
              <Ionicons
                name="chevron-forward"
                size={20}
                color={theme.colors.text.secondary}
              />
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.actionButton}
              onPress={() => {
                Alert.alert(
                  t("profile.dataExport.title"),
                  t("profile.dataExport.message"),
                  [
                    { text: t("common.cancel"), style: "cancel" },
                    {
                      text: t("profile.dataExport.requestButton"),
                      onPress: async () => {
                        const isAvailable =
                          await MailComposer.isAvailableAsync();
                        if (!isAvailable) {
                          Alert.alert(t("settings.noEmailClient"));
                          return;
                        }
                        MailComposer.composeAsync({
                          recipients: ["privacy@decorebator.com"],
                          subject: t("profile.dataExport.emailSubject"),
                          body: t("profile.dataExport.emailBody", {
                            email: profile?.email,
                          }),
                        });
                      },
                    },
                  ],
                );
              }}
            >
              <MaterialIcons
                name="download"
                size={24}
                color={theme.colors.text.secondary}
              />
              <Text style={styles.actionText}>
                {t("profile.dataExport.title")}
              </Text>
              <Ionicons
                name="chevron-forward"
                size={20}
                color={theme.colors.text.secondary}
              />
            </TouchableOpacity>

            <TouchableOpacity
              style={[styles.actionButton, styles.deleteButton]}
              onPress={handleDeleteAccount}
            >
              <MaterialIcons
                name="delete-outline"
                size={24}
                color={theme.colors.error}
              />
              <Text style={[styles.actionText, { color: theme.colors.error }]}>
                {t("settings.account.deleteAccount")}
              </Text>
              <Ionicons
                name="chevron-forward"
                size={20}
                color={theme.colors.error}
              />
            </TouchableOpacity>
          </View>

          {/* Member Since */}
          <View style={styles.memberInfo}>
            <Text style={styles.memberText}>
              {t("profile.memberSince", {
                date: new Date(profile?.createdAt || "").toLocaleDateString(
                  i18n.language.startsWith("en") ? "en-US" : i18n.language,
                  { year: "numeric", month: "long" },
                ),
              })}
            </Text>
          </View>
        </ScrollView>
      </KeyboardAvoidingView>

      {/* Date Picker Modal for iOS */}
      {showDatePicker && Platform.OS === "ios" && (
        <Modal
          transparent
          visible={showDatePicker}
          animationType="slide"
          onRequestClose={() => setShowDatePicker(false)}
        >
          <View style={styles.datePickerModal}>
            <TouchableOpacity
              style={styles.datePickerBackdrop}
              onPress={() => setShowDatePicker(false)}
              activeOpacity={1}
            />
            <View style={styles.datePickerContent}>
              <View style={styles.datePickerHeader}>
                <TouchableOpacity onPress={() => setShowDatePicker(false)}>
                  <Text style={styles.datePickerCancel}>
                    {t("common.cancel")}
                  </Text>
                </TouchableOpacity>
                <Text style={styles.datePickerTitle}>
                  {t("profile.dateOfBirth")}
                </Text>
                <TouchableOpacity onPress={() => setShowDatePicker(false)}>
                  <Text style={styles.datePickerDone}>{t("common.done")}</Text>
                </TouchableOpacity>
              </View>

              <View style={styles.datePickerWrapper}>
                <DateTimePicker
                  value={
                    watchedDateOfBirth
                      ? new Date(watchedDateOfBirth)
                      : new Date()
                  }
                  mode="date"
                  display="spinner"
                  maximumDate={new Date()}
                  onChange={(event, selectedDate) => {
                    if (selectedDate) {
                      setValue(
                        "dateOfBirth",
                        selectedDate.toISOString().split("T")[0],
                        { shouldDirty: true },
                      );
                    }
                  }}
                  style={styles.datePicker}
                  textColor={theme.colors.text.primary}
                />
              </View>
            </View>
          </View>
        </Modal>
      )}
      {/* Android DatePicker */}
      {showDatePicker && Platform.OS === "android" && (
        <DateTimePicker
          value={watchedDateOfBirth ? new Date(watchedDateOfBirth) : new Date()}
          mode="date"
          display="default"
          maximumDate={new Date()}
          onChange={(event, selectedDate) => {
            setShowDatePicker(false);
            if (event.type === "set" && selectedDate) {
              setValue(
                "dateOfBirth",
                selectedDate.toISOString().split("T")[0],
                { shouldDirty: true },
              );
            }
          }}
        />
      )}
      {/* Country Picker Modal */}
      {showCountryPicker && (
        <View style={styles.modalOverlay}>
          <TouchableOpacity
            style={styles.modalBackdrop}
            onPress={() => setShowCountryPicker(false)}
          />
          <View style={styles.countryPickerModal}>
            <View style={styles.modalHeader}>
              <Text style={styles.modalTitle}>
                {t("profile.selectCountry")}
              </Text>
              <TouchableOpacity onPress={() => setShowCountryPicker(false)}>
                <Ionicons
                  name="close"
                  size={24}
                  color={theme.colors.text.secondary}
                />
              </TouchableOpacity>
            </View>
            <ScrollView style={styles.countryList}>
              {COUNTRIES.map((country) => (
                <TouchableOpacity
                  key={country.code}
                  style={styles.countryItem}
                  onPress={() => {
                    setValue("country", country.code, { shouldDirty: true });
                    setShowCountryPicker(false);
                  }}
                >
                  <Text style={styles.countryFlag}>{country.flag}</Text>
                  <Text style={styles.countryName}>{country.name}</Text>
                </TouchableOpacity>
              ))}
            </ScrollView>
          </View>
        </View>
      )}

      {/* Language Picker Modal */}
      {showLanguagePicker && (
        <View style={styles.modalOverlay}>
          <TouchableOpacity
            style={styles.modalBackdrop}
            onPress={() => setShowLanguagePicker(false)}
          />
          <View style={styles.countryPickerModal}>
            <View style={styles.modalHeader}>
              <Text style={styles.modalTitle}>
                {t("settings.preferences.language")}
              </Text>
              <TouchableOpacity onPress={() => setShowLanguagePicker(false)}>
                <Ionicons
                  name="close"
                  size={24}
                  color={theme.colors.text.secondary}
                />
              </TouchableOpacity>
            </View>
            <ScrollView style={styles.countryList}>
              {supportedLanguages.map((language) => (
                <TouchableOpacity
                  key={language.code}
                  style={styles.countryItem}
                  onPress={() => {
                    setValue("preferredLanguage", language.code, {
                      shouldDirty: true,
                    });
                    setShowLanguagePicker(false);
                  }}
                >
                  <Text style={styles.countryName}>{language.nativeName}</Text>
                  <Text style={styles.languageCode}>({language.name})</Text>
                </TouchableOpacity>
              ))}
            </ScrollView>
          </View>
        </View>
      )}

      {showChangePasswordModal && (
        <ChangePasswordModal
          visible
          onClose={() => setShowChangePasswordModal(false)}
        />
      )}
    </SafeAreaView>
  );
};

export default ProfileSettingsScreen;

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    container: {
      flex: 1,
      backgroundColor: theme.colors.background.default,
    },
    loadingContainer: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      backgroundColor: theme.colors.background.default,
    },
    scrollContent: {
      paddingBottom: 30,
    },
    header: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      paddingHorizontal: 20,
      paddingVertical: 16,
    },
    backButton: {
      width: 40,
      height: 40,
      borderRadius: 20,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
    },
    headerTitle: {
      fontSize: 20,
      fontWeight: "600",
      color: theme.colors.text.primary,
    },
    profilePictureSection: {
      alignItems: "center",
      paddingVertical: 24,
    },
    profilePictureContainer: {
      position: "relative",
      marginBottom: 12,
    },
    profilePicture: {
      width: 120,
      height: 120,
      borderRadius: 60,
      backgroundColor: theme.colors.background.surface,
    },
    uploadingOverlay: {
      width: 120,
      height: 120,
      borderRadius: 60,
      backgroundColor: theme.colors.overlay.backdropHeavy,
      justifyContent: "center",
      alignItems: "center",
    },
    editBadge: {
      position: "absolute",
      bottom: 0,
      right: 0,
      width: 36,
      height: 36,
      borderRadius: 18,
      backgroundColor: theme.colors.primary,
      justifyContent: "center",
      alignItems: "center",
      borderWidth: 3,
      borderColor: theme.colors.background.surface,
    },
    emailText: {
      fontSize: 16,
      color: theme.colors.text.secondary,
    },
    formSection: {
      paddingHorizontal: 20,
      marginBottom: 24,
    },
    inputGroup: {
      marginBottom: 20,
    },
    inputLabel: {
      fontSize: 14,
      fontWeight: "500",
      color: theme.colors.text.primary,
      marginBottom: 8,
    },
    input: {
      backgroundColor: theme.colors.ui.inputBackground,
      borderRadius: 12,
      paddingHorizontal: 16,
      paddingVertical: 14,
      fontSize: 16,
      color: theme.colors.text.primary,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
    },
    inputError: {
      borderColor: theme.colors.error,
    },
    inputText: {
      fontSize: 16,
      color: theme.colors.text.primary,
      flex: 1,
    },
    placeholderText: {
      color: theme.colors.text.placeholder,
    },
    errorText: {
      color: theme.colors.error,
      fontSize: 12,
      marginTop: 4,
    },
    saveButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: 12,
      paddingVertical: 16,
      alignItems: "center",
      marginTop: 8,
      shadowColor: theme.colors.primary,
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.2,
      shadowRadius: 8,
      elevation: 5,
    },
    saveButtonText: {
      color: theme.colors.text.inverse,
      fontSize: 16,
      fontWeight: "600",
    },
    buttonDisabled: {
      opacity: 0.7,
    },
    actionsSection: {
      paddingHorizontal: 20,
      marginBottom: 24,
    },
    actionButton: {
      flexDirection: "row",
      alignItems: "center",
      backgroundColor: theme.colors.background.surface,
      borderRadius: 12,
      padding: 16,
      marginBottom: 12,
    },
    actionText: {
      flex: 1,
      fontSize: 16,
      color: theme.colors.text.primary,
      marginLeft: 12,
    },
    deleteButton: {
      backgroundColor:
        theme.mode === "light"
          ? "rgba(255, 107, 107, 0.1)"
          : theme.colors.background.subtle,
    },
    memberInfo: {
      alignItems: "center",
      paddingVertical: 16,
    },
    memberText: {
      fontSize: 14,
      color: theme.colors.text.secondary,
    },
    modalOverlay: {
      position: "absolute",
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor: theme.colors.overlay.backdropHeavy,
      justifyContent: "flex-end",
    },
    modalBackdrop: {
      flex: 1,
    },
    countryPickerModal: {
      backgroundColor: theme.colors.background.surface,
      borderTopLeftRadius: 24,
      borderTopRightRadius: 24,
      maxHeight: SCREEN_WIDTH * 0.8,
    },
    modalHeader: {
      flexDirection: "row",
      justifyContent: "space-between",
      alignItems: "center",
      padding: 20,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.ui.divider,
    },
    modalTitle: {
      fontSize: 18,
      fontWeight: "600",
      color: theme.colors.text.primary,
    },
    countryList: {
      paddingHorizontal: 20,
    },
    countryItem: {
      flexDirection: "row",
      alignItems: "center",
      paddingVertical: 16,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.ui.divider,
    },
    countryFlag: {
      fontSize: 24,
      marginRight: 12,
    },
    countryName: {
      fontSize: 16,
      color: theme.colors.text.primary,
      flex: 1,
    },
    languageCode: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      marginLeft: 8,
    },
    datePickerContainer: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: 12,
      marginTop: 12,
      marginBottom: 20,
      overflow: "hidden",
    },
    datePickerModal: {
      flex: 1,
      justifyContent: "flex-end",
      backgroundColor: theme.colors.overlay.backdropLight,
    },
    datePickerBackdrop: {
      flex: 1,
    },
    datePickerContent: {
      backgroundColor: theme.colors.background.surface,
      borderTopLeftRadius: 20,
      borderTopRightRadius: 20,
      paddingBottom: Platform.OS === "ios" ? 34 : 20, // Account for safe area
      ...theme.shadows.md,
    },
    datePickerHeader: {
      flexDirection: "row",
      justifyContent: "space-between",
      alignItems: "center",
      paddingHorizontal: 20,
      paddingVertical: 16,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.ui.border,
    },
    datePickerTitle: {
      fontSize: 17,
      fontWeight: "600",
      color: theme.colors.text.primary,
    },
    datePickerDone: {
      fontSize: 17,
      color: theme.colors.primary,
      fontWeight: "600",
    },
    datePickerCancel: {
      fontSize: 17,
      color: theme.colors.text.secondary,
    },
    datePickerWrapper: {
      height: 216, // Standard iOS picker height
      justifyContent: "center",
      overflow: "hidden",
    },
    datePicker: {
      width: "100%",
      backgroundColor: theme.colors.background.surface,
    },
  });
