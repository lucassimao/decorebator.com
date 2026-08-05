import React from "react";
import { fireEvent, render } from "@testing-library/react-native";

import { FlashcardNavigation } from "../FlashcardNavigation";
import { FlashcardProgressBar } from "../FlashcardProgressBar";

jest.mock("react-native-reanimated", () => ({
  __esModule: true,
  default: { View: "Animated.View" },
  useAnimatedStyle: (fn: () => unknown) => fn(),
  useReducedMotion: () => true,
  useSharedValue: (value: number) => ({ value }),
  withSpring: (value: number) => value,
}));

describe("FlashcardNavigation", () => {
  it("keeps one wide action while changing Next card to Finish wordlist", () => {
    const onNavigate = jest.fn();
    const onFinish = jest.fn();
    const view = render(
      <FlashcardNavigation
        currentIndex={0}
        totalWords={2}
        onNavigate={onNavigate}
        onFinish={onFinish}
      />,
    );

    expect(view.getByLabelText("Previous card").props.disabled).toBe(true);
    fireEvent.press(view.getByLabelText("Next card"));
    expect(onNavigate).toHaveBeenCalledWith("next");

    view.rerender(
      <FlashcardNavigation
        currentIndex={1}
        totalWords={2}
        onNavigate={onNavigate}
        onFinish={onFinish}
      />,
    );
    expect(view.getByLabelText("Previous card").props.disabled).toBe(false);
    fireEvent.press(view.getByLabelText("Finish wordlist"));
    expect(onFinish).toHaveBeenCalledTimes(1);
  });

  it("exposes progress semantics including a one-card wordlist", () => {
    const view = render(
      <FlashcardProgressBar currentIndex={0} totalWords={1} />,
    );
    const progress = view.getByLabelText("Flashcard progress");

    expect(progress.props.accessibilityRole).toBe("progressbar");
    expect(progress.props.accessibilityValue).toEqual({
      min: 0,
      max: 1,
      now: 1,
    });
  });
});
