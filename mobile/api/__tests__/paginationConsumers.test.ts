import { callAPIWithMetadata } from "@/api/api";
import { getProgressSummary, getWordlistWordMastery } from "@/api/analytics";
import { getAuthenticationSessionEpoch } from "@/api/users";
import {
  getProcessingStatus,
  getDefinitionsForWords,
  getUserWordlists,
  getWordDefinitions,
  getWords,
} from "@/api/wordlists";

jest.mock("@/api/api", () => ({
  callAPIWithMetadata: jest.fn(),
}));

jest.mock("@/api/users", () => ({
  getAuthenticationSessionEpoch: jest.fn(),
}));

jest.mock("@/api/baseUrl", () => ({
  getApiBaseUrl: () => "https://api.test",
}));

describe("paginated mobile API consumers", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.mocked(getAuthenticationSessionEpoch).mockReturnValue("native:1");
  });

  it("returns complete wordlist, word, and definition arrays in their existing shapes", async () => {
    jest
      .mocked(callAPIWithMetadata)
      .mockResolvedValueOnce({
        body: [{ id: 3, name: "Newest" }],
        nextCursor: "wl-next",
      })
      .mockResolvedValueOnce({
        body: [{ id: 2, name: "Older" }],
        nextCursor: null,
      });
    await expect(getUserWordlists()).resolves.toEqual([
      { id: 3, name: "Newest" },
      { id: 2, name: "Older" },
    ]);

    jest
      .mocked(callAPIWithMetadata)
      .mockResolvedValueOnce({
        body: [{ id: 9, name: "alpha" }],
        nextCursor: "word-next",
      })
      .mockResolvedValueOnce({
        body: [{ id: 8, name: "beta" }],
        nextCursor: null,
      });
    await expect(getWords(7, true)).resolves.toEqual([
      { id: 9, name: "alpha" },
      { id: 8, name: "beta" },
    ]);

    jest
      .mocked(callAPIWithMetadata)
      .mockResolvedValueOnce({
        body: [{ id: 14, meaning: "first" }],
        nextCursor: "definition-next",
      })
      .mockResolvedValueOnce({
        body: [{ id: 13, meaning: "second" }],
        nextCursor: null,
      });
    await expect(getWordDefinitions(7, 9)).resolves.toEqual([
      { id: 14, meaning: "first" },
      { id: 13, meaning: "second" },
    ]);
  });

  it("returns the complete processing words and server-wide latest summary", async () => {
    jest
      .mocked(callAPIWithMetadata)
      .mockResolvedValueOnce({
        body: {
          words: [{ id: 3, name: "alpha", processingStatus: "pending" }],
          summary: {
            total: 2,
            pending: 1,
            processing: 0,
            completed: 1,
            failed: 0,
          },
        },
        nextCursor: "processing-next",
      })
      .mockResolvedValueOnce({
        body: {
          words: [{ id: 2, name: "beta", processingStatus: "completed" }],
          summary: {
            total: 2,
            pending: 1,
            processing: 0,
            completed: 1,
            failed: 0,
          },
        },
        nextCursor: null,
      });

    await expect(getProcessingStatus(7)).resolves.toEqual({
      words: [
        { id: 3, name: "alpha", processingStatus: "pending" },
        { id: 2, name: "beta", processingStatus: "completed" },
      ],
      summary: { total: 2, pending: 1, processing: 0, completed: 1, failed: 0 },
    });
  });

  it("aggregates mastery and progress-summary envelopes without changing their public shape", async () => {
    jest
      .mocked(callAPIWithMetadata)
      .mockResolvedValueOnce({
        body: { stats: [{ wordId: 9, masteryLevel: 1 }] },
        nextCursor: "mastery-next",
      })
      .mockResolvedValueOnce({
        body: { stats: [{ wordId: 8, masteryLevel: 0.5 }] },
        nextCursor: null,
      });
    await expect(getWordlistWordMastery(7)).resolves.toEqual([
      { wordId: 9, masteryLevel: 1 },
      { wordId: 8, masteryLevel: 0.5 },
    ]);

    jest
      .mocked(callAPIWithMetadata)
      .mockResolvedValueOnce({
        body: { wordlists: [{ wordlistId: 9, wordlistName: "A" }] },
        nextCursor: "progress-next",
      })
      .mockResolvedValueOnce({
        body: { wordlists: [{ wordlistId: 8, wordlistName: "B" }] },
        nextCursor: null,
      });
    await expect(getProgressSummary()).resolves.toEqual({
      wordlists: [
        { wordlistId: 9, wordlistName: "A" },
        { wordlistId: 8, wordlistName: "B" },
      ],
    });
  });

  it("fails instead of replacing a malformed paginated envelope with an empty result", async () => {
    jest.mocked(callAPIWithMetadata).mockResolvedValueOnce({
      body: { stats: null },
      nextCursor: null,
    });

    await expect(getWordlistWordMastery(7)).rejects.toThrow(
      "Invalid paginated word mastery stats response: expected an array",
    );
  });

  it("echoes cumulative batched-definition continuations until every word completes", async () => {
    const alphaPage = (start: number, end: number) =>
      Array.from({ length: end - start + 1 }, (_, index) => ({
        id: start + index,
        meaning: `alpha-${start + index}`,
      }));
    const betaPage = (start: number, end: number) =>
      Array.from({ length: end - start + 1 }, (_, index) => ({
        id: start + index,
        meaning: `beta-${start + index}`,
      }));
    // Base64url for {"8":50,"9":50} and {"8":60,"9":100}.
    const firstCumulativeCursor = "eyI4Ijo1MCwiOSI6NTB9";
    const secondCumulativeCursor = "eyI4Ijo2MCwiOSI6MTAwfQ";

    jest
      .mocked(callAPIWithMetadata)
      .mockResolvedValueOnce({
        body: [
          {
            wordId: 9,
            name: "alpha",
            definitions: alphaPage(1, 50),
          },
          {
            wordId: 8,
            name: "beta",
            definitions: betaPage(1, 50),
          },
        ],
        nextCursor: null,
        definitionsContinuation: firstCumulativeCursor,
      })
      .mockResolvedValueOnce({
        body: [
          {
            wordId: 9,
            name: "alpha",
            definitions: alphaPage(51, 100),
          },
          {
            wordId: 8,
            name: "beta",
            definitions: betaPage(51, 60),
          },
        ],
        nextCursor: null,
        definitionsContinuation: secondCumulativeCursor,
      })
      .mockResolvedValueOnce({
        body: [
          {
            wordId: 9,
            name: "alpha",
            definitions: alphaPage(101, 120),
          },
        ],
        nextCursor: null,
        definitionsContinuation: null,
      });

    const result = await getDefinitionsForWords(7, [9, 8]);
    expect(result.map((word) => [word.name, word.definitions.length])).toEqual([
      ["alpha", 120],
      ["beta", 60],
    ]);
    expect(
      new Set(result[0].definitions.map((definition) => definition.id)).size,
    ).toBe(120);
    expect(
      new Set(result[1].definitions.map((definition) => definition.id)).size,
    ).toBe(60);
    expect(callAPIWithMetadata).toHaveBeenNthCalledWith(
      2,
      "GET",
      `https://api.test/wordlists/7/words/definitions?ids=9%2C8&definitionCursors=${firstCumulativeCursor}`,
    );
    expect(callAPIWithMetadata).toHaveBeenNthCalledWith(
      3,
      "GET",
      `https://api.test/wordlists/7/words/definitions?ids=9%2C8&definitionCursors=${secondCumulativeCursor}`,
    );
  });

  it("completes uneven 11-page and 13-page words without duplicates", async () => {
    const cumulativeCursors = [
      "eyI4IjoyMDAxLCI5IjoxMDAxfQ",
      "eyI4IjoyMDAyLCI5IjoxMDAyfQ",
      "eyI4IjoyMDAzLCI5IjoxMDAzfQ",
      "eyI4IjoyMDA0LCI5IjoxMDA0fQ",
      "eyI4IjoyMDA1LCI5IjoxMDA1fQ",
      "eyI4IjoyMDA2LCI5IjoxMDA2fQ",
      "eyI4IjoyMDA3LCI5IjoxMDA3fQ",
      "eyI4IjoyMDA4LCI5IjoxMDA4fQ",
      "eyI4IjoyMDA5LCI5IjoxMDA5fQ",
      "eyI4IjoyMDEwLCI5IjoxMDEwfQ",
      "eyI4IjoyMDExLCI5IjoxMDExfQ",
      "eyI4IjoyMDEyLCI5IjoxMDExfQ",
    ];

    for (let page = 1; page <= 13; page += 1) {
      const body = [
        ...(page <= 11
          ? [
              {
                wordId: 9,
                name: "alpha",
                definitions: [{ id: 1_000 + page, meaning: `alpha-${page}` }],
              },
            ]
          : []),
        {
          wordId: 8,
          name: "beta",
          definitions: [{ id: 2_000 + page, meaning: `beta-${page}` }],
        },
      ];
      jest.mocked(callAPIWithMetadata).mockResolvedValueOnce({
        body,
        nextCursor: null,
        definitionsContinuation: page < 13 ? cumulativeCursors[page - 1] : null,
      });
    }

    const result = await getDefinitionsForWords(7, [9, 8]);
    expect(result.map((word) => [word.name, word.definitions.length])).toEqual([
      ["alpha", 11],
      ["beta", 13],
    ]);
    expect(
      new Set(result[0].definitions.map((definition) => definition.id)).size,
    ).toBe(11);
    expect(
      new Set(result[1].definitions.map((definition) => definition.id)).size,
    ).toBe(13);
    expect(callAPIWithMetadata).toHaveBeenCalledTimes(13);
    expect(callAPIWithMetadata).toHaveBeenNthCalledWith(
      13,
      "GET",
      `https://api.test/wordlists/7/words/definitions?ids=9%2C8&definitionCursors=${cumulativeCursors[11]}`,
    );
  });

  it("splits batched definitions into server-bounded requests", async () => {
    jest
      .mocked(callAPIWithMetadata)
      .mockResolvedValueOnce({
        body: [
          {
            wordId: 1,
            name: "first",
            definitions: [{ id: 1, meaning: "one" }],
          },
        ],
        nextCursor: null,
      })
      .mockResolvedValueOnce({
        body: [
          {
            wordId: 51,
            name: "last",
            definitions: [{ id: 51, meaning: "fifty one" }],
          },
        ],
        nextCursor: null,
      });

    await expect(
      getDefinitionsForWords(
        7,
        Array.from({ length: 51 }, (_, index) => index + 1),
      ),
    ).resolves.toHaveLength(2);
    expect(callAPIWithMetadata).toHaveBeenCalledTimes(2);
    expect(callAPIWithMetadata).toHaveBeenNthCalledWith(
      2,
      "GET",
      "https://api.test/wordlists/7/words/definitions?ids=51",
    );
  });
});
