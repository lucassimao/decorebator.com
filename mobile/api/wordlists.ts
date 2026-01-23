import { callAPI } from "./api";
import { getApiBaseUrl } from "./baseUrl";

const API_URL = getApiBaseUrl();

export type PronunciationSystem =
  | "ipa"
  | "romaji"
  | "hiragana"
  | "pinyin"
  | "hangul";

export type Wordlist = {
  createdAt: string;
  description: string;
  id: number;
  name: string;
  updatedAt: string;
  userId: number;
  languageCode: string;
  languageName: string;
  pronunciationSystem: PronunciationSystem;
  wordsCount: number;
  wordsLearnedCount: number;
};

export type Word = {
  id: number;
  name: string;
  wordlistId: number;
  learned: boolean;
  pronunciation?: string;
  notes?: string;
  audioURL?: string;
  processingStatus?: string;
  processingError?: string;
  processingStartedAt?: string;
  processingCompletedAt?: string;
};

export type Definition = {
  id: number;
  token: string;
  language: string;
  meaning: string;
  partOfSpeech?: string;
  isVerbType?: boolean; // computed flag indicating if this is a verb/phrasal verb
  examples?: string[];
  inflections?: {
    inflection: string;
    tense: string;
    examples: string[];
  }[];
  source: string;
  sourceId?: string;
  sounds?: {
    accent: string;
    link: string;
  }[];
  phoneticNotations?: {
    ipa: string;
    accent: string;
  }[];
  createdAt: string;
  updatedAt: string;
};

export type Quiz = {
  value: string;
  options: string[];
  answerIndex: number;
  id: number;
  type:
    | "GUESS_MEANING"
    | "COMPLETE_SENTENCE"
    | "WORD_FROM_MEANING"
    | "WORD_FROM_IMAGE"
    | "WORD_FROM_AUDIO"
    | "MEANING_FROM_AUDIO"
    | "WORD_FROM_MEANING_AUDIO"
    | "WRITE_WORD_FROM_DEFINITION"
    | "WORD_FROM_EXAMPLE_AUDIO";
  pos: string; // part of speech
  isVerbType: boolean; // computed flag indicating if this is a verb/phrasal verb
  pronunciation?: string; // IPA pronunciation
  audioURL?: string; //only present in MeaningFromAudio, WordFromMeaningAudio, GUESS_MEANING, WordFromAudio, and WordFromExampleAudio quizes
  imageDescription: string;
  definitionId: number;
  wordId: number;
};

export type QuizType = Quiz["type"];

// Public Quiz (MVP)
// Public quiz publishing removed from mobile UI; keep read-only endpoints elsewhere

export type CreateWordDTO = Pick<Word, "wordlistId" | "name" | "notes"> & {
  hasOptimisticSubscription?: boolean;
};
export type CreateWordlistDTO = Pick<
  Wordlist,
  "description" | "name" | "languageCode"
> & {
  pronunciationSystem?: PronunciationSystem;
  hasOptimisticSubscription?: boolean;
};

export async function getUserWordlists() {
  const endpoint = API_URL + "/wordlists";
  const body = await callAPI<Wordlist[]>("GET", endpoint);
  return body;
}

export async function getWords(
  wordlistId: number,
  onlyWithDefinitions = false,
): Promise<Word[]> {
  const baseEndpoint = API_URL + `/wordlists/${wordlistId}/words`;
  const endpoint = onlyWithDefinitions
    ? `${baseEndpoint}?onlyWithDefinitions=true`
    : baseEndpoint;

  const body = await callAPI<Word[]>("GET", endpoint);
  return body;
}

export async function deleteWordlist(wordlistId: number): Promise<void> {
  const endpoint = API_URL + `/wordlists/${wordlistId}`;
  await callAPI("DELETE", endpoint);
}

export async function getWordlist(wordlistId: number): Promise<Wordlist> {
  const endpoint = API_URL + `/wordlists/${wordlistId}`;
  return await callAPI<Wordlist>("GET", endpoint);
}

export async function addWordlist(dto: CreateWordlistDTO): Promise<Wordlist> {
  let endpoint = API_URL + `/wordlists`;

  // Add optimistic flag as query parameter if present
  if (dto.hasOptimisticSubscription) {
    endpoint += `?hasOptimisticSubscription=true`;
  }

  // Remove flag from body since it's now in query params
  const { hasOptimisticSubscription, ...bodyData } = dto;

  return await callAPI<Wordlist>("POST", endpoint, JSON.stringify(bodyData));
}

export async function addWord(dto: CreateWordDTO): Promise<void> {
  let endpoint = API_URL + `/wordlists/${dto.wordlistId}/words`;

  // Add optimistic flag as query parameter if present
  if (dto.hasOptimisticSubscription) {
    endpoint += `?hasOptimisticSubscription=true`;
  }

  // Remove flag from body since it's now in query params
  const { hasOptimisticSubscription, ...bodyData } = dto;

  await callAPI("POST", endpoint, JSON.stringify(bodyData));
}

