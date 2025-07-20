import { Wordlist } from "@/api/wordlists";
import { WordlistProgress } from "@/api/analytics";
import { MaterialIcons, Ionicons } from "@expo/vector-icons";
import React, { useState } from "react";
import {
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
  Alert,
  ActivityIndicator,
  Modal,
  TouchableWithoutFeedback,
} from "react-native";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { LANGUAGES } from "./CreateWordlistModal";
import * as wordlistsApi from "@/api/wordlists";
import { useRouter } from "expo-router";
import { useTranslation } from "react-i18next";
import { useUserSession } from "@/hooks/useUserSession";
import { LinearGradient } from "expo-linear-gradient";
import { useTheme } from "@/contexts/ThemeContext";

type WordlistItemProps = {
  item: Wordlist;
  progress?: WordlistProgress;
  onQuizStart?: (wordlist: Wordlist) => void;
  onPressed?: () => void;
  onUpgradePress?: () => void;
};

const WordlistItem: React.FC<WordlistItemProps> = ({
  item,
  progress,
  onQuizStart,
  onPressed,
  onUpgradePress,
}) => {
  const queryClient = useQueryClient();
  const [showMenu, setShowMenu] = useState(false);
  const [showPremiumModal, setShowPremiumModal] = useState(false);
  const router = useRouter();
  const { t } = useTranslation();
  const { isPremium } = useUserSession();
  const language = LANGUAGES.find((l) => item.languageCode === l.code)!;
  const { theme } = useTheme();
  const styles = createStyles(theme);

  // Use progress from props
  const progressPercentage = progress?.progressPercent ?? 0;

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: () => wordlistsApi.deleteWordlist(item.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wordlists"] });
      queryClient.invalidateQueries({ queryKey: ["dashboardStats"] });
      Alert.alert(t("common.success"), t("wordlistItem.deleteSuccess"));
    },
    onError: (error) => {
      console.error(error);
      Alert.alert(t("common.error"), t("wordlistItem.deleteError"));
    },
  });

  const handleDelete = () => {
    setShowMenu(false);

    setTimeout(() => {
      Alert.alert(
        t("wordlistItem.deleteTitle"),
        t("wordlistItem.deleteConfirmMessage", { name: item.name }),
        [
          {
            text: t("common.cancel"),
            style: "cancel",
          },
          {
            text: t("common.delete"),
            style: "destructive",
            onPress: () => {
              deleteMutation.mutate();
            },
          },
        ],
        { cancelable: true },
      );
    }, 100);
  };

  const handleQuizStart = () => {
    setShowMenu(false);

    if (item.wordsCount === 0) {
      Alert.alert(
        t("wordlistItem.noWordsTitle"),
        t("wordlistItem.noWordsMessage"),
        [{ text: t("common.ok") }],
      );
      return;
    }

    if (onQuizStart) {
      onQuizStart(item);
    } else {
      router.push(`/quiz?wordlistId=${item.id}&wordlistName=${item.name}`);
    }
  };

  const handlePractice = () => {
    setShowMenu(false);

    if (item.wordsCount === 0) {
      Alert.alert(
        t("wordlistItem.noWordsTitle"),
        t("wordlistItem.noWordsMessage"),
        [{ text: t("common.ok") }],
      );
      return;
    }

    router.push(`/flashcard?wordlistId=${item.id}&wordlistName=${item.name}`);
  };

  const handleAnalytics = () => {
    setShowMenu(false);

    if (!isPremium) {
      setShowPremiumModal(true);
      return;
    }

    router.push(`/analytics?wordlistId=${item.id}`);
  };

  const handleEdit = () => {
    setShowMenu(false);
    onPressed?.();
  };

  return (
    <>
      <TouchableOpacity
        style={styles.wordlistCard}
        onPress={onPressed}
        activeOpacity={0.7}
        onLongPress={() => setShowMenu(true)}
        accessibilityRole="button"
        accessibilityLabel={`${item.name} wordlist. ${item.wordsCount} words. ${progressPercentage > 0 ? `${Math.round(progressPercentage)}% learned.` : "No progress yet."} Double tap to open details, long press for menu.`}
        accessibilityHint="Open wordlist details or long press for actions menu"
      >
        <View style={styles.cardHeader}>
          <Text
            style={styles.languageFlag}
            accessibilityLabel={`Language: ${t(`dashboard.languages.${language.name.toLowerCase()}`)}`}
          >
            {language.flag}
          </Text>
          <View style={styles.cardTitleContainer}>
            <Text
              style={styles.wordlistTitle}
              accessibilityRole="header"
              accessibilityLabel={item.name}
            >
              {item.name}
            </Text>
            {item.description && (
              <Text
                style={styles.wordlistDescription}
                numberOfLines={2}
                accessibilityRole="text"
              >
                {item.description}
              </Text>
            )}
          </View>
        </View>

        <View style={styles.cardStats}>
          <View style={styles.cardStat}>
            <MaterialIcons
              name="library-books"
              size={16}
              color={theme.colors.text.secondary}
            />
            <Text style={styles.cardStatText}>
              {t("wordlistItem.wordCount", { count: item.wordsCount ?? 0 })}
            </Text>
          </View>
          <View style={styles.cardStat}>
            <Text style={styles.languageName}>
              {t(`dashboard.languages.${language.name.toLowerCase()}`)}
            </Text>
          </View>
          {progressPercentage > 0 && (
            <View style={styles.cardStat}>
              <MaterialIcons
                name="school"
                size={16}
                color={theme.colors.text.secondary}
              />
              <Text style={styles.cardStatText}>
                {t("wordlistItem.percentLearned", {
                  percent: Math.round(progressPercentage),
                })}
              </Text>
            </View>
          )}
        </View>

        <View style={styles.progressBar}>
          <View
            style={[styles.progressFill, { width: `${progressPercentage}%` }]}
          />
        </View>

        {/* Action Buttons Row */}
        <View style={styles.actionButtonsRow}>
          <TouchableOpacity
            style={styles.actionButtonLarge}
            onPress={handleAnalytics}
            accessibilityRole="button"
            accessibilityLabel={t("wordlistItem.analytics")}
            accessibilityHint={
              isPremium
                ? "View detailed learning analytics"
                : "Premium feature - tap to upgrade"
            }
          >
            <MaterialIcons
              name="bar-chart"
              size={20}
              color={theme.colors.premium}
            />
            <Text style={styles.actionButtonText}>
              {t("wordlistItem.analytics")}
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.actionButtonLarge}
            onPress={handlePractice}
            accessibilityRole="button"
            accessibilityLabel={t("wordlistItem.flashcards")}
            accessibilityHint="Practice with flashcards"
          >
            <MaterialIcons
              name="style"
              size={20}
              color={theme.colors.semantic.info}
            />
            <Text style={styles.actionButtonText}>
              {t("wordlistItem.flashcards")}
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.actionButtonLarge}
            onPress={handleQuizStart}
            accessibilityRole="button"
            accessibilityLabel={t("wordlistItem.quiz")}
            accessibilityHint="Start quiz session"
          >
            <MaterialIcons
              name="play-circle-filled"
              size={20}
              color={theme.colors.success}
            />
            <Text style={styles.actionButtonText}>
              {t("wordlistItem.quiz")}
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.actionButtonLarge}
            onPress={() => setShowMenu(true)}
            accessibilityRole="button"
            accessibilityLabel={t("common.more")}
            accessibilityHint="Show more options for this wordlist"
          >
            <MaterialIcons
              name="more-horiz"
              size={20}
              color={theme.colors.text.secondary}
            />
            <Text style={styles.actionButtonText}>{t("common.more")}</Text>
          </TouchableOpacity>
        </View>
      </TouchableOpacity>

      {/* Action Menu Modal */}
      <Modal
        visible={showMenu}
        transparent
        animationType="fade"
        onRequestClose={() => setShowMenu(false)}
        accessibilityViewIsModal={true}
      >
        <TouchableWithoutFeedback onPress={() => setShowMenu(false)}>
          <View style={styles.modalOverlay}>
            <TouchableWithoutFeedback>
              <View style={styles.menuContainer}>
                <Text
                  style={styles.menuTitle}
                  accessibilityRole="header"
                  accessibilityLabel={item.name}
                >
                  {item.name}
                </Text>

                <TouchableOpacity
                  style={styles.menuItem}
                  onPress={() => {
                    setShowMenu(false);
                    router.push(`/analytics?wordlistId=${item.id}`);
                  }}
                  accessibilityRole="button"
                  accessibilityLabel={t("wordlistItem.viewAnalytics")}
                  accessibilityHint="View detailed learning analytics for this wordlist"
                >
                  <MaterialIcons
                    name="analytics"
                    size={24}
                    color={theme.colors.premium}
                  />
                  <Text style={styles.menuItemText}>
                    {t("wordlistItem.viewAnalytics")}
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity
                  style={styles.menuItem}
                  onPress={handlePractice}
                  accessibilityRole="button"
                  accessibilityLabel={t("wordlistItem.practiceFlashcards")}
                  accessibilityHint="Practice with interactive flashcards"
                >
                  <MaterialIcons
                    name="style"
                    size={24}
                    color={theme.colors.semantic.info}
                  />
                  <Text style={styles.menuItemText}>
                    {t("wordlistItem.practiceFlashcards")}
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity
                  style={styles.menuItem}
                  onPress={handleQuizStart}
                  accessibilityRole="button"
                  accessibilityLabel={t("wordlistItem.startQuiz")}
                  accessibilityHint="Start an interactive quiz session"
                >
                  <MaterialIcons
                    name="quiz"
                    size={24}
                    color={theme.colors.success}
                  />
                  <Text style={styles.menuItemText}>
                    {t("wordlistItem.startQuiz")}
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity
                  style={styles.menuItem}
                  onPress={handleEdit}
                  accessibilityRole="button"
                  accessibilityLabel={t("wordlistItem.editWordlist")}
                  accessibilityHint="Edit wordlist details and add or remove words"
                >
                  <MaterialIcons
                    name="edit"
                    size={24}
                    color={theme.colors.primary}
                  />
                  <Text style={styles.menuItemText}>
                    {t("wordlistItem.editWordlist")}
                  </Text>
                </TouchableOpacity>

                <View style={styles.menuDivider} />

                <TouchableOpacity
                  style={[styles.menuItem, styles.deleteMenuItem]}
                  onPress={handleDelete}
                  disabled={deleteMutation.isPending}
                  accessibilityRole="button"
                  accessibilityLabel={
                    deleteMutation.isPending
                      ? "Deleting wordlist..."
                      : t("wordlistItem.deleteWordlist")
                  }
                  accessibilityHint="Permanently delete this wordlist and all its words"
                  accessibilityState={{ disabled: deleteMutation.isPending }}
                >
                  {deleteMutation.isPending ? (
                    <ActivityIndicator
                      size="small"
                      color={theme.colors.error}
                    />
                  ) : (
                    <Ionicons
                      name="trash-outline"
                      size={24}
                      color={theme.colors.error}
                    />
                  )}
                  <Text
                    style={[styles.menuItemText, styles.deleteMenuItemText]}
                  >
                    {t("wordlistItem.deleteWordlist")}
                  </Text>
                </TouchableOpacity>
              </View>
            </TouchableWithoutFeedback>
          </View>
        </TouchableWithoutFeedback>
      </Modal>

      {/* Premium Analytics Upsell Modal */}
      <Modal
        visible={showPremiumModal}
        transparent
        animationType="fade"
        onRequestClose={() => setShowPremiumModal(false)}
        accessibilityViewIsModal={true}
      >
        <TouchableWithoutFeedback onPress={() => setShowPremiumModal(false)}>
          <View style={styles.modalOverlay}>
            <TouchableWithoutFeedback>
              <View style={styles.premiumModalContainer}>
                <LinearGradient
                  colors={[theme.colors.premium, theme.colors.semantic.warning]}
                  style={styles.premiumGradient}
                  start={{ x: 0, y: 0 }}
                  end={{ x: 1, y: 0 }}
                >
                  <View style={styles.premiumContent}>
                    <View style={styles.premiumIconContainer}>
                      <MaterialIcons
                        name="analytics"
                        size={32}
                        color={theme.colors.text.inverse}
                      />
                    </View>
                    <Text
                      style={styles.premiumTitle}
                      accessibilityRole="header"
                      accessibilityLabel={item.name}
                    >
                      {t("dashboard.stats.premium.title")}
                    </Text>
                    <Text
                      style={styles.premiumSubtitle}
                      accessibilityRole="text"
                    >
                      {t("dashboard.stats.premium.subtitle")}
                    </Text>

                    <View style={styles.premiumButtons}>
                      <TouchableOpacity
                        style={styles.upgradeButton}
                        onPress={() => {
                          setShowPremiumModal(false);
                          onUpgradePress?.();
                        }}
                        accessibilityRole="button"
                        accessibilityLabel={t(
                          "settings.subscription.upgradeButton",
                        )}
                        accessibilityHint="Upgrade to premium to access analytics features"
                      >
                        <Text style={styles.upgradeButtonText}>
                          {t("settings.subscription.upgradeButton")}
                        </Text>
                      </TouchableOpacity>

                      <TouchableOpacity
                        style={styles.cancelButton}
                        onPress={() => setShowPremiumModal(false)}
                        accessibilityRole="button"
                        accessibilityLabel={t("upgradePrompt.maybeLater")}
                        accessibilityHint="Close this upgrade prompt"
                      >
                        <Text style={styles.cancelButtonText}>
                          {t("upgradePrompt.maybeLater")}
                        </Text>
                      </TouchableOpacity>
                    </View>
                  </View>
                </LinearGradient>
              </View>
            </TouchableWithoutFeedback>
          </View>
        </TouchableWithoutFeedback>
      </Modal>
    </>
  );
};

