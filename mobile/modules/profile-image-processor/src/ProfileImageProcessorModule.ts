import { requireNativeModule } from "expo";

export type PreparedProfileImage = {
  uri: string;
  width: number;
  height: number;
  sourceWidth: number;
  sourceHeight: number;
  size: number;
};

type ProfileImageProcessorModule = {
  prepare(uri: string): Promise<PreparedProfileImage>;
};

export default requireNativeModule<ProfileImageProcessorModule>(
  "ProfileImageProcessor",
);
