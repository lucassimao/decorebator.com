import React, { useState, useCallback } from "react";
import { useLocalSearchParams, useRouter, Stack } from "expo-router";
import {
  View,
  StyleSheet,
  Text,
  TouchableOpacity,
  SafeAreaView,
  StatusBar,
  FlatList,
  Alert,
} from "react-native";
import { MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import { useQuery } from "@tanstack/react-query";
import { LoadingWithTimeout } from "@/components/LoadingWithTimeout";
import * as wordlistsApi from "@/api/wordlists";
import { Word } from "@/api/wordlists";

const MAX_SELECTED_WORDS = 10;

const WordSelectionScreen: React.FC = () => {
  const { wordlistId, wordlistName } = useLocalSearchParams<{
    wordlistId: string;
    wordlistName: string;
  }>();
  const router = useRouter();
  const { t } = useTranslation();
  const { theme, responsive } = useTheme();
  const styles = createStyles(theme, responsive);

  // State for word selection
  const [selectedWordIds, setSelectedWordIds] = useState<Set<number>>(
    new Set(),
  );
  const [loading, setLoading] = useState(false);

  // Fetch words from the wordlist
  const {
    data: words,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ["words", wordlistId, "withDefinitions"],
    queryFn: () => wordlistsApi.getWords(Number(wordlistId), true),
    enabled: !!wordlistId,
    retry: 2,
  });

  // Filter words that have definitions (words without definitions can't be used for chat)
  const wordsWithDefinitions =
    words?.filter((word) => word.processingStatus === "completed") || [];

  const handleWordToggle = useCallback(
    (wordId: number) => {
      setSelectedWordIds((prev) => {
        const newSet = new Set(prev);

        if (newSet.has(wordId)) {
          newSet.delete(wordId);
        } else {
          if (newSet.size >= MAX_SELECTED_WORDS) {
            Alert.alert(
              t("wordSelection.maxWordsTitle", "Maximum Words Reached"),
              t(
                "wordSelection.maxWordsMessage",
                "You can only select up to {{count}} words for chat practice.",
                { count: MAX_SELECTED_WORDS },
              ),
              [{ text: t("common.ok", "OK") }],
            );
            return prev;
          }
          newSet.add(wordId);
        }

        return newSet;
      });
    },
    [t],
  );

  const handleSelectAll = useCallback(() => {
    const wordsToSelect = wordsWithDefinitions.slice(0, MAX_SELECTED_WORDS);
    setSelectedWordIds(new Set(wordsToSelect.map((word) => word.id)));
  }, [wordsWithDefinitions]);

  const handleClearAll = useCallback(() => {
    setSelectedWordIds(new Set());
  }, []);

  const handleContinueToChat = useCallback(async () => {
    if (selectedWordIds.size === 0) {
      Alert.alert(
        t("wordSelection.noWordsTitle", "No Words Selected"),
        t(
          "wordSelection.noWordsMessage",
          "Please select at least one word to continue.",
        ),
        [{ text: t("common.ok", "OK") }],
      );
      return;
    }

    setLoading(true);

    try {
      // Navigate to realtime chat with selected word IDs
      const selectedIds = Array.from(selectedWordIds).join(",");
      router.push(
        `/realtime-chat?wordlistId=${wordlistId}&wordlistName=${encodeURIComponent(wordlistName || "")}&selectedWordIds=${selectedIds}`,
      );
    } catch (error) {
      console.error("Failed to navigate to chat:", error);
      Alert.alert(
        t("common.error", "Error"),
        t(
          "wordSelection.navigationError",
          "Failed to start chat session. Please try again.",
        ),
        [{ text: t("common.ok", "OK") }],
      );
    } finally {
      setLoading(false);
    }
  }, [selectedWordIds, wordlistId, wordlistName, router, t]);

  const renderWordItem = ({ item }: { item: Word }) => {
    const isSelected = selectedWordIds.has(item.id);

    return (
      <TouchableOpacity
        style={[styles.wordItem, isSelected && styles.wordItemSelected]}
        onPress={() => handleWordToggle(item.id)}
        activeOpacity={0.7}
        accessibilityRole="checkbox"
        accessibilityState={{ checked: isSelected }}
        accessibilityLabel={`${item.name}${isSelected ? t("wordSelection.selected", " - selected") : ""}`}
      >
        <View style={styles.wordItemContent}>
          <View style={styles.wordInfo}>
            <Text
              style={[styles.wordName, isSelected && styles.wordNameSelected]}
            >
              {item.name}
            </Text>
            {item.pronunciation && (
              <Text
                style={[
                  styles.wordPronunciation,
                  isSelected && styles.wordPronunciationSelected,
                ]}
              >
                {item.pronunciation}
              </Text>
            )}
          </View>
          <View
            style={[styles.checkbox, isSelected && styles.checkboxSelected]}
          >
            {isSelected && (
              <MaterialIcons
                name="check"
                size={responsive.getValueForSize(16, 18, 20, 22)}
                color={theme.colors.text.inverse}
              />
            )}
          </View>
        </View>
      </TouchableOpacity>
    );
  };

  const headerTitle = t("wordSelection.title", "Select Words for Chat");

  if (isLoading) {
    return (
      <SafeAreaView style={styles.container}>
        <Stack.Screen options={{ title: headerTitle }} />
        <StatusBar
          barStyle={theme.mode === "light" ? "dark-content" : "light-content"}
        />
        <LoadingWithTimeout
          isLoading={isLoading}
          hasTimeout={false}
          loadingMessage={t("wordSelection.loading", "Loading words...")}
          onRetry={() => refetch()}
          onGoBack={() => router.back()}
        />
      </SafeAreaView>
    );
  }

  if (error) {
    return (
      <SafeAreaView style={styles.container}>
        <Stack.Screen options={{ title: headerTitle }} />
        <StatusBar
          barStyle={theme.mode === "light" ? "dark-content" : "light-content"}
        />
        <View style={styles.errorContainer}>
          <MaterialIcons
            name="error-outline"
            size={responsive.getValueForSize(48, 52, 56, 60)}
            color={theme.colors.error}
          />
          <Text style={styles.errorTitle}>{t("common.error", "Error")}</Text>
          <Text style={styles.errorMessage}>
            {t(
              "wordSelection.loadError",
              "Failed to load words. Please try again.",
            )}
          </Text>
          <TouchableOpacity
            style={styles.retryButton}
            onPress={() => refetch()}
          >
            <Text style={styles.retryButtonText}>
              {t("common.retry", "Retry")}
            </Text>
          </TouchableOpacity>
        </View>
      </SafeAreaView>
    );
  }

  if (wordsWithDefinitions.length === 0) {
    return (
      <SafeAreaView style={styles.container}>
        <Stack.Screen options={{ title: headerTitle }} />
        <StatusBar
          barStyle={theme.mode === "light" ? "dark-content" : "light-content"}
        />
        <View style={styles.emptyContainer}>
          <MaterialIcons
            name="chat-bubble-outline"
            size={responsive.getValueForSize(64, 68, 72, 76)}
            color={theme.colors.text.secondary}
          />
          <Text style={styles.emptyTitle}>
            {t("wordSelection.noWordsAvailableTitle", "No Words Available")}
          </Text>
          <Text style={styles.emptyMessage}>
            {t(
              "wordSelection.noWordsAvailableMessage",
              "This wordlist doesn't have any words with definitions ready for chat practice.",
            )}
          </Text>
          <TouchableOpacity
            style={styles.backButton}
            onPress={() => router.back()}
          >
            <Text style={styles.backButtonText}>
              {t("common.goBack", "Go Back")}
            </Text>
          </TouchableOpacity>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <Stack.Screen options={{ title: headerTitle }} />
      <StatusBar
        barStyle={theme.mode === "light" ? "dark-content" : "light-content"}
      />

      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.subtitle}>
          {t(
            "wordSelection.subtitle",
            "Choose up to {{count}} words to practice",
            { count: MAX_SELECTED_WORDS },
          )}
        </Text>
        <View style={styles.actionButtons}>
          <TouchableOpacity
            style={styles.actionButton}
            onPress={handleSelectAll}
            disabled={wordsWithDefinitions.length === 0}
          >
            <Text style={styles.actionButtonText}>
              {t("wordSelection.selectAll", "Select All")}
            </Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[styles.actionButton, styles.clearButton]}
            onPress={handleClearAll}
            disabled={selectedWordIds.size === 0}
          >
            <Text style={[styles.actionButtonText, styles.clearButtonText]}>
              {t("wordSelection.clearAll", "Clear All")}
            </Text>
          </TouchableOpacity>
        </View>
      </View>

      {/* Words List */}
      <FlatList
        data={wordsWithDefinitions}
        renderItem={renderWordItem}
        keyExtractor={(item) => item.id.toString()}
        style={styles.wordsList}
        contentContainerStyle={styles.wordsListContent}
        showsVerticalScrollIndicator={false}
        initialNumToRender={20}
        maxToRenderPerBatch={20}
        windowSize={10}
      />

      {/* Bottom Controls */}
      <View style={styles.bottomControls}>
        <View style={styles.selectionIndicator}>
          <Text style={styles.selectionCount}>
            {t(
              "wordSelection.wordsSelected",
              "{{count}}/{{max}} words selected",
              {
                count: selectedWordIds.size,
                max: MAX_SELECTED_WORDS,
              },
            )}
          </Text>
        </View>
        <TouchableOpacity
          style={[
            styles.continueButton,
            selectedWordIds.size === 0 && styles.continueButtonDisabled,
          ]}
          onPress={handleContinueToChat}
          disabled={selectedWordIds.size === 0 || loading}
          accessibilityRole="button"
          accessibilityLabel={t(
            "wordSelection.continueToChat",
            "Continue to Chat",
          )}
        >
          {loading ? (
            <MaterialIcons
              name="hourglass-empty"
              size={responsive.getValueForSize(20, 22, 24, 26)}
              color={theme.colors.text.inverse}
            />
          ) : (
            <MaterialIcons
              name="chat"
              size={responsive.getValueForSize(20, 22, 24, 26)}
              color={
                selectedWordIds.size > 0
                  ? theme.colors.text.inverse
                  : theme.colors.text.disabled
              }
            />
          )}
          <Text
            style={[
              styles.continueButtonText,
              selectedWordIds.size === 0 && styles.continueButtonTextDisabled,
            ]}
          >
            {t("wordSelection.continueToChat", "Continue to Chat")}
          </Text>
        </TouchableOpacity>
      </View>
    </SafeAreaView>
  );
};

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  responsive: ReturnType<typeof useTheme>["responsive"],
) => {
  return StyleSheet.create({
    container: {
      flex: 1,
      backgroundColor: theme.colors.background.default,
    },
    header: {
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.vertical,
      backgroundColor: theme.colors.background.surface,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.ui.divider,
    },
    subtitle: {
      fontSize: responsive.getScaledFont("body"),
      color: theme.colors.text.secondary,
      textAlign: "center",
      marginBottom: responsive.spacing.elementSpacing,
    },
    actionButtons: {
      flexDirection: "row",
      justifyContent: "space-evenly",
    },
    actionButton: {
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.elementSpacing / 2,
      borderRadius: theme.borderRadius.md,
      backgroundColor: theme.colors.primary,
      minWidth: responsive.getValueForSize(80, 90, 100, 110),
    },
    clearButton: {
      backgroundColor: theme.colors.background.elevated,
      borderWidth: 1,
      borderColor: theme.colors.border.medium,
    },
    actionButtonText: {
      fontSize: responsive.getScaledFont("body"),
      fontWeight: "600",
      color: theme.colors.text.inverse,
      textAlign: "center",
    },
    clearButtonText: {
      color: theme.colors.text.secondary,
    },
    wordsList: {
      flex: 1,
    },
    wordsListContent: {
      paddingHorizontal: responsive.spacing.horizontal,
      paddingTop: responsive.spacing.vertical / 2,
      paddingBottom: responsive.spacing.vertical,
    },
    wordItem: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: theme.borderRadius.md,
      marginBottom: responsive.spacing.elementSpacing / 2,
      ...theme.shadows.sm,
    },
    wordItemSelected: {
      backgroundColor: theme.colors.primary,
      borderColor: theme.colors.primary,
      borderWidth: 2,
    },
    wordItemContent: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.elementSpacing,
    },
    wordInfo: {
      flex: 1,
    },
    wordName: {
      fontSize: responsive.getScaledFont("headline"),
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: responsive.spacing.elementSpacing / 4,
    },
    wordNameSelected: {
      color: theme.colors.text.inverse,
    },
    wordPronunciation: {
      fontSize: responsive.getScaledFont("body"),
      color: theme.colors.text.secondary,
    },
    wordPronunciationSelected: {
      color: theme.colors.text.inverse,
      opacity: 0.8,
    },
    checkbox: {
      width: responsive.getValueForSize(24, 26, 28, 30),
      height: responsive.getValueForSize(24, 26, 28, 30),
      borderRadius: responsive.getValueForSize(12, 13, 14, 15),
      borderWidth: 2,
      borderColor: theme.colors.border.medium,
      backgroundColor: theme.colors.background.default,
      justifyContent: "center",
      alignItems: "center",
    },
    checkboxSelected: {
      backgroundColor: theme.colors.text.inverse,
      borderColor: theme.colors.text.inverse,
    },
    bottomControls: {
      backgroundColor: theme.colors.background.surface,
      borderTopWidth: 1,
      borderTopColor: theme.colors.ui.divider,
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.vertical,
    },
    selectionIndicator: {
      alignItems: "center",
      marginBottom: responsive.spacing.elementSpacing,
    },
    selectionCount: {
      fontSize: responsive.getScaledFont("body"),
      color: theme.colors.text.secondary,
      fontWeight: "500",
    },
    continueButton: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      backgroundColor: theme.colors.primary,
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.vertical,
      borderRadius: theme.borderRadius.lg,
      ...theme.shadows.md,
      gap: responsive.spacing.elementSpacing / 2,
    },
    continueButtonDisabled: {
      backgroundColor: theme.colors.ui.disabled,
    },
    continueButtonText: {
      fontSize: responsive.getScaledFont("headline"),
      fontWeight: "600",
      color: theme.colors.text.inverse,
    },
    continueButtonTextDisabled: {
      color: theme.colors.text.disabled,
    },
    errorContainer: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      paddingHorizontal: responsive.spacing.horizontal,
    },
    errorTitle: {
      fontSize: responsive.getScaledFont("title"),
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginTop: responsive.spacing.vertical,
      marginBottom: responsive.spacing.elementSpacing,
    },
    errorMessage: {
      fontSize: responsive.getScaledFont("body"),
      color: theme.colors.text.secondary,
      textAlign: "center",
      marginBottom: responsive.spacing.vertical,
    },
    retryButton: {
      backgroundColor: theme.colors.primary,
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.elementSpacing,
      borderRadius: theme.borderRadius.md,
    },
    retryButtonText: {
      fontSize: responsive.getScaledFont("body"),
      fontWeight: "600",
      color: theme.colors.text.inverse,
    },
    emptyContainer: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      paddingHorizontal: responsive.spacing.horizontal,
    },
    emptyTitle: {
      fontSize: responsive.getScaledFont("title"),
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginTop: responsive.spacing.vertical,
      marginBottom: responsive.spacing.elementSpacing,
      textAlign: "center",
    },
    emptyMessage: {
      fontSize: responsive.getScaledFont("body"),
      color: theme.colors.text.secondary,
      textAlign: "center",
      marginBottom: responsive.spacing.vertical,
      lineHeight: responsive.getScaledFont("body") * 1.4,
    },
    backButton: {
      backgroundColor: theme.colors.background.elevated,
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.elementSpacing,
      borderRadius: theme.borderRadius.md,
      borderWidth: 1,
      borderColor: theme.colors.border.medium,
    },
    backButtonText: {
      fontSize: responsive.getScaledFont("body"),
      fontWeight: "600",
      color: theme.colors.text.secondary,
    },
  });
};

export default WordSelectionScreen;
