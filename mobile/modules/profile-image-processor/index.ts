import ProfileImageProcessor from "./src/ProfileImageProcessorModule";

export const prepareProfileImage = ProfileImageProcessor.prepare.bind(
  ProfileImageProcessor,
);
export type { PreparedProfileImage } from "./src/ProfileImageProcessorModule";
