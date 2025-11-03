import { callAPI } from "./api";
import { RealtimeChatTelemetryPayload } from "@/utils/realtimeTelemetry";

export async function sendRealtimeChatTelemetry(
  payload: RealtimeChatTelemetryPayload,
) {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + "/telemetry/realtime-chat";
  await callAPI("POST", endpoint, JSON.stringify(payload));
}
