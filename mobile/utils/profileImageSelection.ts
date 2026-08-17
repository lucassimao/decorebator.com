import type { ImagePickerAsset, ImagePickerResult } from "expo-image-picker";

type ProfileImageSelectionDependencies = {
  sessionEpoch: string;
  getSessionEpoch: () => string | null;
  requestPermission: () => Promise<{ granted: boolean }>;
  launchPicker: () => Promise<ImagePickerResult>;
  onPermissionDenied: () => void;
  onSelectionError: () => void;
  onDiscarded: (asset: ImagePickerAsset) => void;
  onSelected: (asset: ImagePickerAsset, sessionEpoch: string) => void;
};

export async function runProfileImageSelection(
  dependencies: ProfileImageSelectionDependencies,
): Promise<void> {
  const sessionIsCurrent = () =>
    dependencies.sessionEpoch.length > 0 &&
    dependencies.getSessionEpoch() === dependencies.sessionEpoch;
  if (!sessionIsCurrent()) return;

  try {
    const permission = await dependencies.requestPermission();
    if (!sessionIsCurrent()) return;
    if (!permission.granted) {
      dependencies.onPermissionDenied();
      return;
    }

    const result = await dependencies.launchPicker();
    if (result.canceled || !result.assets[0]) return;
    const asset = result.assets[0];
    if (!sessionIsCurrent()) {
      dependencies.onDiscarded(asset);
      return;
    }
    dependencies.onSelected(asset, dependencies.sessionEpoch);
  } catch {
    if (sessionIsCurrent()) dependencies.onSelectionError();
  }
}