export async function deleteWord(
  word: Pick<Word, "id" | "wordlistId">,
): Promise<void> {
  const { wordlistId, id: wordId } = word;

  const endpoint = API_URL + `/wordlists/${wordlistId}/words/${wordId}`;

  await callAPI("DELETE", endpoint);
}

export async function updateWord(
  dto: Pick<Word, "id" | "wordlistId" | "learned" | "name" | "notes">,
) {
  const endpoint = API_URL + `/wordlists/${dto.wordlistId}/words/${dto.id}`;
  await callAPI("PUT", endpoint, JSON.stringify(dto));
}
export async function newQuiz(
  wordlistId: number,
  quizTypes?: QuizType[],
): Promise<Quiz> {
  let endpoint = API_URL + `/wordlists/${wordlistId}/quizzes`;
  if (quizTypes && quizTypes.length > 0) {
    endpoint += `?quizTypes=${encodeURIComponent(quizTypes.join(","))}`;
  }
  return await callAPI<Quiz>("POST", endpoint);
}

export type AnswerQuizInput = {
  wordlistID: number;
  wordID: number;
  definitionID: number;
  leitnerSystemTrackingID: number;
  quizType: string;
  isCorrect: boolean;
  responseTimeMs: number;
};
export async function answerQuiz(input: AnswerQuizInput): Promise<void> {
  const endpoint = API_URL + `/wordlists/${input.wordlistID}/quizzes`;

  await callAPI<Quiz>("PATCH", endpoint, JSON.stringify(input));
}

export async function getWordDefinitions(
  wordlistId: number,
  wordId: number,
): Promise<Definition[]> {
  const endpoint =
    API_URL + `/wordlists/${wordlistId}/words/${wordId}/definitions`;

  const body = await callAPI<Definition[]>("GET", endpoint);
  return body;
}

export type PronunciationSystemsResponse = {
  supportedSystems: PronunciationSystem[];
  defaultSystem: PronunciationSystem;
  canChange: boolean;
};

export async function getPronunciationSystems(
  languageCode: string,
): Promise<PronunciationSystemsResponse> {
  const endpoint =
    API_URL + `/wordlists/pronunciation-systems?languageCode=${languageCode}`;

  const body = await callAPI<PronunciationSystemsResponse>("GET", endpoint);
  return body;
}

export type ProcessingInfo = {
  id: number;
  name: string;
  processingStatus: string;
  processingError?: string;
  processingStartedAt?: string;
  processingCompletedAt?: string;
};

export type ProcessingStatusResponse = {
  words: ProcessingInfo[];
  summary: {
    total: number;
    pending: number;
    processing: number;
    completed: number;
    failed: number;
  };
};

export async function getProcessingStatus(
  wordlistId: number,
): Promise<ProcessingStatusResponse> {
  const endpoint = API_URL + `/wordlists/${wordlistId}/processing-status`;

  const body = await callAPI<ProcessingStatusResponse>("GET", endpoint);
  return body;
}

// Realtime Chat Session Types and API
export interface ChatDefinition {
  id: number;
  meaning: string;
  partOfSpeech: string;
  examples?: string[];
}

export interface WordWithDefinitions {
  name: string;
  definitions: ChatDefinition[];
}

export interface ChatSessionData {
  token: string;
  expiresAt: number;
  webrtcConfig: {
    baseUrl: string;
    model: string;
    iceServers?: {
      urls: string | string[];
      username?: string;
      credential?: string;
    }[];
  };
}

// Create a new realtime chat session for vocabulary practice
export async function createChatSession(
  wordlistId: number,
): Promise<ChatSessionData> {
  const endpoint = API_URL + `/wordlists/${wordlistId}/chat/session`;
  return await callAPI<ChatSessionData>("POST", endpoint);
}

// Server response for batched definitions endpoint
type WordDefinitionsBatchResponse = {
  wordId: number;
  name: string;
  definitions: Definition[];
};

// Fetch definitions for multiple words in a single request and map to chat shape
export async function getDefinitionsForWords(
  wordlistId: number,
  wordIds: number[],
): Promise<WordWithDefinitions[]> {
  if (!wordIds || wordIds.length === 0) return [];

  const idsParam = wordIds.join(",");
  const endpoint =
    API_URL + `/wordlists/${wordlistId}/words/definitions?ids=${idsParam}`;

  const resp = await callAPI<WordDefinitionsBatchResponse[]>("GET", endpoint);

  // Map server definition model to chat-friendly minimal shape
  return resp.map((item) => ({
    name: item.name,
    definitions: (item.definitions || []).map((def) => ({
      id: def.id,
      meaning: def.meaning,
      partOfSpeech: def.partOfSpeech || "",
      examples: def.examples || [],
    })),
  }));
}
