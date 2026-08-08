import AsyncStorage from "@react-native-async-storage/async-storage";
import * as FileSystem from "expo-file-system/legacy";
import type { Wordlist } from "@/api/wordlists";

jest.mock("expo-file-system/legacy", () => ({
  documentDirectory: "file:///documents/",
  deleteAsync: jest.fn(() => Promise.resolve()),
  getInfoAsync: jest.fn(() => Promise.resolve({ exists: true })),
  makeDirectoryAsync: jest.fn(() => Promise.resolve()),
}));

jest.mock("@react-native-community/netinfo", () => ({
  __esModule: true,
  default: {
    configure: jest.fn(),
    fetch: jest.fn(() =>
      Promise.resolve({ isConnected: true, isInternetReachable: true }),
    ),
    addEventListener: jest.fn(() => jest.fn()),
  },
}));

const offlineManager = jest.requireActual<
  typeof import("@/utils/offlineManager")
>("@/utils/offlineManager").default;

describe("offline cache cleanup", () => {
  afterAll(() => offlineManager.destroy());

  it("still deletes asset files when AsyncStorage listing fails", async () => {
    (AsyncStorage.getAllKeys as jest.Mock).mockRejectedValueOnce(
      new Error("storage unavailable"),
    );
    (FileSystem.deleteAsync as jest.Mock).mockClear();

    await expect(offlineManager.clearCache()).rejects.toThrow(
      "Offline cache cleanup incomplete",
    );

    expect(FileSystem.deleteAsync).toHaveBeenCalledWith(
      "file:///documents/decorebator_assets/",
      { idempotent: true },
    );
  });

  it("waits for and removes an identity-A write that finishes during cleanup", async () => {
    const originalSetItem = (
      AsyncStorage.setItem as jest.Mock
    ).getMockImplementation();
    let releaseWrite!: () => void;
    let writeStarted!: () => void;
    const started = new Promise<void>((resolve) => {
      writeStarted = resolve;
    });
    const gate = new Promise<void>((resolve) => {
      releaseWrite = resolve;
    });
    (AsyncStorage.setItem as jest.Mock).mockImplementationOnce(
      async (key: string, value: string) => {
        writeStarted();
        await gate;
        return originalSetItem?.(key, value);
      },
    );

    const staleWrite = offlineManager.cacheWordlists([
      { id: 1, name: "Identity A" },
    ] as unknown as Wordlist[]);
    await started;
    const cleanup = offlineManager.clearCache();
    releaseWrite();

    await expect(staleWrite).rejects.toThrow(
      "Offline cache identity changed during write",
    );
    await expect(cleanup).resolves.toBeUndefined();
    expect(
      (await AsyncStorage.getAllKeys()).filter((key) =>
        key.startsWith("decorebator_offline_"),
      ),
    ).toEqual([]);
  });

  it("rejects an identity-A continuation that starts writing after cleanup", async () => {
    const identityAGeneration = offlineManager.captureCacheGeneration();
    await offlineManager.clearCache();

    await expect(
      offlineManager.cacheWordlists(
        [{ id: 1, name: "Late Identity A" }] as unknown as Wordlist[],
        identityAGeneration,
      ),
    ).rejects.toThrow("Offline cache identity changed during write");
    expect(
      (await AsyncStorage.getAllKeys()).filter((key) =>
        key.startsWith("decorebator_offline_"),
      ),
    ).toEqual([]);
  });

  it("rejects identity A from reading identity B's replacement cache", async () => {
    const identityAGeneration = offlineManager.captureCacheGeneration();
    await offlineManager.clearCache();
    await offlineManager.cacheWordlists([
      { id: 2, name: "Identity B" },
    ] as unknown as Wordlist[]);

    await expect(
      offlineManager.getCachedWordlists(identityAGeneration),
    ).rejects.toThrow("Offline cache identity changed during write");
  });
});
