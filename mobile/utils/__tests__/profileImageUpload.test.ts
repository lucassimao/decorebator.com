import {
  MAX_PROFILE_IMAGE_BYTES,
  uploadProfilePicture,
} from "../profileImageUpload";

jest.mock("@/modules/profile-image-processor", () => ({
  prepareProfileImage: jest.fn(),
}));
jest.mock("expo-file-system/legacy", () => ({
  EncodingType: { Base64: "base64" },
  readAsStringAsync: jest.fn(),
  deleteAsync: jest.fn(),
}));
jest.mock("@/api/users", () => ({
  update: jest.fn(),
  getAuthenticationSessionEpoch: jest.fn(),
}));

const prepared = {
  uri: "file:///bounded.jpg",
  width: 2048,
  height: 1536,
  sourceWidth: 8000,
  sourceHeight: 6000,
  size: 18,
};

function dependencies(overrides: Record<string, unknown> = {}) {
  return {
    prepareImage: jest.fn().mockResolvedValue(prepared),
    readBase64: jest.fn().mockResolvedValue("Y2Fub25pY2FsLWpwZWc="),
    deleteTemporaryFile: jest.fn().mockResolvedValue(undefined),
    reportCleanupFailure: jest.fn(),
    updateProfile: jest.fn().mockResolvedValue({
      profilePictureUrl: "https://images.example/users/1-server.jpg",
    }),
    getSessionEpoch: jest.fn(() => "identity-a"),
    ...overrides,
  } as any;
}

