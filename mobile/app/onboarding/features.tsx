import React, { useMemo, useRef, useState, useCallback } from "react";
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  Dimensions,
  TouchableOpacity,
  Image,
  ViewToken,
} from "react-native";
import { useRouter } from "expo-router";
import { useTheme } from "@/contexts/ThemeContext";
import { useTranslation } from "react-i18next";
import { usePostHog } from "posthog-react-native";
import { OnboardingLayout } from "@/components/onboarding/OnboardingLayout";
import LottieView, { AnimationObject } from "lottie-react-native";

const { width } = Dimensions.get("window");

// Mirror the horizontal padding applied in OnboardingLayout.content for accurate slide sizing.
const LAYOUT_HORIZONTAL_PADDING = 24;
const CAROUSEL_HORIZONTAL_PADDING = 20;
const SLIDE_HORIZONTAL_MARGIN = 12;
const TOTAL_SLIDE_INSETS =
  2 *
  (LAYOUT_HORIZONTAL_PADDING +
    CAROUSEL_HORIZONTAL_PADDING +
    SLIDE_HORIZONTAL_MARGIN);
const slideWidth = Math.max(width - TOTAL_SLIDE_INSETS, 1);

type Slide = {
  key: string;
  title: string;
  subtitle: string;
  bullets: string[];
  icon: any;
  iconAccent: string;
  animation?: string | { uri: string } | AnimationObject;
};

