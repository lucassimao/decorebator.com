import React, { useState } from "react";
import {
  Modal,
  TouchableWithoutFeedback,
  View,
  Text,
  TextInput,
  TouchableOpacity,
  ActivityIndicator,
  StyleSheet,
} from "react-native";
import { MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useMutation } from "@tanstack/react-query";
import { useTheme } from "@/contexts/ThemeContext";
import * as wordlistsApi from "@/api/wordlists";

type Props = {
  visible: boolean;
  wordlistId: number;
  defaultTitle: string;
  onClose: () => void;
  onSuccess?: (slug: string) => void;
};

const PublishPublicQuizModal: React.FC<Props> = ({
  visible,
  wordlistId,
  defaultTitle,
  onClose,
  onSuccess,
}) => {
  const { t } = useTranslation();
  const { theme, responsive } = useTheme();
  const styles = createStyles(theme, responsive);

  const [title, setTitle] = useState<string>(defaultTitle);
  const [description, setDescription] = useState<string>("");
  const [difficulty, setDifficulty] = useState<
    wordlistsApi.PublicQuizDifficulty
  >("medium");
  const [timeLimit, setTimeLimit] = useState<number>(10);

  const publishMutation = useMutation({
    mutationFn: () =>
      wordlistsApi.publishPublicQuiz(wordlistId, {
        title,
        description: description || undefined,
        difficulty,
        timeLimitMinutes: timeLimit,
      }),
    onSuccess: (res) => {
      onClose();
      onSuccess?.(res.slug);
    },
  });

  const handleConfirm = () => {
    if (!title?.trim()) {
      return; // Parent handles alerts; keep component lean
    }
    if (timeLimit < 1 || timeLimit > 15) {
      return;
    }
    publishMutation.mutate();
  };

  return (
    <Modal
      visible={visible}
      transparent
      animationType="fade"
      onRequestClose={onClose}
      accessibilityViewIsModal={true}
    >
      <TouchableWithoutFeedback onPress={onClose}>
        <View style={styles.modalOverlay}>
          <TouchableWithoutFeedback>
            <View style={styles.container}>
              <Text style={styles.title}>
                {t("wordlistItem.publishTitle", "Publish public quiz")}
              </Text>

              <Text style={styles.inputLabel}>
                {t("wordlistItem.quizTitle", "Quiz title")}
              </Text>
              <TextInput
                style={styles.input}
                value={title}
                onChangeText={(text) => setTitle(text)}
                placeholder={t(
                  "wordlistItem.quizTitlePlaceholder",
                  "Enter a title",
                )}
                accessibilityLabel="Quiz title"
              />

              <Text style={styles.inputLabel}>
                {t("wordlistItem.quizDescription", "Description (optional)")}
              </Text>
              <TextInput
                style={[styles.input, styles.textArea]}
                value={description}
                onChangeText={(text) => setDescription(text)}
                placeholder={t(
                  "wordlistItem.quizDescriptionPlaceholder",
                  "Short description",
                )}
                multiline
                numberOfLines={3}
                accessibilityLabel="Quiz description"
              />

              <Text style={styles.inputLabel}>
                {t("wordlistItem.difficulty", "Difficulty")}
              </Text>
              <View style={styles.segmentRow}>
                {["easy", "medium", "hard"].map((level) => (
                  <TouchableOpacity
                    key={level}
                    style={[
                      styles.segmentButton,
                      difficulty === level && styles.segmentButtonActive,
                    ]}
                    onPress={() => setDifficulty(level as any)}
                    accessibilityRole="button"
                    accessibilityLabel={`Difficulty ${level}`}
                  >
                    <Text
                      style={[
                        styles.segmentButtonText,
                        difficulty === level && styles.segmentButtonTextActive,
                      ]}
                    >
                      {t(`wordlistItem.diff.${level}`, level)}
                    </Text>
                  </TouchableOpacity>
                ))}
              </View>

              <Text style={styles.inputLabel}>
                {t("wordlistItem.timeLimit", "Time limit (minutes)")}
              </Text>
              <View style={styles.timeRow}>
                <TouchableOpacity
                  style={styles.timeButton}
                  onPress={() => setTimeLimit((m) => Math.max(1, m - 1))}
                  accessibilityLabel="Decrease time limit"
                  accessibilityRole="button"
                >
                  <MaterialIcons
                    name="remove"
                    size={20}
                    color={theme.colors.text.primary}
                  />
                </TouchableOpacity>
                <Text style={styles.timeValue}>{timeLimit}</Text>
                <TouchableOpacity
                  style={styles.timeButton}
                  onPress={() => setTimeLimit((m) => Math.min(15, m + 1))}
                  accessibilityLabel="Increase time limit"
                  accessibilityRole="button"
                >
                  <MaterialIcons
                    name="add"
                    size={20}
                    color={theme.colors.text.primary}
                  />
                </TouchableOpacity>
              </View>

              <View style={styles.actionsRow}>
                <TouchableOpacity
                  style={[styles.actionButton, styles.cancelButton]}
                  onPress={onClose}
                  accessibilityRole="button"
                  accessibilityLabel={t("common.cancel")}
                >
                  <Text style={styles.cancelButtonText}>{t("common.cancel")}</Text>
                </TouchableOpacity>
                <TouchableOpacity
                  style={[
                    styles.actionButton,
                    publishMutation.isPending && { opacity: 0.7 },
                  ]}
                  onPress={handleConfirm}
                  disabled={publishMutation.isPending}
                  accessibilityRole="button"
                  accessibilityLabel={t("wordlistItem.publish", "Publish")}
                >
                  {publishMutation.isPending ? (
                    <ActivityIndicator
                      size="small"
                      color={theme.colors.text.inverse}
                    />
                  ) : (
                    <Text style={styles.publishButtonText}>
                      {t("wordlistItem.publish", "Publish")}
                    </Text>
                  )}
                </TouchableOpacity>
              </View>
            </View>
          </TouchableWithoutFeedback>
        </View>
      </TouchableWithoutFeedback>
    </Modal>
  );
};

