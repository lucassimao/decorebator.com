import { callAPI } from "./api";

export enum ErrorType {
  UnrelatedImage = "_unrelated_image",
  MissingImage = "_missing_image",
  UnrelatedMeaning = "_unrelated_meaning",
  UnrelatedExample = "_unrelated_example",
  SoundNotPlaying = "_sound_not_playing",
}

export interface ErrorReportRequest {
  wordId: number;
  definitionId: number;
  errorType: ErrorType;
}

export async function reportError(request: ErrorReportRequest): Promise<void> {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + `/errorReports`;
  await callAPI("POST", endpoint, JSON.stringify(request));
}
