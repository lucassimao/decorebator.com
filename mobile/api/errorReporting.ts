import { callAPI } from "./api";

export enum ErrorType {
  UnrelatedImage = "_unrelated_image",
  MissingImage = "_missing_image",
  UnrelatedMeaning = "_unrelated_meaning",
  UnrelatedExample = "_unrelated_example",
  SoundNotPlaying = "_sound_not_playing",
  ProcessingFailed = "_processing_failed",
}

export interface QuizDetails {
  quizType: string;
  value: string;
  options: string[];
  answerIndex: number;
  context: "quiz" | "flashcard";
}

export interface ErrorReportRequest {
  wordId: number;
  definitionId: number | null;
  errorType: ErrorType;
  quizDetails?: QuizDetails;
}

export interface RateLimitError {
  error: string;
  cooldownUntil?: number;
  retryAfter?: number;
  limit?: number;
  remaining?: number;
  windowType?: string;
}

export class ErrorReportRateLimitError extends Error {
  public cooldownUntil?: number;
  public retryAfter?: number;
  public limit?: number;
  public remaining?: number;
  public windowType?: string;

  constructor(data: RateLimitError) {
    super(data.error);
    this.name = "ErrorReportRateLimitError";
    this.cooldownUntil = data.cooldownUntil;
    this.retryAfter = data.retryAfter;
    this.limit = data.limit;
    this.remaining = data.remaining;
    this.windowType = data.windowType;
  }
}

export async function reportError(request: ErrorReportRequest): Promise<void> {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + `/errorReports`;

  try {
    await callAPI("POST", endpoint, JSON.stringify(request));
  } catch (error: any) {
    // Check if it's a rate limit error (429 status)
    if (error.status === 429 && error.data) {
      throw new ErrorReportRateLimitError(error.data);
    }
    throw error;
  }
}
