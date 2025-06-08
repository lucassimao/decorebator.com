import React, { useRef, useEffect } from "react";
import {
  View,
  Text,
  TouchableOpacity,
  Animated,
  ScrollView,
  ActivityIndicator,
  StyleSheet,
} from "react-native";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { Definition, Word } from "../../api/wordlists";

interface FlashcardContentProps {
  currentWord: Word;
  definitions: Definition[];
  isFlipped: boolean;
  shouldFetchDefinitions: boolean;
  loadingDefinitions: boolean;
  definitionsError: any;
  slideAnimation: Animated.Value;
  scaleAnimation: Animated.Value;
  onFlip: () => void;
  onPlayAudio: () => void;
  onRefetchDefinitions: () => void;
}

const colors = {
  primary: "#FF7B54",
  error: "#FF6B6B",
  backgroundPeach: "#FFE8D6",
  textDark: "#2D3436",
  textMedium: "#636E72",
  textLight: "#B2BEC3",
  white: "#FFFFFF",
  divider: "#F0F0F0",
};

export const FlashcardContent: React.FC<FlashcardContentProps> = ({
  currentWord,
  definitions,
  isFlipped,
  shouldFetchDefinitions,
  loadingDefinitions,
  definitionsError,
  slideAnimation,
  scaleAnimation,
  onFlip,
  onPlayAudio,
  onRefetchDefinitions,
}) => {
  const { t } = useTranslation();
  const flipAnimation = useRef(new Animated.Value(0)).current;

  // Reset flip animation when word changes
  useEffect(() => {
    flipAnimation.setValue(0);
  }, [currentWord?.id]);

  // Handle flip animation
  useEffect(() => {
    Animated.timing(flipAnimation, {
      toValue: isFlipped ? 180 : 0,
      duration: 600,
      useNativeDriver: true,
    }).start();
  }, [isFlipped, flipAnimation]);

  // Animation interpolations
  const frontInterpolate = flipAnimation.interpolate({
    inputRange: [0, 180],
    outputRange: ["0deg", "180deg"],
  });

  const backInterpolate = flipAnimation.interpolate({
    inputRange: [0, 180],
    outputRange: ["180deg", "360deg"],
  });

  const frontAnimatedStyle = {
    transform: [{ rotateY: frontInterpolate }],
  };

  const backAnimatedStyle = {
    transform: [{ rotateY: backInterpolate }],
  };

  const renderDefinitionsContent = () => {
    if (loadingDefinitions) {
      return (
        <View style={styles.loadingDefinitionsContainer}>
          <ActivityIndicator size="large" color={colors.primary} />
          <Text style={styles.loadingDefinitionsText}>
            Loading definitions...
          </Text>
        </View>
      );
    }

    if (definitionsError) {
      return (
        <View style={styles.errorDefinitionsContainer}>
          <MaterialIcons name="error-outline" size={48} color={colors.error} />
          <Text style={styles.errorDefinitionsText}>
            Failed to load definitions
          </Text>
          <TouchableOpacity
            style={styles.retryButton}
            onPress={onRefetchDefinitions}
          >
            <Text style={styles.retryButtonText}>Retry</Text>
          </TouchableOpacity>
        </View>
      );
    }

    if (definitions.length > 0) {
      return definitions.map((definition, index) => (
        <View key={definition.id} style={styles.definitionBlock}>
          {index > 0 && <View style={styles.definitionDivider} />}

          {definition.partOfSpeech && (
            <View style={styles.partOfSpeechBadge}>
              <Text style={styles.partOfSpeechText}>
                {definition.partOfSpeech}
              </Text>
            </View>
          )}

          <Text style={styles.meaningText}>{definition.meaning}</Text>

          {currentWord?.pronunciation && (
            <Text style={styles.phoneticText}>
              /{currentWord?.pronunciation}/
            </Text>
          )}

          {(() => {
            const isVerb = definition.isVerbType || false;
            const hasInflections =
              definition.inflections && definition.inflections.length > 0;
            const hasExamples =
              definition.examples && definition.examples.length > 0;

            if (isVerb && hasInflections) {
              const allInflectionExamples = definition.inflections!.flatMap(
                (inf) =>
                  inf.examples.map((ex) => ({
                    example: ex,
                    tense: inf.tense,
                  })),
              );

              return (
                allInflectionExamples.length > 0 && (
                  <View style={styles.examplesContainer}>
                    <Text style={styles.examplesTitle}>
                      {t("flashcards.examples")}:
                    </Text>
                    {allInflectionExamples.map((item, idx) => (
                      <View key={idx} style={styles.inflectionExampleContainer}>
                        <Text style={styles.exampleText}>
                          • {item.example.replace(/\[|\]/g, "")}
                        </Text>
                        <Text style={styles.tenseLabel}>({item.tense})</Text>
                      </View>
                    ))}
                  </View>
                )
              );
            } else if (hasExamples) {
              return (
                <View style={styles.examplesContainer}>
                  <Text style={styles.examplesTitle}>
                    {t("flashcards.examples")}:
                  </Text>
                  {definition.examples!.map((example, idx) => (
                    <Text key={idx} style={styles.exampleText}>
                      • {example.replace(/\[|\]/g, "")}
                    </Text>
                  ))}
                </View>
              );
            }

            return null;
          })()}
        </View>
      ));
    }

    if (shouldFetchDefinitions) {
      return (
        <View style={styles.noDefinitionsContainer}>
          <Text style={styles.noDefinitionsText}>
            No definitions available for this word.
          </Text>
        </View>
      );
    }

    return (
      <View style={styles.flipPromptContainer}>
        <Text style={styles.flipPromptText}>{t("flashcards.tapToFlip")}</Text>
      </View>
    );
  };

  return (
    <View style={styles.cardContainer}>
      <View style={styles.cardTouchable}>
        {/* Front of card */}
        <TouchableOpacity
          activeOpacity={0.9}
          onPress={onFlip}
          style={[
            styles.card,
            frontAnimatedStyle,
            {
              transform: [
                { translateX: slideAnimation },
                { scale: scaleAnimation },
                { rotateY: frontInterpolate },
              ],
            },
            !isFlipped ? styles.cardFront : null,
          ]}
        >
          <View style={styles.cardContent}>
            <Text style={styles.wordText}>{currentWord?.name}</Text>
            {currentWord?.audioURL && (
              <TouchableOpacity
                style={styles.audioButton}
                onPress={onPlayAudio}
              >
                <Ionicons name="volume-high" size={32} color={colors.primary} />
              </TouchableOpacity>
            )}
          </View>
          <Text style={styles.flipHint}>{t("flashcards.tapToFlip")}</Text>
        </TouchableOpacity>

        {/* Back of card */}
        <Animated.View
          style={[
            styles.card,
            styles.cardBack,
            backAnimatedStyle,
            {
              transform: [
                { translateX: slideAnimation },
                { scale: scaleAnimation },
                { rotateY: backInterpolate },
              ],
            },
            isFlipped ? styles.cardBackVisible : null,
          ]}
        >
          <ScrollView
            style={styles.definitionsScroll}
            contentContainerStyle={styles.definitionsScrollContent}
            showsVerticalScrollIndicator={true}
            bounces={true}
            alwaysBounceVertical={false}
            nestedScrollEnabled={true}
          >
            {renderDefinitionsContent()}
          </ScrollView>
          <TouchableOpacity
            style={styles.flipHintTouchable}
            onPress={onFlip}
            activeOpacity={0.7}
          >
            <Text style={styles.flipHint}>{t("flashcards.tapToFlipBack")}</Text>
          </TouchableOpacity>
        </Animated.View>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  cardContainer: {
    flex: 1,
    paddingHorizontal: 20,
    justifyContent: "center",
  },
  cardTouchable: {
    height: 400,
    position: "relative",
  },
  card: {
    position: "absolute",
    width: "100%",
    height: "100%",
    backgroundColor: colors.white,
    borderRadius: 20,
    padding: 24,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: 12,
    elevation: 5,
    backfaceVisibility: "hidden",
  },
  cardFront: {
    zIndex: 2,
  },
  cardBack: {
    zIndex: 1,
  },
  cardBackVisible: {
    zIndex: 3,
  },
  cardContent: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  wordText: {
    fontSize: 48,
    fontWeight: "bold",
    color: colors.textDark,
    textAlign: "center",
    marginBottom: 20,
  },
  audioButton: {
    width: 60,
    height: 60,
    borderRadius: 30,
    backgroundColor: colors.backgroundPeach,
    justifyContent: "center",
    alignItems: "center",
  },
  flipHint: {
    fontSize: 14,
    color: colors.textLight,
    fontStyle: "italic",
    textAlign: "center",
  },
  flipHintTouchable: {
    position: "absolute",
    bottom: 10,
    left: 0,
    right: 0,
    paddingVertical: 10,
    paddingHorizontal: 20,
    alignItems: "center",
  },
  definitionsScroll: {
    flex: 1,
    marginBottom: 40,
  },
  definitionsScrollContent: {
    flexGrow: 1,
    paddingBottom: 20,
  },
  definitionBlock: {
    marginBottom: 20,
  },
  definitionDivider: {
    height: 1,
    backgroundColor: colors.divider,
    marginBottom: 16,
  },
  partOfSpeechBadge: {
    backgroundColor: colors.backgroundPeach,
    paddingHorizontal: 12,
    paddingVertical: 4,
    borderRadius: 12,
    alignSelf: "flex-start",
    marginBottom: 8,
  },
  partOfSpeechText: {
    fontSize: 12,
    color: colors.primary,
    fontWeight: "600",
  },
  meaningText: {
    fontSize: 20,
    color: colors.textDark,
    lineHeight: 28,
    marginBottom: 8,
  },
  phoneticText: {
    fontSize: 16,
    color: colors.textMedium,
    fontStyle: "italic",
    marginBottom: 12,
  },
  examplesContainer: {
    marginTop: 12,
  },
  examplesTitle: {
    fontSize: 14,
    fontWeight: "600",
    color: colors.textMedium,
    marginBottom: 8,
  },
  exampleText: {
    fontSize: 16,
    color: colors.textDark,
    lineHeight: 24,
    marginBottom: 6,
    paddingLeft: 8,
  },
  inflectionExampleContainer: {
    flexDirection: "row",
    alignItems: "flex-start",
    marginBottom: 6,
    paddingLeft: 8,
  },
  tenseLabel: {
    fontSize: 12,
    color: colors.textMedium,
    fontStyle: "italic",
    marginLeft: 8,
    marginTop: 2,
  },
  loadingDefinitionsContainer: {
    minHeight: 200,
    justifyContent: "center",
    alignItems: "center",
    paddingVertical: 40,
  },
  loadingDefinitionsText: {
    fontSize: 16,
    color: colors.textMedium,
    marginTop: 16,
    textAlign: "center",
  },
  noDefinitionsContainer: {
    minHeight: 200,
    justifyContent: "center",
    alignItems: "center",
    paddingVertical: 40,
  },
  noDefinitionsText: {
    fontSize: 16,
    color: colors.textMedium,
    textAlign: "center",
    fontStyle: "italic",
  },
  errorDefinitionsContainer: {
    minHeight: 200,
    justifyContent: "center",
    alignItems: "center",
    paddingVertical: 40,
  },
  errorDefinitionsText: {
    fontSize: 16,
    color: colors.error,
    textAlign: "center",
    marginTop: 16,
    marginBottom: 20,
  },
  retryButton: {
    backgroundColor: colors.primary,
    paddingHorizontal: 20,
    paddingVertical: 10,
    borderRadius: 20,
  },
  retryButtonText: {
    color: colors.white,
    fontSize: 14,
    fontWeight: "600",
  },
  flipPromptContainer: {
    minHeight: 200,
    justifyContent: "center",
    alignItems: "center",
    paddingVertical: 40,
  },
  flipPromptText: {
    fontSize: 18,
    color: colors.textLight,
    textAlign: "center",
    fontStyle: "italic",
  },
});
