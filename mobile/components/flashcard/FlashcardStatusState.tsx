import React from "react";
import { ScrollView, StyleSheet, View } from "react-native";
import { MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";

import { Button, UiText } from "@/components/ui";
import { useTheme } from "@/contexts/ThemeContext";

interface FlashcardStatusStateProps {
  icon: React.ComponentProps<typeof MaterialIcons>["name"];
  title: string;
  message: string;
  onBack: () => void;
  onRetry?: () => void;
  assertive?: boolean;
}

export function FlashcardStatusState({
  icon,
  title,
  message,
  onBack,
  onRetry,
  assertive = false,
}: FlashcardStatusStateProps) {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);

  return (
    <ScrollView contentContainerStyle={styles.scrollContent}>
      <View
        testID="flashcard-status-content"
        style={styles.content}
        accessibilityLiveRegion={assertive ? "assertive" : "polite"}
      >
        <View
          style={styles.iconRing}
          accessibilityElementsHidden
          importantForAccessibility="no-hide-descendants"
        >
          <MaterialIcons
            name={icon}
            size={38}
            color={
              assertive ? theme.colors.roles.error : theme.colors.roles.action
            }
          />
        </View>
        <UiText variant="title" accessibilityRole="header" style={styles.title}>
          {title}
        </UiText>
        <UiText tone="secondary" style={styles.message}>
          {message}
        </UiText>
        <View style={styles.actions}>
          {onRetry ? (
            <Button onPress={onRetry} style={styles.action}>
              {t("common.tryAgain")}
            </Button>
          ) : null}
          <Button variant="secondary" onPress={onBack} style={styles.action}>
            {t("flashcards.backToWordlists")}
          </Button>
        </View>
      </View>
    </ScrollView>
  );
}

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    scrollContent: {
      flexGrow: 1,
      justifyContent: "center",
      padding: theme.spacing.lg,
    },
    content: { alignItems: "center" },
    iconRing: {
      width: 72,
      height: 72,
      alignItems: "center",
      justifyContent: "center",
      borderWidth: 2,
      borderColor: theme.colors.roles.controlBorder,
      borderRadius: theme.borderRadius.full,
      backgroundColor: theme.colors.background.surface,
    },
    title: {
      marginTop: theme.spacing.lg,
      textAlign: "center",
    },
    message: {
      maxWidth: 340,
      marginTop: theme.spacing.sm,
      textAlign: "center",
    },
    actions: {
      width: "100%",
      maxWidth: 320,
      gap: theme.spacing.compact,
      marginTop: theme.spacing.lg,
    },
    action: { width: "100%" },
  });