export default PublishPublicQuizModal;

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  responsive: ReturnType<typeof useTheme>["responsive"],
) =>
  StyleSheet.create({
    modalOverlay: {
      flex: 1,
      backgroundColor: theme.colors.overlay.backdrop,
      justifyContent: "center",
      alignItems: "center",
      padding: responsive.spacing.horizontal,
    },
    container: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: theme.borderRadius.lg,
      padding: responsive.spacing.formPadding,
      width: "100%",
      maxWidth: responsive.getValueForSize(320, 360, 400, 440),
      ...theme.shadows.lg,
    },
    title: {
      fontSize: responsive.getScaledFont("headline"),
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: responsive.spacing.elementSpacing,
      textAlign: "center",
    },
    inputLabel: {
      fontSize: responsive.getScaledFont("label"),
      color: theme.colors.text.secondary,
      marginBottom: responsive.spacing.elementSpacing / 4,
    },
    input: {
      borderWidth: 1,
      borderColor: theme.colors.ui.divider,
      borderRadius: theme.borderRadius.sm,
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.elementSpacing / 1.5,
      color: theme.colors.text.primary,
      marginBottom: responsive.spacing.elementSpacing,
      backgroundColor: theme.colors.background.surface,
    },
    textArea: {
      minHeight: responsive.getValueForSize(64, 80, 96, 112),
      textAlignVertical: "top",
    },
    segmentRow: {
      flexDirection: "row",
      gap: responsive.spacing.elementSpacing / 2,
      marginBottom: responsive.spacing.elementSpacing,
    },
    segmentButton: {
      flex: 1,
      alignItems: "center",
      paddingVertical: responsive.getValueForSize(8, 10, 12, 12),
      borderRadius: theme.borderRadius.sm,
      borderWidth: 1,
      borderColor: theme.colors.ui.divider,
      backgroundColor: theme.colors.background.subtle,
    },
    segmentButtonActive: {
      backgroundColor: theme.colors.primary,
      borderColor: theme.colors.primary,
    },
    segmentButtonText: {
      color: theme.colors.text.primary,
      fontWeight: "500",
    },
    segmentButtonTextActive: {
      color: theme.colors.text.inverse,
    },
    timeRow: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      gap: responsive.spacing.elementSpacing,
      marginBottom: responsive.spacing.elementSpacing,
    },
    timeButton: {
      borderWidth: 1,
      borderColor: theme.colors.ui.divider,
      borderRadius: theme.borderRadius.sm,
      padding: responsive.getValueForSize(6, 8, 10, 10),
      backgroundColor: theme.colors.background.subtle,
    },
    timeValue: {
      fontSize: responsive.getScaledFont("headline"),
      fontWeight: "600",
      color: theme.colors.text.primary,
      minWidth: 32,
      textAlign: "center",
    },
    actionsRow: {
      flexDirection: "row",
      gap: responsive.spacing.elementSpacing,
    },
    actionButton: {
      flex: 1,
      alignItems: "center",
      justifyContent: "center",
      borderRadius: theme.borderRadius.sm,
      paddingVertical: responsive.getValueForSize(10, 12, 14, 14),
      backgroundColor: theme.colors.primary,
    },
    publishButtonText: {
      color: theme.colors.text.inverse,
      fontWeight: "600",
      fontSize: responsive.getScaledFont("body"),
    },
    cancelButton: {
      backgroundColor: "transparent",
      borderWidth: 1,
      borderColor: theme.colors.ui.divider,
    },
    cancelButtonText: {
      fontSize: responsive.getScaledFont("body"),
      fontWeight: "500",
      color: theme.colors.text.secondary,
    },
  });


