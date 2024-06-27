import { callAPI } from "./api";

export type Wordlist = {
  createdAt: string;
  description: string;
  id: number;
  name: string;
  updatedAt: string;
  userId: number;
};

export type Word = {
  id: number;
  name: string;
  wordlistId: number;
};

export type Quiz = {
  value: string;
  options: string[];
  answerIndex: number;
  id: number;
  type: number;
  pos: string // part of speech
};

export type CreateWordDTO = Pick<Word, "wordlistId" | "name">;
export type CreateWordlistDTO = Pick<Wordlist, "description" | "name">;

export async function getUserWordlists() {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + "/wordlists";
  const body = await callAPI("GET", endpoint);
  return body;
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

export async function addWordlist(dto: CreateWordlistDTO): Promise<void> {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + `/wordlists`;
  await callAPI("POST", endpoint, JSON.stringify(dto));
}

export async function addWord(dto: CreateWordDTO): Promise<void> {
  const endpoint =
    process.env.EXPO_PUBLIC_API_URL + `/wordlists/${dto.wordlistId}/words`;

  await callAPI("POST", endpoint, JSON.stringify(dto));
}

export async function deleteWord(word: Word): Promise<void> {
  const { wordlistId, id: wordId } = word;

  const endpoint =
    process.env.EXPO_PUBLIC_API_URL +
    `/wordlists/${wordlistId}/words/${wordId}`;

  await callAPI("DELETE", endpoint);
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
