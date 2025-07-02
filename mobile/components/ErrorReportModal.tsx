import React from "react";
import {
  Modal,
  View,
  Text,
  TouchableOpacity,
  ActivityIndicator,
  StyleSheet,
} from "react-native";
import { MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { ErrorType } from "@/api/errorReporting";
import { useTheme } from "@/contexts/ThemeContext";
import {
  useResponsive,
  useResponsiveSpacing,
} from "@/hooks/useResponsive";

interface ErrorReportModalProps {
  visible: boolean;
  onClose: () => void;
  onReportError: (errorType: ErrorType) => void;
  isLoading?: boolean;
  context?: "quiz" | "flashcards";
  wordName?: string;
  errorTypes?: ErrorType[];
}

export const ErrorReportModal: React.FC<ErrorReportModalProps> = ({
  visible,
  onClose,
  onReportError,
  isLoading = false,
  context = "quiz",
  wordName,
  errorTypes,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet, contentWidth } = useResponsive();
  const spacing = useResponsiveSpacing();
  const styles = createStyles(theme, isTablet, contentWidth, spacing);

  const isUnrelatedImageOptionAvailable =
    !errorTypes || errorTypes.includes(ErrorType.UnrelatedImage);
  const isMissingImageOptionAvailable =
    !errorTypes || errorTypes.includes(ErrorType.MissingImage);
  const isSoundNotPlayingOptionAvailable =
    !errorTypes || errorTypes.includes(ErrorType.SoundNotPlaying);
  const isUnrelatedExampleOptionAvailable =
    !errorTypes || errorTypes.includes(ErrorType.UnrelatedExample);
  const isUnrelatedMeaningOptionAvailable =
    !errorTypes || errorTypes.includes(ErrorType.UnrelatedMeaning);

  return (
    <Modal
      visible={visible}
      transparent
      animationType="fade"
      onRequestClose={onClose}
    >
      <View style={styles.modalOverlay}>
        <View
          style={[styles.reportModal, isTablet && styles.tabletReportModal]}
        >
          <Text style={styles.reportModalTitle}>
            {t(`${context}.reportIssue`)}
          </Text>
          {wordName && (
            <Text style={styles.reportModalSubtitle}>
              {t(`${context}.reportWord`, { word: wordName })}
            </Text>
          )}

          <View
            style={[
              styles.reportOptions,
              isTablet && styles.tabletReportOptions,
            ]}
          >
            {isUnrelatedImageOptionAvailable && (
              <TouchableOpacity
                style={styles.reportOption}
                onPress={() => onReportError(ErrorType.UnrelatedImage)}
                disabled={isLoading}
              >
                <MaterialIcons
                  name="image"
                  size={24}
                  color={theme.colors.text.secondary}
                />
                <Text style={styles.reportOptionText}>
                  {t(`${context}.reportOptions.imageDoesntMatch`)}
                </Text>
              </TouchableOpacity>
            )}

            {isMissingImageOptionAvailable && (
              <TouchableOpacity
                style={styles.reportOption}
                onPress={() => onReportError(ErrorType.MissingImage)}
                disabled={isLoading}
              >
                <MaterialIcons
                  name="broken-image"
                  size={24}
                  color={theme.colors.text.secondary}
                />
                <Text style={styles.reportOptionText}>
                  {t(`${context}.reportOptions.imageNotLoading`)}
                </Text>
              </TouchableOpacity>
            )}

            {isUnrelatedMeaningOptionAvailable && (
              <TouchableOpacity
                style={styles.reportOption}
                onPress={() => onReportError(ErrorType.UnrelatedMeaning)}
                disabled={isLoading}
              >
                <MaterialIcons
                  name="help-outline"
                  size={24}
                  color={theme.colors.text.secondary}
                />
                <Text style={styles.reportOptionText}>
                  {t(`${context}.reportOptions.wrongMeaning`)}
                </Text>
              </TouchableOpacity>
            )}

            {isUnrelatedExampleOptionAvailable && (
              <TouchableOpacity
                style={styles.reportOption}
                onPress={() => onReportError(ErrorType.UnrelatedExample)}
                disabled={isLoading}
              >
                <MaterialIcons
                  name="format-quote"
                  size={24}
                  color={theme.colors.text.secondary}
                />
                <Text style={styles.reportOptionText}>
                  {t(`${context}.reportOptions.wrongExample`)}
                </Text>
              </TouchableOpacity>
            )}

            {isSoundNotPlayingOptionAvailable && (
              <TouchableOpacity
                style={styles.reportOption}
                onPress={() => onReportError(ErrorType.SoundNotPlaying)}
                disabled={isLoading}
              >
                <MaterialIcons
                  name="volume-off"
                  size={24}
                  color={theme.colors.text.secondary}
                />
                <Text style={styles.reportOptionText}>
                  {t(`${context}.reportOptions.soundNotPlaying`)}
                </Text>
              </TouchableOpacity>
            )}
          </View>

          <View style={styles.reportModalActions}>
            <TouchableOpacity
              style={styles.reportCancelButton}
              onPress={onClose}
              disabled={isLoading}
            >
              <Text style={styles.reportCancelText}>{t("common.cancel")}</Text>
            </TouchableOpacity>
          </View>

          {isLoading && (
            <View style={styles.reportLoadingOverlay}>
              <ActivityIndicator size="large" color={theme.colors.primary} />
            </View>
          )}
        </View>
      </View>
    </Modal>
  );
};

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  isTablet: boolean,
  contentWidth: number,
  spacing: ReturnType<typeof useResponsiveSpacing>,
) =>
  StyleSheet.create({
    modalOverlay: {
      flex: 1,
      backgroundColor: "rgba(0, 0, 0, 0.5)",
      justifyContent: "center",
      alignItems: "center",
      padding: spacing.lg,
    },
    reportModal: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: theme.borderRadius.lg,
      padding: spacing.lg,
      width: "100%",
      maxWidth: isTablet ? Math.min(contentWidth * 0.7, 600) : 400,
      position: "relative",
      ...theme.shadows.lg,
    },
    tabletReportModal: {
      padding: spacing.xl,
    },
    reportModalTitle: {
      fontSize: 20,
      fontWeight: "600",
      color: theme.colors.text.primary,
      textAlign: "center",
      marginBottom: spacing.sm,
    },
    reportModalSubtitle: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      textAlign: "center",
      marginBottom: spacing.lg,
    },
    reportOptions: {
      gap: spacing.sm,
      marginBottom: spacing.lg,
    },
    tabletReportOptions: {
      flexDirection: isTablet ? "row" : "column",
      flexWrap: isTablet ? "wrap" : "nowrap",
      gap: isTablet ? spacing.md : spacing.sm,
    },
    reportOption: {
      flexDirection: "row",
      alignItems: "center",
      padding: spacing.md,
      backgroundColor: theme.colors.background.default,
      borderRadius: theme.borderRadius.md,
      gap: spacing.sm,
      flex: isTablet ? 1 : undefined,
      minWidth: isTablet ? "45%" : undefined,
    },
    reportOptionText: {
      fontSize: 16,
      color: theme.colors.text.primary,
      flex: 1,
    },
    reportModalActions: {
      flexDirection: "row",
      justifyContent: "center",
    },
    reportCancelButton: {
      paddingHorizontal: spacing.lg,
      paddingVertical: spacing.sm,
      borderRadius: theme.borderRadius.sm,
    },
    reportCancelText: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      fontWeight: "500",
    },
    reportLoadingOverlay: {
      position: "absolute",
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor:
        theme.mode === "light"
          ? "rgba(255, 255, 255, 0.8)"
          : "rgba(0, 0, 0, 0.8)",
      justifyContent: "center",
      alignItems: "center",
      borderRadius: theme.borderRadius.lg,
    },
  });
