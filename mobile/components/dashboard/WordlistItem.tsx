import { Wordlist } from "@/api/wordlists";
import { WordlistProgress } from "@/api/analytics";
import { MaterialIcons, Ionicons } from "@expo/vector-icons";
import React, { useState } from "react";
import { StyleSheet, Text, TouchableOpacity, View, Alert, ActivityIndicator } from "react-native";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { LANGUAGES } from "./CreateWordlistModal";
import * as wordlistsApi from "@/api/wordlists";
import { useRouter } from "expo-router";
import { useTranslation } from "react-i18next";
import { useUserSession } from "@/hooks/useUserSession";
import PremiumUpsellModal, {
  PremiumUpsellContext,
} from "@/components/common/PremiumUpsellModal";
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
  const [showPremiumModal, setShowPremiumModal] = useState(false);
  const [premiumModalContext, setPremiumModalContext] = useState<PremiumUpsellContext>(
    "analytics",
  );
  const router = useRouter();
  const { t } = useTranslation();
  const { isPremium } = useUserSession();
  const language = LANGUAGES.find((l) => item.languageCode === l.code)!;
  const { theme, responsive } = useTheme();
  const styles = createStyles(theme, responsive);

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
    Alert.alert(
      t("wordlistItem.deleteTitle"),
      t("wordlistItem.deleteConfirmMessage", { name: item.name }),
      [
        { text: t("common.cancel"), style: "cancel" },
        {
          text: t("common.delete"),
          style: "destructive",
          onPress: () => deleteMutation.mutate(),
        },
      ],
      { cancelable: true },
    );
  };

  const handleQuizStart = () => {
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

  const handleChatStart = () => {
    if (item.wordsCount === 0) {
      Alert.alert(
        t("wordlistItem.noWordsTitle"),
        t("wordlistItem.noWordsMessage"),
        [{ text: t("common.ok") }],
      );
      return;
    }

    // if (!isPremium) {
    //   setPremiumModalContext("chat");
    //   setShowPremiumModal(true);
    //   return;
    // }
    router.push(
      `/word-selection?wordlistId=${item.id}&wordlistName=${encodeURIComponent(item.name)}`,
    );
  };

  const handleAnalytics = () => {
    if (!isPremium) {
      setPremiumModalContext("analytics");
      setShowPremiumModal(true);
      return;
    }

    router.push(`/analytics?wordlistId=${item.id}`);
  };

  // Removed publish/share functionality

  // Removed menu-only: view on web handled elsewhere if needed

  return (
    <>
      <TouchableOpacity
        style={styles.wordlistCard}
        onPress={onPressed}
        activeOpacity={0.7}
        // removed menu long-press
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
              numberOfLines={responsive.getValueForSize(1, 2, 2, 2)}
              ellipsizeMode="tail"
              accessibilityRole="header"
              accessibilityLabel={item.name}
            >
              {item.name}
            </Text>
            {/* Public badge removed */}
          </View>
          <TouchableOpacity
            style={styles.headerMoreButton}
            onPress={handleDelete}
            accessibilityRole="button"
            accessibilityLabel={t("wordlistItem.deleteWordlist")}
            accessibilityHint={t("wordlistItem.deleteConfirmMessage", {
              name: item.name,
            })}
            disabled={deleteMutation.isPending}
            accessibilityState={{ disabled: deleteMutation.isPending }}
          >
            {deleteMutation.isPending ? (
              <ActivityIndicator size="small" color={theme.colors.error} />
            ) : (
              <Ionicons
                name="trash-outline"
                size={responsive.getValueForSize(18, 20, 22, 24)}
                color={theme.colors.error}
              />
            )}
          </TouchableOpacity>
        </View>

        <View style={styles.cardStats}>
          <View style={styles.cardStat}>
            <MaterialIcons
              name="library-books"
              size={responsive.getValueForSize(14, 16, 18, 20)}
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
                size={responsive.getValueForSize(14, 16, 18, 20)}
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

        {/* Action Buttons - Two Row Layout */}
        <View style={styles.actionButtonsContainer}>
          {/* Primary Learning Actions Row */}
          <View style={styles.primaryActionRow}>
            <TouchableOpacity
              style={styles.actionButtonPrimary}
              onPress={handleQuizStart}
              accessibilityRole="button"
              accessibilityLabel={t("wordlistItem.quiz")}
              accessibilityHint="Start quiz session"
            >
              <MaterialIcons
                name="play-circle-filled"
                size={responsive.getValueForSize(20, 22, 24, 26)}
                color={theme.colors.success}
              />
              <Text style={styles.actionButtonPrimaryText} numberOfLines={1}>
                {t("wordlistItem.quiz")}
              </Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.actionButtonPrimary}
              onPress={handlePractice}
              accessibilityRole="button"
              accessibilityLabel={t("wordlistItem.flashcards")}
              accessibilityHint="Practice with flashcards"
            >
              <MaterialIcons
                name="style"
                size={responsive.getValueForSize(20, 22, 24, 26)}
                color={theme.colors.semantic.info}
              />
              <Text style={styles.actionButtonPrimaryText} numberOfLines={1}>
                {t("wordlistItem.flashcards")}
              </Text>
            </TouchableOpacity>
          </View>

          {/* Secondary Features Row */}
          <View style={styles.secondaryActionRow}>
            <TouchableOpacity
              style={styles.actionButtonSecondary}
              onPress={handleChatStart}
              accessibilityRole="button"
              accessibilityLabel={t("wordlistItem.chat", "Chat")}
              accessibilityHint={t(
                "wordlistItem.chatHint",
                "Start a conversation using the words in this list",
              )}
            >
              <MaterialIcons
                name="chat"
                size={responsive.getValueForSize(16, 18, 20, 22)}
                color={theme.colors.primary}
              />
              <Text style={styles.actionButtonSecondaryText} numberOfLines={1}>
                {t("wordlistItem.chat", "Chat")}
              </Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.actionButtonSecondary}
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
                size={responsive.getValueForSize(16, 18, 20, 22)}
                color={theme.colors.premium}
              />
              <Text style={styles.actionButtonSecondaryText} numberOfLines={1}>
                {t("wordlistItem.analytics")}
              </Text>
            </TouchableOpacity>

            {/* Publish/Share action removed */}
          </View>
        </View>
      </TouchableOpacity>

      <PremiumUpsellModal
        visible={showPremiumModal}
        onClose={() => setShowPremiumModal(false)}
        context={premiumModalContext}
        onUpgradePress={onUpgradePress}
      />
    </>
  );
};

