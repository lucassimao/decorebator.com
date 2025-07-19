import React from "react";
import { View, Text, Image, StyleSheet } from "react-native";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";

interface LoginHeaderProps {
  keyboardVisible: boolean;
}

export const LoginHeader: React.FC<LoginHeaderProps> = ({
  keyboardVisible,
}) => {
  const { t } = useTranslation();
  const { theme, responsive } = useTheme();

  const styles = StyleSheet.create({
    headerSection: {
      alignItems: "center",
      marginBottom: responsive.spacing.elementSpacing,
    },
    appName: {
      fontSize: responsive.fontSizes.title,
      fontWeight: "700",
      color: theme.colors.text.primary,
      textAlign: "center",
      marginBottom: responsive.spacing.elementSpacing / 2,
    },
    tagline: {
      fontSize: responsive.fontSizes.headline,
      color: theme.colors.text.secondary,
      textAlign: "center",
    },
    illustrationContainer: {
      height: responsive.screenHeight * 0.15,
      justifyContent: "center",
      alignItems: "center",
      marginBottom: responsive.spacing.elementSpacing,
    },
    illustration: {
      width: responsive.screenWidth * 0.8,
      height: "100%",
      maxWidth: 350,
    },
  });

  if (keyboardVisible) {
    return null;
  }

  return (
    <>
      <View style={styles.headerSection}>
        <Text style={styles.appName}>{t("common.appName")}</Text>
        <Text style={styles.tagline}>{t("auth.tagline")}</Text>
      </View>

      {/* Hide foreground image on small devices to save space */}
      {!responsive.isSmallPhone && (
        <View style={styles.illustrationContainer}>
          <Image
            source={require("@/assets/images/login-fg.png")}
            style={styles.illustration}
            resizeMode="contain"
          />
        </View>
      )}
    </>
  );
};
