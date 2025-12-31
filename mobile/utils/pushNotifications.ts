import AsyncStorage from "@react-native-async-storage/async-storage";
import * as Device from "expo-device";
import * as Notifications from "expo-notifications";
import Constants from "expo-constants";
import { Platform } from "react-native";
import i18n from "@/i18n";
import { getDeviceTimezone } from "@/utils/dateUtils";
import * as pushApi from "@/api/pushNotifications";

const PUSH_TOKEN_STORAGE_KEY = "expoPushToken";

export async function getStoredPushToken(): Promise<string | null> {
  return AsyncStorage.getItem(PUSH_TOKEN_STORAGE_KEY);
}

export async function registerDevicePushToken(options: {
  prompt: boolean;
}): Promise<string | null> {
  if (Platform.OS === "web") {
    return null;
  }

  if (!Device.isDevice) {
    return null;
  }

  const existingToken = await getStoredPushToken();
  if (existingToken) {
    await pushApi.registerPushToken({
      expoPushToken: existingToken,
      platform: Platform.OS === "ios" ? "ios" : "android",
      deviceId: Device.osInternalBuildId || Device.modelId || undefined,
      timezone: getDeviceTimezone(),
      locale: i18n.language,
    });
    return existingToken;
  }

  const permissions = await Notifications.getPermissionsAsync();
  let status = permissions.status;
  if (status !== "granted" && options.prompt) {
    const requested = await Notifications.requestPermissionsAsync();
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

  const expoPushToken = tokenResponse.data;
  if (!expoPushToken) {
    return null;
  }

  const deviceId = Device.osInternalBuildId || Device.modelId || undefined;

  await pushApi.registerPushToken({
    expoPushToken,
    platform: Platform.OS === "ios" ? "ios" : "android",
    deviceId,
    timezone: getDeviceTimezone(),
    locale: i18n.language,
  });

  await AsyncStorage.setItem(PUSH_TOKEN_STORAGE_KEY, expoPushToken);
  return expoPushToken;
}

export async function unregisterDevicePushToken(): Promise<void> {
  const storedToken = await getStoredPushToken();
  if (!storedToken) {
    return;
  }

  await pushApi.unregisterPushToken(storedToken);
  await AsyncStorage.removeItem(PUSH_TOKEN_STORAGE_KEY);
}
