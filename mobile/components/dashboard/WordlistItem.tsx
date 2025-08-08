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
  Linking,
  Share,
} from "react-native";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { LANGUAGES } from "./CreateWordlistModal";
import * as wordlistsApi from "@/api/wordlists";
import PublishPublicQuizModal from "./PublishPublicQuizModal";
import * as Clipboard from "expo-clipboard";
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
  const [showPublishModal, setShowPublishModal] = useState(false);
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

  const openPublishModal = () => {
    setShowMenu(false);
    setShowPublishModal(true);
  };

  const getPublicUrl = (slug: string) => {
    const appDomain = process.env.EXPO_PUBLIC_APP_DOMAIN?.trim();
    const baseUrl = appDomain && appDomain.startsWith("http") ? appDomain : "http://localhost:3000";
    return `${baseUrl}/q/${slug}`;
  };

  const copyPublicLink = async (slug: string) => {
    const fullUrl = getPublicUrl(slug);
    const copyLabel = t("wordlistItem.copyUrl", "Copy link");
    const okLabel = t("common.ok");
    Alert.alert(
      t("wordlistItem.publishSuccessTitle", "Published"),
      t("wordlistItem.publishReadyToShare", "Your quiz is ready to be shared!"),
      [
        { text: okLabel, style: "default" },
        {
          text: copyLabel,
          onPress: () => Clipboard.setStringAsync(fullUrl),
        },
      ],
    );
  };

  const handlePublishSuccess = async (slug: string) => {
    await copyPublicLink(slug);
    // refresh lists to pull in publicQuizSlug
    queryClient.invalidateQueries({ queryKey: ["wordlists"] });
  };

  const handleViewOnWeb = (slug: string) => {
    const url = getPublicUrl(slug);
    Linking.openURL(url).catch((e) => console.warn("Failed to open URL", e));
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
              numberOfLines={responsive.getValueForSize(1, 2, 2, 2)}
              ellipsizeMode="tail"
              accessibilityRole="header"
              accessibilityLabel={item.name}
            >
              {item.name}
            </Text>
            {item.publicQuizSlug ? (
              <View style={styles.publicBadge} accessibilityLabel={t("wordlistItem.publicBadge", "Public")}>
                <MaterialIcons
                  name="public"
                  size={responsive.getValueForSize(12, 14, 14, 16)}
                  color={theme.colors.text.inverse}
                />
                <Text style={styles.publicBadgeText}>{t("wordlistItem.publicBadge", "Public")}</Text>
              </View>
            ) : null}
          </View>
          <TouchableOpacity
            style={styles.headerMoreButton}
            onPress={() => setShowMenu(true)}
            accessibilityRole="button"
            accessibilityLabel={t("common.more")}
          >
            <MaterialIcons
              name="more-vert"
              size={responsive.getValueForSize(18, 20, 22, 24)}
              color={theme.colors.text.secondary}
            />
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
              size={responsive.getValueForSize(18, 20, 22, 24)}
              color={theme.colors.premium}
            />
            <Text style={styles.actionButtonText} numberOfLines={1}>
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
              size={responsive.getValueForSize(18, 20, 22, 24)}
              color={theme.colors.semantic.info}
            />
            <Text style={styles.actionButtonText} numberOfLines={1}>
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
              size={responsive.getValueForSize(18, 20, 22, 24)}
              color={theme.colors.success}
            />
            <Text style={styles.actionButtonText} numberOfLines={1}>
              {t("wordlistItem.quiz")}
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.actionButtonLarge}
            onPress={() =>
              item.publicQuizSlug
                ? handleViewOnWeb(item.publicQuizSlug)
                : openPublishModal()
            }
            accessibilityRole="button"
            accessibilityLabel={
              item.publicQuizSlug
                ? t("wordlistItem.viewOnWeb", "View on web")
                : t("wordlistItem.publishPublicQuiz", "Publish public quiz")
            }
            accessibilityHint={
              item.publicQuizSlug
                ? "Open this public quiz on the web"
                : "Publish this wordlist as a public quiz"
            }
          >
            <MaterialIcons
              name={item.publicQuizSlug ? "public" : "upload"}
              size={responsive.getValueForSize(18, 20, 22, 24)}
              color={theme.colors.primary}
            />
            <Text style={styles.actionButtonText} numberOfLines={1}>
              {item.publicQuizSlug
                ? t("wordlistItem.public", "Public")
                : t("wordlistItem.publish", "Publish")}
            </Text>
            {item.publicQuizSlug ? (
              <TouchableOpacity
                onPress={() => {
                  const url = getPublicUrl(item.publicQuizSlug!)
                  const cta = t(
                    "wordlistItem.shareCta",
                    "I just published a vocab quiz—can you beat it?",
                  )
                  const message = `${cta} ${url}`
                  Share.share({ message, title: item.name })
                }}
                accessibilityLabel={t("wordlistItem.shareLink", "Share link")}
                style={styles.shareOverlay}
                hitSlop={{ top: 6, right: 6, bottom: 6, left: 6 }}
              >
                <MaterialIcons
                  name="share"
                  size={responsive.getValueForSize(14, 16, 18, 18)}
                  color={theme.colors.text.secondary}
                />
              </TouchableOpacity>
            ) : null}
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
                  numberOfLines={responsive.getValueForSize(2, 2, 3, 3)}
                  ellipsizeMode="tail"
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
                    size={responsive.getValueForSize(20, 24, 26, 28)}
                    color={theme.colors.premium}
                  />
                  <Text style={styles.menuItemText}>
                    {t("wordlistItem.viewAnalytics")}
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity
                  style={styles.menuItem}
                  onPress={openPublishModal}
                  accessibilityRole="button"
                  accessibilityLabel={
                    item.publicQuizSlug
                      ? t("wordlistItem.editPublicQuiz", "Edit public quiz")
                      : t("wordlistItem.publishPublicQuiz", "Publish public quiz")
                  }
                  accessibilityHint={
                    item.publicQuizSlug
                      ? "Update the public quiz settings"
                      : "Publish this wordlist as a shareable quiz"
                  }
                >
                  <MaterialIcons
                    name="public"
                    size={responsive.getValueForSize(20, 24, 26, 28)}
                    color={theme.colors.primary}
                  />
                  <Text style={styles.menuItemText}>
                    {item.publicQuizSlug
                      ? t("wordlistItem.editPublicQuiz", "Edit public quiz")
                      : t("wordlistItem.publishPublicQuiz", "Publish public quiz")}
                  </Text>
                </TouchableOpacity>

                {item.publicQuizSlug ? (
                  <TouchableOpacity
                    style={styles.menuItem}
                    onPress={() => copyPublicLink(item.publicQuizSlug!)}
                    accessibilityRole="button"
                    accessibilityLabel={t("wordlistItem.copyPublicLink", "Copy public link")}
                    accessibilityHint="Copy the shareable link for this public quiz"
                  >
                    <MaterialIcons
                      name="link"
                      size={responsive.getValueForSize(20, 24, 26, 28)}
                      color={theme.colors.success}
                    />
                    <Text style={styles.menuItemText}>
                      {t("wordlistItem.copyPublicLink", "Copy public link")}
                    </Text>
                  </TouchableOpacity>
                ) : null}

                {item.publicQuizSlug ? (
                  <TouchableOpacity
                    style={styles.menuItem}
                    onPress={async () => {
                      await wordlistsApi.unpublishPublicQuiz(item.id)
                      queryClient.invalidateQueries({ queryKey: ["wordlists"] })
                      setShowMenu(false)
                    }}
                    accessibilityRole="button"
                    accessibilityLabel={t("wordlistItem.unpublishPublicQuiz", "Unpublish public quiz")}
                    accessibilityHint="Disable public access to this quiz"
                  >
                    <MaterialIcons
                      name="visibility-off"
                      size={responsive.getValueForSize(20, 24, 26, 28)}
                      color={theme.colors.error}
                    />
                    <Text style={styles.menuItemText}>
                      {t("wordlistItem.unpublishPublicQuiz", "Unpublish public quiz")}
                    </Text>
                  </TouchableOpacity>
                ) : null}

                {item.publicQuizSlug ? (
                  <TouchableOpacity
                    style={styles.menuItem}
                    onPress={() => handleViewOnWeb(item.publicQuizSlug!)}
                    accessibilityRole="button"
                    accessibilityLabel={t("wordlistItem.viewOnWeb", "View on web")}
                    accessibilityHint="Open this public quiz in the browser"
                  >
                    <MaterialIcons
                      name="open-in-new"
                      size={responsive.getValueForSize(20, 24, 26, 28)}
                      color={theme.colors.primary}
                    />
                    <Text style={styles.menuItemText}>
                      {t("wordlistItem.viewOnWeb", "View on web")}
                    </Text>
                  </TouchableOpacity>
                ) : null}

                <TouchableOpacity
                  style={styles.menuItem}
                  onPress={handlePractice}
                  accessibilityRole="button"
                  accessibilityLabel={t("wordlistItem.practiceFlashcards")}
                  accessibilityHint="Practice with interactive flashcards"
                >
                  <MaterialIcons
                    name="style"
                    size={responsive.getValueForSize(20, 24, 26, 28)}
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
                    size={responsive.getValueForSize(20, 24, 26, 28)}
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
                    size={responsive.getValueForSize(20, 24, 26, 28)}
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
                      size={responsive.getValueForSize(20, 24, 26, 28)}
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

      <PublishPublicQuizModal
        visible={showPublishModal}
        wordlistId={item.id}
        defaultTitle={item.name}
        onClose={() => setShowPublishModal(false)}
        onSuccess={handlePublishSuccess}
      />

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
                        size={responsive.getValueForSize(28, 32, 36, 40)}
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

      {/* Publish Public Quiz Modal */}
      <PublishPublicQuizModal
        visible={showPublishModal}
        wordlistId={item.id}
        defaultTitle={item.name}
        onClose={() => setShowPublishModal(false)}
        onSuccess={handlePublishSuccess}
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
    publicBadge: {
      flexDirection: "row",
      alignSelf: "flex-start",
      alignItems: "center",
      gap: 6,
      backgroundColor: theme.colors.primary,
      paddingHorizontal: 8,
      paddingVertical: 4,
      borderRadius: theme.borderRadius.sm,
    },
    publicBadgeText: {
      color: theme.colors.text.inverse,
      fontSize: responsive.getScaledFont("label"),
      fontWeight: "600",
    },
    headerMoreButton: {
      padding: 4,
      borderRadius: theme.borderRadius.sm,
    },
    actionButtonsRow: {
      flexDirection: "row",
      marginTop: responsive.spacing.elementSpacing,
      gap: responsive.spacing.elementSpacing,
      paddingTop: responsive.spacing.elementSpacing / 2,
      borderTopWidth: 1,
      borderTopColor: theme.colors.ui.divider,
    },
    actionButtonLarge: {
      flex: 1,
      flexDirection: "column",
      alignItems: "center",
      justifyContent: "center",
      paddingVertical: responsive.getValueForSize(8, 10, 12, 14),
      paddingHorizontal: responsive.getValueForSize(4, 6, 8, 10),
      borderRadius: theme.borderRadius.sm,
      backgroundColor:
        theme.mode === "light"
          ? "rgba(253, 246, 227, 0.2)" // Very subtle web beige tint
          : theme.colors.background.subtle,
      gap: responsive.spacing.elementSpacing / 4,
      minHeight: responsive.getValueForSize(60, 64, 68, 72),
      position: "relative",
    },
    actionButtonText: {
      fontSize: responsive.getValueForSize(10, 12, 13, 14),
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
    shareOverlay: {
      position: "absolute",
      top: responsive.getValueForSize(6, 6, 6, 8),
      right: responsive.getValueForSize(6, 6, 6, 8),
      padding: 4,
      borderRadius: theme.borderRadius.sm,
      backgroundColor:
        theme.mode === "light" ? "rgba(0,0,0,0.02)" : "rgba(255,255,255,0.06)",
    },
  });
