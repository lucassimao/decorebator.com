import { callAPIWithMetadata } from "@/api/api";
import {
  MAX_PAGINATION_ITEMS,
  MAX_PAGINATION_PAGES,
  getAllPages,
} from "@/api/pagination";
import { getAuthenticationSessionEpoch } from "@/api/users";

jest.mock("@/api/api", () => ({
  callAPIWithMetadata: jest.fn(),
}));

jest.mock("@/api/users", () => ({
  getAuthenticationSessionEpoch: jest.fn(),
}));

type Page = { items: { id: number; value: string }[] };

const getItems = (page: Page) => page.items;
const getItemKey = (item: { id: number }) => String(item.id);

describe("bounded API pagination", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.mocked(getAuthenticationSessionEpoch).mockReturnValue("native:1");
  });

  it("follows opaque cursors, preserves query parameters, and aggregates every page", async () => {
    jest
      .mocked(callAPIWithMetadata)
      .mockResolvedValueOnce({
        body: { items: [{ id: 3, value: "first" }] },
        nextCursor: "opaque+/=cursor",
      })
      .mockResolvedValueOnce({
        body: { items: [{ id: 2, value: "second" }] },
        nextCursor: null,
      });

    await expect(
      getAllPages({
        endpoint: "https://api.test/words?onlyWithDefinitions=true",
        getItems,
        getItemKey,
      }),
    ).resolves.toEqual([
      { id: 3, value: "first" },
      { id: 2, value: "second" },
    ]);

    expect(callAPIWithMetadata).toHaveBeenNthCalledWith(
      1,
      "GET",
      "https://api.test/words?onlyWithDefinitions=true&limit=100",
    );
    expect(callAPIWithMetadata).toHaveBeenNthCalledWith(
      2,
      "GET",
      "https://api.test/words?onlyWithDefinitions=true&limit=100&cursor=opaque%2B%2F%3Dcursor",
    );
  });

  it("rejects a repeated continuation cursor instead of looping", async () => {
    jest
      .mocked(callAPIWithMetadata)
      .mockResolvedValueOnce({
        body: { items: [{ id: 3, value: "first" }] },
        nextCursor: "same-cursor",
      })
      .mockResolvedValueOnce({
        body: { items: [{ id: 2, value: "second" }] },
        nextCursor: "same-cursor",
      });

    await expect(
      getAllPages({
        endpoint: "https://api.test/words",
        getItems,
        getItemKey,
      }),
    ).rejects.toThrow("repeated continuation cursor");
  });

  it("rejects duplicate items across pages instead of double-counting them", async () => {
    jest
      .mocked(callAPIWithMetadata)
      .mockResolvedValueOnce({
        body: { items: [{ id: 3, value: "first" }] },
        nextCursor: "next",
      })
      .mockResolvedValueOnce({
        body: { items: [{ id: 3, value: "duplicate" }] },
        nextCursor: null,
      });

    await expect(
      getAllPages({
        endpoint: "https://api.test/words",
        getItems,
        getItemKey,
      }),
    ).rejects.toThrow("duplicate item");
  });

  it("fails explicitly when an API error occurs during a later page", async () => {
    const failure = new Error("request failed");
    jest
      .mocked(callAPIWithMetadata)
      .mockResolvedValueOnce({
        body: { items: [{ id: 3, value: "first" }] },
        nextCursor: "next",
      })
      .mockRejectedValueOnce(failure);

    await expect(
      getAllPages({
        endpoint: "https://api.test/words",
        getItems,
        getItemKey,
      }),
    ).rejects.toBe(failure);
  });

  it("fails explicitly at the configured item and page ceilings", async () => {
    jest.mocked(callAPIWithMetadata).mockResolvedValue({
      body: {
        items: Array.from({ length: 100 }, (_, index) => ({
          id: index + 1,
          value: String(index + 1),
        })),
      },
      nextCursor: "next",
    });

    await expect(
      getAllPages({
        endpoint: "https://api.test/words",
        getItems,
        getItemKey,
        maxItems: 99,
      }),
    ).rejects.toThrow("item ceiling");

    jest.mocked(callAPIWithMetadata).mockReset();
    jest.mocked(callAPIWithMetadata).mockResolvedValue({
      body: { items: [{ id: 1, value: "one" }] },
      nextCursor: "next",
    });

    await expect(
      getAllPages({
        endpoint: "https://api.test/words",
        getItems,
        getItemKey,
        maxPages: 1,
      }),
    ).rejects.toThrow("page ceiling");

    expect(MAX_PAGINATION_ITEMS).toBeGreaterThan(0);
    expect(MAX_PAGINATION_PAGES).toBeGreaterThan(0);
  });

  it("does not aggregate pages after the authenticated identity changes", async () => {
    jest.mocked(callAPIWithMetadata).mockResolvedValueOnce({
      body: { items: [{ id: 3, value: "first" }] },
      nextCursor: "next",
    });
    jest
      .mocked(getAuthenticationSessionEpoch)
      .mockReturnValueOnce("native:1")
      .mockReturnValueOnce("native:1")
      .mockReturnValueOnce("native:1")
      .mockReturnValue("native:2");

    await expect(
      getAllPages({
        endpoint: "https://api.test/words",
        getItems,
        getItemKey,
      }),
    ).rejects.toThrow("Session changed during paginated request");
    expect(callAPIWithMetadata).toHaveBeenCalledTimes(1);
  });
});
