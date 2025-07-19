import { useState, useCallback } from "react";
import * as Updates from "expo-updates";

export const useUpdates = () => {
  const [isChecking, setIsChecking] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [updateAvailable, setUpdateAvailable] = useState(false);

  const checkForUpdates = useCallback(async () => {
    if (__DEV__) {
      // In development, simulate no updates available
      return false;
    }

    setIsChecking(true);
    try {
      const update = await Updates.checkForUpdateAsync();
      setUpdateAvailable(update.isAvailable);
      return update.isAvailable;
    } catch (error) {
      console.error("Error checking for updates:", error);
      return false;
    } finally {
      setIsChecking(false);
    }
  }, []);

  const applyUpdate = useCallback(async () => {
    if (__DEV__) {
      // In development, just log
      console.log("Update would be applied in production");
      return;
    }

    setIsUpdating(true);
    try {
      await Updates.fetchUpdateAsync();
      await Updates.reloadAsync();
    } catch (error) {
      console.error("Error applying update:", error);
      setIsUpdating(false);
      throw error;
    }
  }, []);

  const getUpdateInfo = useCallback(async () => {
    try {
      return {
        updateId: Updates.updateId,
        // releaseChannel: Updates.releaseChannel, // Not available in current version
        createdAt: Updates.createdAt,
        isEmbeddedLaunch: Updates.isEmbeddedLaunch,
      };
    } catch (error) {
      console.error("Error getting update info:", error);
      return null;
    }
  }, []);

  return {
    isChecking,
    isUpdating,
    updateAvailable,
    checkForUpdates,
    applyUpdate,
    getUpdateInfo,
  };
};
