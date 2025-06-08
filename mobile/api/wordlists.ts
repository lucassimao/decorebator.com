import { callAPI } from "./api";

export type Wordlist = {
  createdAt: string;
  description: string;
  id: number;
  name: string;
  updatedAt: string;
  userId: number;
  languageCode: string;
  wordsCount: number;
  wordsLearnedCount: number;
};

export type UserStats = {
  totalWords: number;
  wordlists: number;
  wordsLearned: number;
  currentStreak?: number;
};

export type Word = {
  id: number;
  name: string;
  wordlistId: number;
  learned: boolean;
  pronunciation?: string;
  notes?: string;
  audioURL?: string;
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
    | "WRITE_WORD_FROM_DEFINITION"
    | "WORD_FROM_EXAMPLE_AUDIO";
  pos: string; // part of speech
  isVerbType: boolean; // computed flag indicating if this is a verb/phrasal verb
  pronunciation?: string; // IPA pronunciation
  audioURL?: string; //only present in MeaningFromAudio, GUESS_MEANING, WordFromAudio, and WordFromExampleAudio quizes
  imageDescription: string;
  definitionId: number;
  wordId: number;
};

export type CreateWordDTO = Pick<Word, "wordlistId" | "name" | "notes">;
export type CreateWordlistDTO = Pick<
  Wordlist,
  "description" | "name" | "languageCode"
>;

export async function getUserWordlists() {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + "/wordlists";
  const body = await callAPI<Wordlist[]>("GET", endpoint);
  return body;
}

export async function getUserStats(): Promise<UserStats> {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + `/wordlists/stats`;

  const body = await callAPI<UserStats>("GET", endpoint);

  return { ...body, currentStreak: 10 };
}

export async function getWords(wordlistId: number): Promise<Word[]> {
  const endpoint =
    process.env.EXPO_PUBLIC_API_URL + `/wordlists/${wordlistId}/words`;

  const body = await callAPI<Word[]>("GET", endpoint);
  return body;
}

export async function getWordsWithDefinitions(
  wordlistId: number,
): Promise<Word[]> {
  const endpoint =
    process.env.EXPO_PUBLIC_API_URL +
    `/wordlists/${wordlistId}/words?onlyWithDefinitions=true`;

  const body = await callAPI<Word[]>("GET", endpoint);
  return body;
}

export async function deleteWordlist(wordlistId: number): Promise<void> {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + `/wordlists/${wordlistId}`;
  await callAPI("DELETE", endpoint);
}

export async function getWordlist(wordlistId: number): Promise<Wordlist> {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + `/wordlists/${wordlistId}`;
  return await callAPI<Wordlist>("GET", endpoint);
}

export async function addWordlist(dto: CreateWordlistDTO): Promise<Wordlist> {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + `/wordlists`;
  return await callAPI<Wordlist>("POST", endpoint, JSON.stringify(dto));
}

export async function addWord(dto: CreateWordDTO): Promise<void> {
  const endpoint =
    process.env.EXPO_PUBLIC_API_URL + `/wordlists/${dto.wordlistId}/words`;

  await callAPI("POST", endpoint, JSON.stringify(dto));
}

export async function deleteWord(
  word: Pick<Word, "id" | "wordlistId">,
): Promise<void> {
  const { wordlistId, id: wordId } = word;

  const endpoint =
    process.env.EXPO_PUBLIC_API_URL +
    `/wordlists/${wordlistId}/words/${wordId}`;

  await callAPI("DELETE", endpoint);
}

export async function updateWord(
  dto: Pick<Word, "id" | "wordlistId" | "learned" | "name" | "notes">,
) {
  const endpoint =
    process.env.EXPO_PUBLIC_API_URL +
    `/wordlists/${dto.wordlistId}/words/${dto.id}`;
  await callAPI("PUT", endpoint, JSON.stringify(dto));
}
export async function newQuiz(wordlistId: number): Promise<Quiz> {
  const endpoint =
    process.env.EXPO_PUBLIC_API_URL + `/wordlists/${wordlistId}/quizzes`;
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
  const endpoint =
    process.env.EXPO_PUBLIC_API_URL + `/wordlists/${input.wordlistID}/quizzes`;

  await callAPI<Quiz>("PATCH", endpoint, JSON.stringify(input));
}

export async function getWordDefinitions(
  wordlistId: number,
  wordId: number,
): Promise<Definition[]> {
  const endpoint =
    process.env.EXPO_PUBLIC_API_URL +
    `/wordlists/${wordlistId}/words/${wordId}/definitions`;

  const body = await callAPI<Definition[]>("GET", endpoint);
  return body;
}
