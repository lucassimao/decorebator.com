import React from "react";
import {
  View,
  Text,
  ActivityIndicator,
  TouchableOpacity,
  StyleSheet,
} from "react-native";
import { MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";

interface LoadingWithTimeoutProps {
  isLoading: boolean;
  hasTimeout: boolean;
  error?: Error | null;
  primaryColor?: string;
  backgroundColor?: string;
  textColor?: string;
  loadingMessage?: string;
  timeoutMessage?: string;
  onRetry: () => void;
  onGoBack: () => void;
  showTimeoutActions?: boolean;
}

export const LoadingWithTimeout: React.FC<LoadingWithTimeoutProps> = ({
  isLoading,
  hasTimeout,
  error,
  primaryColor = "#FF7B54",
  backgroundColor = "#FFFFFF",
  textColor = "#636E72",
  loadingMessage,
  timeoutMessage,
  onRetry,
  onGoBack,
  showTimeoutActions = true,
}) => {
  const { t } = useTranslation();

  if (!isLoading) return null;

  return (
    <View style={styles.loadingContainer}>
      <ActivityIndicator size="large" color={primaryColor} />
      <Text style={[styles.loadingText, { color: textColor }]}>
        {hasTimeout ? timeoutMessage : loadingMessage}
      </Text>
      
      {hasTimeout && showTimeoutActions && (
        <View style={styles.timeoutActions}>
          <Text style={[styles.timeoutMessage, { color: textColor }]}>
            {error?.message.includes('timeout') 
              ? t("quiz.requestTimedOut")
              : t("quiz.slowConnection")}
          </Text>
          <View style={styles.actionButtons}>
            <TouchableOpacity 
              style={[styles.retryButton, { backgroundColor: primaryColor }]} 
              onPress={onRetry}
            >
              <MaterialIcons name="refresh" size={20} color="#FFFFFF" />
              <Text style={styles.retryButtonText}>{t("common.retry")}</Text>
            </TouchableOpacity>
            <TouchableOpacity 
              style={[styles.goBackButton, { backgroundColor }]} 
              onPress={onGoBack}
            >
              <MaterialIcons name="arrow-back" size={20} color={textColor} />
              <Text style={[styles.goBackButtonText, { color: textColor }]}>
                {t("common.goBack")}
              </Text>
            </TouchableOpacity>
          </View>
        </View>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  loadingContainer: {
    minHeight: 300,
    justifyContent: "center",
    alignItems: "center",
    paddingHorizontal: 20,
  },
  loadingText: {
    marginTop: 16,
    fontSize: 16,
    textAlign: "center",
  },
  timeoutActions: {
    marginTop: 24,
    alignItems: "center",
  },
  timeoutMessage: {
    fontSize: 14,
    textAlign: "center",
    marginBottom: 20,
    lineHeight: 20,
  },
  actionButtons: {
    flexDirection: "row",
    gap: 12,
  },
  retryButton: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 20,
    paddingVertical: 12,
    borderRadius: 25,
    gap: 8,
  },
  retryButtonText: {
    color: "#FFFFFF",
    fontSize: 16,
    fontWeight: "600",
  },
  goBackButton: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 20,
    paddingVertical: 12,
    borderRadius: 25,
    borderWidth: 1,
    borderColor: "#E9ECEF",
    gap: 8,
  },
  goBackButtonText: {
    fontSize: 16,
    fontWeight: "600",
  },
});