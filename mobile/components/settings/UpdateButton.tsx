import React from "react";
import {
  TouchableOpacity,
  Text,
  ActivityIndicator,
  Alert,
  View,
} from "react-native";
import { MaterialIcons, Ionicons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import { useUpdates } from "@/hooks/useUpdates";
import Constants from "expo-constants";

export const UpdateButton: React.FC = () => {
  const { t } = useTranslation();
  const { theme, responsive } = useTheme();
  const {
    isChecking,
    isUpdating,
    updateAvailable,
    checkForUpdates,
    applyUpdate,
  } = useUpdates();

  const handlePress = async () => {
    if (updateAvailable) {
      Alert.alert(
        t("settings.update.confirmTitle", "Apply Update"),
        t(
          "settings.update.confirmMessage",
          "The app will restart to apply the update. Any unsaved progress will be lost.",
        ),
        [
          {
            text: t("common.cancel", "Cancel"),
            style: "cancel",
          },
          {
            text: t("settings.update.updateNow", "Update Now"),
            onPress: async () => {
              try {
                await applyUpdate();
              } catch {
                Alert.alert(
                  t("common.error", "Error"),
                  t(
                    "settings.update.updateFailed",
                    "Failed to apply update. Please try again later.",
                  ),
                );
              }
            },
          },
        ],
      );
    } else {
      const hasUpdate = await checkForUpdates();
      if (!hasUpdate) {
        Alert.alert(
          t("settings.update.noUpdatesTitle", "No Updates"),
          t(
            "settings.update.noUpdatesMessage",
            "You have the latest version of the app.",
          ),
        );
      }
    }
  };

  const getStatusText = () => {
    if (isUpdating) {
      return t("settings.update.updating", "Updating...");
    }
    if (isChecking) {
      return t("settings.update.checking", "Checking...");
    }
    if (updateAvailable) {
      return t("settings.update.updateAvailable", "Update Available");
    }
    return t("settings.update.checkForUpdates", "Check for Updates");
  };

  const getStatusIcon = () => {
    if (isUpdating || isChecking) {
      return null; // ActivityIndicator will be shown instead
    }
    if (updateAvailable) {
      return (
        <MaterialIcons
          name="system-update"
          size={responsive.getValueForSize(20, 22, 24, 24)}
          color={theme.colors.success}
        />
      );
    }
    return (
      <MaterialIcons
        name="refresh"
        size={responsive.getValueForSize(20, 22, 24, 24)}
        color={theme.colors.text.secondary}
      />
    );
  };

  // Get runtime version info
  const getRuntimeVersion = () => {
    return (
      Constants.expoConfig?.runtimeVersion ||
      Constants.expoConfig?.version ||
      "Unknown"
    );
  };

  const styles = {
    updateItem: {
      flexDirection: "row" as const,
      alignItems: "center" as const,
      backgroundColor: theme.colors.background.surface,
      marginHorizontal: responsive.spacing.horizontal,
      padding: responsive.spacing.formPadding,
      borderRadius: theme.borderRadius.md,
      marginBottom: responsive.spacing.elementSpacing,
      minHeight: responsive.spacing.minTouchTarget,
      ...theme.shadows.sm,
    },
    updateContent: {
      flex: 1,
      flexDirection: "column" as const,
      gap: responsive.spacing.elementSpacing / 2,
    },
    updateMainRow: {
      flexDirection: "row" as const,
      alignItems: "center" as const,
      gap: responsive.spacing.elementSpacing,
    },
    updateText: {
      flex: 1,
      fontSize: responsive.fontSizes.body,
      color: updateAvailable ? theme.colors.success : theme.colors.text.primary,
      fontWeight: updateAvailable ? ("600" as const) : ("400" as const),
    },
    versionText: {
      fontSize: responsive.fontSizes.small,
      color: theme.colors.text.secondary,
      marginLeft: responsive.getValueForSize(28, 30, 32, 32), // Align with main text (icon width + gap)
    },
    updateStatus: {
      flexDirection: "row" as const,
      alignItems: "center" as const,
      gap: responsive.spacing.elementSpacing / 2,
    },
  };

  return (
    <TouchableOpacity
      style={styles.updateItem}
      onPress={handlePress}
      disabled={isChecking || isUpdating}
      activeOpacity={0.7}
      accessible={true}
      accessibilityLabel={getStatusText()}
      accessibilityRole="button"
      accessibilityHint={
        updateAvailable
          ? t(
              "settings.update.updateAvailableHint",
              "Tap to apply the available update",
            )
          : t(
              "settings.update.checkUpdatesHint",
              "Tap to check for app updates",
            )
      }
    >
      <View style={styles.updateContent}>
        <View style={styles.updateMainRow}>
          {isUpdating || isChecking ? (
            <ActivityIndicator size="small" color={theme.colors.primary} />
          ) : (
            getStatusIcon()
          )}

          <Text style={styles.updateText}>{getStatusText()}</Text>
        </View>

        <Text style={styles.versionText}>
          Runtime Version: {getRuntimeVersion()}
        </Text>
      </View>

      <View style={styles.updateStatus}>
        {updateAvailable && (
          <View
            style={{
              backgroundColor: theme.colors.success,
              borderRadius: responsive.getValueForSize(8, 10, 12, 12),
              paddingHorizontal: responsive.spacing.elementSpacing / 2,
              paddingVertical: responsive.spacing.elementSpacing / 4,
            }}
          >
            <Text
              style={{
                fontSize: responsive.fontSizes.micro,
                color: theme.colors.text.inverse,
                fontWeight: "600" as const,
              }}
            >
              {t("settings.update.new", "NEW")}
            </Text>
          </View>
        )}

        {!isUpdating && !isChecking && (
          <Ionicons
            name="chevron-forward"
            size={responsive.getValueForSize(18, 20, 20, 22)}
            color={theme.colors.text.secondary}
          />
        )}
      </View>
    </TouchableOpacity>
  );
};
