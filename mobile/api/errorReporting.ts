import { callAPI } from "./api";
import { Quiz } from "./wordlists";

export enum ErrorType {
  UnrelatedImage = "_unrelated_image",
  UnrelatedMeaning = "_unrelated_meaning",
  UnrelatedExample = "_unrelated_example",
  SoundNotPlaying = "_sound_not_playing",
}

export async function reportError(
  errorType: ErrorType,
  quiz: Quiz,
): Promise<void> {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + `/errorReport`;
  await callAPI("POST", endpoint, JSON.stringify({ errorType, quiz }));
}
