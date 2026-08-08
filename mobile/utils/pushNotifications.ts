import * as Device from "expo-device";
import * as Notifications from "expo-notifications";
import Constants from "expo-constants";
import { Platform } from "react-native";
import i18n from "@/i18n";
import { getDeviceTimezone } from "@/utils/dateUtils";
import * as pushApi from "@/api/pushNotifications";
import {
  beginPushTokenRegistration,
  completePushTokenRegistration,
  deactivateStoredPushToken,
  getStoredPushToken,
  scheduleStoredPushTokenDeactivation,
} from "@/utils/pushTokenStorage";
import { getAuthenticationSessionEpoch } from "@/api/users";

export { getStoredPushToken } from "@/utils/pushTokenStorage";

export async function registerDevicePushToken(options: {
  prompt: boolean;
}): Promise<string | null> {
  if (Platform.OS === "web") {
    return null;
  }

  if (!Device.isDevice) {
    return null;
  }

  const sessionOrigin = getAuthenticationSessionEpoch();
  const assertSessionOrigin = () => {
    if (!sessionOrigin || getAuthenticationSessionEpoch() !== sessionOrigin) {
      throw new Error("Session changed during push registration");
    }
  };

  const existingToken = await getStoredPushToken();
  assertSessionOrigin();
  if (existingToken) {
    const registrationGeneration = beginPushTokenRegistration();
    const timezone = getDeviceTimezone();
    try {
      const registered = await completePushTokenRegistration(
        registrationGeneration,
        existingToken,
        (signal) => {
          assertSessionOrigin();
          return pushApi.registerPushToken(
            {
              expoPushToken: existingToken,
              platform: Platform.OS === "ios" ? "ios" : "android",
              deviceId: Device.osInternalBuildId || Device.modelId || undefined,
              timezone,
              locale: i18n.language,
            },
            signal,
          );
        },
      );
      if (!registered) return null;
    } catch (error) {
      console.warn("Failed to register push token", {
        timezone,
        error,
      });
      throw error;
    }
    return existingToken;
  }

  const permissions = await Notifications.getPermissionsAsync();
  assertSessionOrigin();
  let status = permissions.status;
  if (status !== "granted" && options.prompt) {
    const requested = await Notifications.requestPermissionsAsync();
    assertSessionOrigin();
    status = requested.status;
  }
  if (status !== "granted") {
    return null;
  }

  const projectId =
    Constants.expoConfig?.extra?.eas?.projectId ??
    Constants.easConfig?.projectId;
  const tokenResponse = projectId
    ? await Notifications.getExpoPushTokenAsync({ projectId })
    : await Notifications.getExpoPushTokenAsync();
  assertSessionOrigin();

  const expoPushToken = tokenResponse.data;
  if (!expoPushToken) {
    return null;
  }

  const registrationGeneration = beginPushTokenRegistration();

  const deviceId = Device.osInternalBuildId || Device.modelId || undefined;
  const timezone = getDeviceTimezone();

  try {
    const registered = await completePushTokenRegistration(
      registrationGeneration,
      expoPushToken,
      (signal) => {
        assertSessionOrigin();
        return pushApi.registerPushToken(
          {
            expoPushToken,
            platform: Platform.OS === "ios" ? "ios" : "android",
            deviceId,
            timezone,
            locale: i18n.language,
          },
          signal,
        );
      },
    );
    if (!registered) return null;
  } catch (error) {
    console.warn("Failed to register push token", {
      timezone,
      error,
    });
    throw error;
  }

  return expoPushToken;
}

export async function unregisterDevicePushToken(): Promise<void> {
  await deactivateStoredPushToken();
}

export function resumeStoredPushTokenDeactivation(): void {
  if (Platform.OS !== "web") {
    scheduleStoredPushTokenDeactivation();
  }
}
