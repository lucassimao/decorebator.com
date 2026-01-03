import { callAPI } from "./api";
import { RealtimeChatTelemetryPayload } from "@/utils/realtimeTelemetry";
import { getApiBaseUrl } from "./baseUrl";

export async function sendRealtimeChatTelemetry(
  payload: RealtimeChatTelemetryPayload,
) {
  const endpoint = getApiBaseUrl() + "/telemetry/realtime-chat";
  await callAPI("POST", endpoint, JSON.stringify(payload));
}
