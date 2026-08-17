import { runProfileImageSelection } from "@/utils/profileImageSelection";

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((next) => {
    resolve = next;
  });
  return { promise, resolve };
}

describe("ProfileSettings profile-image session boundary", () => {
  it("does not launch user A's library picker after permission resolves in user B's session", async () => {
    const permission = deferred<{ granted: boolean }>();
    let epoch = "identity-a";
    const launchPicker = jest.fn();
    const onSelected = jest.fn();
    const pending = runProfileImageSelection({
      sessionEpoch: "identity-a",
      getSessionEpoch: () => epoch,
      requestPermission: () => permission.promise,
      launchPicker,
      onPermissionDenied: jest.fn(),
      onSelectionError: jest.fn(),
      onDiscarded: jest.fn(),
      onSelected,
    });

    epoch = "identity-b";
    permission.resolve({ granted: true });
    await pending;

    expect(launchPicker).not.toHaveBeenCalled();
    expect(onSelected).not.toHaveBeenCalled();
  });

  it("does not consume user A's camera result after the picker resolves in user B's session", async () => {
    const picker = deferred<any>();
    let epoch = "identity-a";
    const onSelected = jest.fn();
    const onDiscarded = jest.fn();
    const pending = runProfileImageSelection({
      sessionEpoch: "identity-a",
      getSessionEpoch: () => epoch,
      requestPermission: async () => ({ granted: true }),
      launchPicker: () => picker.promise,
      onPermissionDenied: jest.fn(),
      onSelectionError: jest.fn(),
      onDiscarded,
      onSelected,
    });
    await Promise.resolve();

    epoch = "identity-b";
    picker.resolve({
      canceled: false,
      assets: [{ uri: "file:///user-a.jpg", width: 4000, height: 3000 }],
    });
    await pending;

    expect(onSelected).not.toHaveBeenCalled();
    expect(onDiscarded).toHaveBeenCalledWith(
      expect.objectContaining({ uri: "file:///user-a.jpg" }),
    );
  });

  it("reports a picker failure only while the originating session is current", async () => {
    const onSelectionError = jest.fn();
    await runProfileImageSelection({
      sessionEpoch: "identity-a",
      getSessionEpoch: () => "identity-a",
      requestPermission: async () => ({ granted: true }),
      launchPicker: async () => {
        throw new Error("oversized picker source");
      },
      onPermissionDenied: jest.fn(),
      onSelectionError,
      onDiscarded: jest.fn(),
      onSelected: jest.fn(),
    });

    expect(onSelectionError).toHaveBeenCalledTimes(1);
  });
});
