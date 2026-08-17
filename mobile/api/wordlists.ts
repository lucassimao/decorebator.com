import {
  callAPI,
  callAPIWithMetadata,
  type APIResponseWithMetadata,
} from "./api";
import { getApiBaseUrl } from "./baseUrl";
import {
  createPaginationSessionGuard,
  getAllPages,
  PaginationError,
} from "./pagination";

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
  return getAllPages<Wordlist[], Wordlist>({
    endpoint,
    getItems: (body) => requireArray(body, "wordlists"),
    getItemKey: (wordlist) => requireItemKey(wordlist.id, "wordlist"),
  });
}

export async function getWords(
  wordlistId: number,
  onlyWithDefinitions = false,
): Promise<Word[]> {
  const baseEndpoint = API_URL + `/wordlists/${wordlistId}/words`;
  const endpoint = onlyWithDefinitions
    ? `${baseEndpoint}?onlyWithDefinitions=true`
    : baseEndpoint;

  return getAllPages<Word[], Word>({
    endpoint,
    getItems: (body) => requireArray(body, "words"),
    getItemKey: (word) => requireItemKey(word.id, "word"),
  });
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
  dto: Pick<Word, "id" | "wordlistId"> &
    Partial<Pick<Word, "learned" | "name" | "notes">>,
) {
  const endpoint = API_URL + `/wordlists/${dto.wordlistId}/words/${dto.id}`;
  const { learned, name, notes, wordlistId } = dto;
  await callAPI(
    "PUT",
    endpoint,
    JSON.stringify({ learned, name, notes, wordlistId }),
  );
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

  return getAllPages<Definition[], Definition>({
    endpoint,
    getItems: (body) => requireArray(body, "definitions"),
    getItemKey: (definition) => requireItemKey(definition.id, "definition"),
  });
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
  let latestSummary: ProcessingStatusResponse["summary"] | undefined;

  const words = await getAllPages<ProcessingStatusResponse, ProcessingInfo>({
    endpoint,
    getItems: (body) => requireArray(body?.words, "processing status words"),
    getItemKey: (word) => requireItemKey(word.id, "processing-status word"),
    onPage: (body) => {
      latestSummary = requireProcessingSummary(body?.summary);
    },
  });

  if (latestSummary === undefined) {
    throw new PaginationError(
      "Processing status response is missing a summary",
    );
  }
  return {
    words,
    summary: latestSummary,
  };
}

function requireArray<T>(value: unknown, responseField: string): T[] {
  if (!Array.isArray(value)) {
    throw new PaginationError(
      `Invalid paginated ${responseField} response: expected an array`,
    );
  }
  return value as T[];
}

function requireItemKey(value: unknown, itemName: string): string {
  if (typeof value !== "number" || !Number.isSafeInteger(value) || value <= 0) {
    throw new PaginationError(
      `Invalid paginated ${itemName} response: expected a positive ID`,
    );
  }
  return String(value);
}

