import React, { useEffect, useRef } from "react";
import {
  Animated,
  Modal,
  StyleSheet,
  Text,
  TouchableOpacity,
  TouchableWithoutFeedback,
  View,
} from "react-native";
import { MaterialIcons } from "@expo/vector-icons";
import { router } from "expo-router";
import { useTranslation } from "react-i18next";

type Props = {
  visible: boolean;
  onDismiss: () => void;
};

const UpgradePromptDialog = ({ visible, onDismiss }: Props) => {
  const scaleAnim = useRef(new Animated.Value(0)).current;
  const opacityAnim = useRef(new Animated.Value(0)).current;
  const { t } = useTranslation();

  useEffect(() => {
    if (visible) {
      Animated.parallel([
        Animated.spring(scaleAnim, {
          toValue: 1,
          friction: 8,
          tension: 40,
          useNativeDriver: true,
        }),
        Animated.timing(opacityAnim, {
          toValue: 1,
          duration: 200,
          useNativeDriver: true,
        }),
      ]).start();
    } else {
      Animated.parallel([
        Animated.timing(scaleAnim, {
          toValue: 0,
          duration: 200,
          useNativeDriver: true,
        }),
        Animated.timing(opacityAnim, {
          toValue: 0,
          duration: 200,
          useNativeDriver: true,
        }),
      ]).start();
    }
  }, [visible]);

  const handleUpgrade = () => {
    onDismiss();
    router.push({ pathname: "/settings" });
  };

  return (
    <Modal
      transparent
      visible={visible}
      animationType="none"
      onRequestClose={onDismiss}
    >
      <TouchableWithoutFeedback onPress={onDismiss}>
        <Animated.View
          style={[
            styles.overlay,
            {
              opacity: opacityAnim,
            },
          ]}
        >
          <TouchableWithoutFeedback>
            <Animated.View
              style={[
                styles.dialogContainer,
                {
                  transform: [{ scale: scaleAnim }],
                  opacity: opacityAnim,
                },
              ]}
            >
              {/* Icon */}
              <View style={styles.iconContainer}>
                <View style={styles.iconBackground}>
                  <MaterialIcons
                    name="workspace-premium"
                    size={32}
                    color="#FFD700"
                  />
                </View>
              </View>

              {/* Title */}
              <Text style={styles.title}>{t("upgradePrompt.title")}</Text>

              {/* Content */}
              <View style={styles.content}>
                <Text style={styles.contentText}>
                  {t("upgradePrompt.limitReached")}
                </Text>
                <Text style={styles.contentText}>
                  {t("upgradePrompt.upgradeMessage")}
                </Text>
              </View>

              {/* Actions */}
              <View style={styles.actions}>
                <TouchableOpacity
                  style={styles.laterButton}
                  onPress={onDismiss}
                  activeOpacity={0.8}
                >
                  <Text style={styles.laterButtonText}>
                    {t("upgradePrompt.maybeLater")}
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity
                  style={styles.upgradeButton}
                  onPress={handleUpgrade}
                  activeOpacity={0.8}
                >
                  <Text style={styles.upgradeButtonText}>
                    {t("upgradePrompt.viewPlans")}
                  </Text>
                </TouchableOpacity>
              </View>
            </Animated.View>
          </TouchableWithoutFeedback>
        </Animated.View>
      </TouchableWithoutFeedback>
    </Modal>
  );
};

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: "rgba(0, 0, 0, 0.5)",
    justifyContent: "center",
    alignItems: "center",
    padding: 20,
  },
  dialogContainer: {
    backgroundColor: "#FFFFFF",
    borderRadius: 24,
    padding: 24,
    width: "100%",
    maxWidth: 340,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 10 },
    shadowOpacity: 0.15,
    shadowRadius: 20,
    elevation: 10,
  },
  iconContainer: {
    alignItems: "center",
    marginBottom: 16,
  },
  iconBackground: {
    width: 72,
    height: 72,
    borderRadius: 36,
    backgroundColor: "rgba(255, 215, 0, 0.15)",
    justifyContent: "center",
    alignItems: "center",
  },
  title: {
    fontSize: 22,
    fontWeight: "700",
    color: "#2D3436",
    textAlign: "center",
    marginBottom: 16,
  },
  content: {
    marginBottom: 24,
  },
  contentText: {
    fontSize: 16,
    color: "#636E72",
    textAlign: "center",
    lineHeight: 24,
    marginBottom: 10,
  },
  actions: {
    flexDirection: "row",
    gap: 12,
  },
  laterButton: {
    flex: 1,
    paddingVertical: 14,
    borderRadius: 12,
    backgroundColor: "#F5F5F5",
    alignItems: "center",
  },
  laterButtonText: {
    fontSize: 16,
    fontWeight: "600",
    color: "#636E72",
  },
  upgradeButton: {
    flex: 1,
    paddingVertical: 14,
    borderRadius: 12,
    backgroundColor: "#FF7B54",
    alignItems: "center",
    shadowColor: "#FF7B54",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.2,
    shadowRadius: 8,
    elevation: 5,
  },
  upgradeButtonText: {
    fontSize: 16,
    fontWeight: "600",
    color: "#FFFFFF",
  },
});

export default UpgradePromptDialog;
