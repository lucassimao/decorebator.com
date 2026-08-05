import React from "react";
import { fireEvent, render, waitFor } from "@testing-library/react-native";
import { AccessibilityInfo } from "react-native";

import { FlashcardContent } from "../FlashcardContent";

let mockReducedMotion = false;
const mockWithTiming = jest.fn((value: number) => value);
const mockWithSpring = jest.fn((value: number) => value);
const mockCancelAnimation = jest.fn();
const mockPlay = jest.fn();

jest.mock("expo-audio", () => ({
  setAudioModeAsync: jest.fn(() => Promise.resolve()),
  useAudioPlayer: () => ({
    pause: jest.fn(),
    play: mockPlay,
    replace: jest.fn(),
    seekTo: jest.fn(),
  }),
  useAudioPlayerStatus: () => ({ playing: false, didJustFinish: false }),
}));

jest.mock("react-native-reanimated", () => ({
  __esModule: true,
  default: { View: "Animated.View" },
  Easing: { cubic: "cubic", out: (value: unknown) => value },
  cancelAnimation: (...args: unknown[]) => mockCancelAnimation(...args),
  interpolate: (
    value: number,
    input: [number, number],
    output: [number, number],
  ) =>
    output[0] +
    ((value - input[0]) / (input[1] - input[0])) * (output[1] - output[0]),
  useAnimatedStyle: (fn: () => unknown) => fn(),
  useReducedMotion: () => mockReducedMotion,
  useSharedValue: (value: number) =>
    jest.requireActual("react").useRef({ value }).current,
  withSpring: (value: number) => mockWithSpring(value),
  withTiming: (value: number) => mockWithTiming(value),
}));

const word = {
  id: 7,
  name: "extraordinariamente",
  pronunciation: "extraordinaria",
  audioURL: "https://example.test/audio.mp3",
} as any;

const baseProps = {
  currentWord: word,
  definitions: [],
  isFlipped: false,
  focusRequestKey: 0,
  navigationDirection: "next" as const,
  shouldFetchDefinitions: false,
  loadingDefinitions: false,
  definitionsError: null,
  onFlip: jest.fn(),
  onRefetchDefinitions: jest.fn(),
};

describe("FlashcardContent", () => {
  beforeEach(() => {
    mockReducedMotion = false;
    mockWithTiming.mockClear();
    mockWithSpring.mockClear();
    mockCancelAnimation.mockClear();
    mockPlay.mockClear();
    baseProps.onFlip.mockClear();
    jest
      .mocked(AccessibilityInfo.isScreenReaderEnabled)
      .mockResolvedValue(false);
    jest.mocked(AccessibilityInfo.isScreenReaderEnabled).mockClear();
    jest.mocked(AccessibilityInfo.setAccessibilityFocus).mockClear();
  });

  it("exposes separate flip and audio actions and hides the inactive face", async () => {
    const view = render(<FlashcardContent {...baseProps} />);

    fireEvent.press(
      view.getByLabelText("extraordinariamente. Show definitions"),
    );
    fireEvent.press(
      view.getByLabelText("Play pronunciation for extraordinariamente"),
    );

    expect(baseProps.onFlip).toHaveBeenCalledTimes(1);
    await waitFor(() => expect(mockPlay).toHaveBeenCalledTimes(1));
    expect(view.getByTestId("flashcard-front").props).toEqual(
      expect.objectContaining({ accessibilityElementsHidden: false }),
    );
    expect(
      view.getByTestId("flashcard-back", { includeHiddenElements: true }).props,
    ).toEqual(
      expect.objectContaining({
        accessibilityElementsHidden: true,
        importantForAccessibility: "no-hide-descendants",
      }),
    );

    view.rerender(<FlashcardContent {...baseProps} isFlipped />);
    expect(
      view.getByTestId("flashcard-front", {
        includeHiddenElements: true,
      }).props,
    ).toEqual(
      expect.objectContaining({
        accessibilityElementsHidden: true,
        importantForAccessibility: "no-hide-descendants",
      }),
    );
    expect(view.getByTestId("flashcard-back").props).toEqual(
      expect.objectContaining({ accessibilityElementsHidden: false }),
    );
  });

  it("keeps a 220-point card floor and permits at least 155% text", () => {
    const view = render(<FlashcardContent {...baseProps} />);

    expect(view.getByTestId("flashcard-card-region").props.style).toEqual(
      expect.arrayContaining([expect.objectContaining({ minHeight: 220 })]),
    );
    expect(view.getByText("extraordinariamente").props).toEqual(
      expect.objectContaining({
        allowFontScaling: true,
        maxFontSizeMultiplier: 2,
      }),
    );
  });

  it("sets animation values immediately when reduced motion is enabled", () => {
    mockReducedMotion = true;
    const view = render(<FlashcardContent {...baseProps} />);
    mockWithTiming.mockClear();

    fireEvent(
      view.getByLabelText("extraordinariamente. Show definitions"),
      "pressIn",
    );
    expect(mockWithSpring).not.toHaveBeenCalled();
    view.rerender(<FlashcardContent {...baseProps} isFlipped />);
    expect(mockWithTiming).not.toHaveBeenCalled();
  });

  it("runs and cancels flip, settle, and press animations", () => {
    const view = render(<FlashcardContent {...baseProps} />);
    expect(mockWithTiming).toHaveBeenCalledWith(0);

    fireEvent(
      view.getByLabelText("Play pronunciation for extraordinariamente"),
      "pressIn",
    );
    expect(mockWithSpring).toHaveBeenCalledWith(0.97);

    mockWithTiming.mockClear();
    view.rerender(<FlashcardContent {...baseProps} isFlipped />);
    expect(mockWithTiming).toHaveBeenCalledWith(1);

    mockCancelAnimation.mockClear();
    view.unmount();
    expect(mockCancelAnimation).toHaveBeenCalledTimes(5);
  });

  it("moves screen-reader focus only for an explicit flip request", async () => {
    jest
      .mocked(AccessibilityInfo.isScreenReaderEnabled)
      .mockResolvedValue(true);
    const animationFrame = jest
      .spyOn(global, "requestAnimationFrame")
      .mockImplementation((callback) => {
        callback(0);
        return 1;
      });
    const view = render(<FlashcardContent {...baseProps} />);

    view.rerender(<FlashcardContent {...baseProps} isFlipped />);
    expect(AccessibilityInfo.isScreenReaderEnabled).not.toHaveBeenCalled();
    expect(AccessibilityInfo.setAccessibilityFocus).not.toHaveBeenCalled();

    view.rerender(
      <FlashcardContent {...baseProps} isFlipped focusRequestKey={1} />,
    );
    await waitFor(() =>
      expect(AccessibilityInfo.isScreenReaderEnabled).toHaveBeenCalledTimes(1),
    );
    animationFrame.mockRestore();
  });
});
