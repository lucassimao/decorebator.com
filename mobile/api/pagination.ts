import { callAPIWithMetadata, type APIResponseWithMetadata } from "./api";
import { getAuthenticationSessionEpoch } from "./users";

export const PAGINATION_PAGE_SIZE = 100;
export const MAX_PAGINATION_PAGES = 100;
export const MAX_PAGINATION_ITEMS = 10_000;

export class PaginationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "PaginationError";
  }
}

type GetAllPagesOptions<Page, Item> = {
  endpoint: string;
  getItems: (page: Page) => Item[];
  getItemKey: (item: Item) => string;
  onPage?: (page: Page) => void;
  maxPages?: number;
  maxItems?: number;
};

function pageEndpoint(endpoint: string, cursor: string | null): string {
  const url = new URL(endpoint);
  url.searchParams.set("limit", String(PAGINATION_PAGE_SIZE));
  if (cursor !== null) url.searchParams.set("cursor", cursor);
  return url.toString();
}

export function createPaginationSessionGuard(): () => void {
  const expectedSessionEpoch = getAuthenticationSessionEpoch();
  return () => {
    if (
      expectedSessionEpoch === null ||
      getAuthenticationSessionEpoch() !== expectedSessionEpoch
    ) {
      throw new PaginationError("Session changed during paginated request");
    }
  };
}

// The API emits an opaque X-Next-Cursor header while keeping its established
// response bodies. This helper owns the continuation protocol so callers can
// retain those response shapes without silently dropping later pages.
export async function getAllPages<Page, Item>(
  options: GetAllPagesOptions<Page, Item>,
): Promise<Item[]> {
  const maxPages = options.maxPages ?? MAX_PAGINATION_PAGES;
  const maxItems = options.maxItems ?? MAX_PAGINATION_ITEMS;
  if (!Number.isInteger(maxPages) || maxPages < 1) {
    throw new PaginationError("Pagination page ceiling must be at least one");
  }
  if (!Number.isInteger(maxItems) || maxItems < 1) {
    throw new PaginationError("Pagination item ceiling must be at least one");
  }

  const assertSessionUnchanged = createPaginationSessionGuard();
  assertSessionUnchanged();

  const items: Item[] = [];
  const seenItemKeys = new Set<string>();
  const seenCursors = new Set<string>();
  let cursor: string | null = null;

  for (let page = 0; page < maxPages; page += 1) {
    assertSessionUnchanged();
    const response: APIResponseWithMetadata<Page> =
      await callAPIWithMetadata<Page>(
        "GET",
        pageEndpoint(options.endpoint, cursor),
      );
    assertSessionUnchanged();

    const pageItems = options.getItems(response.body);
    if (!Array.isArray(pageItems)) {
      throw new PaginationError("Pagination response items must be an array");
    }
    if (pageItems.length === 0 && response.nextCursor !== null) {
      throw new PaginationError(
        "Pagination response cannot continue after an empty page",
      );
    }
    if (items.length + pageItems.length > maxItems) {
      throw new PaginationError("Pagination item ceiling reached");
    }

    for (const item of pageItems) {
      const itemKey = options.getItemKey(item);
      if (!itemKey) {
        throw new PaginationError(
          "Pagination response contains an invalid item",
        );
      }
      if (seenItemKeys.has(itemKey)) {
        throw new PaginationError(
          "Pagination response contains a duplicate item",
        );
      }
      seenItemKeys.add(itemKey);
      items.push(item);
    }
    options.onPage?.(response.body);

    const nextCursor = response.nextCursor;
    if (nextCursor === null) return items;
    if (seenCursors.has(nextCursor)) {
      throw new PaginationError(
        "Pagination response contains a repeated continuation cursor",
      );
    }
    seenCursors.add(nextCursor);
    cursor = nextCursor;
  }

  throw new PaginationError("Pagination page ceiling reached");
}
