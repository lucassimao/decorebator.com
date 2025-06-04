import { Wordlist } from "@/api/wordlists";
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

type WordlistItemProps = {
  item: Wordlist;
  onQuizStart?: (wordlist: Wordlist) => void;
  onPressed?: () => void;
};

const WordlistItem: React.FC<WordlistItemProps> = ({
  item,
  onQuizStart,
  onPressed,
}) => {
  const queryClient = useQueryClient();
  const [showMenu, setShowMenu] = useState(false);
  const router = useRouter();
  const { t } = useTranslation();

  const language = LANGUAGES.find((l) => item.languageCode === l.code)!;

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: () => wordlistsApi.deleteWordlist(item.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wordlists"] });
      queryClient.invalidateQueries({ queryKey: ["dashboardStats"] });
      ``;
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
      // router.push(`/analytics?wordlistId=${item.id}`);
    }
  };

  const handleEdit = () => {
    setShowMenu(false);
    onPressed?.();
  };

  const progressPercentage = (item.wordsLearnedCount / item.wordsCount) * 100;

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
            <Text style={styles.wordlistTitle} numberOfLines={1}>
              {item.name}
            </Text>
            {item.description && (
              <Text style={styles.wordlistDescription} numberOfLines={2}>
                {item.description}
              </Text>
            )}
          </View>

          {/* Action Buttons */}
          <View style={styles.actionButtons}>
            <TouchableOpacity
              style={styles.actionButton}
              onPress={handleQuizStart}
              hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
            >
              <MaterialIcons
                name="play-circle-filled"
                size={28}
                color="#4CAF50"
              />
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.actionButton}
              onPress={() => setShowMenu(true)}
              hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
            >
              <MaterialIcons name="more-vert" size={24} color="#636E72" />
            </TouchableOpacity>
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
  actionButtons: {
    flexDirection: "row",
    alignItems: "center",
    gap: 4,
  },
  actionButton: {
    padding: 4,
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
});
