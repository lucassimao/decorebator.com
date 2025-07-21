import * as wordlistsApi from "@/api/wordlists";
import * as offlineWordsApi from "@/api/offlineWords";
import { Wordlist, getProcessingStatus } from "@/api/wordlists";
import {
  reportError,
  ErrorType,
  ErrorReportRateLimitError,
} from "@/api/errorReporting";
import { useUpgradePromptDialog } from "@/hooks/useUpgradePromptDialog";
import { useOffline } from "@/hooks/useOffline";
import { OfflineIndicator } from "@/components/OfflineIndicator";
import { useAnalytics } from "@/hooks/useAnalytics";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import React, { useEffect, useRef, useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import {
  ActivityIndicator,
  Alert,
  Animated,
  Dimensions,
  FlatList,
  Keyboard,
  KeyboardAvoidingView,
  Modal,
  PanResponder,
  Platform,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  TouchableWithoutFeedback,
  View,
} from "react-native";
import { LANGUAGES } from "./CreateWordlistModal";
import { useUserSession } from "@/hooks/useUserSession";
import { useTheme } from "@/contexts/ThemeContext";

const { height: SCREEN_HEIGHT } = Dimensions.get("window");

interface WordlistDetailModalProps {
  visible: boolean;
  onClose: () => void;
  wordlist: Wordlist;
}

export const WordlistDetailModal: React.FC<WordlistDetailModalProps> = ({
  visible,
  onClose,
  wordlist,
}) => {
  const slideAnim = useRef(new Animated.Value(SCREEN_HEIGHT)).current;
  const backdropAnim = useRef(new Animated.Value(0)).current;
  const queryClient = useQueryClient();
  const [showAddForm, setShowAddForm] = useState(false);
  const { theme } = useTheme();
  const styles = createStyles(theme);
  const [searchQuery, setSearchQuery] = useState("");
  const [filterLearned, setFilterLearned] = useState<
    "all" | "learned" | "unlearned"
  >("all");
  const language = LANGUAGES.find((l) => wordlist.languageCode === l.code)!;
  const updatePromptDialog = useUpgradePromptDialog();
  const { isOnline } = useOffline();
  const { t } = useTranslation();
  const { isPremium, hasOptimisticSubscription } = useUserSession();

  // Handle modal close with animation
  const handleClose = () => {
    Animated.parallel([
      Animated.timing(slideAnim, {
        toValue: SCREEN_HEIGHT,
        duration: 300,
        useNativeDriver: true,
      }),
      Animated.timing(backdropAnim, {
        toValue: 0,
        duration: 300,
        useNativeDriver: true,
      }),
    ]).start(() => {
      onClose();
    });
  };

  // Pan responder for swipe to dismiss
  const panResponder = useRef(
    PanResponder.create({
      onStartShouldSetPanResponder: () => false,
      onMoveShouldSetPanResponder: (_, gestureState) => {
        // Only respond to downward swipes with sufficient movement
        return (
          gestureState.dy > 10 &&
          Math.abs(gestureState.dx) < Math.abs(gestureState.dy)
        );
      },
      onPanResponderMove: (_, gestureState) => {
        // Only allow downward movement
        if (gestureState.dy > 0) {
          slideAnim.setValue(gestureState.dy);
        }
      },
      onPanResponderRelease: (_, gestureState) => {
        // If swipe is significant enough, close the modal
        if (gestureState.dy > 100 || gestureState.vy > 0.5) {
          handleClose();
        } else {
          // Otherwise, spring back to original position
          Animated.spring(slideAnim, {
            toValue: 0,
            useNativeDriver: true,
            tension: 40,
            friction: 8,
          }).start();
        }
      },
    }),
  ).current;

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<wordlistsApi.CreateWordDTO>({
    defaultValues: {
      name: "",
      notes: "",
    },
  });

  // Fetch words (with offline support)
  const { data: words = [], isLoading } = useQuery({
    queryKey: ["words", wordlist.id],
    queryFn: () => offlineWordsApi.getWords(wordlist.id),
    enabled: visible,
    retry: isOnline ? 3 : 0, // Don't retry in offline mode
  });

  // Get analytics data for this wordlist
  const { wordlistProgress } = useAnalytics(wordlist.id);

  // Fetch processing status for words (only when online)
  const { data: processingStatus } = useQuery({
    queryKey: ["processingStatus", wordlist.id],
    queryFn: () => getProcessingStatus(wordlist.id),
    enabled: visible && isOnline,
    refetchInterval: ({ state: { data } }) => {
      // Refetch every 5 seconds if there are pending or processing words
      if (!data) return false;
      const hasPendingOrProcessing =
        data.summary.pending > 0 || data.summary.processing > 0;
      return hasPendingOrProcessing ? 5000 : false;
    },
  });

  // Add word mutation
  const addWordMutation = useMutation({
    mutationFn: (data: wordlistsApi.CreateWordDTO) =>
      wordlistsApi.addWord({ ...data, wordlistId: wordlist.id }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["words", wordlist.id] });
      queryClient.invalidateQueries({ queryKey: ["wordlists"] });
      queryClient.invalidateQueries({ queryKey: ["dashboardStats"] });
      reset();
      setShowAddForm(false);
    },
    onError: (error) => {
      console.error("Failed to add word:", error);
      // The error will be displayed in the UI using getUserFriendlyErrorMessage
    },
  });

  const deleteWordMutation = useMutation({
    mutationFn: (wordId: number) =>
      wordlistsApi.deleteWord({ id: wordId, wordlistId: wordlist.id }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["words", wordlist.id] });
      queryClient.invalidateQueries({ queryKey: ["wordlists"] });
      queryClient.invalidateQueries({ queryKey: ["dashboardStats"] });
    },
    onError: console.error,
  });

  const toggleLearnedMutation = useMutation({
    mutationFn: (word: wordlistsApi.Word) =>
      wordlistsApi.updateWord({ ...word }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["words", wordlist.id] });
      queryClient.invalidateQueries({ queryKey: ["dashboardStats"] });
      queryClient.invalidateQueries({ queryKey: ["wordlists"] });
    },
    onError: console.error,
  });

  // Animation
  useEffect(() => {
    if (visible) {
      Animated.parallel([
        Animated.timing(slideAnim, {
          toValue: 0,
          duration: 300,
          useNativeDriver: true,
        }),
        Animated.timing(backdropAnim, {
          toValue: 1,
          duration: 300,
          useNativeDriver: true,
        }),
      ]).start();
    } else {
      Animated.parallel([
        Animated.timing(slideAnim, {
          toValue: SCREEN_HEIGHT,
          duration: 300,
          useNativeDriver: true,
        }),
        Animated.timing(backdropAnim, {
          toValue: 0,
          duration: 300,
          useNativeDriver: true,
        }),
      ]).start();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [visible]);

  // Filter words
  const filteredWords = words.filter((word) => {
    const matchesSearch = word.name
      .toLowerCase()
      .includes(searchQuery.toLowerCase());

    const matchesFilter =
      filterLearned === "all" ||
      (filterLearned === "learned" && word.learned) ||
      (filterLearned === "unlearned" && !word.learned);

    return matchesSearch && matchesFilter;
  });

  const handleDeleteWord = (word: wordlistsApi.Word) => {
    Alert.alert(
      t("wordDetail.deleteWord"),
      t("wordDetail.deleteWordConfirm", { word: word.name }),
      [
        { text: t("common.cancel"), style: "cancel" },
        {
          text: t("common.delete"),
          style: "destructive",
          onPress: () => deleteWordMutation.mutate(word.id),
        },
      ],
    );
  };

  // Helper function to determine if an error is retryable
  const isRetryableError = (error: string | undefined): boolean => {
    if (!error) return false;

    const errorLower = error.toLowerCase();

    // Non-retryable errors (content issues and validation errors)
    const nonRetryablePatterns = [
      "no definitions found",
      "no definition found",
      "word not found",
      "validation failed",
      "definition validation failed",
      // Content moderation patterns (same as getUserFriendlyErrorMessage)
      "not appropriate for educational use",
      "content not appropriate",
      "hate speech",
      "threatening language",
      "harassment",
      "threatening behavior",
      "violent",
      "graphic material",
      "self-harm",
      "flagged as inappropriate",
    ];

    return !nonRetryablePatterns.some((pattern) =>
      errorLower.includes(pattern),
    );
  };

  // Helper function to get user-friendly error message
  const getUserFriendlyErrorMessage = (error: string | undefined): string => {
    if (!error) return t("wordDetail.processingError");

    const errorLower = error.toLowerCase();

    // Check for specific non-retryable errors
    if (
      errorLower.includes("no definitions found") ||
      errorLower.includes("no definition found")
    ) {
      return t("wordDetail.noDefinitionsFound");
    }

    // Check for any content moderation violation
    const moderationPatterns = [
      "not appropriate for educational use", // Sexual content
      "content not appropriate", // Alternative sexual content format
      "hate speech", // Hate speech
      "threatening language", // Hate speech variant
      "harassment", // Harassment
      "threatening behavior", // Harassment variant
      "violent", // Violence
      "graphic material", // Violence variant
      "self-harm", // Self-harm
      "flagged as inappropriate", // Generic fallback
    ];

    const isModerationError = moderationPatterns.some((pattern) =>
      errorLower.includes(pattern),
    );

    if (isModerationError) {
      return t("wordDetail.contentNotAppropriate");
    }

    if (errorLower.includes("validation failed")) {
      return t("wordDetail.validationFailed");
    }

    // Check for character limit error
    if (errorLower.includes("words must be limited to 15 chars")) {
      return t("wordDetail.wordTooLong");
    }

    // For technical errors, show a generic message
    return t("wordDetail.technicalError");
  };

  const handleRetryWord = async (word: wordlistsApi.Word) => {
    const processingInfo = processingStatus?.words.find(
      (w) => w.id === word.id,
    );
    const isRetryable = isRetryableError(processingInfo?.processingError);

    if (!isRetryable) {
      Alert.alert(
        t("wordDetail.cannotRetry"),
        t("wordDetail.cannotRetryMessage"),
        [{ text: t("common.ok"), style: "default" }],
      );
      return;
    }

    Alert.alert(
      t("wordDetail.retryProcessing"),
      t("wordDetail.retryProcessingConfirm", { word: word.name }),
      [
        { text: t("common.cancel"), style: "cancel" },
        {
          text: t("common.retry"),
          style: "default",
          onPress: async () => {
            try {
              // Use error reports endpoint to trigger retry with rate limiting
              await reportError({
                wordId: word.id,
                definitionId: null, // No definition ID for processing errors
                errorType: ErrorType.ProcessingFailed,
              });

              // Invalidate queries to refresh the list
              queryClient.invalidateQueries({
                queryKey: ["words", wordlist.id],
              });
              queryClient.invalidateQueries({
                queryKey: ["processingStatus", wordlist.id],
              });

              Alert.alert(t("success.title"), t("wordDetail.retryStarted"), [
                { text: t("common.ok"), style: "default" },
              ]);
            } catch (error) {
              console.error("Retry failed:", error);

              if (error instanceof ErrorReportRateLimitError) {
                const retryAfter = error.retryAfter || 60;
                Alert.alert(
                  t("rateLimit.title"),
                  error.windowType === "cooldown"
                    ? t("rateLimit.cooldownMessage", {
                        minutes: Math.ceil(retryAfter / 60),
                      })
                    : t("rateLimit.limitReachedMessage"),
                  [{ text: t("common.ok"), style: "default" }],
                );
              } else {
                Alert.alert(t("error.title"), t("wordDetail.retryFailed"), [
                  { text: t("common.ok"), style: "default" },
                ]);
              }
            }
          },
        },
      ],
    );
  };

  const handleAddWord = (data: wordlistsApi.CreateWordDTO) => {
    Keyboard.dismiss();
    const payload = {
      ...data,
      hasOptimisticSubscription,
    };
    addWordMutation.mutate(payload);
  };

  const onPressAddWord = () => {
    if (words.length >= 10 && !isPremium) {
      onClose();
      updatePromptDialog.show();
    } else {
      setShowAddForm(true);
    }
  };

  // Helper function to get processing info for a word
  const getWordProcessingInfo = (wordId: number) => {
    return processingStatus?.words.find((w) => w.id === wordId);
  };

  const renderWordItem = ({ item }: { item: wordlistsApi.Word }) => {
    const processingInfo = getWordProcessingInfo(item.id);
    const status =
      processingInfo?.processingStatus || item.processingStatus || "completed";

    return (
      <View style={styles.wordCard}>
        <TouchableOpacity
          style={[styles.learnedToggle, !isOnline && styles.disabledButton]}
          onPress={() =>
            isOnline &&
            toggleLearnedMutation.mutate({
              ...item,
              learned: !item.learned,
            })
          }
          disabled={!isOnline}
        >
          <MaterialIcons
            name={item.learned ? "check-circle" : "radio-button-unchecked"}
            size={24}
            color={item.learned ? "#4CAF50" : "#DFE6E9"}
          />
        </TouchableOpacity>

        <View style={styles.wordContent}>
          <View style={styles.wordHeader}>
            <Text style={styles.wordTerm}>{item.name}</Text>
            {status === "pending" && (
              <View style={styles.statusIndicator}>
                <ActivityIndicator size="small" color={theme.colors.primary} />
                <Text style={styles.statusText}>
                  {t("wordDetail.processing")}
                </Text>
              </View>
            )}
            {status === "processing" && (
              <View style={styles.statusIndicator}>
                <ActivityIndicator size="small" color={theme.colors.primary} />
                <Text style={styles.statusText}>
                  {t("wordDetail.processing")}
                </Text>
              </View>
            )}
            {status === "failed" && (
              <View style={styles.statusIndicator}>
                <MaterialIcons name="error-outline" size={20} color="#FF6B6B" />
                <Text style={[styles.statusText, { color: "#FF6B6B" }]}>
                  {t("wordDetail.failed")}
                </Text>
              </View>
            )}
          </View>
          {item.pronunciation && (
            <Text style={styles.wordPronunciation}>[{item.pronunciation}]</Text>
          )}
          {item.notes && <Text style={styles.wordNotes}>{item.notes}</Text>}
          {status === "failed" && processingInfo?.processingError && (
            <Text style={styles.wordErrorMessage}>
              {getUserFriendlyErrorMessage(processingInfo.processingError)}
            </Text>
          )}
        </View>

        {isOnline && status === "failed" && (
          <>
            {isRetryableError(processingInfo?.processingError) ? (
              <TouchableOpacity
                style={styles.retryButton}
                onPress={() => handleRetryWord(item)}
              >
                <MaterialIcons
                  name="refresh"
                  size={20}
                  color={theme.colors.primary}
                />
              </TouchableOpacity>
            ) : (
              <TouchableOpacity
                style={styles.deleteButton}
                onPress={() => handleDeleteWord(item)}
              >
                <Ionicons name="trash-outline" size={20} color="#FF6B6B" />
              </TouchableOpacity>
            )}
          </>
        )}

        {isOnline && status === "completed" && (
          <TouchableOpacity
            style={styles.deleteButton}
            onPress={() => handleDeleteWord(item)}
          >
            <Ionicons name="trash-outline" size={20} color="#FF6B6B" />
          </TouchableOpacity>
        )}
      </View>
    );
  };

  // Use analytics data for learned count if available, fallback to local calculation
  const learnedFromAnalytics = wordlistProgress?.wordsMastered ?? 0;
  const progressFromAnalytics = wordlistProgress?.progressPercentage ?? 0;

  const stats = {
    total: words.length,
    learned: learnedFromAnalytics,
    progress: progressFromAnalytics,
  };

  if (!visible) return null;

  return (
    <Modal
      transparent
      visible={visible}
      animationType="none"
      onRequestClose={handleClose}
    >
      <TouchableWithoutFeedback onPress={handleClose}>
        <Animated.View style={[styles.backdrop, { opacity: backdropAnim }]} />
      </TouchableWithoutFeedback>

      <View style={styles.container}>
        <Animated.View
          style={[
            styles.modalContent,
            { transform: [{ translateY: slideAnim }] },
          ]}
          {...panResponder.panHandlers}
        >
          <OfflineIndicator />

          {/* Header */}
          <View style={styles.header}>
            <View style={styles.handle} />
            <View style={styles.titleRow}>
              <Text style={styles.languageFlag}>{language.flag}</Text>
              <View style={styles.titleContainer}>
                <Text style={styles.title}>{wordlist.name}</Text>
                {wordlist.description && (
                  <Text style={styles.subtitle}>{wordlist.description}</Text>
                )}
              </View>
              <TouchableOpacity
                style={styles.closeButton}
                onPress={handleClose}
              >
                <Ionicons name="close" size={24} color="#636E72" />
              </TouchableOpacity>
            </View>
          </View>

          {/* Stats */}
          <View style={styles.statsRow}>
            <View style={styles.stat}>
              <Text style={styles.statValue}>{stats.total}</Text>
              <Text style={styles.statLabel}>{t("wordDetail.total")}</Text>
            </View>
            <View style={styles.stat}>
              <Text style={[styles.statValue, { color: theme.colors.success }]}>
                {stats.learned}
              </Text>
              <Text style={styles.statLabel}>
                {learnedFromAnalytics !== null
                  ? t("wordDetail.mastered")
                  : t("wordDetail.learned")}
              </Text>
            </View>
            <View style={styles.stat}>
              <Text style={[styles.statValue, { color: theme.colors.primary }]}>
                {Math.round(stats.progress)}%
              </Text>
              <Text style={styles.statLabel}>
                {progressFromAnalytics !== null
                  ? t("wordDetail.masteryProgress")
                  : t("wordDetail.progress")}
              </Text>
            </View>
          </View>

          {/* Stats Explanation */}
          {learnedFromAnalytics !== null && (
            <View style={styles.statsExplanation}>
              <MaterialIcons
                name="info-outline"
                size={16}
                color={theme.colors.text.secondary}
              />
              <Text style={styles.explanationText}>
                {t("wordDetail.masteryExplanation")}
              </Text>
            </View>
          )}

          {/* Learned Toggle Explanation */}
          <View style={styles.toggleExplanation}>
            <MaterialIcons
              name="check-circle-outline"
              size={16}
              color={theme.colors.text.secondary}
            />
            <Text style={styles.explanationText}>
              {t("wordDetail.learnedToggleExplanation")}
            </Text>
          </View>

          {/* Search and Filter */}
          <View style={{ flex: 1 }}>
            <View style={styles.searchContainer}>
              <View style={styles.searchBox}>
                <Ionicons
                  name="search"
                  size={20}
                  color={theme.colors.text.secondary}
                />
                <TextInput
                  style={styles.searchInput}
                  placeholder={t("wordDetail.searchPlaceholder")}
                  placeholderTextColor={theme.colors.text.tertiary}
                  value={searchQuery}
                  onChangeText={setSearchQuery}
                />
              </View>
            </View>

            <ScrollView
              horizontal
              showsHorizontalScrollIndicator={false}
              style={styles.filterContainer}
            >
              <TouchableOpacity
                style={[
                  styles.filterChip,
                  filterLearned === "all" && styles.filterChipActive,
                ]}
                onPress={() => setFilterLearned("all")}
              >
                <Text
                  style={[
                    styles.filterChipText,
                    filterLearned === "all" && styles.filterChipTextActive,
                  ]}
                >
                  {t("wordDetail.filterAll", { count: words.length })}
                </Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[
                  styles.filterChip,
                  filterLearned === "learned" && styles.filterChipActive,
                ]}
                onPress={() => setFilterLearned("learned")}
              >
                <Text
                  style={[
                    styles.filterChipText,
                    filterLearned === "learned" && styles.filterChipTextActive,
                  ]}
                >
                  {t("wordDetail.filterLearned", {
                    count: words.filter((w) => w.learned).length,
                  })}
                </Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[
                  styles.filterChip,
                  filterLearned === "unlearned" && styles.filterChipActive,
                ]}
                onPress={() => setFilterLearned("unlearned")}
              >
                <Text
                  style={[
                    styles.filterChipText,
                    filterLearned === "unlearned" &&
                      styles.filterChipTextActive,
                  ]}
                >
                  {t("wordDetail.filterToLearn", {
                    count: words.filter((w) => !w.learned).length,
                  })}
                </Text>
              </TouchableOpacity>
            </ScrollView>

            {/* Words List */}
            {isLoading ? (
              <View style={styles.loadingContainer}>
                <ActivityIndicator size="large" color="#FF7B54" />
              </View>
            ) : (
              <FlatList
                data={filteredWords}
                keyboardShouldPersistTaps="handled"
                renderItem={renderWordItem}
                keyExtractor={(item) => String(item.id)}
                contentContainerStyle={styles.listContent}
                style={{ flex: 1 }}
                showsVerticalScrollIndicator={false}
                ListEmptyComponent={
                  <View style={styles.emptyState}>
                    <MaterialIcons
                      name="library-books"
                      size={48}
                      color={theme.colors.text.tertiary}
                    />
                    <Text style={styles.emptyText}>
                      {searchQuery
                        ? t("wordDetail.noWordsFound")
                        : t("wordDetail.noWordsYet")}
                    </Text>
                    {!searchQuery && isOnline && (
                      <Text style={styles.emptySubtext}>
                        {t("dashboard.wordlists.addFirstWord")}
                      </Text>
                    )}
                    {!isOnline && (
                      <Text style={styles.emptySubtext}>
                        {t("offline.featureUnavailable")}
                      </Text>
                    )}
                  </View>
                }
              />
            )}
          </View>

          {/* Add Word Form */}
          {showAddForm && (
            <KeyboardAvoidingView
              behavior={Platform.OS === "ios" ? "padding" : "height"}
              keyboardVerticalOffset={Platform.OS === "ios" ? 0 : 20}
              style={styles.keyboardAvoidingView}
            >
              <View style={styles.addFormContainer}>
                <View style={styles.formHeader}>
                  <Text style={styles.formTitle}>
                    {t("wordDetail.addNewWord")}
                  </Text>
                  <TouchableOpacity
                    onPress={() => {
                      setShowAddForm(false);
                      reset();
                    }}
                  >
                    <Ionicons
                      name="close"
                      size={24}
                      color={theme.colors.text.secondary}
                    />
                  </TouchableOpacity>
                </View>

                <Controller
                  control={control}
                  name="name"
                  rules={{
                    required: t("wordDetail.termRequired"),
                    minLength: {
                      value: 1,
                      message: t("wordDetail.termTooShort"),
                    },
                  }}
                  render={({ field: { onChange, onBlur, value } }) => (
                    <View style={styles.inputGroup}>
                      <Text style={styles.inputLabel}>
                        {t("wordDetail.wordPhraseLabel")}
                      </Text>
                      <TextInput
                        autoFocus
                        style={[styles.input, errors.name && styles.inputError]}
                        placeholder={t("wordDetail.wordPhrasePlaceholder")}
                        placeholderTextColor={theme.colors.text.tertiary}
                        value={value}
                        onChangeText={(text) => {
                          onChange(text);
                          addWordMutation.reset();
                        }}
                        onBlur={onBlur}
                      />
                      {errors.name && (
                        <Text style={styles.errorText}>
                          {errors.name.message}
                        </Text>
                      )}
                    </View>
                  )}
                />

                <Controller
                  control={control}
                  name="notes"
                  render={({ field: { onChange, onBlur, value } }) => (
                    <View style={styles.inputGroup}>
                      <Text style={styles.inputLabel}>
                        {t("wordDetail.notesLabel")}
                      </Text>
                      <TextInput
                        style={[styles.input, styles.textArea]}
                        placeholder={t("wordDetail.notesPlaceholder")}
                        placeholderTextColor={theme.colors.text.tertiary}
                        value={value}
                        onChangeText={(text) => {
                          onChange(text);
                          addWordMutation.reset();
                        }}
                        onBlur={onBlur}
                        multiline
                        numberOfLines={3}
                        textAlignVertical="top"
                      />
                    </View>
                  )}
                />

                {addWordMutation.error && (
                  <View style={styles.errorContainer}>
                    <MaterialIcons
                      name="error-outline"
                      size={20}
                      color={theme.colors.error}
                    />
                    <Text style={styles.errorMessage}>
                      {getUserFriendlyErrorMessage(
                        addWordMutation.error?.message,
                      )}
                    </Text>
                  </View>
                )}
                <TouchableOpacity
                  style={[
                    styles.submitButton,
                    addWordMutation.isPending && styles.buttonDisabled,
                  ]}
                  onPress={handleSubmit(handleAddWord)}
                  disabled={addWordMutation.isPending}
                >
                  {addWordMutation.isPending ? (
                    <ActivityIndicator
                      color={theme.colors.text.inverse}
                      size="small"
                    />
                  ) : (
                    <Text style={styles.submitButtonText}>
                      {t("wordDetail.addWordButton")}
                    </Text>
                  )}
                </TouchableOpacity>
              </View>
            </KeyboardAvoidingView>
          )}

          {/* FAB */}
          {!showAddForm && isOnline && (
            <TouchableOpacity
              style={styles.fab}
              onPress={onPressAddWord}
              activeOpacity={0.8}
            >
              <Ionicons
                name="add"
                size={28}
                color={theme.colors.text.inverse}
              />
            </TouchableOpacity>
          )}
        </Animated.View>
      </View>
    </Modal>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    container: {
      flex: 1,
      justifyContent: "flex-end",
    },
    backdrop: {
      position: "absolute",
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor:
        theme.mode === "light" ? "rgba(0, 0, 0, 0.4)" : "rgba(0, 0, 0, 0.6)",
    },
    modalContent: {
      backgroundColor: theme.colors.background.surface,
      borderTopLeftRadius: 24,
      borderTopRightRadius: 24,
      height: SCREEN_HEIGHT * 0.95, // Changed from height to maxHeight
      shadowColor: theme.colors.text.primary,
      shadowOffset: { width: 0, height: -4 },
      shadowOpacity: 0.1,
      shadowRadius: 12,
      elevation: 8,
    },
    header: {
      paddingTop: 12,
      paddingBottom: 16,
      paddingHorizontal: 20,
      backgroundColor: theme.colors.background.surface,
      borderTopLeftRadius: 24,
      borderTopRightRadius: 24,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.border.light,
    },
    handle: {
      width: 40,
      height: 4,
      backgroundColor: theme.colors.text.tertiary,
      borderRadius: 2,
      alignSelf: "center",
      marginBottom: 16,
    },
    titleRow: {
      flexDirection: "row",
      alignItems: "center",
    },
    languageFlag: {
      fontSize: 32,
      marginRight: 12,
    },
    titleContainer: {
      flex: 1,
    },
    title: {
      fontSize: 24,
      fontWeight: "600",
      color: theme.colors.text.primary,
    },
    subtitle: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      marginTop: 2,
    },
    closeButton: {
      width: 32,
      height: 32,
      borderRadius: 16,
      backgroundColor: theme.colors.background.default,
      justifyContent: "center",
      alignItems: "center",
    },
    statsRow: {
      flexDirection: "row",
      justifyContent: "space-around",
      paddingVertical: 16,
      paddingHorizontal: 20,
      backgroundColor: theme.colors.background.surface,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.border.light,
    },
    stat: {
      alignItems: "center",
    },
    statValue: {
      fontSize: 24,
      fontWeight: "700",
      color: theme.colors.text.primary,
    },
    statLabel: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      marginTop: 4,
    },
    statsExplanation: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: 20,
      paddingVertical: 12,
      backgroundColor: theme.colors.background.secondary,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.border.light,
      gap: 8,
    },
    toggleExplanation: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: 20,
      paddingVertical: 12,
      backgroundColor: theme.colors.background.secondary,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.border.light,
      gap: 8,
    },
    explanationText: {
      flex: 1,
      fontSize: 13,
      color: theme.colors.text.secondary,
      lineHeight: 18,
    },
    searchContainer: {
      padding: 20,
      paddingBottom: 0,
    },
    searchBox: {
      flexDirection: "row",
      alignItems: "center",
      backgroundColor: theme.colors.ui.inputBackground,
      borderRadius: 12,
      paddingHorizontal: 16,
      paddingVertical: 12,
      gap: 12,
    },
    searchInput: {
      flex: 1,
      fontSize: 16,
      color: theme.colors.text.primary,
    },
    filterContainer: {
      paddingHorizontal: 20,
      paddingVertical: 12,
      maxHeight: 60,
    },
    filterChip: {
      paddingHorizontal: 16,
      paddingVertical: 8,
      borderRadius: 20,
      backgroundColor: theme.colors.background.surface,
      marginRight: 8,
      borderWidth: 1,
      borderColor: theme.colors.border.light,
    },
    filterChipActive: {
      backgroundColor: theme.colors.primary,
      borderColor: theme.colors.primary,
    },
    filterChipText: {
      fontSize: 14,
      color: theme.colors.text.secondary,
    },
    filterChipTextActive: {
      color: theme.colors.text.inverse,
      fontWeight: "500",
    },
    listContent: {
      paddingHorizontal: 20,
      paddingBottom: 100,
      paddingTop: 10, // Add this
    },
    wordCard: {
      flexDirection: "row",
      alignItems: "center",
      backgroundColor: theme.colors.background.surface,
      borderRadius: 12,
      padding: 16,
      marginBottom: 8,
      shadowColor: theme.colors.text.primary,
      shadowOffset: { width: 0, height: 2 },
      shadowOpacity: 0.05,
      shadowRadius: 4,
      elevation: 2,
    },
    learnedToggle: {
      marginRight: 12,
    },
    wordContent: {
      flex: 1,
    },
    wordTerm: {
      fontSize: 16,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: 4,
    },
    wordTranslation: {
      fontSize: 16,
      color: theme.colors.primary,
      marginBottom: 2,
    },
    wordPronunciation: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      fontStyle: "italic",
    },
    wordNotes: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      marginTop: 4,
    },
    deleteButton: {
      padding: 8,
    },
    retryButton: {
      padding: 8,
    },
    wordHeader: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
    },
    statusIndicator: {
      flexDirection: "row",
      alignItems: "center",
      gap: 4,
    },
    statusText: {
      fontSize: 12,
      color: theme.colors.text.secondary,
    },
    wordErrorMessage: {
      fontSize: 12,
      color: "#FF6B6B",
      marginTop: 4,
    },
    loadingContainer: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
    },
    emptyState: {
      alignItems: "center",
      paddingVertical: 60,
    },
    emptyText: {
      fontSize: 18,
      color: theme.colors.text.secondary,
      marginTop: 16,
    },
    emptySubtext: {
      fontSize: 14,
      color: theme.colors.text.tertiary,
      marginTop: 8,
    },
    fab: {
      position: "absolute",
      right: 20,
      bottom: 20,
      width: 56,
      height: 56,
      borderRadius: 28,
      backgroundColor: theme.colors.primary,
      justifyContent: "center",
      alignItems: "center",
      shadowColor: theme.colors.primary,
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.3,
      shadowRadius: 8,
      elevation: 8,
    },
    addFormContainer: {
      // position: "absolute",
      bottom: 0,
      left: 0,
      right: 0,
      backgroundColor: theme.colors.background.surface,
      borderTopLeftRadius: 24,
      borderTopRightRadius: 24,
      padding: 20,
      paddingBottom: 40,
      shadowColor: theme.colors.text.primary,
      shadowOffset: { width: 0, height: -4 },
      shadowOpacity: 0.1,
      shadowRadius: 12,
      elevation: 8,
    },
    keyboardAvoidingView: {
      position: "absolute",
      bottom: 0,
      left: 0,
      right: 0,
    },
    formHeader: {
      flexDirection: "row",
      justifyContent: "space-between",
      alignItems: "center",
      marginBottom: 20,
    },
    formTitle: {
      fontSize: 20,
      fontWeight: "600",
      color: theme.colors.text.primary,
    },
    inputGroup: {
      marginBottom: 16,
    },
    inputLabel: {
      fontSize: 14,
      fontWeight: "500",
      color: theme.colors.text.primary,
      marginBottom: 8,
    },
    textArea: {
      height: 80,
      paddingTop: 12,
    },
    input: {
      borderWidth: 1,
      borderColor: theme.colors.border.light,
      borderRadius: 12,
      paddingHorizontal: 16,
      paddingVertical: 12,
      fontSize: 16,
      color: theme.colors.text.primary,
      backgroundColor: theme.colors.background.default,
    },
    inputError: {
      borderColor: theme.colors.error,
    },
    errorText: {
      color: theme.colors.error,
      fontSize: 12,
      marginTop: 4,
    },
    submitButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: 12,
      paddingVertical: 16,
      alignItems: "center",
      marginTop: 8,
    },
    submitButtonText: {
      color: theme.colors.text.inverse,
      fontSize: 16,
      fontWeight: "600",
    },
    buttonDisabled: {
      opacity: 0.7,
    },
    errorContainer: {
      flexDirection: "row",
      alignItems: "center",
      backgroundColor:
        theme.mode === "light" ? "#FFF5F5" : theme.colors.background.secondary,
      borderWidth: 1,
      borderColor:
        theme.mode === "light" ? "#FFE0E0" : theme.colors.border.medium,
      borderRadius: 8,
      padding: 12,
      marginBottom: 16,
      gap: 8,
    },
    errorMessage: {
      flex: 1,
      color: theme.colors.error,
      fontSize: 14,
    },
    disabledButton: {
      opacity: 0.5,
    },
  });
