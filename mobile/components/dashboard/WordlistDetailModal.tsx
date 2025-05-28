// types/word.types.ts
export interface Word {
  id: number;
  term: string;
  translation: string;
  pronunciation?: string;
  notes?: string;
  learned: boolean;
  createdAt: string;
}

export interface WordFormData {
  term: string;
  translation: string;
  pronunciation?: string;
  notes?: string;
}

import * as wordlistsApi from "@/api/wordlists";
import React, { useRef, useEffect, useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  Modal,
  TouchableOpacity,
  TouchableWithoutFeedback,
  FlatList,
  TextInput,
  Animated,
  Dimensions,
  KeyboardAvoidingView,
  Platform,
  ActivityIndicator,
  Alert,
  ScrollView,
} from 'react-native';
import { Ionicons, MaterialIcons } from '@expo/vector-icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm, Controller } from 'react-hook-form';
import { Wordlist } from '@/api/wordlists';
import { LANGUAGES } from './CreateWordlistModal';

const { height: SCREEN_HEIGHT, width: SCREEN_WIDTH } = Dimensions.get('window');

// API calls (replace with actual endpoints)
const fetchWords = async (wordlistId: number): Promise<Word[]> => {
  await new Promise(resolve => setTimeout(resolve, 500));
  
  // Replace with actual API call
  // const response = await fetch(`/api/wordlists/${wordlistId}/words`);
  // return response.json();
  
  return [
    {
      id: 1,
      term: 'Hello',
      translation: 'Hola',
      pronunciation: 'OH-lah',
      learned: true,
      createdAt: '2024-01-15T10:00:00Z',
    },
    {
      id: 2,
      term: 'Thank you',
      translation: 'Gracias',
      pronunciation: 'GRAH-see-ahs',
      notes: 'Common courtesy phrase',
      learned: true,
      createdAt: '2024-01-15T10:00:00Z',
    },
    {
      id: 3,
      term: 'Goodbye',
      translation: 'Adiós',
      pronunciation: 'ah-dee-OHS',
      learned: false,
      createdAt: '2024-01-16T10:00:00Z',
    },
  ];
};

const addWord = async (wordlistId: number, data: WordFormData): Promise<Word> => {
  await new Promise(resolve => setTimeout(resolve, 500));
  
  // Replace with actual API call
  // const response = await fetch(`/api/wordlists/${wordlistId}/words`, {
  //   method: 'POST',
  //   headers: { 'Content-Type': 'application/json' },
  //   body: JSON.stringify(data),
  // });
  // return response.json();
  
  return {
    id: Date.now(),
    ...data,
    learned: false,
    createdAt: new Date().toISOString(),
  };
};


