import * as userApi from "@/api/users";
import * as FileSystem from "expo-file-system/legacy";
import {
  prepareProfileImage,
  type PreparedProfileImage,
} from "@/modules/profile-image-processor";

export const MAX_PROFILE_IMAGE_BYTES = 5 * 1024 * 1024;
export const MAX_PROFILE_IMAGE_DIMENSION = 2048;
export const MAX_PROFILE_IMAGE_SOURCE_DIMENSION = 20_000;
export const MAX_PROFILE_IMAGE_SOURCE_PIXELS = 100_000_000;

export type ProfileImageSource = {
  uri: string;
  width: number;
  height: number;
};

export type ProfileImageUploadRequest = {
  source: ProfileImageSource;
  sessionEpoch: string;
};

type ProfileImageUploadDependencies = {
  prepareImage: (uri: string) => Promise<PreparedProfileImage>;
  readBase64: (uri: string) => Promise<string>;
  deleteTemporaryFile: (uri: string) => Promise<void>;
  reportCleanupFailure: (error: unknown) => void;
  updateProfile: typeof userApi.update;
  getSessionEpoch: typeof userApi.getAuthenticationSessionEpoch;
};

const defaultDependencies: ProfileImageUploadDependencies = {
  prepareImage: prepareProfileImage,
  readBase64: (uri) =>
    FileSystem.readAsStringAsync(uri, {
      encoding: FileSystem.EncodingType.Base64,
    }),
  deleteTemporaryFile: (uri) =>
    FileSystem.deleteAsync(uri, { idempotent: true }),
  reportCleanupFailure: (error) =>
    console.warn("Failed to remove profile image temporary file", error),
  updateProfile: userApi.update,
  getSessionEpoch: userApi.getAuthenticationSessionEpoch,
};

export async function uploadProfilePicture(
  request: ProfileImageUploadRequest,
  dependencies: ProfileImageUploadDependencies = defaultDependencies,
): Promise<string> {
  const { source, sessionEpoch } = request;
  let prepared: PreparedProfileImage | undefined;
  try {
    assertProfileUploadIdentity(dependencies, sessionEpoch);
    validateReportedSourceDimensions(source.width, source.height);
    // The native module probes metadata before allocating pixels, rejects
    // pathological sources, and performs decode-time downsampling exactly once.
    prepared = await dependencies.prepareImage(source.uri);
    assertProfileUploadIdentity(dependencies, sessionEpoch);
    validatePreparedImage(prepared);
    const base64Data = await dependencies.readBase64(prepared.uri);
    assertProfileUploadIdentity(dependencies, sessionEpoch);
    if (decodedBase64Length(base64Data) > MAX_PROFILE_IMAGE_BYTES) {
      throw new Error("Profile image is too large");
    }

    const result = await dependencies.updateProfile({
      updateProfilePicture: { base64Data },
    });
    return result.profilePictureUrl;
  } finally {
    const temporaryUris = new Set([source.uri, prepared?.uri].filter(Boolean));
    for (const uri of temporaryUris) {
      await dependencies
        .deleteTemporaryFile(uri as string)
        .catch(dependencies.reportCleanupFailure);
    }
  }
}

export function discardProfileImageSource(uri: string): void {
  void defaultDependencies
    .deleteTemporaryFile(uri)
    .catch(defaultDependencies.reportCleanupFailure);
}

function validateReportedSourceDimensions(width: number, height: number) {
  if (width === 0 || height === 0) {
    if (
      Number.isSafeInteger(width) &&
      Number.isSafeInteger(height) &&
      width >= 0 &&
      height >= 0
    ) {
      return;
    }
    throw new Error("Profile image dimensions are unsupported");
  }
  validateSourceDimensions(width, height);
}

function validatePreparedImage(image: PreparedProfileImage) {
  validateSourceDimensions(image.sourceWidth, image.sourceHeight);
  if (
    !Number.isSafeInteger(image.width) ||
    !Number.isSafeInteger(image.height) ||
    image.width <= 0 ||
    image.height <= 0 ||
    image.width > MAX_PROFILE_IMAGE_DIMENSION ||
    image.height > MAX_PROFILE_IMAGE_DIMENSION ||
    !Number.isSafeInteger(image.size) ||
    image.size <= 0 ||
    image.size > MAX_PROFILE_IMAGE_BYTES
  ) {
    throw new Error("Invalid prepared profile image");
  }
}

function validateSourceDimensions(width: number, height: number) {
  if (
    !Number.isSafeInteger(width) ||
    !Number.isSafeInteger(height) ||
    width <= 0 ||
    height <= 0 ||
    width > MAX_PROFILE_IMAGE_SOURCE_DIMENSION ||
    height > MAX_PROFILE_IMAGE_SOURCE_DIMENSION ||
    width > Math.floor(MAX_PROFILE_IMAGE_SOURCE_PIXELS / height)
  ) {
    throw new Error("Profile image dimensions are unsupported");
  }
}

function assertProfileUploadIdentity(
  dependencies: ProfileImageUploadDependencies,
  expectedEpoch: string,
) {
  if (!expectedEpoch || dependencies.getSessionEpoch() !== expectedEpoch) {
    throw new Error(
      "Authentication session changed during profile image upload",
    );
  }
}

function decodedBase64Length(value: string): number {
  const padding = value.endsWith("==") ? 2 : value.endsWith("=") ? 1 : 0;
  return Math.floor((value.length * 3) / 4) - padding;
}
