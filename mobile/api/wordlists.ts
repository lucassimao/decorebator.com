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
    | 'WRITE_WORD_FROM_DEFINITION'
  pos: string; // part of speech
  audioURL?: string; //only present in MeaningFromAudio, GUESS_MEANING and WordFromAudio quizes
  imageDescription: string;
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

export async function answerQuiz(
  wordlistId: number,
  quizId: number,
  success: boolean,
): Promise<void> {
  const endpoint =
    process.env.EXPO_PUBLIC_API_URL + `/wordlists/${wordlistId}/quizzes`;

  await callAPI<Quiz>(
    "PATCH",
    endpoint,
    JSON.stringify({
      id: quizId,
      success,
    }),
  );
}