describe("uploadProfilePicture", () => {
  it.each([
    ["known", { uri: "file:///camera.jpg", width: 8000, height: 6000 }],
    ["unknown", { uri: "file:///provider.jpg", width: 0, height: 0 }],
    [
      "unknown width",
      { uri: "file:///provider-width.jpg", width: 0, height: 3000 },
    ],
    [
      "unknown height",
      { uri: "file:///provider-height.jpg", width: 4000, height: 0 },
    ],
  ])(
    "uses one metadata-first sampled native pass for %s picker dimensions",
    async (_name, source) => {
      const deps = dependencies();

      await expect(
        uploadProfilePicture({ source, sessionEpoch: "identity-a" }, deps),
      ).resolves.toBe("https://images.example/users/1-server.jpg");

      expect(deps.prepareImage).toHaveBeenCalledTimes(1);
      expect(deps.prepareImage).toHaveBeenCalledWith(source.uri);
      expect(deps.readBase64).toHaveBeenCalledWith(prepared.uri);
      expect(deps.updateProfile).toHaveBeenCalledWith({
        updateProfilePicture: { base64Data: "Y2Fub25pY2FsLWpwZWc=" },
      });
      expect(deps.deleteTemporaryFile).toHaveBeenCalledWith(prepared.uri);
    },
  );

  it.each([
    { width: -1, height: 0 },
    { width: 0, height: -1 },
    { width: Number.NaN, height: 0 },
    { width: 0, height: 1.5 },
  ])(
    "rejects invalid partial picker metadata %#",
    async ({ width, height }) => {
      const deps = dependencies();
      await expect(
        uploadProfilePicture(
          {
            source: { uri: "file:///invalid.jpg", width, height },
            sessionEpoch: "identity-a",
          },
          deps,
        ),
      ).rejects.toThrow("dimensions are unsupported");
      expect(deps.prepareImage).not.toHaveBeenCalled();
    },
  );

  it("rejects unsafe picker dimensions before native raster work", async () => {
    const deps = dependencies();
    await expect(
      uploadProfilePicture(
        {
          source: { uri: "file:///hostile.jpg", width: 20_000, height: 20_000 },
          sessionEpoch: "identity-a",
        },
        deps,
      ),
    ).rejects.toThrow("dimensions are unsupported");
    expect(deps.prepareImage).not.toHaveBeenCalled();
  });

  it("rejects native metadata that exceeds the overflow-safe source budget", async () => {
    const deps = dependencies({
      prepareImage: jest.fn().mockResolvedValue({
        ...prepared,
        sourceWidth: 20_000,
        sourceHeight: 20_000,
      }),
    });
    await expect(
      uploadProfilePicture(
        {
          source: { uri: "file:///unknown.jpg", width: 0, height: 0 },
          sessionEpoch: "identity-a",
        },
        deps,
      ),
    ).rejects.toThrow("dimensions are unsupported");
    expect(deps.readBase64).not.toHaveBeenCalled();
    expect(deps.deleteTemporaryFile).toHaveBeenCalled();
  });

  it.each([0, -1, 1.5, Number.NaN, MAX_PROFILE_IMAGE_BYTES + 1])(
    "rejects unsafe native output size %s before Base64 allocation",
    async (size) => {
      const deps = dependencies({
        prepareImage: jest.fn().mockResolvedValue({ ...prepared, size }),
      });
      await expect(
        uploadProfilePicture(
          {
            source: { uri: "file:///large.jpg", width: 4000, height: 3000 },
            sessionEpoch: "identity-a",
          },
          deps,
        ),
      ).rejects.toThrow("Invalid prepared profile image");
      expect(deps.readBase64).not.toHaveBeenCalled();
      expect(deps.deleteTemporaryFile).toHaveBeenCalledWith(prepared.uri);
    },
  );

  it("rejects Base64 output above the server byte limit", async () => {
    const oversizedBase64 = "A".repeat(
      Math.ceil(((MAX_PROFILE_IMAGE_BYTES + 1) * 4) / 3),
    );
    const deps = dependencies({
      readBase64: jest.fn().mockResolvedValue(oversizedBase64),
    });
    await expect(
      uploadProfilePicture(
        {
          source: { uri: "file:///large.jpg", width: 4000, height: 3000 },
          sessionEpoch: "identity-a",
        },
        deps,
      ),
    ).rejects.toThrow("Profile image is too large");
    expect(deps.updateProfile).not.toHaveBeenCalled();
  });

  it("never dispatches user A's photo after the native pass crosses an auth epoch", async () => {
    let resolvePreparation!: (value: typeof prepared) => void;
    let currentEpoch = "identity-a";
    const deps = dependencies({
      prepareImage: jest.fn().mockReturnValue(
        new Promise<typeof prepared>((resolve) => {
          resolvePreparation = resolve;
        }),
      ),
      getSessionEpoch: () => currentEpoch,
    });
    const pending = uploadProfilePicture(
      {
        source: { uri: "file:///user-a.jpg", width: 4000, height: 3000 },
        sessionEpoch: "identity-a",
      },
      deps,
    );

    currentEpoch = "identity-b";
    resolvePreparation(prepared);

    await expect(pending).rejects.toThrow("Authentication session changed");
    expect(deps.readBase64).not.toHaveBeenCalled();
    expect(deps.updateProfile).not.toHaveBeenCalled();
    expect(deps.deleteTemporaryFile).toHaveBeenCalled();
  });

  it("removes the picker source when the upload starts in a stale session", async () => {
    const deps = dependencies({ getSessionEpoch: () => "identity-b" });

    await expect(
      uploadProfilePicture(
        {
          source: {
            uri: "file:///stale-picker.jpg",
            width: 4000,
            height: 3000,
          },
          sessionEpoch: "identity-a",
        },
        deps,
      ),
    ).rejects.toThrow("Authentication session changed");
    expect(deps.prepareImage).not.toHaveBeenCalled();
    expect(deps.deleteTemporaryFile).toHaveBeenCalledWith(
      "file:///stale-picker.jpg",
    );
  });

  it("reports temporary cleanup failures without replacing a successful upload", async () => {
    const cleanupError = new Error("filesystem unavailable");
    const deps = dependencies({
      deleteTemporaryFile: jest.fn().mockRejectedValue(cleanupError),
    });

    await expect(
      uploadProfilePicture(
        {
          source: { uri: "file:///camera.jpg", width: 8000, height: 6000 },
          sessionEpoch: "identity-a",
        },
        deps,
      ),
    ).resolves.toBe("https://images.example/users/1-server.jpg");
    expect(deps.reportCleanupFailure).toHaveBeenCalledWith(cleanupError);
  });
});