const toggleWordLearned = async (wordlistId: number, wordId: number, learned: boolean): Promise<void> => {
  await new Promise(resolve => setTimeout(resolve, 300));
  
  // Replace with actual API call
  // await fetch(`/api/wordlists/${wordlistId}/words/${wordId}`, {
  //   method: 'PATCH',
  //   headers: { 'Content-Type': 'application/json' },
  //   body: JSON.stringify({ learned }),
  // });
};

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
  const [searchQuery, setSearchQuery] = useState('');
  const [filterLearned, setFilterLearned] = useState<'all' | 'learned' | 'unlearned'>('all');
  const language = LANGUAGES.find(l => (wordlist.languageCode) === l.code)!;

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<WordFormData>({
    defaultValues: {
      term: '',
      translation: '',
      pronunciation: '',
      notes: '',
    },
  });

  // Fetch words
  const { data: words = [], isLoading } = useQuery({
    queryKey: ['words', wordlist.id],
    queryFn: () => fetchWords(wordlist.id),
    enabled: visible,
  });

  // Add word mutation
  const addWordMutation = useMutation({
    mutationFn: (data: WordFormData) => addWord(wordlist.id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['words', wordlist.id] });
      queryClient.invalidateQueries({ queryKey: ['wordlists'] });
      queryClient.invalidateQueries({ queryKey: ['dashboardStats'] });
      reset();
      setShowAddForm(false);
    },
  });

  const deleteWordMutation = useMutation({
    mutationFn: (wordId: number) => wordlistsApi.deleteWord({id: wordId,wordlistId: wordlist.id}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['words', wordlist.id] });
      queryClient.invalidateQueries({ queryKey: ['wordlists'] });
      queryClient.invalidateQueries({ queryKey: ['dashboardStats'] });
    },
  });

  // Toggle learned mutation
  const toggleLearnedMutation = useMutation({
    mutationFn: ({ wordId, learned }: { wordId: number; learned: boolean }) => 
      toggleWordLearned(wordlist.id, wordId, learned),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['words', wordlist.id] });
      queryClient.invalidateQueries({ queryKey: ['dashboardStats'] });
    },
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
  const filteredWords = words.filter(word => {
    const matchesSearch = 
      word.term.toLowerCase().includes(searchQuery.toLowerCase()) ||
      word.translation.toLowerCase().includes(searchQuery.toLowerCase());
    
    const matchesFilter = 
      filterLearned === 'all' ||
      (filterLearned === 'learned' && word.learned) ||
      (filterLearned === 'unlearned' && !word.learned);
    
    return matchesSearch && matchesFilter;
  });

  const handleDeleteWord = (word: Word) => {
    Alert.alert(
      'Delete Word',
      `Are you sure you want to delete "${word.term}"?`,
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete',
          style: 'destructive',
          onPress: () => deleteWordMutation.mutate(word.id),
        },
      ]
    );
  };

  const handleAddWord = (data: WordFormData) => {
    addWordMutation.mutate(data);
  };

  const renderWordItem = ({ item }: { item: Word }) => (
    <View style={styles.wordCard}>
      <TouchableOpacity
        style={styles.learnedToggle}
        onPress={() => toggleLearnedMutation.mutate({ wordId: item.id, learned: !item.learned })}
      >
        <MaterialIcons
          name={item.learned ? "check-circle" : "radio-button-unchecked"}
          size={24}
          color={item.learned ? "#4CAF50" : "#DFE6E9"}
        />
      </TouchableOpacity>

      <View style={styles.wordContent}>
        <Text style={styles.wordTerm}>{item.term}</Text>
        <Text style={styles.wordTranslation}>{item.translation}</Text>
        {item.pronunciation && (
          <Text style={styles.wordPronunciation}>[{item.pronunciation}]</Text>
        )}
        {item.notes && (
          <Text style={styles.wordNotes}>{item.notes}</Text>
        )}
      </View>

      <TouchableOpacity
        style={styles.deleteButton}
        onPress={() => handleDeleteWord(item)}
      >
        <Ionicons name="trash-outline" size={20} color="#FF6B6B" />
      </TouchableOpacity>
    </View>
  );

  const stats = {
    total: words.length,
    learned: words.filter(w => w.learned).length,
    progress: words.length > 0 ? (words.filter(w => w.learned).length / words.length) * 100 : 0,
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
        <Animated.View 
          style={[
            styles.backdrop,
            { opacity: backdropAnim },
          ]}
        />
      </TouchableWithoutFeedback>

      <KeyboardAvoidingView 
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        style={styles.container}
      >
        <Animated.View
          style={[
            styles.modalContent,
            { transform: [{ translateY: slideAnim }] },
          ]}
        >
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
                onPress={onClose}
              >
                <Ionicons name="close" size={24} color="#636E72" />
              </TouchableOpacity>
            </View>
          </View>

          {/* Stats */}
          <View style={styles.statsRow}>
            <View style={styles.stat}>
              <Text style={styles.statValue}>{stats.total}</Text>
              <Text style={styles.statLabel}>Total</Text>
            </View>
            <View style={styles.stat}>
              <Text style={[styles.statValue, { color: '#4CAF50' }]}>{stats.learned}</Text>
              <Text style={styles.statLabel}>Learned</Text>
            </View>
            <View style={styles.stat}>
              <Text style={[styles.statValue, { color: '#FF7B54' }]}>
                {Math.round(stats.progress)}%
              </Text>
              <Text style={styles.statLabel}>Progress</Text>
            </View>
          </View>

          {/* Search and Filter */}
          <View style={styles.searchContainer}>
            <View style={styles.searchBox}>
              <Ionicons name="search" size={20} color="#636E72" />
              <TextInput
                style={styles.searchInput}
                placeholder="Search words..."
                placeholderTextColor="#B2BEC3"
                value={searchQuery}
                onChangeText={setSearchQuery}
              />
            </View>
          </View>

          <ScrollView horizontal showsHorizontalScrollIndicator={false} style={styles.filterContainer}>
            <TouchableOpacity
              style={[styles.filterChip, filterLearned === 'all' && styles.filterChipActive]}
              onPress={() => setFilterLearned('all')}
            >
              <Text style={[styles.filterChipText, filterLearned === 'all' && styles.filterChipTextActive]}>
                All ({words.length})
              </Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={[styles.filterChip, filterLearned === 'learned' && styles.filterChipActive]}
              onPress={() => setFilterLearned('learned')}
            >
              <Text style={[styles.filterChipText, filterLearned === 'learned' && styles.filterChipTextActive]}>
                Learned ({words.filter(w => w.learned).length})
              </Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={[styles.filterChip, filterLearned === 'unlearned' && styles.filterChipActive]}
              onPress={() => setFilterLearned('unlearned')}
            >
              <Text style={[styles.filterChipText, filterLearned === 'unlearned' && styles.filterChipTextActive]}>
                To Learn ({words.filter(w => !w.learned).length})
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
              renderItem={renderWordItem}
              keyExtractor={(item) => String(item.id)}
              contentContainerStyle={styles.listContent}
              showsVerticalScrollIndicator={false}
              ListEmptyComponent={
                <View style={styles.emptyState}>
                  <MaterialIcons name="library-books" size={48} color="#DFE6E9" />
                  <Text style={styles.emptyText}>
                    {searchQuery ? 'No words found' : 'No words yet'}
                  </Text>
                  {!searchQuery && (
                    <Text style={styles.emptySubtext}>
                      Add your first word to get started
                    </Text>
                  )}
                </View>
              }
            />
          )}

          {/* Add Word Form */}
          {showAddForm && (
            <View style={styles.addFormContainer}>
              <View style={styles.formHeader}>
                <Text style={styles.formTitle}>Add New Word</Text>
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
                name="term"
                rules={{
                  required: 'Term is required',
                  minLength: { value: 1, message: 'Term is too short' },
                }}
                render={({ field: { onChange, onBlur, value } }) => (
                  <View style={styles.inputGroup}>
                    <Text style={styles.inputLabel}>Word/Phrase *</Text>
                    <TextInput
                      style={[styles.input, errors.term && styles.inputError]}
                      placeholder="e.g., Hello"
                      placeholderTextColor="#B2BEC3"
                      value={value}
                      onChangeText={onChange}
                      onBlur={onBlur}
                    />
                    {errors.term && (
                      <Text style={styles.errorText}>{errors.term.message}</Text>
                    )}
                  </View>
                )}
              />

              <Controller
                control={control}
                name="translation"
                rules={{
                  required: 'Translation is required',
                  minLength: { value: 1, message: 'Translation is too short' },
                }}
                render={({ field: { onChange, onBlur, value } }) => (
                  <View style={styles.inputGroup}>
                    <Text style={styles.inputLabel}>Translation *</Text>
                    <TextInput
                      style={[styles.input, errors.translation && styles.inputError]}
                      placeholder="e.g., Hola"
                      placeholderTextColor="#B2BEC3"
                      value={value}
                      onChangeText={onChange}
                      onBlur={onBlur}
                    />
                    {errors.translation && (
                      <Text style={styles.errorText}>{errors.translation.message}</Text>
                    )}
                  </View>
                )}
              />

              <Controller
                control={control}
                name="pronunciation"
                render={({ field: { onChange, onBlur, value } }) => (
                  <View style={styles.inputGroup}>
                    <Text style={styles.inputLabel}>Pronunciation (Optional)</Text>
                    <TextInput
                      style={styles.input}
                      placeholder="e.g., OH-lah"
                      placeholderTextColor="#B2BEC3"
                      value={value}
                      onChangeText={onChange}
                      onBlur={onBlur}
                    />
                  </View>
                )}
              />

              <TouchableOpacity
                style={[styles.submitButton, addWordMutation.isPending && styles.buttonDisabled]}
                onPress={handleSubmit(handleAddWord)}
                disabled={addWordMutation.isPending}
              >
                {addWordMutation.isPending ? (
                  <ActivityIndicator color="#FFFFFF" size="small" />
                ) : (
                  <Text style={styles.submitButtonText}>Add Word</Text>
                )}
              </TouchableOpacity>
            </View>
          )}

          {/* FAB */}
          {!showAddForm && (
            <TouchableOpacity
              style={styles.fab}
              onPress={() => setShowAddForm(true)}
              activeOpacity={0.8}
            >
              <Ionicons name="add" size={28} color="#FFFFFF" />
            </TouchableOpacity>
          )}
        </Animated.View>
      </KeyboardAvoidingView>
    </Modal>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'flex-end',
  },
  backdrop: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0, 0, 0, 0.4)',
  },
  modalContent: {
    backgroundColor: '#FDF6E3',
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    height: SCREEN_HEIGHT * 0.95,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: -4 },
    shadowOpacity: 0.1,
    shadowRadius: 12,
    elevation: 8,
  },
  header: {
    paddingTop: 12,
    paddingBottom: 16,
    paddingHorizontal: 20,
    backgroundColor: '#FFFFFF',
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    borderBottomWidth: 1,
    borderBottomColor: '#F0F0F0',
  },
  handle: {
    width: 40,
    height: 4,
    backgroundColor: '#DFE6E9',
    borderRadius: 2,
    alignSelf: 'center',
    marginBottom: 16,
  },
  titleRow: {
    flexDirection: 'row',
    alignItems: 'center',
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
    fontWeight: '600',
    color: '#2D3436',
  },
  subtitle: {
    fontSize: 14,
    color: '#636E72',
    marginTop: 2,
  },
  closeButton: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: '#F5F5F5',
    justifyContent: 'center',
    alignItems: 'center',
  },
  statsRow: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    paddingVertical: 16,
    paddingHorizontal: 20,
    backgroundColor: '#FFFFFF',
    borderBottomWidth: 1,
    borderBottomColor: '#F0F0F0',
  },
  stat: {
    alignItems: 'center',
  },
  statValue: {
    fontSize: 24,
    fontWeight: '700',
    color: '#2D3436',
  },
  statLabel: {
    fontSize: 14,
    color: '#636E72',
    marginTop: 4,
  },
  searchContainer: {
    padding: 20,
    paddingBottom: 0,
  },
  searchBox: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    paddingHorizontal: 16,
    paddingVertical: 12,
    gap: 12,
  },
  searchInput: {
    flex: 1,
    fontSize: 16,
    color: '#2D3436',
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
    backgroundColor: '#FFFFFF',
    marginRight: 8,
    borderWidth: 1,
    borderColor: '#E0E0E0',
  },
  filterChipActive: {
    backgroundColor: '#FF7B54',
    borderColor: '#FF7B54',
  },
  filterChipText: {
    fontSize: 14,
    color: '#636E72',
  },
  filterChipTextActive: {
    color: '#FFFFFF',
    fontWeight: '500',
  },
  listContent: {
    paddingHorizontal: 20,
    paddingBottom: 100,
  },
  wordCard: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    padding: 16,
    marginBottom: 8,
    shadowColor: '#000',
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
    fontWeight: '600',
    color: '#2D3436',
    marginBottom: 4,
  },
  wordTranslation: {
    fontSize: 16,
    color: '#FF7B54',
    marginBottom: 2,
  },
  wordPronunciation: {
    fontSize: 14,
    color: '#636E72',
    fontStyle: 'italic',
  },
  wordNotes: {
    fontSize: 14,
    color: '#636E72',
    marginTop: 4,
  },
  deleteButton: {
    padding: 8,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  emptyState: {
    alignItems: 'center',
    paddingVertical: 60,
  },
  emptyText: {
    fontSize: 18,
    color: '#636E72',
    marginTop: 16,
  },
  emptySubtext: {
    fontSize: 14,
    color: '#B2BEC3',
    marginTop: 8,
  },
  fab: {
    position: 'absolute',
    right: 20,
    bottom: 20,
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: '#FF7B54',
    justifyContent: 'center',
    alignItems: 'center',
    shadowColor: '#FF7B54',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 8,
  },
  addFormContainer: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    backgroundColor: '#FFFFFF',
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    padding: 20,
    paddingBottom: 40,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: -4 },
    shadowOpacity: 0.1,
    shadowRadius: 12,
    elevation: 8,
  },
  formHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 20,
  },
  formTitle: {
    fontSize: 20,
    fontWeight: '600',
    color: '#2D3436',
  },
  inputGroup: {
    marginBottom: 16,
  },
  inputLabel: {
    fontSize: 14,
    fontWeight: '500',
    color: '#2D3436',
    marginBottom: 8,
  },
  input: {
    borderWidth: 1,
    borderColor: '#E0E0E0',
    borderRadius: 12,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 16,
    color: '#2D3436',
    backgroundColor: '#FAFAFA',
  },
  inputError: {
    borderColor: '#FF6B6B',
  },
  errorText: {
    color: '#FF6B6B',
    fontSize: 12,
    marginTop: 4,
  },
  submitButton: {
    backgroundColor: '#FF7B54',
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
    marginTop: 8,
  },
  submitButtonText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
  },
  buttonDisabled: {
    opacity: 0.7,
  },
});