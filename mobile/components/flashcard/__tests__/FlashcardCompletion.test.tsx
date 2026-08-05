import React from "react";
import { act, fireEvent, render } from "@testing-library/react-native";
import { AccessibilityInfo } from "react-native";

import { FlashcardCompletion } from "../FlashcardCompletion";

let mockReducedMotion = false;
const mockWithDelay = jest.fn((_delay: number, value: number) => value);

jest.mock("react-native-reanimated", () => ({
  __esModule: true,
  default: { View: "Animated.View" },
  useAnimatedStyle: (fn: () => unknown) => fn(),
  useReducedMotion: () => mockReducedMotion,
  useSharedValue: (value: number) => ({ value }),
  withDelay: (delay: number, value: number) => mockWithDelay(delay, value),
  withSpring: (value: number) => value,
  withTiming: (value: number) => value,
}));

describe("FlashcardCompletion", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    mockReducedMotion = false;
    mockWithDelay.mockClear();
    (AccessibilityInfo.setAccessibilityFocus as jest.Mock).mockClear();
    (AccessibilityInfo.announceForAccessibility as jest.Mock).mockClear();
  });

  afterEach(() => jest.useRealTimers());

  it("renders honest visit data, focuses the heading, and exposes all exits", () => {
    const onReviewAgain = jest.fn();
    const onBackToWordlists = jest.fn();
    const onReturnToLastCard = jest.fn();
    const view = render(
      <FlashcardCompletion
        wordlistName="Travel Italian"
        viewedCount={4}
        shouldAnimate
        onReviewAgain={onReviewAgain}
        onBackToWordlists={onBackToWordlists}
        onReturnToLastCard={onReturnToLastCard}
      />,
    );

    expect(
      view.getByRole("header").props.accessibilityLiveRegion,
    ).toBeUndefined();
    expect(view.getByText("4")).toBeTruthy();
    expect(view.getByText("cards viewed this visit")).toBeTruthy();

    fireEvent.press(view.getByText("Review again"));
    fireEvent.press(view.getByText("Back to wordlists"));
    fireEvent.press(view.getByText("Return to last card"));
    expect(onReviewAgain).toHaveBeenCalledTimes(1);
    expect(onBackToWordlists).toHaveBeenCalledTimes(1);
    expect(onReturnToLastCard).toHaveBeenCalledTimes(1);

    act(() => jest.runOnlyPendingTimers());
    expect(AccessibilityInfo.setAccessibilityFocus).toHaveBeenCalledWith(1);
    expect(AccessibilityInfo.announceForAccessibility).not.toHaveBeenCalled();
    expect(mockWithDelay).toHaveBeenCalledTimes(2);
  });

  it("renders the static end state when reduced motion is enabled", () => {
    mockReducedMotion = true;
    const view = render(
      <FlashcardCompletion
        wordlistName="Travel Italian"
        viewedCount={1}
        shouldAnimate
        onReviewAgain={jest.fn()}
        onBackToWordlists={jest.fn()}
        onReturnToLastCard={jest.fn()}
      />,
    );

    expect(mockWithDelay).not.toHaveBeenCalled();
    expect(view.getByText("card viewed this visit")).toBeTruthy();
  });
});
