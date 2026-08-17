import { fireEvent, render } from "@testing-library/react-native";
import * as ImagePicker from "expo-image-picker";
import { Platform } from "react-native";
import ProfileSettingsScreen from "../profileSettings";

jest.mock("expo-image-picker", () => ({
  requestMediaLibraryPermissionsAsync: jest.fn(),
  requestCameraPermissionsAsync: jest.fn(),
  launchImageLibraryAsync: jest.fn(),
  launchCameraAsync: jest.fn(),
}));
jest.mock("@/modules/profile-image-processor", () => ({
  prepareProfileImage: jest.fn(),
}));
jest.mock("expo-file-system/legacy", () => ({
  EncodingType: { Base64: "base64" },
  readAsStringAsync: jest.fn(),
  deleteAsync: jest.fn(),
}));
jest.mock("@/components/profileSettings/ChangePasswordModal", () => ({
  ChangePasswordModal: () => null,
}));
jest.mock("@/hooks/useUserSession", () => {
  const user = {
    email: "test@example.test",
    firstName: "Test",
    lastName: "User",
    country: "US",
    dateOfBirth: "2000-01-01",
    preferredLanguage: "en",
  };
  return { useUserSession: () => ({ user, isLoading: false }) };
});
jest.mock("@/contexts/ThemeContext", () => {
  const actual = jest.requireActual("@/contexts/ThemeContext");
  return { ...actual, useTheme: () => ({ theme: actual.lightTheme }) };
});
jest.mock("@react-navigation/native", () => ({
  useNavigation: () => ({ goBack: jest.fn() }),
}));
jest.mock("@tanstack/react-query", () => ({
  useQueryClient: () => ({ clear: jest.fn(), invalidateQueries: jest.fn() }),
  useMutation: () => ({ mutate: jest.fn(), isPending: false }),
}));

describe("ProfileSettings web profile image boundary", () => {
  it("renders the route with selection disabled before ImagePicker", () => {
    const originalPlatform = Platform.OS;
    Object.defineProperty(Platform, "OS", { configurable: true, value: "web" });
    try {
      const screen = render(<ProfileSettingsScreen />);
      const button = screen.getByTestId("profile-picture-button");
      expect(
        button.props.accessibilityState?.disabled ?? button.props.disabled,
      ).toBe(true);
      fireEvent.press(button);
      expect(ImagePicker.launchImageLibraryAsync).not.toHaveBeenCalled();
      expect(ImagePicker.launchCameraAsync).not.toHaveBeenCalled();
    } finally {
      Object.defineProperty(Platform, "OS", {
        configurable: true,
        value: originalPlatform,
      });
    }
  });
});
