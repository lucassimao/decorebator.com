import React, { useState, useEffect } from "react";
import {
  View,
  Text,
  StyleSheet,
  SafeAreaView,
  ScrollView,
  TouchableOpacity,
  ImageBackground,
  Dimensions,
  Alert,
  ActivityIndicator,
  TextInput,
  Image,
  KeyboardAvoidingView,
  Platform,
  Modal,
} from "react-native";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { useNavigation } from "@react-navigation/native";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm, Controller } from "react-hook-form";
import * as ImagePicker from "expo-image-picker";
import * as userApi from "@/api/users";
import DateTimePicker from "@react-native-community/datetimepicker";

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

  const base64String = await blobToBase64(blob);

  const parts = uri.split(".");
  const ext = parts[parts.length - 1].toLowerCase();
  const res = await userApi.update({
    profilePicture: base64String,
    profilePictureFileExtension: ext,
  });

  return res.profilePictureUrl;
};

const deleteAccount = async (): Promise<void> => {
  await new Promise((resolve) => setTimeout(resolve, 1000));
  // API call to delete account
};

const ProfileSettingsScreen: React.FC = () => {
  const navigation = useNavigation();
  const queryClient = useQueryClient();
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [showCountryPicker, setShowCountryPicker] = useState(false);
  const [tempProfilePicture, setTempProfilePicture] = useState<string | null>(
    null,
  );

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
      });
    }
  }, [profile, reset]);

  // Update profile mutation
  const updateMutation = useMutation({
    mutationFn: userApi.update,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["userProfile"] });
      Alert.alert("Success", "Profile updated successfully");
    },
    onError: () => {
      Alert.alert("Error", "Failed to update profile. Please try again.");
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
      Alert.alert(
        "Error",
        "Failed to upload profile picture. Please try again.",
      );
      setTempProfilePicture(null);
    },
  });

  // Delete account mutation
  const deleteAccountMutation = useMutation({
    mutationFn: deleteAccount,
    onSuccess: () => {
      // Navigate to login/auth screen
      Alert.alert(
        "Account Deleted",
        "Your account has been permanently deleted.",
      );
    },
    onError: () => {
      Alert.alert("Error", "Failed to delete account. Please try again.");
    },
  });

  const handlePickImage = async () => {
    const permissionResult =
      await ImagePicker.requestMediaLibraryPermissionsAsync();

    if (!permissionResult.granted) {
      Alert.alert(
        "Permission Required",
        "Please allow access to your photo library.",
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
      Alert.alert("Permission Required", "Please allow access to your camera.");
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
    Alert.alert("Change Profile Picture", "Choose an option", [
      { text: "Take Photo", onPress: handleTakePhoto },
      { text: "Choose from Library", onPress: handlePickImage },
      { text: "Cancel", style: "cancel" },
    ]);
  };

  const handleSubmitProfile = (data: userApi.UpdateInput) => {
    updateMutation.mutate(data);
  };

  const handleChangePassword = () => {
    // Navigate to change password screen
  };

  const handleDeleteAccount = () => {
    Alert.alert(
      "Delete Account",
      "Are you sure you want to delete your account? This action cannot be undone and you will lose all your data.",
      [
        { text: "Cancel", style: "cancel" },
        {
          text: "Delete Account",
          style: "destructive",
          onPress: () => {
            Alert.alert(
              "Confirm Deletion",
              'Please type "DELETE" to confirm account deletion.',
              [
                { text: "Cancel", style: "cancel" },
                {
                  text: "Confirm",
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
        <ActivityIndicator size="large" color="#FF7B54" />
      </View>
    );
  }

  const watchedDateOfBirth = watch("dateOfBirth");
  const profilePictureUri = tempProfilePicture || profile?.profilePictureUrl;

  return (
    <ImageBackground
      source={require("@/assets/images/dashboard-bg.png")}
      style={styles.backgroundImage}
      resizeMode="cover"
    >
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
                <Ionicons name="arrow-back" size={24} color="#2D3436" />
              </TouchableOpacity>
              <Text style={styles.headerTitle}>Profile Settings</Text>
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
                    <ActivityIndicator size="large" color="#FFFFFF" />
                  </View>
                ) : (
                  <>
                    <Image
                      source={{
                        uri:
                          profilePictureUri ||
                          "https://via.placeholder.com/150",
                      }}
                      style={styles.profilePicture}
                    />
                    <View style={styles.editBadge}>
                      <MaterialIcons
                        name="camera-alt"
                        size={20}
                        color="#FFFFFF"
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
                <Text style={styles.inputLabel}>First Name</Text>
                <Controller
                  control={control}
                  name="firstName"
                  rules={{
                    required: "First name is required",
                    minLength: { value: 2, message: "Too short" },
                  }}
                  render={({ field: { onChange, onBlur, value } }) => (
                    <TextInput
                      style={[
                        styles.input,
                        errors.firstName && styles.inputError,
                      ]}
                      placeholder="Enter your first name"
                      placeholderTextColor="#B2BEC3"
                      value={value}
                      onChangeText={onChange}
                      onBlur={onBlur}
                      autoCapitalize="words"
                    />
                  )}
                />
                {errors.firstName && (
                  <Text style={styles.errorText}>
                    {errors.firstName.message}
                  </Text>
                )}
              </View>

              {/* Last Name */}
              <View style={styles.inputGroup}>
                <Text style={styles.inputLabel}>Last Name</Text>
                <Controller
                  control={control}
                  name="lastName"
                  rules={{
                    required: "Last name is required",
                    minLength: { value: 2, message: "Too short" },
                  }}
                  render={({ field: { onChange, onBlur, value } }) => (
                    <TextInput
                      style={[
                        styles.input,
                        errors.lastName && styles.inputError,
                      ]}
                      placeholder="Enter your last name"
                      placeholderTextColor="#B2BEC3"
                      value={value}
                      onChangeText={onChange}
                      onBlur={onBlur}
                      autoCapitalize="words"
                    />
                  )}
                />
                {errors.lastName && (
                  <Text style={styles.errorText}>
                    {errors.lastName.message}
                  </Text>
                )}
              </View>

              {/* Country */}
              <View style={styles.inputGroup}>
                <Text style={styles.inputLabel}>Country (Optional)</Text>
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
                        {value ? getCountryName(value) : "Select your country"}
                      </Text>
                      <Ionicons name="chevron-down" size={20} color="#636E72" />
                    </TouchableOpacity>
                  )}
                />
              </View>

              {/* Date of Birth */}
              <View style={styles.inputGroup}>
                <Text style={styles.inputLabel}>Date of Birth (Optional)</Text>
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
                        {value
                          ? formatDate(value)
                          : "Select your date of birth"}
                      </Text>
                      <Ionicons
                        name="calendar-outline"
                        size={20}
                        color="#636E72"
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
                    <ActivityIndicator size="small" color="#FFFFFF" />
                  ) : (
                    <Text style={styles.saveButtonText}>Save Changes</Text>
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
                <MaterialIcons name="lock-outline" size={24} color="#636E72" />
                <Text style={styles.actionText}>Change Password</Text>
                <Ionicons name="chevron-forward" size={20} color="#636E72" />
              </TouchableOpacity>

              <TouchableOpacity
                style={[styles.actionButton, styles.deleteButton]}
                onPress={handleDeleteAccount}
              >
                <MaterialIcons
                  name="delete-outline"
                  size={24}
                  color="#FF6B6B"
                />
                <Text style={[styles.actionText, { color: "#FF6B6B" }]}>
                  Delete Account
                </Text>
                <Ionicons name="chevron-forward" size={20} color="#FF6B6B" />
              </TouchableOpacity>
            </View>

            {/* Member Since */}
            <View style={styles.memberInfo}>
              <Text style={styles.memberText}>
                Member since{" "}
                {new Date(profile?.createdAt || "").toLocaleDateString(
                  "en-US",
                  { year: "numeric", month: "long" },
                )}
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
                    <Text style={styles.datePickerCancel}>Cancel</Text>
                  </TouchableOpacity>
                  <Text style={styles.datePickerTitle}>Date of Birth</Text>
                  <TouchableOpacity onPress={() => setShowDatePicker(false)}>
                    <Text style={styles.datePickerDone}>Done</Text>
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
                    textColor="#000000" // Ensure text is visible
                  />
                </View>
              </View>
            </View>
          </Modal>
        )}
        {/* Android DatePicker */}
        {showDatePicker && Platform.OS === "android" && (
          <DateTimePicker
            value={
              watchedDateOfBirth ? new Date(watchedDateOfBirth) : new Date()
            }
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
                <Text style={styles.modalTitle}>Select Country</Text>
                <TouchableOpacity onPress={() => setShowCountryPicker(false)}>
                  <Ionicons name="close" size={24} color="#636E72" />
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
      </SafeAreaView>
    </ImageBackground>
  );
};

export default ProfileSettingsScreen;

const styles = StyleSheet.create({
  backgroundImage: {
    flex: 1,
    width: SCREEN_WIDTH,
  },
  container: {
    flex: 1,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    backgroundColor: "#FDF6E3",
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
    backgroundColor: "rgba(255, 255, 255, 0.9)",
    justifyContent: "center",
    alignItems: "center",
  },
  headerTitle: {
    fontSize: 20,
    fontWeight: "600",
    color: "#2D3436",
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
    backgroundColor: "#FFFFFF",
  },
  uploadingOverlay: {
    width: 120,
    height: 120,
    borderRadius: 60,
    backgroundColor: "rgba(0, 0, 0, 0.7)",
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
    backgroundColor: "#FF7B54",
    justifyContent: "center",
    alignItems: "center",
    borderWidth: 3,
    borderColor: "#FFFFFF",
  },
  emailText: {
    fontSize: 16,
    color: "#636E72",
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
    color: "#2D3436",
    marginBottom: 8,
  },
  input: {
    backgroundColor: "rgba(255, 255, 255, 0.9)",
    borderRadius: 12,
    paddingHorizontal: 16,
    paddingVertical: 14,
    fontSize: 16,
    color: "#2D3436",
    borderWidth: 1,
    borderColor: "#E0E0E0",
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
  },
  inputError: {
    borderColor: "#FF6B6B",
  },
  inputText: {
    fontSize: 16,
    color: "#2D3436",
    flex: 1,
  },
  placeholderText: {
    color: "#B2BEC3",
  },
  errorText: {
    color: "#FF6B6B",
    fontSize: 12,
    marginTop: 4,
  },
  saveButton: {
    backgroundColor: "#FF7B54",
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: "center",
    marginTop: 8,
    shadowColor: "#FF7B54",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.2,
    shadowRadius: 8,
    elevation: 5,
  },
  saveButtonText: {
    color: "#FFFFFF",
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
    backgroundColor: "rgba(255, 255, 255, 0.9)",
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
  },
  actionText: {
    flex: 1,
    fontSize: 16,
    color: "#2D3436",
    marginLeft: 12,
  },
  deleteButton: {
    backgroundColor: "rgba(255, 107, 107, 0.1)",
  },
  memberInfo: {
    alignItems: "center",
    paddingVertical: 16,
  },
  memberText: {
    fontSize: 14,
    color: "#636E72",
  },
  modalOverlay: {
    position: "absolute",
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: "rgba(0, 0, 0, 0.5)",
    justifyContent: "flex-end",
  },
  modalBackdrop: {
    flex: 1,
  },
  countryPickerModal: {
    backgroundColor: "#FFFFFF",
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
    borderBottomColor: "#F0F0F0",
  },
  modalTitle: {
    fontSize: 18,
    fontWeight: "600",
    color: "#2D3436",
  },
  countryList: {
    paddingHorizontal: 20,
  },
  countryItem: {
    flexDirection: "row",
    alignItems: "center",
    paddingVertical: 16,
    borderBottomWidth: 1,
    borderBottomColor: "#F0F0F0",
  },
  countryFlag: {
    fontSize: 24,
    marginRight: 12,
  },
  countryName: {
    fontSize: 16,
    color: "#2D3436",
  },
  datePickerContainer: {
    backgroundColor: "#FFFFFF",
    borderRadius: 12,
    marginTop: 12,
    marginBottom: 20,
    overflow: "hidden",
  },
  datePickerModal: {
    flex: 1,
    justifyContent: "flex-end",
    backgroundColor: "rgba(0, 0, 0, 0.3)", // Move backdrop color here
  },
  datePickerBackdrop: {
    flex: 1,
  },
  datePickerContent: {
    backgroundColor: "#FFFFFF",
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    paddingBottom: Platform.OS === "ios" ? 34 : 20, // Account for safe area
    shadowColor: "#000",
    shadowOffset: { width: 0, height: -2 },
    shadowOpacity: 0.1,
    shadowRadius: 8,
    elevation: 5,
  },
  datePickerHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    paddingHorizontal: 20,
    paddingVertical: 16,
    borderBottomWidth: 1,
    borderBottomColor: "#E0E0E0",
  },
  datePickerTitle: {
    fontSize: 17,
    fontWeight: "600",
    color: "#000000",
  },
  datePickerDone: {
    fontSize: 17,
    color: "#FF7B54",
    fontWeight: "600",
  },
  datePickerCancel: {
    fontSize: 17,
    color: "#636E72",
  },
  datePickerWrapper: {
    height: 216, // Standard iOS picker height
    justifyContent: "center",
    overflow: "hidden",
  },
  datePicker: {
    width: "100%",
    backgroundColor: "#FFFFFF",
  },
});
