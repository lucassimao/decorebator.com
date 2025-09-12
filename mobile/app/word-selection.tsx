import React, { useState, useCallback, useMemo } from "react";
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
  TextInput,
} from "react-native";
import { MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import { useQuery } from "@tanstack/react-query";
import { LoadingWithTimeout } from "@/components/LoadingWithTimeout";
import ScreenHeader from "@/components/common/ScreenHeader";
import * as wordlistsApi from "@/api/wordlists";
import { Word } from "@/api/wordlists";

const MAX_SELECTED_WORDS = 5;

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
  const [searchQuery, setSearchQuery] = useState("");

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

  // Words available for selection (memoized)
  const selectableWords = useMemo(
    () => words?.filter((w) => w.processingStatus === "completed") || [],
    [words],
  );

  // Search filter (memoized)
  const filteredWords = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();
    if (!q) return selectableWords;
    return selectableWords.filter((w) => w.name.toLowerCase().includes(q));
  }, [selectableWords, searchQuery]);

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
    const wordsToSelect = selectableWords.slice(0, MAX_SELECTED_WORDS);
    setSelectedWordIds(new Set(wordsToSelect.map((word) => word.id)));
  }, [selectableWords]);

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

  const renderWordItem = useCallback(
    ({ item }: { item: Word }) => {
      const isSelected = selectedWordIds.has(item.id);

      return (
        <TouchableOpacity
          style={[styles.wordCard, isSelected && styles.wordCardSelected]}
          onPress={() => handleWordToggle(item.id)}
          activeOpacity={0.8}
          accessibilityRole="checkbox"
          accessibilityState={{ checked: isSelected }}
          accessibilityLabel={`${item.name}${isSelected ? t("wordSelection.selected", " - selected") : ""}`}
        >
          <View style={styles.wordContent}>
            <Text style={styles.wordTerm}>{item.name}</Text>
          </View>
          <View
            style={[styles.checkbox, isSelected && styles.checkboxSelected]}
          >
            {isSelected && (
              <MaterialIcons
                name="check"
                size={responsive.getValueForSize(18, 20, 22, 24)}
                color={theme.colors.text.inverse}
              />
            )}
          </View>
        </TouchableOpacity>
      );
    },
    [
      selectedWordIds,
      styles,
      responsive,
      theme.colors.text.inverse,
      handleWordToggle,
      t,
    ],
  );

  const headerTitle = t("wordSelection.title", "Select Words for Chat");
  const headerSubtitle = t(
    "wordSelection.subtitle",
    "Choose up to {{count}} words to practice",
    { count: MAX_SELECTED_WORDS },
  );

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

  if (selectableWords.length === 0) {
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
      <Stack.Screen options={{ headerShown: false }} />
      <StatusBar
        barStyle={theme.mode === "light" ? "dark-content" : "light-content"}
      />

      {/* Header (standardized) */}
      <ScreenHeader
        title={headerTitle}
        subtitle={headerSubtitle}
        onBackPress={() => router.back()}
      />

      {/* Header extras: search and actions (contained) */}
      <View style={styles.searchPanel}>
        <Text style={styles.searchHelpText}>
          {t(
            "wordSelection.searchHelp",
            "Search and select up to {{count}} words to practice",
            { count: MAX_SELECTED_WORDS },
          )}
        </Text>
        <View style={styles.searchBox}>
          <MaterialIcons
            name="search"
            size={18}
            color={theme.colors.text.secondary}
          />
          <TextInput
            style={styles.searchInput}
            placeholder={t("wordSelection.search", "Search words")}
            placeholderTextColor={theme.colors.text.tertiary}
            value={searchQuery}
            onChangeText={setSearchQuery}
            autoCorrect={false}
            autoCapitalize="none"
          />
        </View>
        <View style={styles.headerActions}>
          <TouchableOpacity
            style={[styles.chip, styles.chipSpacing]}
            onPress={handleSelectAll}
            disabled={selectableWords.length === 0}
          >
            <Text style={styles.chipText}>
              {t("wordSelection.selectAll", "Select All")}
            </Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[styles.chip, styles.chipSecondary]}
            onPress={handleClearAll}
            disabled={selectedWordIds.size === 0}
          >
            <Text style={[styles.chipText, styles.chipSecondaryText]}>
              {t("wordSelection.clearAll", "Clear All")}
            </Text>
          </TouchableOpacity>
        </View>
      </View>

      {/* Words List */}
      <FlatList
        data={filteredWords}
        renderItem={renderWordItem}
        keyExtractor={(item) => item.id.toString()}
        style={styles.wordsList}
        contentContainerStyle={styles.wordsListContent}
        showsVerticalScrollIndicator={false}
        initialNumToRender={12}
        maxToRenderPerBatch={12}
        updateCellsBatchingPeriod={50}
        windowSize={10}
        removeClippedSubviews
        extraData={selectedWordIds}
        ListEmptyComponent={
          <View style={styles.emptyState}>
            <MaterialIcons
              name="library-books"
              size={48}
              color={theme.colors.text.tertiary}
            />
            <Text style={styles.emptyText}>
              {searchQuery
                ? t("wordSelection.noMatch", "No words match your search")
                : t(
                    "wordSelection.noWordsAvailableTitle",
                    "No Words Available",
                  )}
            </Text>
            {!searchQuery && (
              <Text style={styles.emptySubtext}>
                {t(
                  "wordSelection.noWordsAvailableMessage",
                  "This wordlist doesn't have any words with definitions ready for chat practice.",
                )}
              </Text>
            )}
          </View>
        }
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
    // header removed in favor of ScreenHeader
    searchPanel: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: theme.borderRadius.lg,
      borderWidth: 1,
      borderColor: theme.colors.ui.divider,
      marginHorizontal: responsive.spacing.horizontal,
      marginTop: responsive.spacing.elementSpacing / 2,
      marginBottom: responsive.spacing.elementSpacing,
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.getValueForSize(8, 10, 10, 12),
      ...theme.shadows.sm,
    },
    searchHelpText: {
      fontSize: responsive.getScaledFont("label"),
      color: theme.colors.text.secondary,
      marginBottom: responsive.getValueForSize(6, 8, 8, 10),
      textAlign: "left",
    },
    handle: {
      alignSelf: "center",
      width: 40,
      height: 4,
      borderRadius: 2,
      backgroundColor: theme.colors.border.light,
      marginBottom: responsive.getValueForSize(8, 10, 12, 12),
    },
    title: {
      fontSize: responsive.getScaledFont("headline"),
      fontWeight: "700",
      color: theme.colors.text.primary,
      textAlign: "center",
      marginBottom: responsive.spacing.elementSpacing / 2,
    },
    searchContainer: {
      marginTop: 0,
      gap: responsive.spacing.elementSpacing / 2,
      marginBottom: 0,
    },
    searchBox: {
      flexDirection: "row",
      alignItems: "center",
      backgroundColor: theme.colors.background.elevated,
      borderRadius: theme.borderRadius.md,
      borderWidth: 1,
      borderColor: theme.colors.ui.divider,
      paddingHorizontal: 12,
      paddingVertical: responsive.getValueForSize(4, 6, 6, 8),
      // Tighter vertical separation to bring chips closer to the field
      marginBottom: responsive.getValueForSize(6, 8, 8, 10),
    },
    searchInput: {
      flex: 1,
      marginLeft: 8,
      color: theme.colors.text.primary,
      fontSize: responsive.getScaledFont("label"),
      lineHeight: Math.round(responsive.getScaledFont("label") * 1.2),
    },
    headerActions: {
      flexDirection: "row",
      justifyContent: "flex-start",
      // Slightly tighter spacing between chips
      gap: responsive.spacing.elementSpacing / 2,
      marginTop: responsive.getValueForSize(6, 8, 10, 12),
    },
    chip: {
      paddingHorizontal: responsive.getValueForSize(12, 14, 14, 16),
      paddingVertical: responsive.getValueForSize(8, 9, 10, 10),
      borderRadius: theme.borderRadius.full,
      backgroundColor: theme.colors.primary,
    },
    chipText: {
      color: theme.colors.text.inverse,
      fontWeight: "700",
      fontSize: responsive.getScaledFont("body"),
    },
    chipSecondary: {
      backgroundColor: theme.colors.background.elevated,
      borderWidth: 1,
      borderColor: theme.colors.ui.divider,
    },
    chipSecondaryText: {
      color: theme.colors.text.secondary,
    },
    chipSpacing: {
      marginRight: responsive.getValueForSize(6, 8, 10, 12),
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
    emptyState: {
      alignItems: "center",
      paddingVertical: responsive.getValueForSize(40, 48, 56, 64),
      gap: responsive.spacing.elementSpacing / 2,
    },
    emptyText: {
      fontSize: responsive.getScaledFont("headline"),
      color: theme.colors.text.secondary,
      marginTop: responsive.spacing.elementSpacing / 2,
      textAlign: "center",
    },
    emptySubtext: {
      fontSize: responsive.getScaledFont("body"),
      color: theme.colors.text.tertiary,
      textAlign: "center",
      marginTop: 4,
      paddingHorizontal: responsive.spacing.horizontal,
    },
    wordCard: {
      flexDirection: "row",
      alignItems: "center",
      backgroundColor: theme.colors.background.surface,
      borderRadius: 12,
      padding: 16,
      marginHorizontal: responsive.spacing.horizontal,
      marginVertical: 6,
      borderWidth: 1,
      borderColor: theme.colors.ui.divider,
      shadowColor: theme.colors.text.primary,
      shadowOffset: { width: 0, height: 2 },
      shadowOpacity: 0.05,
      shadowRadius: 4,
      elevation: 2,
    },
    wordCardSelected: {
      borderColor: theme.colors.primary,
      backgroundColor: theme.colors.background.elevated,
    },
    wordContent: { flex: 1 },
    wordTerm: {
      fontSize: responsive.getScaledFont("headline"),
      fontWeight: "600",
      color: theme.colors.text.primary,
      // single-line display; no extra margin under term
      marginBottom: 0,
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
    // Removed pronunciation from word selection list for a cleaner UI
    checkbox: {
      width: responsive.getValueForSize(28, 32, 36, 40),
      height: responsive.getValueForSize(28, 32, 36, 40),
      borderRadius: theme.borderRadius.full,
      borderWidth: 2,
      borderColor: theme.colors.border.medium,
      backgroundColor: theme.colors.background.elevated,
      justifyContent: "center",
      alignItems: "center",
    },
    checkboxSelected: {
      backgroundColor: theme.colors.primary,
      borderColor: theme.colors.primary,
    },
    bottomControls: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      backgroundColor: theme.colors.background.surface,
      borderTopWidth: 1,
      borderTopColor: theme.colors.ui.divider,
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.vertical,
    },
    selectionIndicator: {
      backgroundColor: theme.colors.background.elevated,
      borderRadius: theme.borderRadius.full,
      paddingHorizontal: responsive.getValueForSize(12, 14, 16, 18),
      paddingVertical: responsive.getValueForSize(6, 8, 8, 10),
      borderWidth: 1,
      borderColor: theme.colors.ui.divider,
    },
    selectionCount: {
      fontSize: responsive.getScaledFont("label"),
      color: theme.colors.text.secondary,
      fontWeight: "600",
    },
    continueButton: {
      flexDirection: "row",
      alignItems: "center",
      gap: 8,
      backgroundColor: theme.colors.primary,
      paddingHorizontal: responsive.getValueForSize(14, 16, 18, 20),
      paddingVertical: responsive.getValueForSize(10, 12, 12, 14),
      borderRadius: theme.borderRadius.full,
      ...theme.shadows.sm,
    },
    continueButtonDisabled: {
      backgroundColor: theme.colors.ui.disabled,
    },
    continueButtonText: {
      fontSize: responsive.getScaledFont("label"),
      fontWeight: "700",
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