function requireProcessingSummary(
  value: unknown,
): ProcessingStatusResponse["summary"] {
  if (typeof value !== "object" || value === null) {
    throw new PaginationError(
      "Processing status response is missing a summary",
    );
  }
  const summary = value as ProcessingStatusResponse["summary"];
  for (const field of [
    "total",
    "pending",
    "processing",
    "completed",
    "failed",
  ] as const) {
    if (!Number.isInteger(summary[field]) || summary[field] < 0) {
      throw new PaginationError(
        "Processing status response has an invalid summary",
      );
    }
  }
  return summary;
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

const MAX_DEFINITION_BATCH_WORD_IDS = 50;
const MAX_DEFINITION_BATCH_WORDS = 500;
const MAX_DEFINITION_BATCH_PAGES = 100;
const MAX_DEFINITION_BATCH_ITEMS = 10_000;

// Fetch definitions for multiple words in a single request and map to chat shape
export async function getDefinitionsForWords(
  wordlistId: number,
  wordIds: number[],
): Promise<WordWithDefinitions[]> {
  if (!wordIds || wordIds.length === 0) return [];

  const uniqueWordIds = validateDefinitionBatchWordIDs(wordIds);
  const assertSessionUnchanged = createPaginationSessionGuard();
  assertSessionUnchanged();
  const definitionsByWord = new Map<number, WordDefinitionsBatchResponse>();
  const definitionIDsByWord = new Map<number, Set<number>>();
  let totalDefinitions = 0;

  for (
    let start = 0;
    start < uniqueWordIds.length;
    start += MAX_DEFINITION_BATCH_WORD_IDS
  ) {
    const wordIDs = uniqueWordIds.slice(
      start,
      start + MAX_DEFINITION_BATCH_WORD_IDS,
    );
    const requestedWordIDs = new Set(wordIDs);
    const seenContinuations = new Set<string>();
    let continuation: string | null = null;

    for (let page = 0; page < MAX_DEFINITION_BATCH_PAGES; page += 1) {
      assertSessionUnchanged();
      const response: APIResponseWithMetadata<WordDefinitionsBatchResponse[]> =
        await callAPIWithMetadata<WordDefinitionsBatchResponse[]>(
          "GET",
          getDefinitionBatchEndpoint(wordlistId, wordIDs, continuation),
        );
      assertSessionUnchanged();

      const responseItems = requireArray<WordDefinitionsBatchResponse>(
        response.body,
        "batched definitions",
      );
      const pageDefinitionIDs = new Set<string>();
      for (const item of responseItems) {
        if (
          !Number.isSafeInteger(item?.wordId) ||
          !requestedWordIDs.has(item.wordId)
        ) {
          throw new PaginationError(
            "Batched definitions response contains an unexpected word",
          );
        }
        const definitions = requireArray<Definition>(
          item.definitions,
          "batched definition items",
        );
        let aggregate = definitionsByWord.get(item.wordId);
        if (aggregate === undefined) {
          aggregate = { wordId: item.wordId, name: item.name, definitions: [] };
          definitionsByWord.set(item.wordId, aggregate);
          definitionIDsByWord.set(item.wordId, new Set());
        } else if (aggregate.name !== item.name) {
          throw new PaginationError(
            "Batched definitions response changed a word name mid-request",
          );
        }

        const seenDefinitionIDs = definitionIDsByWord.get(item.wordId)!;
        for (const definition of definitions) {
          if (!Number.isSafeInteger(definition?.id) || definition.id <= 0) {
            throw new PaginationError(
              "Batched definitions response contains an invalid definition",
            );
          }
          const pageDefinitionKey = `${item.wordId}:${definition.id}`;
          if (pageDefinitionIDs.has(pageDefinitionKey)) {
            throw new PaginationError(
              "Batched definitions response contains a duplicate definition",
            );
          }
          pageDefinitionIDs.add(pageDefinitionKey);
          // Repeating the original IDs is allowed by the server continuation
          // contract. Those earlier definitions may be returned again, so do
          // not aggregate them twice.
          if (seenDefinitionIDs.has(definition.id)) continue;
          if (totalDefinitions >= MAX_DEFINITION_BATCH_ITEMS) {
            throw new PaginationError(
              "Batched definitions item ceiling reached",
            );
          }
          seenDefinitionIDs.add(definition.id);
          aggregate.definitions.push(definition);
          totalDefinitions += 1;
        }
      }

      continuation = response.definitionsContinuation?.trim() || null;
      if (continuation === null) break;
      if (seenContinuations.has(continuation)) {
        throw new PaginationError(
          "Batched definitions response contains a repeated continuation",
        );
      }
      seenContinuations.add(continuation);
      if (page === MAX_DEFINITION_BATCH_PAGES - 1) {
        throw new PaginationError("Batched definitions page ceiling reached");
      }
    }
  }

  // Map server definition model to chat-friendly minimal shape
  return Array.from(definitionsByWord.values()).map((item) => ({
    name: item.name,
    definitions: item.definitions.map((def) => ({
      id: def.id,
      meaning: def.meaning,
      partOfSpeech: def.partOfSpeech || "",
      examples: def.examples || [],
    })),
  }));
}

function validateDefinitionBatchWordIDs(wordIds: number[]): number[] {
  if (wordIds.length > MAX_DEFINITION_BATCH_WORDS) {
    throw new PaginationError("Batched definitions word ceiling reached");
  }
  const seenWordIDs = new Set<number>();
  for (const wordId of wordIds) {
    if (!Number.isSafeInteger(wordId) || wordId <= 0) {
      throw new PaginationError(
        "Batched definitions request contains an invalid word",
      );
    }
    if (seenWordIDs.has(wordId)) {
      throw new PaginationError(
        "Batched definitions request contains a duplicate word",
      );
    }
    seenWordIDs.add(wordId);
  }
  return wordIds;
}

function getDefinitionBatchEndpoint(
  wordlistId: number,
  wordIds: number[],
  continuation: string | null,
): string {
  const endpoint = new URL(
    API_URL + `/wordlists/${wordlistId}/words/definitions`,
  );
  endpoint.searchParams.set("ids", wordIds.join(","));
  if (continuation !== null) {
    endpoint.searchParams.set("definitionCursors", continuation);
  }
  return endpoint.toString();
}
