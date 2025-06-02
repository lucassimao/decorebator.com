import React, { useEffect, useRef } from "react";
import { View, Text, StyleSheet, Animated } from "react-native";
import { MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useOffline } from "@/hooks/useOffline";

export const OfflineIndicator: React.FC = () => {
  const { t } = useTranslation();
  const { isOnline, isOfflineAvailable } = useOffline();
  const slideAnim = useRef(new Animated.Value(-60)).current;

  useEffect(() => {
    if (!isOnline) {
      // Slide in
      Animated.timing(slideAnim, {
        toValue: 0,
        duration: 300,
        useNativeDriver: true,
      }).start();
    } else {
      // Slide out
      Animated.timing(slideAnim, {
        toValue: -60,
        duration: 300,
        useNativeDriver: true,
      }).start();
    }
  }, [isOnline, slideAnim]);

  if (isOnline) return null;

  return (
    <Animated.View
      style={[
        styles.container,
        {
          transform: [{ translateY: slideAnim }],
          backgroundColor: isOfflineAvailable ? "#FF7B54" : "#636E72",
        },
      ]}
    >
      <MaterialIcons
        name={isOfflineAvailable ? "offline-bolt" : "wifi-off"}
        size={20}
        color="#FFFFFF"
      />
      <Text style={styles.text}>
        {isOfflineAvailable
          ? t("offline.availableMessage")
          : t("offline.unavailableMessage")}
      </Text>
    </Animated.View>
  );
};

const styles = StyleSheet.create({
  container: {
    position: "absolute",
    top: 0,
    left: 0,
    right: 0,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    paddingVertical: 8,
    paddingHorizontal: 16,
    zIndex: 1000,
  },
  text: {
    color: "#FFFFFF",
    fontSize: 14,
    fontWeight: "500",
    marginLeft: 8,
  },
});
