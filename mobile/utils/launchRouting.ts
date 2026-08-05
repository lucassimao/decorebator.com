import AsyncStorage from "@react-native-async-storage/async-storage";

const ONBOARDING_SEEN_KEY = "@decorebator/onboarding:v1:seen";

export type InitialRoute = "/dashboard" | "/onboarding" | "/signin";
export type OnboardingExitRoute = "/signup" | "/signin";

interface OnboardingRouter {
  canDismiss: () => boolean;
  dismissAll: () => void;
  replace: (route: OnboardingExitRoute) => void;
}

export function getInitialRoute(input: {
  isAuthenticated: boolean;
  hasAuthError: boolean;
  onboardingSeen: boolean;
}): InitialRoute {
  if (input.isAuthenticated) return "/dashboard";
  if (input.hasAuthError) return "/signin";
  return input.onboardingSeen ? "/signin" : "/onboarding";
}

export async function readOnboardingSeen(): Promise<boolean> {
  return (await AsyncStorage.getItem(ONBOARDING_SEEN_KEY)) === "true";
}

export async function markOnboardingSeen(): Promise<void> {
  await AsyncStorage.setItem(ONBOARDING_SEEN_KEY, "true");
}

export function replaceOnboardingStack(
  router: OnboardingRouter,
  route: OnboardingExitRoute,
): void {
  if (router.canDismiss()) {
    router.dismissAll();
  }
  router.replace(route);
}
