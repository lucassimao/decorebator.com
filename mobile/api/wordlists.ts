import * as SecureStore from "expo-secure-store";
import { AUTH_REQUIRED_ERROR, DEFAULT_ERROR, getAuthorization } from "./users";

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
};

export type CreateWordDTO = Pick<Word, "wordlistId" | "name">;

export async function getUserWordlists() {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + "/wordlists";
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error(AUTH_REQUIRED_ERROR);
  }

  const response = await fetch(endpoint, {
    method: "GET",
    headers: {
      authorization,
    },

    credentials: "include",
  });
  const body = await response.json();
  if (!response.ok) {
    const message =
      body?.error ||
      Object.values(body?.validationErrors)?.[0] ||
      DEFAULT_ERROR;
    throw new Error(message);
  }
  return body;
}

export async function getWords(wordlistId: number): Promise<Word[]> {
  const endpoint =
    process.env.EXPO_PUBLIC_API_URL + `/wordlists/${wordlistId}/words`;
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error(AUTH_REQUIRED_ERROR);
  }

  const response = await fetch(endpoint, {
    method: "GET",
    headers: {
      authorization,
    },
    credentials: "include",
  });
  const body = await response.json();
  if (!response.ok) {
    const message =
      body?.error ||
      Object.values(body?.validationErrors)?.[0] ||
      DEFAULT_ERROR;
    throw new Error(message);
  }
  return body;
}

export async function deleteWordlist(wordlistId: number): Promise<void> {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + `/wordlists/${wordlistId}`;
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error(AUTH_REQUIRED_ERROR);
  }

  const response = await fetch(endpoint, {
    method: "DELETE",
    headers: {
      authorization,
    },
    credentials: "include",
  });
  if (!response.ok) {
    const body = await response.json();
    const message =
      body?.error ||
      Object.values(body?.validationErrors)?.[0] ||
      DEFAULT_ERROR;
    throw new Error(message);
  }
}

export async function getWordlist(wordlistId: number): Promise<Wordlist> {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + `/wordlists/${wordlistId}`;
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error(AUTH_REQUIRED_ERROR);
  }

  const response = await fetch(endpoint, {
    method: "GET",
    headers: {
      authorization,
    },
    credentials: "include",
  });
  const body = await response.json();
  if (!response.ok) {
    const message =
      body?.error ||
      Object.values(body?.validationErrors)?.[0] ||
      DEFAULT_ERROR;
    throw new Error(message);
  }
  return body;
}

export async function addWord(dto: CreateWordDTO): Promise<void> {
  const endpoint =
    process.env.EXPO_PUBLIC_API_URL + `/wordlists/${dto.wordlistId}/words`;
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error(AUTH_REQUIRED_ERROR);
  }

  const response = await fetch(endpoint, {
    method: "POST",
    headers: {
      authorization,
    },
    body: JSON.stringify(dto),
    credentials: "include",
  });
  const body = await response.json();
  if (!response.ok) {
    const message =
      body?.error ||
      Object.values(body?.validationErrors)?.[0] ||
      DEFAULT_ERROR;
    throw new Error(message);
  }
  return body;
}

export async function deleteWord(word: Word): Promise<void> {
  const { wordlistId, id: wordId } = word;

  const endpoint =
    process.env.EXPO_PUBLIC_API_URL +
    `/wordlists/${wordlistId}/words/${wordId}`;
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error(AUTH_REQUIRED_ERROR);
  }

  const response = await fetch(endpoint, {
    method: "DELETE",
    headers: {
      authorization,
    },
    credentials: "include",
  });
  if (!response.ok) {
    const body = await response.json();
    const message =
      body?.error ||
      Object.values(body?.validationErrors)?.[0] ||
      DEFAULT_ERROR;
    throw new Error(message);
  }
}

export async function newQuiz(wordlistId: number): Promise<Quiz> {
  const endpoint =
    process.env.EXPO_PUBLIC_API_URL + `/wordlists/${wordlistId}/quizzes`;
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error(AUTH_REQUIRED_ERROR);
  }

  const response = await fetch(endpoint, {
    method: "POST",
    headers: {
      authorization,
    },
    credentials: "include",
  });
  const body = await response.json();
  if (!response.ok) {
    const message =
      body?.error ||
      Object.values(body?.validationErrors)?.[0] ||
      DEFAULT_ERROR;
    throw new Error(message);
  }
  return body;
}

export async function answerQuiz(
  wordlistId: number,
  quizId: number,
  success: boolean,
): Promise<void> {
  const endpoint =
    process.env.EXPO_PUBLIC_API_URL + `/wordlists/${wordlistId}/quizzes`;
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error(AUTH_REQUIRED_ERROR);
  }

  const response = await fetch(endpoint, {
    method: "PATCH",
    headers: {
      authorization,
    },
    body: JSON.stringify({
      id: quizId,
      success,
    }),
    credentials: "include",
  });
  if (!response.ok) {
    const body = await response.json();
    const message =
      body?.error ||
      Object.values(body?.validationErrors)?.[0] ||
      DEFAULT_ERROR;
    throw new Error(message);
  }
}
