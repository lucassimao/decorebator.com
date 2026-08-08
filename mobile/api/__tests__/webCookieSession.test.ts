import { Platform } from "react-native";

jest.mock("@/api/baseUrl", () => ({
  getApiBaseUrl: () => "https://api.example.test",
}));

jest.mock("@/utils/offlineManager", () => ({
  __esModule: true,
  default: {
    clearCache: jest.fn(() => Promise.resolve()),
    setUserPremiumStatus: jest.fn(),
  },
}));

describe("web cookie session", () => {
  const storage = new Map<string, string>();
  const originalPlatform = Object.getOwnPropertyDescriptor(Platform, "OS");

  beforeAll(() => {
    Object.defineProperty(Platform, "OS", { configurable: true, value: "web" });
    Object.defineProperty(global, "localStorage", {
      configurable: true,
      value: {
        getItem: jest.fn((key: string) => storage.get(key) ?? null),
        setItem: jest.fn((key: string, value: string) =>
          storage.set(key, value),
        ),
        removeItem: jest.fn((key: string) => storage.delete(key)),
      },
    });
  });

  afterAll(() => {
    if (originalPlatform) {
      Object.defineProperty(Platform, "OS", originalPlatform);
    }
  });

  beforeEach(() => {
    storage.clear();
    jest.clearAllMocks();
    Object.defineProperty(global.navigator, "locks", {
      configurable: true,
      value: {
        request: jest.fn(
          (
            _name: string,
            _options: LockOptions,
            callback: () => Promise<unknown>,
          ) => callback(),
        ),
      },
    });
  });

  it("purges legacy browser credentials during the initial session check", async () => {
    storage.set("authorization", "legacy-browser-token");
    storage.set("hasAuthenticatedSession", "existing-session");
    global.fetch = jest.fn(() =>
      Promise.resolve({ ok: true, status: 204 } as Response),
    );
    const users =
      jest.requireActual<typeof import("@/api/users")>("@/api/users");

    expect(users.isAuthenticationSessionReady()).toBe(false);
    expect(users.hasAuthenticationSession()).toBe(false);
    expect(storage.get("authorization")).toBeUndefined();
    await Promise.resolve();
    expect(global.fetch).toHaveBeenCalledWith(
      "https://api.example.test/session/legacy-cookie-cleanup",
      expect.objectContaining({ method: "POST", credentials: "include" }),
    );
    await Promise.resolve();
    expect(users.isAuthenticationSessionReady()).toBe(true);
    expect(users.hasAuthenticationSession()).toBe(true);
  });

  it("aborts an authenticated operation that stalls after acquiring the lock", async () => {
    jest.useFakeTimers();
    storage.set("hasAuthenticatedSession", "identity-a");
    global.fetch = jest.fn(
      (_input: RequestInfo | URL, init?: RequestInit) =>
        new Promise<Response>((_resolve, reject) => {
          init?.signal?.addEventListener(
            "abort",
            () => reject(new Error("request aborted")),
            { once: true },
          );
        }),
    ) as jest.Mock;
    const users =
      jest.requireActual<typeof import("@/api/users")>("@/api/users");

    const request = users.authenticatedFetch(
      "https://api.example.test/stalled",
    );
    const rejection = expect(request).rejects.toThrow("request aborted");
    await Promise.resolve();
    await jest.advanceTimersByTimeAsync(15_000);

    await rejection;
    expect(global.navigator.locks.request).toHaveBeenCalledTimes(1);
    jest.useRealTimers();
  });

  it("stores only a noncredential session marker and sends no authorization header", async () => {
    storage.set("authorization", "legacy-browser-token");
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        headers: { get: () => "server-credential-must-be-ignored" },
      } as unknown as Response),
    );
    const users =
      jest.requireActual<typeof import("@/api/users")>("@/api/users");

    await users.signin({ email: "web@example.com", password: "secret" });

    expect(users.getAuthorization()).toBeNull();
    expect(users.hasAuthenticationSession()).toBe(true);
    expect(storage.get("authorization")).toBeUndefined();
    expect(storage.get("hasAuthenticatedSession")).toBeTruthy();
    expect(global.navigator.locks.request).toHaveBeenCalledWith(
      "decorebator-session-refresh",
      expect.objectContaining({ signal: expect.anything() }),
      expect.any(Function),
    );
    const request = (global.fetch as jest.Mock).mock.calls[0][1] as RequestInit;
    expect(new Headers(request.headers).has("Authorization")).toBe(false);
  });

  it("serializes cookie refresh across browser tabs", async () => {
    storage.set("hasAuthenticatedSession", "true");
    let lockQueue = Promise.resolve();
    Object.defineProperty(global.navigator, "locks", {
      configurable: true,
      value: {
        request: jest.fn(
          (
            _name: string,
            _options: LockOptions,
            callback: () => Promise<Response>,
          ) => {
            const result = lockQueue.then(callback);
            lockQueue = result.then(
              () => undefined,
              () => undefined,
            );
            return result;
          },
        ),
      },
    });
    let freshCookie = false;
    let initialProtectedCalls = 0;
    let refreshCalls = 0;
    global.fetch = jest.fn((input: RequestInfo | URL) => {
      if (String(input).endsWith("/session/refresh")) {
        refreshCalls += 1;
        freshCookie = true;
        return Promise.resolve({
          ok: true,
          status: 200,
          headers: { get: () => null },
        } as unknown as Response);
      }
      initialProtectedCalls += 1;
      return Promise.resolve({
        ok: freshCookie,
        status: freshCookie ? 200 : 401,
        headers: { get: () => null },
      } as unknown as Response);
    }) as jest.Mock;
    const users =
      jest.requireActual<typeof import("@/api/users")>("@/api/users");

    const [first, second] = await Promise.all([
      users.authenticatedFetch("https://api.example.test/users"),
      users.authenticatedFetch("https://api.example.test/users"),
    ]);

    expect(first.status).toBe(200);
    expect(second.status).toBe(200);
    expect(refreshCalls).toBe(1);
    expect(initialProtectedCalls).toBe(3);
  });

  it("serializes browser logout under the credential mutation lock", async () => {
    storage.set("hasAuthenticatedSession", "identity-a");
    global.fetch = jest.fn(() =>
      Promise.resolve({ ok: true, status: 204 } as Response),
    );
    const users =
      jest.requireActual<typeof import("@/api/users")>("@/api/users");

    await users.sigout();

    expect(global.navigator.locks.request).toHaveBeenCalledWith(
      "decorebator-session-refresh",
      expect.objectContaining({ signal: expect.anything() }),
      expect.any(Function),
    );
    expect(storage.get("hasAuthenticatedSession")).toBeUndefined();
  });

  it("does not replay an old browser request under a replacement identity", async () => {
    storage.set("hasAuthenticatedSession", "identity-a");
    let releaseLock!: () => void;
    const lockGate = new Promise<void>((resolve) => {
      releaseLock = resolve;
    });
    Object.defineProperty(global.navigator, "locks", {
      configurable: true,
      value: {
        request: jest.fn(
          async (
            _name: string,
            _options: LockOptions,
            callback: () => Promise<Response>,
          ) => {
            await lockGate;
            return callback();
          },
        ),
      },
    });
    global.fetch = jest.fn(() =>
      Promise.resolve({ ok: false, status: 401 } as Response),
    );
    const users =
      jest.requireActual<typeof import("@/api/users")>("@/api/users");

    const staleRequest = users.authenticatedFetch(
      "https://api.example.test/protected",
    );
    await Promise.resolve();
    storage.set("hasAuthenticatedSession", "identity-b");
    releaseLock();

    await expect(staleRequest).rejects.toThrow(
      "Session changed during request",
    );
    expect(global.fetch).not.toHaveBeenCalled();
  });

  it("rejects a successful response from the previous browser identity", async () => {
    storage.set("hasAuthenticatedSession", "identity-a");
    let resolveRequest!: (response: Response) => void;
    global.fetch = jest.fn(
      () =>
        new Promise<Response>((resolve) => {
          resolveRequest = resolve;
        }),
    );
    const users =
      jest.requireActual<typeof import("@/api/users")>("@/api/users");

    const staleRequest = users.authenticatedFetch(
      "https://api.example.test/protected",
    );
    storage.set("hasAuthenticatedSession", "identity-b");
    resolveRequest({ ok: true, status: 200 } as Response);

    await expect(staleRequest).rejects.toThrow(
      "Session changed during request",
    );
  });

  it("rejects response status and headers after browser identity handoff", async () => {
    storage.set("hasAuthenticatedSession", "identity-a");
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 204,
        headers: new Headers({ "X-Owner": "identity-a" }),
      } as Response),
    );
    const users =
      jest.requireActual<typeof import("@/api/users")>("@/api/users");

    const response = await users.authenticatedFetch(
      "https://api.example.test/destructive",
    );
    const headers = response.headers;
    const entries = headers.entries();
    storage.set("hasAuthenticatedSession", "identity-b");

    expect(() => response.status).toThrow("Session changed during request");
    expect(() => headers.get("X-Owner")).toThrow(
      "Session changed during request",
    );
    expect(() => entries.next()).toThrow("Session changed during request");
  });

  it("guards every header callback and disallows response cloning", async () => {
    storage.set("hasAuthenticatedSession", "identity-a");
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        headers: new Headers({ A: "one", B: "two" }),
      } as Response),
    );
    const users =
      jest.requireActual<typeof import("@/api/users")>("@/api/users");
    const response = await users.authenticatedFetch(
      "https://api.example.test/headers",
    );
    let callbackCount = 0;

    expect(() =>
      response.headers.forEach(() => {
        callbackCount += 1;
        if (callbackCount === 1) {
          storage.set("hasAuthenticatedSession", "identity-b");
        }
      }),
    ).toThrow("Session changed during request");
    expect(callbackCount).toBe(1);

    storage.set("hasAuthenticatedSession", "identity-a");
    expect(() => response.clone()).toThrow(
      "Cloning authenticated responses is unsupported",
    );
  });

  it("rejects a retained raw-stream reader after identity handoff", async () => {
    storage.set("hasAuthenticatedSession", "identity-a");
    const source = new ReadableStream<Uint8Array>({
      start(controller) {
        controller.enqueue(new Uint8Array([65]));
      },
    });
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        headers: new Headers(),
        body: source,
      } as Response),
    );
    const users =
      jest.requireActual<typeof import("@/api/users")>("@/api/users");
    const response = await users.authenticatedFetch(
      "https://api.example.test/stream",
    );
    const reader = response.body!.getReader();
    storage.set("hasAuthenticatedSession", "identity-b");

    await expect(reader.read()).rejects.toThrow(
      "Session changed during request",
    );
    storage.set("hasAuthenticatedSession", "identity-a");
    await reader.cancel();
  });

  it("rejects a request queued before a network-failed logout", async () => {
    storage.set("hasAuthenticatedSession", "identity-a");
    let lockQueue = Promise.resolve();
    Object.defineProperty(global.navigator, "locks", {
      configurable: true,
      value: {
        request: jest.fn(
          (
            _name: string,
            _options: LockOptions,
            callback: () => Promise<Response>,
          ) => {
            const result = lockQueue.then(callback);
            lockQueue = result.then(
              () => undefined,
              () => undefined,
            );
            return result;
          },
        ),
      },
    });
    global.fetch = jest.fn((input: RequestInfo | URL) => {
      if (String(input).endsWith("/logout")) {
        return Promise.reject(new Error("network failed"));
      }
      return Promise.resolve({ ok: true, status: 200 } as Response);
    }) as jest.Mock;
    const users =
      jest.requireActual<typeof import("@/api/users")>("@/api/users");

    const logout = users.sigout();
    const queuedRequest = users.authenticatedFetch(
      "https://api.example.test/protected",
    );
    const logoutRejection = expect(logout).rejects.toThrow("network failed");
    const queuedRejection = expect(queuedRequest).rejects.toThrow(
      "Session changed during request",
    );

    await logoutRejection;
    await queuedRejection;
    expect(global.fetch).toHaveBeenCalledTimes(1);
    expect(users.isAuthenticationSessionReady()).toBe(false);
    expect(users.getAuthenticationSessionEpoch()).toBeNull();
  });

  it("keeps the authenticated timeout active while response JSON stalls", async () => {
    jest.useFakeTimers();
    storage.set("hasAuthenticatedSession", "identity-a");
    global.fetch = jest.fn((_input: RequestInfo | URL, init?: RequestInit) =>
      Promise.resolve({
        ok: true,
        status: 200,
        headers: new Headers(),
        body: {} as ReadableStream<Uint8Array>,
        json: () =>
          new Promise((_resolve, reject) => {
            init?.signal?.addEventListener(
              "abort",
              () => reject(new Error("body aborted")),
              { once: true },
            );
          }),
      } as unknown as Response),
    ) as jest.Mock;
    const users =
      jest.requireActual<typeof import("@/api/users")>("@/api/users");

    const response = await users.authenticatedFetch(
      "https://api.example.test/stalled-body",
    );
    const body = response.json();
    const rejection = expect(body).rejects.toThrow("body aborted");
    await jest.advanceTimersByTimeAsync(15_000);

    await rejection;
    jest.useRealTimers();
  });

  it("rejects a body that finishes under a replacement browser identity", async () => {
    storage.set("hasAuthenticatedSession", "identity-a");
    let resolveBody!: (body: { owner: string }) => void;
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        json: () =>
          new Promise((resolve) => {
            resolveBody = resolve;
          }),
      } as unknown as Response),
    );
    const users =
      jest.requireActual<typeof import("@/api/users")>("@/api/users");

    const response = await users.authenticatedFetch(
      "https://api.example.test/protected",
    );
    const body = response.json();
    storage.set("hasAuthenticatedSession", "identity-b");
    resolveBody({ owner: "identity-a" });

    await expect(body).rejects.toThrow("Session changed during request");
  });
});
