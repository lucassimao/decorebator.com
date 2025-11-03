import * as Crypto from "expo-crypto";
import * as Device from "expo-device";
import Constants from "expo-constants";

type UsageDetails = {
  inputTokens: number;
  outputTokens: number;
  audioInputTokens: number;
  audioOutputTokens: number;
  textInputTokens: number;
  textOutputTokens: number;
};

type RateLimitSnapshot = {
  name: string;
  limit: number;
  remaining: number;
  resetSeconds: number;
  recordedAt: string;
};

type TelemetryTurn = {
  userSpeechMs?: number;
  assistantLatencyMs?: number;
  usage?: UsageDetails;
};

type ActiveTurn = {
  userSpeechStartedAt?: number;
  assistantStartedAt?: number;
  responseId?: string;
  turn: TelemetryTurn;
};

export type RealtimeChatTelemetryPayload = {
  sessionId: string;
  conversationId?: string;
  wordlistId?: number;
  languageCode?: string;
  model?: string;
  startedAt: string;
  endedAt: string;
  durationMs: number;
  totalAudioTokensIn: number;
  totalAudioTokensOut: number;
  totalTextTokensIn: number;
  totalTextTokensOut: number;
  totalTurns: number;
  turns: TelemetryTurn[];
  rateLimitEvents: RateLimitSnapshot[];
  errorCount: number;
  clientVersion?: string;
  deviceInfo?: Record<string, unknown>;
};

type TelemetryConfig = {
  wordlistId?: number;
  languageCode?: string;
  model?: string;
};

export class RealtimeChatTelemetryBuilder {
  private readonly sessionId = Crypto.randomUUID();
  private readonly startedAt = Date.now();
  private endedAt?: number;
  private conversationId?: string;
  private activeTurn?: ActiveTurn;
  private readonly turns: TelemetryTurn[] = [];
  private readonly rateLimits: RateLimitSnapshot[] = [];
  private errorCount = 0;
  private readonly clientVersion =
    Constants.expoConfig?.version ?? (Constants as any)?.expoGoConfig?.version;
  private readonly deviceInfo: Record<string, unknown>;

  private totalAudioIn = 0;
  private totalAudioOut = 0;
  private totalTextIn = 0;
  private totalTextOut = 0;

  private finalized = false;

  constructor(private config: TelemetryConfig) {
    this.deviceInfo = {
      osName: Device.osName,
      osVersion: Device.osVersion,
      modelName: Device.modelName,
      brand: Device.brand,
      deviceYearClass: Device.deviceYearClass,
      isDevice: Device.isDevice,
    };
  }

  updateLanguage(languageCode?: string) {
    if (languageCode) {
      this.config = { ...this.config, languageCode };
    }
  }

  updateModel(model?: string) {
    if (model) {
      this.config = { ...this.config, model };
    }
  }

  markConversationId(id?: string) {
    if (id) {
      this.conversationId = id;
    }
  }

  recordUserSpeechStart() {
    if (!this.activeTurn) {
      this.activeTurn = { turn: {} };
    }
    this.activeTurn.userSpeechStartedAt = Date.now();
  }

  recordUserSpeechStop() {
    if (this.activeTurn?.userSpeechStartedAt) {
      this.activeTurn.turn.userSpeechMs =
        Date.now() - this.activeTurn.userSpeechStartedAt;
    }
  }

  recordResponseCreated(responseId: string | undefined) {
    if (!this.activeTurn) {
      this.activeTurn = { turn: {} };
    }
    this.activeTurn.responseId = responseId;
    this.activeTurn.assistantStartedAt = Date.now();
  }

  recordResponseCompleted(
    responseId: string | undefined,
    usage: any | undefined,
  ) {
    if (!this.activeTurn || this.activeTurn.responseId !== responseId) {
      // Create placeholder turn to keep counts in sync
      this.activeTurn = { responseId, turn: {} };
    }

    if (this.activeTurn.assistantStartedAt) {
      this.activeTurn.turn.assistantLatencyMs =
        Date.now() - this.activeTurn.assistantStartedAt;
    }

    const usageDetails = this.parseUsage(usage);
    if (usageDetails) {
      this.activeTurn.turn.usage = usageDetails;
      this.totalAudioIn += usageDetails.audioInputTokens;
      this.totalAudioOut += usageDetails.audioOutputTokens;
      this.totalTextIn += usageDetails.textInputTokens;
      this.totalTextOut += usageDetails.textOutputTokens;
    }

    this.turns.push(this.activeTurn.turn);
    this.activeTurn = undefined;
  }

  recordRateLimits(
    rateLimits?: {
      name: string;
      limit: number;
      remaining: number;
      reset_seconds: number;
    }[],
  ) {
    if (!rateLimits) {
      return;
    }
    const recordedAt = new Date().toISOString();
    rateLimits.forEach((limit) => {
      this.rateLimits.push({
        name: limit.name,
        limit: limit.limit,
        remaining: limit.remaining,
        resetSeconds: limit.reset_seconds,
        recordedAt,
      });
    });
  }

  recordError() {
    this.errorCount += 1;
  }

  finalize(): RealtimeChatTelemetryPayload | null {
    if (this.finalized) {
      return null;
    }
    this.finalized = true;
    this.endedAt = Date.now();

    if (this.activeTurn) {
      this.turns.push(this.activeTurn.turn);
      this.activeTurn = undefined;
    }

    return {
      sessionId: this.sessionId,
      conversationId: this.conversationId,
      wordlistId: this.config.wordlistId,
      languageCode: this.config.languageCode,
      model: this.config.model,
      startedAt: new Date(this.startedAt).toISOString(),
      endedAt: new Date(this.endedAt).toISOString(),
      durationMs: this.endedAt - this.startedAt,
      totalAudioTokensIn: this.totalAudioIn,
      totalAudioTokensOut: this.totalAudioOut,
      totalTextTokensIn: this.totalTextIn,
      totalTextTokensOut: this.totalTextOut,
      totalTurns: this.turns.length,
      turns: this.turns,
      rateLimitEvents: this.rateLimits,
      errorCount: this.errorCount,
      clientVersion: this.clientVersion ?? undefined,
      deviceInfo: this.deviceInfo,
    };
  }

  private parseUsage(usage: any | undefined): UsageDetails | undefined {
    if (!usage) {
      return undefined;
    }
    const audioIn =
      usage.input_token_details?.audio_tokens ??
      usage.input_token_details?.audio ??
      0;
    const textIn =
      usage.input_token_details?.text_tokens ??
      usage.input_token_details?.text ??
      0;
    const audioOut =
      usage.output_token_details?.audio_tokens ??
      usage.output_token_details?.audio ??
      0;
    const textOut =
      usage.output_token_details?.text_tokens ??
      usage.output_token_details?.text ??
      0;

    return {
      inputTokens: usage.input_tokens ?? audioIn + textIn,
      outputTokens: usage.output_tokens ?? audioOut + textOut,
      audioInputTokens: audioIn,
      audioOutputTokens: audioOut,
      textInputTokens: textIn,
      textOutputTokens: textOut,
    };
  }
}