export default WordlistItem;

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  responsive: ReturnType<typeof useTheme>["responsive"],
) =>
  StyleSheet.create({
    wordlistCard: {
      backgroundColor:
        theme.mode === "light"
          ? theme.colors.background.surface
          : theme.colors.background.elevated,
      borderRadius: theme.borderRadius.lg,
      padding: responsive.spacing.formPadding,
      marginHorizontal: responsive.spacing.horizontal,
      marginBottom: responsive.spacing.vertical,
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
      alignItems: "center",
      justifyContent: "space-between",
      marginBottom: responsive.spacing.elementSpacing,
    },
    languageFlag: {
      fontSize: responsive.getValueForSize(28, 32, 36, 40),
      marginRight: responsive.spacing.elementSpacing,
    },
    cardTitleContainer: {
      flex: 1,
    },
    wordlistTitle: {
      fontSize: responsive.getScaledFont("headline"),
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: responsive.spacing.elementSpacing / 4,
    },
    // public badge styles removed
    headerMoreButton: {
      padding: 4,
      borderRadius: theme.borderRadius.sm,
    },
    actionButtonsContainer: {
      marginTop: responsive.spacing.elementSpacing,
      paddingTop: responsive.spacing.elementSpacing / 2,
      borderTopWidth: 1,
      borderTopColor: theme.colors.ui.divider,
      gap: responsive.spacing.elementSpacing / 2,
    },
    primaryActionRow: {
      flexDirection: "row",
      gap: responsive.spacing.elementSpacing,
    },
    secondaryActionRow: {
      flexDirection: "row",
      gap: responsive.spacing.elementSpacing / 2,
    },
    actionButtonPrimary: {
      flex: 1,
      flexDirection: "column",
      alignItems: "center",
      justifyContent: "center",
      paddingVertical: responsive.getValueForSize(12, 14, 16, 18),
      paddingHorizontal: responsive.getValueForSize(8, 10, 12, 14),
      borderRadius: theme.borderRadius.md,
      backgroundColor:
        theme.mode === "light"
          ? "rgba(253, 246, 227, 0.3)" // Slightly more visible for primary actions
          : theme.colors.background.subtle,
      gap: responsive.spacing.elementSpacing / 3,
      minHeight: responsive.getValueForSize(70, 74, 78, 82),
      ...theme.shadows.sm,
    },
    actionButtonSecondary: {
      flex: 1,
      flexDirection: "column",
      alignItems: "center",
      justifyContent: "center",
      paddingVertical: responsive.getValueForSize(8, 10, 12, 14),
      paddingHorizontal: responsive.getValueForSize(4, 6, 8, 10),
      borderRadius: theme.borderRadius.sm,
      backgroundColor:
        theme.mode === "light"
          ? "rgba(253, 246, 227, 0.15)" // More subtle for secondary actions
          : theme.colors.background.subtle,
      gap: responsive.spacing.elementSpacing / 4,
      minHeight: responsive.getValueForSize(50, 54, 58, 62),
    },
    actionButtonPrimaryText: {
      fontSize: responsive.getValueForSize(12, 14, 15, 16),
      color: theme.colors.text.primary,
      fontWeight: "600",
      textAlign: "center",
    },
    actionButtonSecondaryText: {
      fontSize: responsive.getValueForSize(10, 11, 12, 13),
      color: theme.colors.text.secondary,
      fontWeight: "500",
      textAlign: "center",
    },
    cardStats: {
      flexDirection: "row",
      flexWrap: "wrap",
      justifyContent: "space-between",
      alignItems: "center",
      gap: responsive.getValueForSize(12, 16, 20, 24),
      marginBottom: responsive.spacing.elementSpacing,
      paddingHorizontal: responsive.spacing.elementSpacing / 4,
    },
    cardStat: {
      flexDirection: "row",
      alignItems: "center",
      gap: responsive.getValueForSize(6, 8, 10, 12),
    },
    cardStatText: {
      fontSize: responsive.getScaledFont("label"),
      color: theme.colors.text.secondary,
    },
    languageName: {
      fontSize: responsive.getScaledFont("label"),
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
      padding: responsive.spacing.horizontal,
    },
    menuContainer: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: theme.borderRadius.lg,
      padding: responsive.spacing.elementSpacing / 2,
      width: "100%",
      maxWidth: responsive.getValueForSize(280, 320, 360, 400),
      ...theme.shadows.lg,
    },
    menuTitle: {
      fontSize: responsive.getScaledFont("headline"),
      fontWeight: "600",
      color: theme.colors.text.primary,
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.elementSpacing,
      marginBottom: responsive.spacing.elementSpacing / 4,
    },
    menuItem: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.elementSpacing,
      borderRadius: theme.borderRadius.sm,
      gap: responsive.spacing.elementSpacing,
      minHeight: responsive.spacing.minTouchTarget,
    },
    menuItemText: {
      fontSize: responsive.getScaledFont("body"),
      color: theme.colors.text.primary,
      flex: 1,
    },
    menuDivider: {
      height: 1,
      backgroundColor: theme.colors.ui.divider,
      marginVertical: responsive.spacing.elementSpacing / 2,
      marginHorizontal: responsive.spacing.horizontal,
    },
    deleteMenuItem: {
      marginTop: responsive.spacing.elementSpacing / 4,
    },
    deleteMenuItemText: {
      color: theme.colors.error,
    },
    // Premium Modal Styles
    premiumModalContainer: {
      backgroundColor: "#FFFFFF",
      borderRadius: theme.borderRadius.xl,
      margin: responsive.spacing.horizontal,
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
      borderRadius: theme.borderRadius.lg,
      padding: responsive.spacing.formPadding,
      alignItems: "center",
    },
    premiumIconContainer: {
      width: responsive.getValueForSize(56, 60, 64, 68),
      height: responsive.getValueForSize(56, 60, 64, 68),
      borderRadius: responsive.getValueForSize(28, 30, 32, 34),
      backgroundColor: theme.colors.premium,
      justifyContent: "center",
      alignItems: "center",
      marginBottom: responsive.spacing.elementSpacing,
    },
    premiumTitle: {
      fontSize: responsive.getScaledFont("headline"),
      fontWeight: "600",
      color: "#2D3436",
      marginBottom: responsive.spacing.elementSpacing / 2,
      textAlign: "center",
    },
    premiumSubtitle: {
      fontSize: responsive.getScaledFont("body"),
      color: "#636E72",
      textAlign: "center",
      marginBottom: responsive.spacing.elementSpacing * 2,
      lineHeight: responsive.fontSizes.lineHeight,
    },
    premiumButtons: {
      width: "100%",
      gap: responsive.spacing.elementSpacing,
    },
    upgradeButton: {
      backgroundColor: "#FFD700",
      borderRadius: theme.borderRadius.md,
      paddingVertical: responsive.spacing.elementSpacing,
      paddingHorizontal: responsive.spacing.horizontal,
      alignItems: "center",
      minHeight: responsive.spacing.buttonHeight,
    },
    upgradeButtonText: {
      fontSize: responsive.getScaledFont("body"),
      fontWeight: "600",
      color: "#2D3436",
    },
    cancelButton: {
      backgroundColor: "transparent",
      borderRadius: theme.borderRadius.md,
      paddingVertical: responsive.spacing.elementSpacing,
      paddingHorizontal: responsive.spacing.horizontal,
      alignItems: "center",
      minHeight: responsive.spacing.buttonHeight,
    },
    cancelButtonText: {
      fontSize: responsive.getScaledFont("body"),
      fontWeight: "500",
      color: "#636E72",
    },
    // share overlay style removed
  });
