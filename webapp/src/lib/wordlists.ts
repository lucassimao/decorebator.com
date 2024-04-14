"use client";

type Wordlist = {
  name: string;
  id: number;
};

type Word = {
  name: string;
  id: number;
};
export async function getWordlists(): Promise<Wordlist[]> {
  const result = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/wordlists`, {
    credentials: "include",
  });

  if (!result.ok) {
    throw new Error("Failed to fetch wordlists");
  }
  return (await result.json()) as Wordlist[];
}

export async function getWords(wordlistId: number): Promise<Word[]> {
  const result = await fetch(
    `${process.env.NEXT_PUBLIC_API_URL}/wordlists/${wordlistId}/words`,
    { credentials: "include" },
  );

  if (!result.ok) {
    throw new Error("Failed to fetch words");
  }
  return (await result.json()) as Word[];
}
