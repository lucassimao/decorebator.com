import AsyncStorage from "@react-native-async-storage/async-storage";

import {
  getInitialRoute,
  markOnboardingSeen,
  readOnboardingSeen,
  replaceOnboardingStack,
} from "@/utils/launchRouting";

describe("launch routing", () => {
  beforeEach(async () => {
    jest.clearAllMocks();
    await AsyncStorage.clear();
  });

  it.each([
    [true, false, false, "/dashboard"],
    [false, true, false, "/signin"],
    [false, false, true, "/signin"],
    [false, false, false, "/onboarding"],
  ] as const)(
    "routes authenticated=%s authError=%s onboardingSeen=%s to %s",
    (isAuthenticated, hasAuthError, onboardingSeen, expected) => {
      expect(
        getInitialRoute({
          isAuthenticated,
          hasAuthError,
          onboardingSeen,
        }),
      ).toBe(expected);
    },
  );

  it("persists dismissal with a versioned private key", async () => {
    expect(await readOnboardingSeen()).toBe(false);

    await markOnboardingSeen();
    expect(await readOnboardingSeen()).toBe(true);
    expect(AsyncStorage.setItem).toHaveBeenCalledWith(
      "@decorebator/onboarding:v1:seen",
      "true",
    );
  });

  it("clears nested onboarding history before replacing the terminal route", () => {
    const calls: string[] = [];
    const router = {
      canDismiss: jest.fn(() => true),
      dismissAll: jest.fn(() => calls.push("dismiss")),
      replace: jest.fn(() => calls.push("replace")),
    };

    replaceOnboardingStack(router, "/signup");

    expect(calls).toEqual(["dismiss", "replace"]);
    expect(router.replace).toHaveBeenCalledWith("/signup");
  });

  it("replaces directly when onboarding is the only stack entry", () => {
    const router = {
      canDismiss: jest.fn(() => false),
      dismissAll: jest.fn(),
      replace: jest.fn(),
    };

    replaceOnboardingStack(router, "/signin");

    expect(router.dismissAll).not.toHaveBeenCalled();
    expect(router.replace).toHaveBeenCalledWith("/signin");
  });
});
