import React from "react";
import { fireEvent, render } from "@testing-library/react-native";

import { FlashcardHeader } from "../FlashcardHeader";
import { FlashcardLoadingState } from "../FlashcardLoadingState";
import { FlashcardStatusState } from "../FlashcardStatusState";
import { LoadingWithTimeout } from "@/components/LoadingWithTimeout";

jest.mock("react-native-reanimated", () => ({
  __esModule: true,
  default: { View: "Animated.View" },
  useAnimatedStyle: (fn: () => unknown) => fn(),
  useReducedMotion: () => true,
  useSharedValue: (value: number) => ({ value }),
  withSpring: (value: number) => value,
}));

describe("flashcard accessibility and recovery states", () => {
  it("keeps shared loading recovery generic outside flashcards", () => {
    const view = render(
      <LoadingWithTimeout
        isLoading
        hasTimeout
        error={new Error("timeout")}
        loadingMessage="Loading"
        timeoutMessage="Taking longer"
        onRetry={jest.fn()}
        onGoBack={jest.fn()}
      />,
    );

    expect(view.getByText("Go Back")).toBeTruthy();
    expect(view.queryByText("Back to wordlists")).toBeNull();
  });

  it("keeps processing distinct and exposes retry/back actions", () => {
    const onRetry = jest.fn();
    const onGoBack = jest.fn();
    const view = render(
      <FlashcardLoadingState
        isLoading={false}
        hasTimeout={false}
        error={new Error("no definitions are ready")}
        onRetry={onRetry}
        onGoBack={onGoBack}
      />,
    );

    expect(view.getByText("Words Still Processing")).toBeTruthy();
    fireEvent.press(view.getByText("Try Again"));
    fireEvent.press(view.getByText("Back to wordlists"));
    expect(onRetry).toHaveBeenCalledTimes(1);
    expect(onGoBack).toHaveBeenCalledTimes(1);
  });

  it("renders recoverable status copy in a scroll-safe surface", () => {
    const onRetry = jest.fn();
    const onBack = jest.fn();
    const view = render(
      <FlashcardStatusState
        icon="error-outline"
        title="Couldn't load flashcards"
        message="Check your connection."
        onRetry={onRetry}
        onBack={onBack}
        assertive
      />,
    );

    expect(view.getByRole("header").props.children).toBe(
      "Couldn't load flashcards",
    );
    expect(view.getByTestId("flashcard-status-content").props).toEqual(
      expect.objectContaining({ accessibilityLiveRegion: "assertive" }),
    );
    fireEvent.press(view.getByText("Try Again"));
    fireEvent.press(view.getByText("Back to wordlists"));
    expect(onRetry).toHaveBeenCalledTimes(1);
    expect(onBack).toHaveBeenCalledTimes(1);
  });

  it("caps long headers and localizes control semantics", () => {
    const view = render(
      <FlashcardHeader
        wordlistName="Portuguese words for an unforgettable afternoon"
        currentIndex={7}
        totalWords={12}
        isOnline={false}
        onClose={jest.fn()}
        onReportError={jest.fn()}
        savePosition
        onToggleSavePosition={jest.fn()}
      />,
    );

    expect(
      view.getByRole("header", {
        name: "Portuguese words for an unforgettable afternoon",
      }).props.numberOfLines,
    ).toBe(2);
    expect(
      view.getByLabelText("Report issue (offline)").props.accessibilityState,
    ).toEqual({ disabled: true });
    expect(
      view.getByLabelText("Save position").props.accessibilityState,
    ).toEqual({
      checked: true,
    });
  });
});
