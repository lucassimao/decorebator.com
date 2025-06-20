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
import { useUserInfo } from "@/hooks/users";
import { LinearGradient } from "expo-linear-gradient";

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
  const { isPremium } = useUserInfo();
  const language = LANGUAGES.find((l) => item.languageCode === l.code)!;

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

    router.push(`/practice?wordlistId=${item.id}&wordlistName=${item.name}`);
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
      >
        <View style={styles.cardHeader}>
          <Text style={styles.languageFlag}>{language.flag}</Text>
          <View style={styles.cardTitleContainer}>
            <Text style={styles.wordlistTitle}>{item.name}</Text>
            {item.description && (
              <Text style={styles.wordlistDescription} numberOfLines={2}>
                {item.description}
              </Text>
            )}
          </View>
        </View>

        <View style={styles.cardStats}>
          <View style={styles.cardStat}>
            <MaterialIcons name="library-books" size={16} color="#636E72" />
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
              <MaterialIcons name="school" size={16} color="#636E72" />
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
          >
            <MaterialIcons name="bar-chart" size={20} color="#FFD700" />
            <Text style={styles.actionButtonText}>
              {t("wordlistItem.analytics")}
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.actionButtonLarge}
            onPress={handlePractice}
          >
            <MaterialIcons name="style" size={20} color="#2196F3" />
            <Text style={styles.actionButtonText}>
              {t("wordlistItem.practice")}
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.actionButtonLarge}
            onPress={handleQuizStart}
          >
            <MaterialIcons
              name="play-circle-filled"
              size={20}
              color="#4CAF50"
            />
            <Text style={styles.actionButtonText}>
              {t("wordlistItem.quiz")}
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.actionButtonLarge}
            onPress={() => setShowMenu(true)}
          >
            <MaterialIcons name="more-horiz" size={20} color="#636E72" />
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
      >
        <TouchableWithoutFeedback onPress={() => setShowMenu(false)}>
          <View style={styles.modalOverlay}>
            <TouchableWithoutFeedback>
              <View style={styles.menuContainer}>
                <Text style={styles.menuTitle}>{item.name}</Text>

                <TouchableOpacity
                  style={styles.menuItem}
                  onPress={() => {
                    setShowMenu(false);
                    router.push(`/analytics?wordlistId=${item.id}`);
                  }}
                >
                  <MaterialIcons name="analytics" size={24} color="#FFD700" />
                  <Text style={styles.menuItemText}>
                    {t("wordlistItem.viewAnalytics")}
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity
                  style={styles.menuItem}
                  onPress={handlePractice}
                >
                  <MaterialIcons name="style" size={24} color="#2196F3" />
                  <Text style={styles.menuItemText}>
                    {t("wordlistItem.practiceFlashcards")}
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity
                  style={styles.menuItem}
                  onPress={handleQuizStart}
                >
                  <MaterialIcons name="quiz" size={24} color="#4CAF50" />
                  <Text style={styles.menuItemText}>
                    {t("wordlistItem.startQuiz")}
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity style={styles.menuItem} onPress={handleEdit}>
                  <MaterialIcons name="edit" size={24} color="#FF7B54" />
                  <Text style={styles.menuItemText}>
                    {t("wordlistItem.editWordlist")}
                  </Text>
                </TouchableOpacity>

                <View style={styles.menuDivider} />

                <TouchableOpacity
                  style={[styles.menuItem, styles.deleteMenuItem]}
                  onPress={handleDelete}
                  disabled={deleteMutation.isPending}
                >
                  {deleteMutation.isPending ? (
                    <ActivityIndicator size="small" color="#FF6B6B" />
                  ) : (
                    <Ionicons name="trash-outline" size={24} color="#FF6B6B" />
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
      >
        <TouchableWithoutFeedback onPress={() => setShowPremiumModal(false)}>
          <View style={styles.modalOverlay}>
            <TouchableWithoutFeedback>
              <View style={styles.premiumModalContainer}>
                <LinearGradient
                  colors={["#FFD700", "#FFA500"]}
                  style={styles.premiumGradient}
                  start={{ x: 0, y: 0 }}
                  end={{ x: 1, y: 0 }}
                >
                  <View style={styles.premiumContent}>
                    <View style={styles.premiumIconContainer}>
                      <MaterialIcons name="analytics" size={32} color="#FFF" />
                    </View>
                    <Text style={styles.premiumTitle}>
                      {t("dashboard.stats.premium.title")}
                    </Text>
                    <Text style={styles.premiumSubtitle}>
                      {t("dashboard.stats.premium.subtitle")}
                    </Text>

                    <View style={styles.premiumButtons}>
                      <TouchableOpacity
                        style={styles.upgradeButton}
                        onPress={() => {
                          setShowPremiumModal(false);
                          onUpgradePress?.();
                        }}
                      >
                        <Text style={styles.upgradeButtonText}>
                          {t("settings.subscription.upgradeButton")}
                        </Text>
                      </TouchableOpacity>

                      <TouchableOpacity
                        style={styles.cancelButton}
                        onPress={() => setShowPremiumModal(false)}
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

const styles = StyleSheet.create({
  wordlistCard: {
    backgroundColor: "rgba(255, 255, 255, 0.95)",
    borderRadius: 16,
    padding: 16,
    marginHorizontal: 20,
    marginBottom: 12,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.08,
    shadowRadius: 8,
    elevation: 3,
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
    color: "#2D3436",
    marginBottom: 4,
  },
  wordlistDescription: {
    fontSize: 14,
    color: "#636E72",
    lineHeight: 20,
  },
  actionButtonsRow: {
    flexDirection: "row",
    marginTop: 12,
    gap: 8,
    paddingTop: 8,
    borderTopWidth: 1,
    borderTopColor: "#F0F0F0",
  },
  actionButtonLarge: {
    flex: 1,
    flexDirection: "column",
    alignItems: "center",
    paddingVertical: 12,
    paddingHorizontal: 8,
    borderRadius: 8,
    backgroundColor: "#FAFAFA",
    gap: 4,
  },
  actionButtonText: {
    fontSize: 12,
    color: "#636E72",
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
    color: "#636E72",
  },
  languageName: {
    fontSize: 14,
    color: "#FF7B54",
    fontWeight: "500",
  },
  progressBar: {
    height: 4,
    backgroundColor: "#F0F0F0",
    borderRadius: 2,
    overflow: "hidden",
  },
  progressFill: {
    height: "100%",
    backgroundColor: "#4CAF50",
    borderRadius: 2,
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: "rgba(0, 0, 0, 0.5)",
    justifyContent: "center",
    alignItems: "center",
    padding: 20,
  },
  menuContainer: {
    backgroundColor: "#FFFFFF",
    borderRadius: 16,
    padding: 8,
    width: "100%",
    maxWidth: 320,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.15,
    shadowRadius: 12,
    elevation: 8,
  },
  menuTitle: {
    fontSize: 18,
    fontWeight: "600",
    color: "#2D3436",
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
    color: "#2D3436",
    flex: 1,
  },
  menuDivider: {
    height: 1,
    backgroundColor: "#F0F0F0",
    marginVertical: 8,
    marginHorizontal: 16,
  },
  deleteMenuItem: {
    marginTop: 4,
  },
  deleteMenuItemText: {
    color: "#FF6B6B",
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
    backgroundColor: "#FFD700",
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
