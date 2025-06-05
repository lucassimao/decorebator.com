import React, { useState, useEffect } from "react";
import { View, Text, StyleSheet, TouchableOpacity } from "react-native";
import { MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useOffline } from "@/hooks/useOffline";
import * as offlineWordlistsApi from "@/api/offlineWordlists";

interface OfflinePreloaderProps {
  wordlistId: number;
  wordlistName: string;
  onPreloadComplete?: () => void;
  onPreloadError?: (error: string) => void;
}

export const OfflinePreloader: React.FC<OfflinePreloaderProps> = ({
  wordlistId,
  wordlistName,
  onPreloadComplete,
  onPreloadError,
}) => {
  const { t } = useTranslation();
  const { isOnline, isOfflineAvailable } = useOffline();
  const [isPreloading, setIsPreloading] = useState(false);
  const [isFullyCached, setIsFullyCached] = useState(false);
  const [cacheStats, setCacheStats] = useState({
    totalWords: 0,
    cachedWords: 0,
    cachedDefinitions: 0,
    cachePercentage: 0,
  });

  useEffect(() => {
    checkCacheStatus();
  }, [wordlistId]);

  const checkCacheStatus = async () => {
    try {
      const [isAvailable, stats] = await Promise.all([
        offlineWordlistsApi.isWordlistAvailableOffline(wordlistId),
        offlineWordlistsApi.getWordlistCacheStats(wordlistId),
      ]);

      setIsFullyCached(isAvailable);
      setCacheStats(stats);
    } catch (error) {
      console.error("Error checking cache status:", error);
    }
  };

  const handlePreload = async () => {
    if (!isOnline || !isOfflineAvailable) return;

    setIsPreloading(true);
    try {
      await offlineWordlistsApi.preloadWordlistForOffline(wordlistId);

      // Refresh cache status
      await checkCacheStatus();

      onPreloadComplete?.();
    } catch (error) {
      console.error("Error preloading wordlist:", error);
      onPreloadError?.(
        error instanceof Error ? error.message : "Unknown error",
      );
    } finally {
      setIsPreloading(false);
    }
  };

  // Don't show if user doesn't have offline capabilities
  if (!isOfflineAvailable) return null;

  const getStatusIcon = () => {
    if (isPreloading) return "downloading";
    if (isFullyCached) return "offline-bolt";
    return "cloud-download";
  };

  const getStatusColor = () => {
    if (isPreloading) return "#FF7B54";
    if (isFullyCached) return "#4CAF50";
    return "#636E72";
  };

  const getStatusText = () => {
    if (isPreloading) return t("offline.preloading");
    if (isFullyCached) return t("offline.availableOffline");
    if (!isOnline) return t("offline.preloadWhenOnline");
    return t("offline.downloadForOffline");
  };

  return (
    <View style={styles.container}>
      <TouchableOpacity
        style={[
          styles.button,
          {
            backgroundColor: isFullyCached ? "#E8F5E9" : "#F5F5F5",
            borderColor: getStatusColor(),
          },
        ]}
        onPress={handlePreload}
        disabled={!isOnline || isPreloading || isFullyCached}
      >
        <MaterialIcons
          name={getStatusIcon()}
          size={20}
          color={getStatusColor()}
        />
        <View style={styles.textContainer}>
          <Text style={[styles.statusText, { color: getStatusColor() }]}>
            {getStatusText()}
          </Text>
          {cacheStats.totalWords > 0 && (
            <Text style={styles.progressText}>
              {t("offline.cacheProgress", {
                cached: cacheStats.cachedDefinitions,
                total: cacheStats.totalWords,
                percentage: cacheStats.cachePercentage,
              })}
            </Text>
          )}
        </View>
      </TouchableOpacity>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    marginVertical: 8,
  },
  button: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderRadius: 12,
    borderWidth: 1,
  },
  textContainer: {
    flex: 1,
    marginLeft: 12,
  },
  statusText: {
    fontSize: 14,
    fontWeight: "500",
  },
  progressText: {
    fontSize: 12,
    color: "#636E72",
    marginTop: 2,
  },
});
