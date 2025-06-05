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

  const colors = {
    textDark: "#2D3436",
    textMedium: "#636E72",
    white: "#FFFFFF",
    lightBackground: "#FAFAFA",
    primary: "#FF7B54",
  };

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
        <View style={styles.reportModal}>
          <Text style={styles.reportModalTitle}>
            {t(`${context}.reportIssue`)}
          </Text>
          {wordName && (
            <Text style={styles.reportModalSubtitle}>
              {t(`${context}.reportWord`, { word: wordName })}
            </Text>
          )}

          <View style={styles.reportOptions}>
            {isUnrelatedImageOptionAvailable && (
              <TouchableOpacity
                style={styles.reportOption}
                onPress={() => onReportError(ErrorType.UnrelatedImage)}
                disabled={isLoading}
              >
                <MaterialIcons
                  name="image"
                  size={24}
                  color={colors.textMedium}
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
                  color={colors.textMedium}
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
                  color={colors.textMedium}
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
                  color={colors.textMedium}
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
                  color={colors.textMedium}
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
              <ActivityIndicator size="large" color={colors.primary} />
            </View>
          )}
        </View>
      </View>
    </Modal>
  );
};

const styles = StyleSheet.create({
  modalOverlay: {
    flex: 1,
    backgroundColor: "rgba(0, 0, 0, 0.5)",
    justifyContent: "center",
    alignItems: "center",
    padding: 20,
  },
  reportModal: {
    backgroundColor: "#FFFFFF",
    borderRadius: 16,
    padding: 24,
    width: "100%",
    maxWidth: 400,
    position: "relative",
  },
  reportModalTitle: {
    fontSize: 20,
    fontWeight: "600",
    color: "#2D3436",
    textAlign: "center",
    marginBottom: 8,
  },
  reportModalSubtitle: {
    fontSize: 16,
    color: "#636E72",
    textAlign: "center",
    marginBottom: 24,
  },
  reportOptions: {
    gap: 12,
    marginBottom: 24,
  },
  reportOption: {
    flexDirection: "row",
    alignItems: "center",
    padding: 16,
    backgroundColor: "#FAFAFA",
    borderRadius: 12,
    gap: 12,
  },
  reportOptionText: {
    fontSize: 16,
    color: "#2D3436",
    flex: 1,
  },
  reportModalActions: {
    flexDirection: "row",
    justifyContent: "center",
  },
  reportCancelButton: {
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
  },
  reportCancelText: {
    fontSize: 16,
    color: "#636E72",
    fontWeight: "500",
  },
  reportLoadingOverlay: {
    position: "absolute",
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: "rgba(255, 255, 255, 0.8)",
    justifyContent: "center",
    alignItems: "center",
    borderRadius: 16,
  },
});