export default function OnboardingFeatures() {
  const { theme } = useTheme();
  const styles = createStyles(theme);
  const router = useRouter();
  const { t } = useTranslation();
  const posthog = usePostHog();

  const slides: Slide[] = useMemo(
    () => [
      {
        key: "quizzes",
        title: t("onboarding.features.quizzes.title", "Adaptive quizzes"),
        subtitle: t(
          "onboarding.features.quizzes.subtitle",
          "Switch between meanings, listening, writing, and picture drills in seconds.",
        ),
        bullets: [
          t(
            "onboarding.features.quizzes.bulletOne",
            "Smart difficulty keeps you challenged.",
          ),
          t(
            "onboarding.features.quizzes.bulletTwo",
            "Varied quiz modes ready for micro-sessions.",
          ),
        ],
        icon: require("@/assets/onboarding-icons/feature-quizzes.png"),
        iconAccent: "#FFB38A",
      },
      {
        key: "chat",
        title: t("onboarding.features.chat.title", "Live speaking coach"),
        subtitle: t(
          "onboarding.features.chat.subtitle",
          "Practice speaking out loud with instant AI guidance.",
        ),
        bullets: [
          t(
            "onboarding.features.chat.bulletOne",
            "Launch voice sessions from any wordlist in seconds.",
          ),
          t(
            "onboarding.features.chat.bulletTwo",
            "Realtime transcripts capture every correction.",
          ),
        ],
        icon: require("@/assets/onboarding-icons/feature-voice-coach.png"),
        iconAccent: "#FCD7F6",
        animation:
          require("@/assets/animations/voice-wave.json") as AnimationObject,
      },
      {
        key: "flashcards",
        title: t("onboarding.features.flashcards.title", "Smart flashcards"),
        subtitle: t(
          "onboarding.features.flashcards.subtitle",
          "Flip through definitions, examples, and audio in one tap.",
        ),
        bullets: [
          t(
            "onboarding.features.flashcards.bulletOne",
            "Rich cards keep context at your fingertips.",
          ),
          t(
            "onboarding.features.flashcards.bulletTwo",
            "Perfect for quick review bursts.",
          ),
        ],
        icon: require("@/assets/onboarding-icons/feature-flashcards.png"),
        iconAccent: "#D1C4E9",
      },
      {
        key: "leitner",
        title: t(
          "onboarding.features.leitner.title",
          "Smart spaced repetition",
        ),
        subtitle: t(
          "onboarding.features.leitner.subtitle",
          "Leitner boxes time your reviews so words stick long-term.",
        ),
        bullets: [
          t(
            "onboarding.features.leitner.bulletOne",
            "Deterministic scheduling keeps practice efficient.",
          ),
          t(
            "onboarding.features.leitner.bulletTwo",
            "Mastery levels rise as you keep streaks alive.",
          ),
        ],
        icon: require("@/assets/onboarding-icons/feature-leitner.png"),
        iconAccent: "#C5CAE9",
      },
      {
        key: "progress",
        title: t("onboarding.features.progress.title", "Track your progress"),
        subtitle: t(
          "onboarding.features.progress.subtitle",
          "Celebrate streaks, mastery levels, and weekly momentum with beautiful charts.",
        ),
        bullets: [
          t(
            "onboarding.features.progress.bulletOne",
            "Daily goal reminders keep the streak alive.",
          ),
          t(
            "onboarding.features.progress.bulletTwo",
            "See strengths and gaps by wordlist instantly.",
          ),
        ],
        icon: require("@/assets/onboarding-icons/feature-progress.png"),
        iconAccent: "#C8E6C9",
      },
      {
        key: "enrichment",
        title: t("onboarding.features.enrichment.title", "AI enrichment pack"),
        subtitle: t(
          "onboarding.features.enrichment.subtitle",
          "Definitions, examples, visuals, and audio generated for every word.",
        ),
        bullets: [
          t(
            "onboarding.features.enrichment.bulletOne",
            "Native-quality pronunciations in your target language.",
          ),
          t(
            "onboarding.features.enrichment.bulletTwo",
            "Examples and visuals that anchor meaning fast.",
          ),
        ],
        icon: require("@/assets/onboarding-icons/feature-enrichment.png"),
        iconAccent: "#FFE0B2",
      },
      {
        key: "images",
        title: t("onboarding.features.images.title", "AI visuals & audio"),
        subtitle: t(
          "onboarding.features.images.subtitle",
          "Every word gets context with imagery, pronunciation, and cultural nuance.",
        ),
        bullets: [
          t(
            "onboarding.features.images.bulletOne",
            "Pronunciations voiced by native accents.",
          ),
          t(
            "onboarding.features.images.bulletTwo",
            "Illustrations generated for your language.",
          ),
        ],
        icon: require("@/assets/onboarding-icons/feature-visuals.png"),
        iconAccent: "#FFECB3",
      },
      {
        key: "offline",
        title: t("onboarding.features.offline.title", "Offline mode"),
        subtitle: t(
          "onboarding.features.offline.subtitle",
          "Premium lets you practice anywhere with seamless sync.",
        ),
        bullets: [
          t(
            "onboarding.features.offline.bulletOne",
            "Download wordlists and keep learning offline.",
          ),
          t(
            "onboarding.features.offline.bulletTwo",
            "Progress syncs automatically when you reconnect.",
          ),
        ],
        icon: require("@/assets/onboarding-icons/feature-offline.png"),
        iconAccent: "#B2EBF2",
      },
      {
        key: "reporting",
        title: t("onboarding.features.reporting.title", "Report & regenerate"),
        subtitle: t(
          "onboarding.features.reporting.subtitle",
          "Flag AI mistakes and we regenerate the content fast.",
        ),
        bullets: [
          t(
            "onboarding.features.reporting.bulletOne",
            "Report wrong meanings, images, or audio.",
          ),
          t(
            "onboarding.features.reporting.bulletTwo",
            "Quality improves with every report you send.",
          ),
        ],
        icon: require("@/assets/onboarding-icons/feature-reporting.png"),
        iconAccent: "#FFCDD2",
      },
    ],
    [t],
  );

  const [index, setIndex] = useState(0);
  const ref = useRef<FlatList<Slide>>(null);

  const onViewableItemsChanged = useCallback(
    ({ viewableItems }: { viewableItems: ViewToken[] }) => {
      const visible = viewableItems.find((item) => item.isViewable);
      if (!visible || typeof visible.index !== "number") return;
      if (visible.index !== index) {
        setIndex(visible.index);
        const slide = slides[visible.index];
        if (slide) {
          posthog.capture("onboarding_feature_viewed", {
            slide: slide.key,
          });
        }
      }
    },
    [index, slides, posthog],
  );

  const viewabilityConfig = useRef({
    viewAreaCoveragePercentThreshold: 60,
  }).current;

  const onNext = () => {
    const next = Math.min(index + 1, slides.length - 1);
    if (next !== index) {
      ref.current?.scrollToIndex({ index: next, animated: true });
      setIndex(next);
      posthog.capture("onboarding_feature_viewed", { slide: slides[next].key });
    } else {
      router.replace("/onboarding/account");
    }
  };

  const onSkip = () => router.replace("/onboarding/account");

  const nextLabel =
    index === slides.length - 1
      ? t("common.continue", "Continue")
      : t("common.next", "Next");

  return (
    <OnboardingLayout
      step={2}
      totalSteps={3}
      showSkip
      skipLabel={t("common.skip", "Skip")}
      onSkip={onSkip}
      showBack
      backLabel={t("common.back", "Back")}
      onBack={() => router.replace("/onboarding")}
      stepLabel={t("onboarding.stepIndicator", {
        step: 2,
        total: 3,
        defaultValue: "Step 2 of 3",
      })}
      contentStyle={styles.content}
      footer={
        <TouchableOpacity
          onPress={onNext}
          style={styles.primaryButton}
          accessibilityRole="button"
        >
          <Text style={styles.primaryText}>{nextLabel}</Text>
        </TouchableOpacity>
      }
    >
      <View>
        <Text style={styles.heading}>
          {t("onboarding.features.heading", "See what you can do")}
        </Text>
        <Text style={styles.description}>
          {t(
            "onboarding.features.description",
            "Swipe through the highlights to discover how Decorebator keeps you inspired every day.",
          )}
        </Text>
      </View>
      <FlatList
        ref={ref}
        data={slides}
        horizontal
        pagingEnabled
        showsHorizontalScrollIndicator={false}
        keyExtractor={(item) => item.key}
        renderItem={({ item }) => (
          <View
            style={[
              styles.slide,
              {
                width: slideWidth,
                backgroundColor: theme.colors.background.surface,
              },
            ]}
          >
            <View style={styles.headerRow}>
              <View
                style={[
                  styles.iconBadge,
                  {
                    backgroundColor: theme.colors.background.surface,
                    borderColor: item.iconAccent,
                  },
                ]}
              >
                <Image source={item.icon} style={styles.iconImage} />
              </View>
              <View style={styles.titleContainer}>
                <Text style={styles.title}>{item.title}</Text>
              </View>
            </View>
            <Text style={styles.subtitle}>{item.subtitle}</Text>
            {item.animation ? (
              <View style={styles.animationContainer}>
                <LottieView
                  source={item.animation}
                  autoPlay
                  loop
                  style={styles.animation}
                />
              </View>
            ) : null}
            <View style={styles.bulletList}>
              {item.bullets.map((bullet) => (
                <View key={bullet} style={styles.bulletRow}>
                  <View
                    style={[
                      styles.bulletDot,
                      { backgroundColor: theme.colors.primary },
                    ]}
                  />
                  <Text style={styles.bulletText}>{bullet}</Text>
                </View>
              ))}
            </View>
          </View>
        )}
        contentContainerStyle={styles.carouselContent}
        snapToAlignment="center"
        decelerationRate="fast"
        onViewableItemsChanged={onViewableItemsChanged}
        viewabilityConfig={viewabilityConfig}
      />
      <View style={styles.pagination}>
        {slides.map((_, i) => (
          <View
            key={i}
            style={[styles.dot, i === index && styles.dotActive]}
            accessibilityRole="image"
            accessibilityLabel={`Slide ${i + 1} of ${slides.length}`}
            accessibilityState={{ selected: i === index }}
          />
        ))}
      </View>
    </OnboardingLayout>
  );
}

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    content: {
      paddingBottom: 12,
      gap: 18,
    },
    heading: {
      fontSize: 24,
      fontWeight: "700",
      color: theme.colors.text.primary,
    },
    description: {
      marginTop: 6,
      fontSize: 15,
      color: theme.colors.text.secondary,
      lineHeight: 22,
    },
    carouselContent: {
      paddingVertical: 12,
      paddingHorizontal: CAROUSEL_HORIZONTAL_PADDING,
    },
    slide: {
      borderRadius: 24,
      padding: 24,
      marginHorizontal: SLIDE_HORIZONTAL_MARGIN,
      shadowColor: "#000",
      shadowOpacity: 0.08,
      shadowOffset: { width: 0, height: 12 },
      shadowRadius: 24,
      elevation: 3,
    },
    iconBadge: {
      width: 58,
      height: 58,
      borderRadius: 29,
      alignItems: "center",
      justifyContent: "center",
      borderWidth: 1,
    },
    iconImage: {
      width: 38,
      height: 38,
      resizeMode: "contain",
    },
    headerRow: {
      flexDirection: "row",
      alignItems: "center",
      gap: 14,
      marginBottom: 12,
    },
    titleContainer: {
      flex: 1,
    },
    title: {
      fontSize: 22,
      fontWeight: "700",
      color: theme.colors.text.primary,
      flexShrink: 1,
    },
    subtitle: {
      fontSize: 15,
      color: theme.colors.text.secondary,
      lineHeight: 22,
      marginBottom: 14,
    },
    animationContainer: {
      height: 96,
      justifyContent: "center",
      alignItems: "center",
      marginBottom: 10,
    },
    animation: {
      width: "100%",
      height: "100%",
    },
    bulletList: {
      gap: 10,
    },
    bulletRow: {
      flexDirection: "row",
      alignItems: "flex-start",
      gap: 10,
    },
    bulletDot: {
      width: 8,
      height: 8,
      borderRadius: 4,
      marginTop: 7,
    },
    bulletText: {
      flex: 1,
      fontSize: 14,
      color: theme.colors.text.primary,
      lineHeight: 20,
    },
    pagination: {
      flexDirection: "row",
      justifyContent: "center",
      gap: 6,
    },
    dot: {
      width: 8,
      height: 8,
      borderRadius: 4,
      backgroundColor: theme.colors.ui.divider,
    },
    dotActive: {
      backgroundColor: theme.colors.primary,
      width: 20,
    },
    primaryButton: {
      backgroundColor: theme.colors.primary,
      paddingVertical: 16,
      borderRadius: 18,
      alignItems: "center",
      marginTop: 12,
      shadowColor: "#FF7B54",
      shadowOpacity: 0.32,
      shadowOffset: { width: 0, height: 10 },
      shadowRadius: 18,
      elevation: 3,
    },
    primaryText: {
      color: theme.colors.text.inverse,
      fontWeight: "700",
      fontSize: 16,
    },
  });
