export type { PreparedProfileImage } from "./src/ProfileImageProcessorModule";

export async function prepareProfileImage(): Promise<never> {
  // Web deliberately fails before reading or decoding the selected file. The
  // native implementations can probe metadata and downsample during decode;
  // browsers do not provide an equivalent memory bound consistently.
  throw new Error(
    "Profile image uploads are available in the iOS and Android apps.",
  );
}