export default WordlistItem;

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    wordlistCard: {
      backgroundColor:
        theme.mode === "light"
          ? theme.colors.background.surface
          : theme.colors.background.elevated,
      borderRadius: theme.borderRadius.lg,
      padding: theme.spacing.md,
      marginHorizontal: 20,
      marginBottom: 12,
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? "rgba(253, 246, 227, 0.3)" // Subtle web beige border
          : theme.colors.ui.divider,
      ...theme.shadows.md,
      // Extra elevation for better contrast with subtle orange shadow
      shadowColor:
        theme.mode === "light"
          ? theme.colors.primary
          : theme.colors.text.primary,
      shadowOpacity: theme.mode === "light" ? 0.15 : 0.1,
      elevation: theme.mode === "light" ? 6 : 10,
    },
    cardHeader: {
      flexDirection: "row",
      alignItems: "flex-start",
      marginBottom: 12,
    },
    languageFlag: {
      fontSize: 32,
      marginRight: 12,
    },
    cardTitleContainer: {
      flex: 1,
    },
    wordlistTitle: {
      fontSize: 18,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: 4,
    },
    wordlistDescription: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      lineHeight: 20,
    },
    actionButtonsRow: {
      flexDirection: "row",
      marginTop: 12,
      gap: 8,
      paddingTop: 8,
      borderTopWidth: 1,
      borderTopColor: theme.colors.ui.divider,
    },
    actionButtonLarge: {
      flex: 1,
      flexDirection: "column",
      alignItems: "center",
      paddingVertical: 12,
      paddingHorizontal: 8,
      borderRadius: theme.borderRadius.sm,
      backgroundColor:
        theme.mode === "light"
          ? "rgba(253, 246, 227, 0.2)" // Very subtle web beige tint
          : theme.colors.background.subtle,
      gap: 4,
    },
    actionButtonText: {
      fontSize: 12,
      color: theme.colors.text.secondary,
      fontWeight: "500",
      textAlign: "center",
    },
    cardStats: {
      flexDirection: "row",
      flexWrap: "wrap",
      gap: 16,
      marginBottom: 12,
    },
    cardStat: {
      flexDirection: "row",
      alignItems: "center",
      gap: 4,
    },
    cardStatText: {
      fontSize: 14,
      color: theme.colors.text.secondary,
    },
    languageName: {
      fontSize: 14,
      color: theme.colors.primary,
      fontWeight: "500",
    },
    progressBar: {
      height: 4,
      backgroundColor: theme.colors.ui.divider,
      borderRadius: 2,
      overflow: "hidden",
    },
    progressFill: {
      height: "100%",
      backgroundColor: theme.colors.success,
      borderRadius: 2,
    },
    modalOverlay: {
      flex: 1,
      backgroundColor: theme.colors.overlay.backdrop,
      justifyContent: "center",
      alignItems: "center",
      padding: 20,
    },
    menuContainer: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: theme.borderRadius.lg,
      padding: 8,
      width: "100%",
      maxWidth: 320,
      ...theme.shadows.lg,
    },
    menuTitle: {
      fontSize: 18,
      fontWeight: "600",
      color: theme.colors.text.primary,
      paddingHorizontal: 16,
      paddingVertical: 12,
      marginBottom: 4,
    },
    menuItem: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: 16,
      paddingVertical: 14,
      borderRadius: 8,
      gap: 12,
    },
    menuItemText: {
      fontSize: 16,
      color: theme.colors.text.primary,
      flex: 1,
    },
    menuDivider: {
      height: 1,
      backgroundColor: theme.colors.ui.divider,
      marginVertical: 8,
      marginHorizontal: 16,
    },
    deleteMenuItem: {
      marginTop: 4,
    },
    deleteMenuItemText: {
      color: theme.colors.error,
    },
    // Premium Modal Styles
    premiumModalContainer: {
      backgroundColor: "#FFFFFF",
      borderRadius: 20,
      margin: 20,
      overflow: "hidden",
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.15,
      shadowRadius: 12,
      elevation: 8,
    },
    premiumGradient: {
      padding: 2,
    },
    premiumContent: {
      backgroundColor: "#FFFFFF",
      borderRadius: 18,
      padding: 24,
      alignItems: "center",
    },
    premiumIconContainer: {
      width: 60,
      height: 60,
      borderRadius: 30,
      backgroundColor: theme.colors.premium,
      justifyContent: "center",
      alignItems: "center",
      marginBottom: 16,
    },
    premiumTitle: {
      fontSize: 18,
      fontWeight: "600",
      color: "#2D3436",
      marginBottom: 8,
      textAlign: "center",
    },
    premiumSubtitle: {
      fontSize: 14,
      color: "#636E72",
      textAlign: "center",
      marginBottom: 24,
      lineHeight: 20,
    },
    premiumButtons: {
      width: "100%",
      gap: 12,
    },
    upgradeButton: {
      backgroundColor: "#FFD700",
      borderRadius: 12,
      paddingVertical: 14,
      paddingHorizontal: 24,
      alignItems: "center",
    },
    upgradeButtonText: {
      fontSize: 16,
      fontWeight: "600",
      color: "#2D3436",
    },
    cancelButton: {
      backgroundColor: "transparent",
      borderRadius: 12,
      paddingVertical: 14,
      paddingHorizontal: 24,
      alignItems: "center",
    },
    cancelButtonText: {
      fontSize: 16,
      fontWeight: "500",
      color: "#636E72",
    },
  });
