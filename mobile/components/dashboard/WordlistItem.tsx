import { Wordlist } from "@/api/wordlists";
import { WordlistProgress } from "@/api/analytics";
import { MaterialIcons, Ionicons } from "@expo/vector-icons";
import { LinearGradient } from "expo-linear-gradient";
import React, { useEffect, useRef, useState } from "react";
import {
  Animated,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
  Alert,
  ActivityIndicator,
} from "react-native";
import { useFocusEffect } from "expo-router";
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
  onAddWords?: (wordlist: Wordlist) => void;
  onUpgradePress?: () => void;
};

const WordlistItem: React.FC<WordlistItemProps> = ({
  item,
  progress,
  onQuizStart,
  onPressed,
  onAddWords,
  onUpgradePress,
}) => {
  const queryClient = useQueryClient();
  const [showPremiumModal, setShowPremiumModal] = useState(false);
  const [premiumModalContext, setPremiumModalContext] =
    useState<PremiumUpsellContext>("analytics");
  const router = useRouter();
  const { t } = useTranslation();
  const { isPremium } = useUserSession();
  const language = LANGUAGES.find((l) => item.languageCode === l.code)!;
  const { theme, responsive } = useTheme();
  const styles = createStyles(theme, responsive);
  const shimmerAnim = useRef(new Animated.Value(-1)).current;
  const shimmerLoopRef = useRef<Animated.CompositeAnimation | null>(null);
  const shimmerPlayedRef = useRef(false);
  const [buttonWidth, setButtonWidth] = useState(0);

  // Use progress from props
  const progressPercentage = progress?.progressPercent ?? 0;
  const isEmptyWordlist = (item.wordsCount ?? 0) === 0;

  const startShimmer = React.useCallback(() => {
    if (shimmerLoopRef.current || shimmerPlayedRef.current) return;
    shimmerAnim.setValue(-1);
    const anim = Animated.timing(shimmerAnim, {
      toValue: 1,
      duration: 3200,
      useNativeDriver: true,
    });
    shimmerLoopRef.current = anim;
    anim.start(({ finished }) => {
      shimmerLoopRef.current = null;
      if (finished) {
        shimmerPlayedRef.current = true;
      }
    });
  }, [shimmerAnim]);

  const stopShimmer = React.useCallback(() => {
    if (shimmerLoopRef.current) {
      shimmerLoopRef.current.stop();
      shimmerLoopRef.current = null;
    }
    shimmerAnim.stopAnimation();
    shimmerAnim.setValue(-1);
    shimmerPlayedRef.current = false;
  }, [shimmerAnim]);

  useEffect(() => {
    if (isEmptyWordlist) {
      startShimmer();
    } else {
      stopShimmer();
    }
    return stopShimmer;
  }, [isEmptyWordlist, startShimmer, stopShimmer]);

  useFocusEffect(
    React.useCallback(() => {
      if (isEmptyWordlist) {
        stopShimmer();
        startShimmer();
      }
      return undefined;
    }, [isEmptyWordlist, startShimmer, stopShimmer]),
  );

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: () => wordlistsApi.deleteWordlist(item.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wordlists"] });
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
          <View
            style={styles.flagContainer}
            accessibilityLabel={`Language: ${t(`dashboard.languages.${language.name.toLowerCase()}`)}`}
          >
            <Text style={styles.languageFlag}>{language.flag}</Text>
          </View>
          <View style={styles.cardTitleContainer}>
            <Text
              style={styles.wordlistTitle}
              numberOfLines={1}
              ellipsizeMode="tail"
              accessibilityRole="header"
              accessibilityLabel={item.name}
            >
              {item.name}
            </Text>
            <Text style={styles.wordCountText}>
              {t("wordlistItem.wordCount", { count: item.wordsCount ?? 0 })}
            </Text>
          </View>
          <TouchableOpacity
            style={styles.headerAddButton}
            onPress={() => {
              if (onAddWords) {
                onAddWords(item);
              } else if (onPressed) {
                onPressed();
              }
            }}
            accessibilityRole="button"
            accessibilityLabel={t("wordlistItem.addWords", "Add words")}
            accessibilityHint="Add words to this wordlist"
          >
            <Ionicons
              name="add"
              size={responsive.getValueForSize(21, 22, 23, 24)}
              color="#22C55E"
            />
          </TouchableOpacity>
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
                size={responsive.getValueForSize(19, 20, 21, 22)}
                color={theme.colors.error}
              />
            )}
          </TouchableOpacity>
        </View>

        <View style={styles.progressBar}>
          <View
            style={[styles.progressFill, { width: `${progressPercentage}%` }]}
          />
        </View>

        {isEmptyWordlist ? (
          <View style={styles.emptyActionsContainer}>
            <Text style={styles.emptyActionsHelper}>
              {t("wordlistItem.emptyHelper")}
            </Text>
            <TouchableOpacity
              style={styles.emptyActionButton}
              onLayout={(event) => {
                setButtonWidth(event.nativeEvent.layout.width);
              }}
              onPress={() => {
                if (onAddWords) {
                  onAddWords(item);
                } else if (onPressed) {
                  onPressed();
                }
              }}
              accessibilityRole="button"
              accessibilityLabel={t("dashboard.stats.readyState.cta")}
            >
              {buttonWidth > 0 && (
                <Animated.View
                  pointerEvents="none"
                  style={[
                    styles.emptyActionShimmer,
                    {
                      transform: [
                        {
                          translateX: shimmerAnim.interpolate({
                            inputRange: [-1, 1],
                            outputRange: [
                              -buttonWidth * 0.6,
                              buttonWidth * 0.6,
                            ],
                          }),
                        },
                      ],
                    },
                  ]}
                >
                  <LinearGradient
                    colors={[
                      "rgba(255,255,255,0)",
                      "rgba(255,255,255,0.2)",
                      "rgba(255,255,255,0)",
                    ]}
                    start={{ x: 0, y: 0 }}
                    end={{ x: 1, y: 0 }}
                    style={styles.emptyActionShimmerGradient}
                  />
                </Animated.View>
              )}
              <Ionicons
                name="add"
                size={18}
                color={theme.colors.text.inverse}
              />
              <Text style={styles.emptyActionButtonText} numberOfLines={1}>
                {t("dashboard.stats.readyState.cta")}
              </Text>
            </TouchableOpacity>
          </View>
        ) : (
          <View style={styles.actionButtonsContainer}>
            <TouchableOpacity
              style={styles.actionButton}
              onPress={handleQuizStart}
              accessibilityRole="button"
              accessibilityLabel={t("wordlistItem.quiz")}
              accessibilityHint="Start quiz session"
            >
              <View style={[styles.actionIconWrapper, styles.quizIconBg]}>
                <MaterialIcons
                  name="lightbulb-outline"
                  size={responsive.getValueForSize(22, 23, 24, 26)}
                  color="#000000"
                />
              </View>
              <Text style={styles.actionButtonText} numberOfLines={1} ellipsizeMode="tail">
                {t("wordlistItem.quiz")}
              </Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.actionButton}
              onPress={handlePractice}
              accessibilityRole="button"
              accessibilityLabel={t("wordlistItem.flashcards")}
              accessibilityHint="Practice with flashcards"
            >
              <View style={[styles.actionIconWrapper, styles.flashcardsIconBg]}>
                <MaterialIcons
                  name="menu-book"
                  size={responsive.getValueForSize(22, 23, 24, 26)}
                  color="#2196F3"
                />
              </View>
              <Text style={styles.actionButtonText} numberOfLines={1} ellipsizeMode="tail">
                {t("wordlistItem.cards", "Cards")}
              </Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.actionButton}
              onPress={handleChatStart}
              accessibilityRole="button"
              accessibilityLabel={t("wordlistItem.speak", "Speak")}
              accessibilityHint={t(
                "wordlistItem.speakHint",
                "Practice pronunciation",
              )}
            >
              <View style={[styles.actionIconWrapper, styles.speakIconBg]}>
                <MaterialIcons
                  name="mic"
                  size={responsive.getValueForSize(22, 23, 24, 26)}
                  color="#FF8533"
                />
              </View>
              <Text style={styles.actionButtonText} numberOfLines={1} ellipsizeMode="tail">
                {t("wordlistItem.speak", "Speak")}
              </Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.actionButton}
              onPress={handleAnalytics}
              accessibilityRole="button"
              accessibilityLabel={t("wordlistItem.stats", "Stats")}
              accessibilityHint={
                isPremium
                  ? "View detailed learning analytics"
                  : "Premium feature - tap to upgrade"
              }
            >
              <View style={[styles.actionIconWrapper, styles.statsIconBg]}>
                <MaterialIcons
                  name="bar-chart"
                  size={responsive.getValueForSize(22, 23, 24, 26)}
                  color="#000000"
                />
              </View>
              <Text style={styles.actionButtonText} numberOfLines={1} ellipsizeMode="tail">
                {t("wordlistItem.stats", "Stats")}
              </Text>
            </TouchableOpacity>
          </View>
        )}
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
          ? "#FFFFFF"
          : theme.colors.background.elevated,
      borderRadius: responsive.getValueForSize(20, 22, 24, 26),
      padding: responsive.getValueForSize(16, 18, 20, 22),
      marginHorizontal: responsive.spacing.horizontal,
      marginBottom: responsive.spacing.vertical,
      // Subtle shadow for premium elevated feel
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 3 },
      shadowOpacity: theme.mode === "light" ? 0.06 : 0.2,
      shadowRadius: 10,
      elevation: theme.mode === "light" ? 3 : 8,
    },
    cardHeader: {
      flexDirection: "row",
      alignItems: "flex-start",
      justifyContent: "space-between",
      marginBottom: responsive.getValueForSize(10, 11, 12, 13),
    },
    flagContainer: {
      width: responsive.getValueForSize(46, 50, 54, 58),
      height: responsive.getValueForSize(46, 50, 54, 58),
      borderRadius: responsive.getValueForSize(10, 11, 12, 13),
      backgroundColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.03)"
          : theme.colors.background.subtle,
      alignItems: "center",
      justifyContent: "center",
      marginRight: responsive.spacing.elementSpacing,
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.06)"
          : theme.colors.ui.border,
    },
    languageFlag: {
      fontSize: responsive.getValueForSize(26, 28, 30, 32),
    },
    cardTitleContainer: {
      flex: 1,
      marginRight: responsive.getValueForSize(8, 10, 12, 14),
    },
    wordlistTitle: {
      fontSize: responsive.getValueForSize(18, 20, 22, 24),
      fontWeight: "700",
      color: theme.colors.text.primary,
      marginBottom: responsive.getValueForSize(2, 3, 4, 4),
      letterSpacing: -0.3,
    },
    wordCountText: {
      fontSize: responsive.getValueForSize(12.5, 13.5, 14.5, 15.5),
      color: theme.colors.text.secondary,
      fontWeight: "500",
      letterSpacing: 0.1,
      opacity: 0.9,
    },
    // public badge styles removed
    headerAddButton: {
      width: responsive.getValueForSize(34, 36, 38, 40),
      height: responsive.getValueForSize(34, 36, 38, 40),
      borderRadius: 999,
      alignItems: "center",
      justifyContent: "center",
      backgroundColor:
        theme.mode === "light"
          ? "rgba(34, 197, 94, 0.12)"
          : "rgba(34, 197, 94, 0.2)",
      marginRight: responsive.getValueForSize(8, 9, 10, 11),
    },
    headerMoreButton: {
      width: responsive.getValueForSize(34, 36, 38, 40),
      height: responsive.getValueForSize(34, 36, 38, 40),
      borderRadius: 999,
      alignItems: "center",
      justifyContent: "center",
      backgroundColor:
        theme.mode === "light"
          ? "rgba(255, 133, 51, 0.15)"
          : "rgba(255, 133, 51, 0.25)",
    },
    actionButtonsContainer: {
      flexDirection: "row",
      gap: responsive.getValueForSize(5, 6, 7, 8),
      marginTop: 0,
    },
    emptyActionsContainer: {
      marginTop: responsive.spacing.elementSpacing,
      alignItems: "center",
    },
    emptyActionsHelper: {
      fontSize: responsive.getValueForSize(11, 12, 13, 13),
      color: theme.colors.text.secondary,
      marginBottom: 10,
      textAlign: "center",
    },
    emptyActionButton: {
      flexDirection: "row",
      alignItems: "center",
      gap: 8,
      paddingHorizontal: 28,
      paddingVertical: 14,
      borderRadius: 999,
      backgroundColor: theme.colors.primary,
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 8 },
      shadowOpacity: 0.16,
      shadowRadius: 14,
      elevation: 10,
      overflow: "hidden",
    },
    emptyActionShimmer: {
      position: "absolute",
      left: 0,
      top: 0,
      bottom: 0,
      width: "40%",
      transform: [{ skewX: "-30deg" }],
    },
    emptyActionShimmerGradient: {
      flex: 1,
    },
    emptyActionButtonText: {
      fontSize: responsive.getValueForSize(13, 14, 15, 16),
      fontWeight: "600",
      color: theme.colors.text.inverse,
    },
    actionButton: {
      flex: 1,
      flexBasis: 0,
      flexGrow: 1,
      flexShrink: 1,
      flexDirection: "column",
      alignItems: "center",
      justifyContent: "center",
      paddingVertical: responsive.getValueForSize(10, 11, 12, 13),
      paddingHorizontal: responsive.getValueForSize(2, 3, 4, 5),
      borderRadius: responsive.getValueForSize(14, 15, 16, 17),
      backgroundColor:
        theme.mode === "light"
          ? "#FFFFFF"
          : theme.colors.background.elevated,
      gap: responsive.getValueForSize(4, 5, 6, 7),
      minHeight: responsive.getValueForSize(75, 78, 82, 85),
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.06)"
          : theme.colors.ui.border,
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 1.5 },
      shadowOpacity: theme.mode === "light" ? 0.04 : 0.1,
      shadowRadius: 5,
      elevation: 1.5,
    },
    actionIconWrapper: {
      width: responsive.getValueForSize(40, 42, 44, 46),
      height: responsive.getValueForSize(40, 42, 44, 46),
      borderRadius: responsive.getValueForSize(20, 21, 22, 23),
      alignItems: "center",
      justifyContent: "center",
    },
    quizIconBg: {
      backgroundColor:
        theme.mode === "light"
          ? "rgba(255, 193, 7, 0.15)"
          : "rgba(255, 193, 7, 0.2)",
    },
    flashcardsIconBg: {
      backgroundColor:
        theme.mode === "light"
          ? "rgba(33, 150, 243, 0.15)"
          : "rgba(33, 150, 243, 0.2)",
    },
    speakIconBg: {
      backgroundColor:
        theme.mode === "light"
          ? "rgba(255, 133, 51, 0.15)"
          : "rgba(255, 133, 51, 0.2)",
    },
    statsIconBg: {
      backgroundColor:
        theme.mode === "light"
          ? "rgba(255, 193, 7, 0.15)"
          : "rgba(255, 193, 7, 0.2)",
    },
    actionButtonText: {
      fontSize: responsive.getValueForSize(10.5, 11.5, 12.5, 13.5),
      color: theme.colors.text.primary,
      fontWeight: "600",
      textAlign: "center",
      width: "100%",
      letterSpacing: -0.2,
    },
    progressBar: {
      height: responsive.getValueForSize(5, 6, 7, 7),
      backgroundColor:
        theme.mode === "light"
          ? "rgba(255, 133, 51, 0.12)"
          : theme.colors.ui.divider,
      borderRadius: 4,
      overflow: "hidden",
      marginBottom: responsive.getValueForSize(14, 16, 18, 20),
    },
    progressFill: {
      height: "100%",
      backgroundColor: theme.colors.primary,
      borderRadius: 4,
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
