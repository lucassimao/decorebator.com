import React, { useRef, useEffect } from "react";
import {
  View,
  Text,
  StyleSheet,
  Modal,
  TouchableOpacity,
  ScrollView,
  Animated,
  Dimensions,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";

const { height: SCREEN_HEIGHT } = Dimensions.get("window");

interface ContentGuidelinesModalProps {
  visible: boolean;
  onClose: () => void;
}

export const ContentGuidelinesModal: React.FC<ContentGuidelinesModalProps> = ({
  visible,
  onClose,
}) => {
  const slideAnim = useRef(new Animated.Value(SCREEN_HEIGHT)).current;
  const backdropAnim = useRef(new Animated.Value(0)).current;
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);

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

  if (!visible) return null;

  const guidelines = [
    {
      icon: "school-outline",
      title: t("contentGuidelines.educational.title"),
      description: t("contentGuidelines.educational.description"),
    },
    {
      icon: "people-outline",
      title: t("contentGuidelines.familyFriendly.title"),
      description: t("contentGuidelines.familyFriendly.description"),
    },
    {
      icon: "ban-outline",
      title: t("contentGuidelines.noProfanity.title"),
      description: t("contentGuidelines.noProfanity.description"),
    },
    {
      icon: "shield-outline",
      title: t("contentGuidelines.noPersonalInfo.title"),
      description: t("contentGuidelines.noPersonalInfo.description"),
    },
    {
      icon: "globe-outline",
      title: t("contentGuidelines.respectful.title"),
      description: t("contentGuidelines.respectful.description"),
    },
    {
      icon: "book-outline",
      title: t("contentGuidelines.legitimate.title"),
      description: t("contentGuidelines.legitimate.description"),
    },
  ];

  return (
    <Modal
      transparent
      visible={visible}
      animationType="none"
      onRequestClose={onClose}
      accessibilityViewIsModal={true}
    >
      <TouchableOpacity
        style={styles.backdrop}
        activeOpacity={1}
        onPress={onClose}
      >
        <Animated.View
          style={[
            styles.backdrop,
            {
              opacity: backdropAnim,
            },
          ]}
        />
      </TouchableOpacity>

      <Animated.View
        style={[
          styles.modalContent,
          {
            transform: [{ translateY: slideAnim }],
          },
        ]}
      >
        {/* Header */}
        <View style={styles.header}>
          <View style={styles.handle} />
          <Text
            style={styles.title}
            accessibilityRole="header"
            accessibilityLabel={t("contentGuidelines.title")}
          >
            {t("contentGuidelines.title")}
          </Text>
          <TouchableOpacity
            style={styles.closeButton}
            onPress={onClose}
            accessibilityRole="button"
            accessibilityLabel={t("common.close")}
            accessibilityHint="Close content guidelines"
          >
            <Ionicons
              name="close"
              size={24}
              color={theme.colors.text.secondary}
            />
          </TouchableOpacity>
        </View>

        <ScrollView
          style={styles.content}
          showsVerticalScrollIndicator={false}
          accessibilityLabel="Content guidelines list"
        >
          <Text style={styles.description} accessibilityRole="text">
            {t("contentGuidelines.description")}
          </Text>

          {guidelines.map((guideline, index) => (
            <View
              key={index}
              style={styles.guidelineItem}
              accessibilityRole="text"
            >
              <View style={styles.guidelineHeader}>
                <Ionicons
                  name={guideline.icon as any}
                  size={24}
                  color={theme.colors.primary}
                  style={styles.guidelineIcon}
                />
                <Text style={styles.guidelineTitle}>{guideline.title}</Text>
              </View>
              <Text style={styles.guidelineDescription}>
                {guideline.description}
              </Text>
            </View>
          ))}

          <View style={styles.footer}>
            <View style={styles.footerIcon}>
              <Ionicons
                name="checkmark-circle"
                size={32}
                color={theme.colors.success}
              />
            </View>
            <Text style={styles.footerText} accessibilityRole="text">
              {t("contentGuidelines.footer")}
            </Text>
          </View>
        </ScrollView>

        <View style={styles.actions}>
          <TouchableOpacity
            style={styles.understandButton}
            onPress={onClose}
            accessibilityRole="button"
            accessibilityLabel={t("contentGuidelines.understand")}
            accessibilityHint="Confirm understanding of content guidelines"
          >
            <Text style={styles.understandButtonText}>
              {t("contentGuidelines.understand")}
            </Text>
            <Ionicons
              name="checkmark"
              size={20}
              color={theme.colors.text.inverse}
            />
          </TouchableOpacity>
        </View>
      </Animated.View>
    </Modal>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
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
      position: "absolute",
      bottom: 0,
      left: 0,
      right: 0,
      backgroundColor: theme.colors.background.surface,
      borderTopLeftRadius: 24,
      borderTopRightRadius: 24,
      maxHeight: SCREEN_HEIGHT * 0.9,
      shadowColor: theme.colors.text.primary,
      shadowOffset: { width: 0, height: -4 },
      shadowOpacity: 0.1,
      shadowRadius: 12,
      elevation: 8,
    },
    header: {
      paddingTop: 12,
      paddingBottom: 20,
      paddingHorizontal: 20,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.ui.border,
    },
    handle: {
      width: 40,
      height: 4,
      backgroundColor: theme.colors.ui.disabled,
      borderRadius: 2,
      alignSelf: "center",
      marginBottom: 16,
    },
    title: {
      fontSize: 24,
      fontWeight: "600",
      color: theme.colors.text.primary,
      textAlign: "center",
    },
    closeButton: {
      position: "absolute",
      right: 20,
      top: 32,
      width: 32,
      height: 32,
      borderRadius: 16,
      backgroundColor: theme.colors.background.default,
      justifyContent: "center",
      alignItems: "center",
    },
    content: {
      flex: 1,
      padding: 20,
    },
    description: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      lineHeight: 24,
      marginBottom: 24,
      textAlign: "center",
    },
    guidelineItem: {
      marginBottom: 20,
      backgroundColor:
        theme.mode === "light" ? "#FAFAFA" : theme.colors.background.secondary,
      borderRadius: 12,
      padding: 16,
      borderLeftWidth: 4,
      borderLeftColor: theme.colors.primary,
    },
    guidelineHeader: {
      flexDirection: "row",
      alignItems: "center",
      marginBottom: 8,
    },
    guidelineIcon: {
      marginRight: 12,
    },
    guidelineTitle: {
      fontSize: 16,
      fontWeight: "600",
      color: theme.colors.text.primary,
      flex: 1,
    },
    guidelineDescription: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      lineHeight: 20,
      marginLeft: 36,
    },
    footer: {
      alignItems: "center",
      marginTop: 20,
      marginBottom: 20,
    },
    footerIcon: {
      marginBottom: 12,
    },
    footerText: {
      fontSize: 16,
      color: theme.colors.success,
      textAlign: "center",
      fontWeight: "500",
    },
    actions: {
      padding: 20,
      borderTopWidth: 1,
      borderTopColor: theme.colors.ui.border,
    },
    understandButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: 12,
      paddingVertical: 16,
      paddingHorizontal: 24,
      flexDirection: "row",
      justifyContent: "center",
      alignItems: "center",
      shadowColor: theme.colors.primary,
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.2,
      shadowRadius: 8,
      elevation: 5,
    },
    understandButtonText: {
      color: theme.colors.text.inverse,
      fontSize: 16,
      fontWeight: "600",
      marginRight: 8,
    },
  });
