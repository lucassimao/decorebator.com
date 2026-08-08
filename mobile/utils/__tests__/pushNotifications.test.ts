import { registerDevicePushToken } from "@/utils/pushNotifications";
import {
  beginPushTokenRegistration,
  getStoredPushToken,
} from "@/utils/pushTokenStorage";
import { getAuthenticationSessionEpoch } from "@/api/users";
import * as Device from "expo-device";

jest.mock("@/api/users", () => ({
  getAuthenticationSessionEpoch: jest.fn(() => "identity-a"),
}));

jest.mock("@/utils/pushTokenStorage", () => ({
  beginPushTokenRegistration: jest.fn(),
  completePushTokenRegistration: jest.fn(),
  deactivateStoredPushToken: jest.fn(),
  getStoredPushToken: jest.fn(() => Promise.resolve(null)),
}));

jest.mock("@/api/pushNotifications", () => ({
  registerPushToken: jest.fn(),
}));

describe("push notification registration ownership", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    Object.defineProperty(Device, "isDevice", {
      configurable: true,
      value: false,
    });
    jest.mocked(getAuthenticationSessionEpoch).mockReturnValue("identity-a");
  });

  it("does not cancel pending cleanup when a non-device exits early", async () => {
    await expect(
      registerDevicePushToken({ prompt: false }),
    ).resolves.toBeNull();

    expect(beginPushTokenRegistration).not.toHaveBeenCalled();
  });

  it("rejects identity A before registration when its token lookup resumes under B", async () => {
    Object.defineProperty(Device, "isDevice", {
      configurable: true,
      value: true,
    });
    let finishTokenLookup!: (token: string | null) => void;
    jest.mocked(getStoredPushToken).mockReturnValueOnce(
      new Promise((resolve) => {
        finishTokenLookup = resolve;
      }),
    );

    const registration = registerDevicePushToken({ prompt: false });
    jest.mocked(getAuthenticationSessionEpoch).mockReturnValue("identity-b");
    finishTokenLookup(null);

    await expect(registration).rejects.toThrow(
      "Session changed during push registration",
    );
    expect(beginPushTokenRegistration).not.toHaveBeenCalled();
  });
});
