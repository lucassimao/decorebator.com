import * as wordlistsApi from "@/api/wordlists";
import * as offlineWordsApi from "@/api/offlineWords";
import { Wordlist } from "@/api/wordlists";
import { useUpgradePromptDialog } from "@/hooks/useUpgradePromptDialog";
import { useOffline } from "@/hooks/useOffline";
import { OfflineIndicator } from "@/components/OfflineIndicator";
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
  const [searchQuery, setSearchQuery] = useState("");
  const [filterLearned, setFilterLearned] = useState<
    "all" | "learned" | "unlearned"
  >("all");
  const language = LANGUAGES.find((l) => wordlist.languageCode === l.code)!;
  const updatePromptDialog = useUpgradePromptDialog();
  const { isOnline } = useOffline();
  const { t } = useTranslation();

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
    onError: console.error,
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

  const handleAddWord = (data: wordlistsApi.CreateWordDTO) => {
    Keyboard.dismiss();
    addWordMutation.mutate(data);
  };

  const onPressAddWord = () => {
    if (words.length >= 10) {
      onClose();
      updatePromptDialog.show();
    } else {
      setShowAddForm(true);
    }
  };

  const renderWordItem = ({ item }: { item: wordlistsApi.Word }) => (
    <View style={styles.wordCard}>
      <TouchableOpacity
        style={[styles.learnedToggle, !isOnline && styles.disabledButton]}
        onPress={() =>
          isOnline && toggleLearnedMutation.mutate({
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
        <Text style={styles.wordTerm}>{item.name}</Text>
        {/* <Text style={styles.wordTranslation}>{item.translation}</Text> */}
        {item.pronunciation && (
          <Text style={styles.wordPronunciation}>[{item.pronunciation}]</Text>
        )}
        {item.notes && <Text style={styles.wordNotes}>{item.notes}</Text>}
      </View>

      {isOnline && (
        <TouchableOpacity
          style={styles.deleteButton}
          onPress={() => handleDeleteWord(item)}
        >
          <Ionicons name="trash-outline" size={20} color="#FF6B6B" />
        </TouchableOpacity>
      )}
    </View>
  );

  const stats = {
    total: words.length,
    learned: words.filter((w) => w.learned).length,
    progress:
      words.length > 0
        ? (words.filter((w) => w.learned).length / words.length) * 100
        : 0,
  };

  if (!visible) return null;

  return (
    <Modal
      transparent
      visible={visible}
      animationType="none"
      onRequestClose={onClose}
    >
      <TouchableWithoutFeedback onPress={onClose}>
        <Animated.View style={[styles.backdrop, { opacity: backdropAnim }]} />
      </TouchableWithoutFeedback>

      <View style={styles.container}>
        <Animated.View
          style={[
            styles.modalContent,
            { transform: [{ translateY: slideAnim }] },
          ]}
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
              <TouchableOpacity style={styles.closeButton} onPress={onClose}>
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
              <Text style={[styles.statValue, { color: "#4CAF50" }]}>
                {stats.learned}
              </Text>
              <Text style={styles.statLabel}>{t("wordDetail.learned")}</Text>
            </View>
            <View style={styles.stat}>
              <Text style={[styles.statValue, { color: "#FF7B54" }]}>
                {Math.round(stats.progress)}%
              </Text>
              <Text style={styles.statLabel}>{t("wordDetail.progress")}</Text>
            </View>
          </View>

          {/* Search and Filter */}
          <View style={{ flex: 1 }}>
            <View style={styles.searchContainer}>
              <View style={styles.searchBox}>
                <Ionicons name="search" size={20} color="#636E72" />
                <TextInput
                  style={styles.searchInput}
                  placeholder={t("wordDetail.searchPlaceholder")}
                  placeholderTextColor="#B2BEC3"
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
                  {t("wordDetail.filterLearned", { count: words.filter((w) => w.learned).length })}
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
                  {t("wordDetail.filterToLearn", { count: words.filter((w) => !w.learned).length })}
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
                      color="#DFE6E9"
                    />
                    <Text style={styles.emptyText}>
                      {searchQuery ? t("wordDetail.noWordsFound") : t("wordDetail.noWordsYet")}
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
                  <Text style={styles.formTitle}>{t("wordDetail.addNewWord")}</Text>
                  <TouchableOpacity
                    onPress={() => {
                      setShowAddForm(false);
                      reset();
                    }}
                  >
                    <Ionicons name="close" size={24} color="#636E72" />
                  </TouchableOpacity>
                </View>

                <Controller
                  control={control}
                  name="name"
                  rules={{
                    required: t("wordDetail.termRequired"),
                    minLength: { value: 1, message: t("wordDetail.termTooShort") },
                  }}
                  render={({ field: { onChange, onBlur, value } }) => (
                    <View style={styles.inputGroup}>
                      <Text style={styles.inputLabel}>{t("wordDetail.wordPhraseLabel")}</Text>
                      <TextInput
                        autoFocus
                        style={[styles.input, errors.name && styles.inputError]}
                        placeholder={t("wordDetail.wordPhrasePlaceholder")}
                        placeholderTextColor="#B2BEC3"
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
                      <Text style={styles.inputLabel}>{t("wordDetail.notesLabel")}</Text>
                      <TextInput
                        style={[styles.input, styles.textArea]}
                        placeholder={t("wordDetail.notesPlaceholder")}
                        placeholderTextColor="#B2BEC3"
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
                      color="#FF6B6B"
                    />
                    <Text style={styles.errorMessage}>
                      {t("wordDetail.addWordError")}
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
                    <ActivityIndicator color="#FFFFFF" size="small" />
                  ) : (
                    <Text style={styles.submitButtonText}>{t("wordDetail.addWordButton")}</Text>
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
              <Ionicons name="add" size={28} color="#FFFFFF" />
            </TouchableOpacity>
          )}
        </Animated.View>
      </View>
    </Modal>
  );
};

const styles = StyleSheet.create({
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
    backgroundColor: "rgba(0, 0, 0, 0.4)",
  },
  modalContent: {
    backgroundColor: "#FDF6E3",
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    height: SCREEN_HEIGHT * 0.95, // Changed from height to maxHeight
    shadowColor: "#000",
    shadowOffset: { width: 0, height: -4 },
    shadowOpacity: 0.1,
    shadowRadius: 12,
    elevation: 8,
  },
  header: {
    paddingTop: 12,
    paddingBottom: 16,
    paddingHorizontal: 20,
    backgroundColor: "#FFFFFF",
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    borderBottomWidth: 1,
    borderBottomColor: "#F0F0F0",
  },
  handle: {
    width: 40,
    height: 4,
    backgroundColor: "#DFE6E9",
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
    color: "#2D3436",
  },
  subtitle: {
    fontSize: 14,
    color: "#636E72",
    marginTop: 2,
  },
  closeButton: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: "#F5F5F5",
    justifyContent: "center",
    alignItems: "center",
  },
  statsRow: {
    flexDirection: "row",
    justifyContent: "space-around",
    paddingVertical: 16,
    paddingHorizontal: 20,
    backgroundColor: "#FFFFFF",
    borderBottomWidth: 1,
    borderBottomColor: "#F0F0F0",
  },
  stat: {
    alignItems: "center",
  },
  statValue: {
    fontSize: 24,
    fontWeight: "700",
    color: "#2D3436",
  },
  statLabel: {
    fontSize: 14,
    color: "#636E72",
    marginTop: 4,
  },
  searchContainer: {
    padding: 20,
    paddingBottom: 0,
  },
  searchBox: {
    flexDirection: "row",
    alignItems: "center",
    backgroundColor: "#FFFFFF",
    borderRadius: 12,
    paddingHorizontal: 16,
    paddingVertical: 12,
    gap: 12,
  },
  searchInput: {
    flex: 1,
    fontSize: 16,
    color: "#2D3436",
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
    backgroundColor: "#FFFFFF",
    marginRight: 8,
    borderWidth: 1,
    borderColor: "#E0E0E0",
  },
  filterChipActive: {
    backgroundColor: "#FF7B54",
    borderColor: "#FF7B54",
  },
  filterChipText: {
    fontSize: 14,
    color: "#636E72",
  },
  filterChipTextActive: {
    color: "#FFFFFF",
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
    backgroundColor: "#FFFFFF",
    borderRadius: 12,
    padding: 16,
    marginBottom: 8,
    shadowColor: "#000",
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
    color: "#2D3436",
    marginBottom: 4,
  },
  wordTranslation: {
    fontSize: 16,
    color: "#FF7B54",
    marginBottom: 2,
  },
  wordPronunciation: {
    fontSize: 14,
    color: "#636E72",
    fontStyle: "italic",
  },
  wordNotes: {
    fontSize: 14,
    color: "#636E72",
    marginTop: 4,
  },
  deleteButton: {
    padding: 8,
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
    color: "#636E72",
    marginTop: 16,
  },
  emptySubtext: {
    fontSize: 14,
    color: "#B2BEC3",
    marginTop: 8,
  },
  fab: {
    position: "absolute",
    right: 20,
    bottom: 20,
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: "#FF7B54",
    justifyContent: "center",
    alignItems: "center",
    shadowColor: "#FF7B54",
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
    backgroundColor: "#FFFFFF",
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    padding: 20,
    paddingBottom: 40,
    shadowColor: "#000",
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
    color: "#2D3436",
  },
  inputGroup: {
    marginBottom: 16,
  },
  inputLabel: {
    fontSize: 14,
    fontWeight: "500",
    color: "#2D3436",
    marginBottom: 8,
  },
  textArea: {
    height: 80,
    paddingTop: 12,
  },
  input: {
    borderWidth: 1,
    borderColor: "#E0E0E0",
    borderRadius: 12,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 16,
    color: "#2D3436",
    backgroundColor: "#FAFAFA",
  },
  inputError: {
    borderColor: "#FF6B6B",
  },
  errorText: {
    color: "#FF6B6B",
    fontSize: 12,
    marginTop: 4,
  },
  submitButton: {
    backgroundColor: "#FF7B54",
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: "center",
    marginTop: 8,
  },
  submitButtonText: {
    color: "#FFFFFF",
    fontSize: 16,
    fontWeight: "600",
  },
  buttonDisabled: {
    opacity: 0.7,
  },
  errorContainer: {
    flexDirection: "row",
    alignItems: "center",
    backgroundColor: "#FFF5F5",
    borderWidth: 1,
    borderColor: "#FFE0E0",
    borderRadius: 8,
    padding: 12,
    marginBottom: 16,
    gap: 8,
  },
  errorMessage: {
    flex: 1,
    color: "#FF6B6B",
    fontSize: 14,
  },
  disabledButton: {
    opacity: 0.5,
  },
});
