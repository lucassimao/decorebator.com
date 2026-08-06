import { callAPI } from "@/api/api";
import { updateWord, type Word } from "@/api/wordlists";

jest.mock("@/api/api", () => ({
  callAPI: jest.fn(),
}));
jest.mock("@/api/baseUrl", () => ({
  getApiBaseUrl: () => "https://api.test",
}));

describe("wordlist API", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (callAPI as jest.Mock).mockResolvedValue(undefined);
  });

  it("serializes only user-editable word fields", async () => {
    const serverWord: Word = {
      id: 7,
      wordlistId: 3,
      name: "safe",
      notes: "keep",
      learned: true,
      audioURL: "https://media.example/server.mp3",
      pronunciation: "server-owned",
      processingStatus: "completed",
      processingError: "server-owned",
    };

    await updateWord({ ...serverWord });

    expect(callAPI).toHaveBeenCalledWith(
      "PUT",
      "https://api.test/wordlists/3/words/7",
      JSON.stringify({
        learned: true,
        name: "safe",
        notes: "keep",
        wordlistId: 3,
      }),
    );
  });

  it("preserves explicit zero values while omitting absent patch fields", async () => {
    await updateWord({ id: 9, wordlistId: 4, learned: false, notes: "" });

    expect(callAPI).toHaveBeenCalledWith(
      "PUT",
      "https://api.test/wordlists/4/words/9",
      JSON.stringify({ learned: false, notes: "", wordlistId: 4 }),
    );
  });
});
