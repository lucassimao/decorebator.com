import React, {
  useState,
  useCallback,
  useMemo,
  useRef,
  useEffect,
} from "react";
import { useLocalSearchParams, useRouter, Stack } from "expo-router";
import {
  View,
  StyleSheet,
  Text,
  TouchableOpacity,
  StatusBar,
  FlatList,
  Alert,
  TextInput,
  Animated,
} from "react-native";
import { MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import { useQuery } from "@tanstack/react-query";
import { LoadingWithTimeout } from "@/components/LoadingWithTimeout";
import ScreenHeader from "@/components/common/ScreenHeader";
import * as wordlistsApi from "@/api/wordlists";
import { Word } from "@/api/wordlists";
import { SafeAreaView } from "react-native-safe-area-context";
import { LinearGradient } from "expo-linear-gradient";

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

  // Animation refs
  const fadeAnim = useRef(new Animated.Value(0)).current;
  const slideAnim = useRef(new Animated.Value(30)).current;

  // State for word selection
  const [selectedWordIds, setSelectedWordIds] = useState<Set<number>>(
    new Set(),
  );
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  // Entrance animation
  useEffect(() => {
    Animated.parallel([
      Animated.timing(fadeAnim, {
        toValue: 1,
        duration: 400,
        useNativeDriver: true,
      }),
      Animated.spring(slideAnim, {
        toValue: 0,
        friction: 8,
        tension: 40,
        useNativeDriver: true,
      }),
    ]).start();
  }, [fadeAnim, slideAnim]);

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
    ({ item, index }: { item: Word; index: number }) => {
      const isSelected = selectedWordIds.has(item.id);

      return (
        <Animated.View
          style={{
            opacity: fadeAnim,
            transform: [
              {
                translateY: slideAnim.interpolate({
                  inputRange: [0, 30],
                  outputRange: [0, 10 + index * 2],
                }),
              },
            ],
          }}
        >
          <TouchableOpacity
            style={[styles.wordCard, isSelected && styles.wordCardSelected]}
            onPress={() => handleWordToggle(item.id)}
            activeOpacity={0.8}
            accessibilityRole="checkbox"
            accessibilityState={{ checked: isSelected }}
            accessibilityLabel={`${item.name}${isSelected ? t("wordSelection.selected", " - selected") : ""}`}
          >
            <View style={styles.wordContent}>
              <Text
                style={[styles.wordTerm, isSelected && styles.wordTermSelected]}
              >
                {item.name}
              </Text>
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
          </TouchableOpacity>
        </Animated.View>
      );
    },
    [
      selectedWordIds,
      styles,
      responsive,
      theme.colors.text.inverse,
      handleWordToggle,
      t,
      fadeAnim,
      slideAnim,
    ],
  );

  const headerTitle = t("wordSelection.title", "Select Words for Chat");
  const headerSubtitle = t(
    "wordSelection.subtitle",
    "Choose up to {{count}} words to practice",
    { count: MAX_SELECTED_WORDS },
  );

  // Glassmorphic wrapper component (semi-transparent background with blur effect simulation)
  const GlassContainer: React.FC<{
    children: React.ReactNode;
    style?: any;
    intensity?: number;
  }> = ({ children, style }) => {
    return (
      <View style={[styles.glassBase, styles.glassEffect, style]}>
        {children}
      </View>
    );
  };

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
    <View style={styles.container}>
      {/* Background gradient */}
      <LinearGradient
        colors={
          theme.mode === "light"
            ? ["#FFF8F4", "#FDF6E3", "#FFF0E6"]
            : [
                theme.colors.background.default,
                theme.colors.background.surface,
                theme.colors.background.default,
              ]
        }
        style={StyleSheet.absoluteFill}
        start={{ x: 0, y: 0 }}
        end={{ x: 1, y: 1 }}
      />

      <SafeAreaView style={styles.safeArea}>
        <Stack.Screen options={{ headerShown: false }} />
        <StatusBar
          barStyle={theme.mode === "light" ? "dark-content" : "light-content"}
        />

        {/* Header */}
        <ScreenHeader
          title={headerTitle}
          subtitle={headerSubtitle}
          onBackPress={() => router.back()}
        />

        {/* Search Panel - Glassmorphic */}
        <Animated.View
          style={[
            styles.searchPanelWrapper,
            {
              opacity: fadeAnim,
              transform: [{ translateY: slideAnim }],
            },
          ]}
        >
          <GlassContainer style={styles.searchPanel}>
            <View style={styles.searchBox}>
              <MaterialIcons
                name="search"
                size={20}
                color={theme.colors.text.secondary}
              />
              <TextInput
                style={styles.searchInput}
                placeholder={t("wordSelection.search", "Search words...")}
                placeholderTextColor={theme.colors.text.tertiary}
                value={searchQuery}
                onChangeText={setSearchQuery}
                autoCorrect={false}
                autoCapitalize="none"
              />
              {searchQuery.length > 0 && (
                <TouchableOpacity onPress={() => setSearchQuery("")}>
                  <MaterialIcons
                    name="close"
                    size={18}
                    color={theme.colors.text.tertiary}
                  />
                </TouchableOpacity>
              )}
            </View>

            <View style={styles.headerActions}>
              <TouchableOpacity
                style={[styles.chip, styles.chipPrimary]}
                onPress={handleSelectAll}
                disabled={selectableWords.length === 0}
                activeOpacity={0.7}
              >
                <MaterialIcons
                  name="select-all"
                  size={16}
                  color={theme.colors.text.inverse}
                />
                <Text style={styles.chipPrimaryText}>
                  {t("wordSelection.selectAll", "Select All")}
                </Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[styles.chip, styles.chipSecondary]}
                onPress={handleClearAll}
                disabled={selectedWordIds.size === 0}
                activeOpacity={0.7}
              >
                <MaterialIcons
                  name="clear-all"
                  size={16}
                  color={theme.colors.text.secondary}
                />
                <Text style={styles.chipSecondaryText}>
                  {t("wordSelection.clearAll", "Clear")}
                </Text>
              </TouchableOpacity>
            </View>
          </GlassContainer>
        </Animated.View>

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
            </View>
          }
        />

        {/* Bottom Controls - Glassmorphic */}
        <GlassContainer style={styles.bottomControls} intensity={80}>
          <View style={styles.selectionIndicator}>
            <View style={styles.selectionDots}>
              {[...Array(MAX_SELECTED_WORDS)].map((_, i) => (
                <View
                  key={i}
                  style={[
                    styles.selectionDot,
                    i < selectedWordIds.size && styles.selectionDotActive,
                  ]}
                />
              ))}
            </View>
            <Text style={styles.selectionCount}>
              {selectedWordIds.size}/{MAX_SELECTED_WORDS}
            </Text>
          </View>

          <TouchableOpacity
            style={[
              styles.continueButton,
              selectedWordIds.size === 0 && styles.continueButtonDisabled,
            ]}
            onPress={handleContinueToChat}
            disabled={selectedWordIds.size === 0 || loading}
            activeOpacity={0.8}
            accessibilityRole="button"
            accessibilityLabel={t(
              "wordSelection.continueToChat",
              "Continue to Chat",
            )}
          >
            <LinearGradient
              colors={
                selectedWordIds.size > 0
                  ? [theme.colors.primary, "#FF9966"]
                  : [theme.colors.ui.disabled, theme.colors.ui.disabled]
              }
              style={styles.continueButtonGradient}
              start={{ x: 0, y: 0 }}
              end={{ x: 1, y: 1 }}
            >
              {loading ? (
                <MaterialIcons
                  name="hourglass-empty"
                  size={responsive.getValueForSize(20, 22, 24, 26)}
                  color={theme.colors.text.inverse}
                />
              ) : (
                <MaterialIcons
                  name="mic"
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
                  selectedWordIds.size === 0 &&
                    styles.continueButtonTextDisabled,
                ]}
              >
                {t("wordSelection.startChat", "Start Voice Chat")}
              </Text>
              <MaterialIcons
                name="arrow-forward"
                size={18}
                color={
                  selectedWordIds.size > 0
                    ? theme.colors.text.inverse
                    : theme.colors.text.disabled
                }
              />
            </LinearGradient>
          </TouchableOpacity>
        </GlassContainer>
      </SafeAreaView>
    </View>
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
    safeArea: {
      flex: 1,
    },
    glassBase: {
      overflow: "hidden",
    },
    glassEffect: {
      backgroundColor:
        theme.mode === "light"
          ? "rgba(255, 255, 255, 0.85)"
          : "rgba(30, 30, 30, 0.85)",
    },
    searchPanelWrapper: {
      marginHorizontal: responsive.spacing.horizontal,
      marginTop: responsive.spacing.elementSpacing / 2,
      marginBottom: responsive.spacing.elementSpacing,
    },
    searchPanel: {
      borderRadius: responsive.getValueForSize(16, 18, 20, 22),
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? "rgba(255, 255, 255, 0.6)"
          : "rgba(255, 255, 255, 0.1)",
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.getValueForSize(12, 14, 16, 18),
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 8 },
      shadowOpacity: 0.08,
      shadowRadius: 16,
      elevation: 8,
    },
    searchBox: {
      flexDirection: "row",
      alignItems: "center",
      backgroundColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.04)"
          : "rgba(255, 255, 255, 0.08)",
      borderRadius: responsive.getValueForSize(12, 14, 16, 18),
      paddingHorizontal: 14,
      paddingVertical: responsive.getValueForSize(10, 12, 12, 14),
    },
    searchInput: {
      flex: 1,
      marginLeft: 10,
      color: theme.colors.text.primary,
      fontSize: responsive.getScaledFont("body"),
      lineHeight: Math.round(responsive.getScaledFont("body") * 1.3),
    },
    headerActions: {
      flexDirection: "row",
      marginTop: responsive.getValueForSize(12, 14, 16, 18),
      gap: responsive.spacing.elementSpacing / 2,
    },
    chip: {
      flexDirection: "row",
      alignItems: "center",
      gap: 6,
      paddingHorizontal: responsive.getValueForSize(14, 16, 18, 20),
      paddingVertical: responsive.getValueForSize(10, 11, 12, 13),
      borderRadius: theme.borderRadius.full,
    },
    chipPrimary: {
      backgroundColor: theme.colors.primary,
    },
    chipPrimaryText: {
      color: theme.colors.text.inverse,
      fontWeight: "600",
      fontSize: responsive.getScaledFont("label"),
    },
    chipSecondary: {
      backgroundColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.06)"
          : "rgba(255, 255, 255, 0.1)",
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.08)"
          : "rgba(255, 255, 255, 0.15)",
    },
    chipSecondaryText: {
      color: theme.colors.text.secondary,
      fontWeight: "600",
      fontSize: responsive.getScaledFont("label"),
    },
    wordsList: {
      flex: 1,
    },
    wordsListContent: {
      paddingHorizontal: responsive.spacing.horizontal,
      paddingTop: responsive.spacing.vertical / 2,
      paddingBottom: responsive.spacing.vertical * 2,
    },
    wordCard: {
      flexDirection: "row",
      alignItems: "center",
      backgroundColor:
        theme.mode === "light"
          ? "rgba(255, 255, 255, 0.9)"
          : "rgba(255, 255, 255, 0.08)",
      borderRadius: responsive.getValueForSize(14, 16, 18, 20),
      padding: responsive.getValueForSize(14, 16, 18, 20),
      marginBottom: responsive.getValueForSize(8, 10, 12, 14),
      borderWidth: 1.5,
      borderColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.04)"
          : "rgba(255, 255, 255, 0.08)",
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.06,
      shadowRadius: 8,
      elevation: 3,
    },
    wordCardSelected: {
      borderColor: theme.colors.primary,
      backgroundColor:
        theme.mode === "light"
          ? "rgba(255, 123, 84, 0.08)"
          : "rgba(255, 123, 84, 0.15)",
      shadowColor: theme.colors.primary,
      shadowOpacity: 0.15,
    },
    wordContent: {
      flex: 1,
    },
    wordTerm: {
      fontSize: responsive.getScaledFont("headline"),
      fontWeight: "600",
      color: theme.colors.text.primary,
      letterSpacing: -0.3,
    },
    wordTermSelected: {
      color: theme.colors.primary,
    },
    checkbox: {
      width: responsive.getValueForSize(28, 32, 36, 40),
      height: responsive.getValueForSize(28, 32, 36, 40),
      borderRadius: theme.borderRadius.full,
      borderWidth: 2,
      borderColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.15)"
          : "rgba(255, 255, 255, 0.2)",
      backgroundColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.02)"
          : "rgba(255, 255, 255, 0.05)",
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
      borderTopWidth: 1,
      borderTopColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.06)"
          : "rgba(255, 255, 255, 0.1)",
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.getValueForSize(14, 16, 18, 20),
      borderTopLeftRadius: responsive.getValueForSize(20, 24, 28, 32),
      borderTopRightRadius: responsive.getValueForSize(20, 24, 28, 32),
    },
    selectionIndicator: {
      flexDirection: "row",
      alignItems: "center",
      gap: 10,
    },
    selectionDots: {
      flexDirection: "row",
      gap: 6,
    },
    selectionDot: {
      width: responsive.getValueForSize(8, 10, 12, 14),
      height: responsive.getValueForSize(8, 10, 12, 14),
      borderRadius: responsive.getValueForSize(4, 5, 6, 7),
      backgroundColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.1)"
          : "rgba(255, 255, 255, 0.15)",
    },
    selectionDotActive: {
      backgroundColor: theme.colors.primary,
    },
    selectionCount: {
      fontSize: responsive.getScaledFont("label"),
      color: theme.colors.text.secondary,
      fontWeight: "700",
    },
    continueButton: {
      borderRadius: theme.borderRadius.full,
      overflow: "hidden",
      shadowColor: theme.colors.primary,
      shadowOffset: { width: 0, height: 6 },
      shadowOpacity: 0.3,
      shadowRadius: 12,
      elevation: 8,
    },
    continueButtonDisabled: {
      shadowOpacity: 0,
      elevation: 0,
    },
    continueButtonGradient: {
      flexDirection: "row",
      alignItems: "center",
      gap: 8,
      paddingHorizontal: responsive.getValueForSize(18, 20, 22, 24),
      paddingVertical: responsive.getValueForSize(12, 14, 16, 18),
    },
    continueButtonText: {
      fontSize: responsive.getScaledFont("body"),
      fontWeight: "700",
      color: theme.colors.text.inverse,
    },
    continueButtonTextDisabled: {
      color: theme.colors.text.disabled,
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
