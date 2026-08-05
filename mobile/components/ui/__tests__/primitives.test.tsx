import React from "react";
import { act, fireEvent, render, waitFor } from "@testing-library/react-native";
import { AccessibilityInfo, findNodeHandle } from "react-native";
import type { View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { Button, Card, Input, Sheet, UiText } from "@/components/ui";

let mockReducedMotion = false;
let mockTimingFinished = true;

jest.mock("react-i18next", () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

jest.mock("react-native-reanimated", () => ({
  __esModule: true,
  default: { View: "Animated.View" },
  cancelAnimation: jest.fn(),
  runOnJS: (fn: (...args: unknown[]) => unknown) => fn,
  useAnimatedStyle: (fn: () => unknown) => fn(),
  useReducedMotion: () => mockReducedMotion,
  useSharedValue: (value: number) => ({ value }),
  withSpring: (value: number) => value,
  withTiming: (
    value: number,
    _config?: unknown,
    callback?: (finished: boolean) => void,
  ) => {
    callback?.(mockTimingFinished);
    return value;
  },
}));

describe("UI primitives", () => {
  beforeEach(() => {
    mockReducedMotion = false;
    mockTimingFinished = true;
    (useSafeAreaInsets as jest.Mock).mockReturnValue({
      top: 0,
      right: 0,
      bottom: 0,
      left: 0,
    });
    (AccessibilityInfo.setAccessibilityFocus as jest.Mock).mockClear();
    (AccessibilityInfo.announceForAccessibility as jest.Mock).mockClear();
  });

  it("exposes button press, busy, and disabled semantics", () => {
    const onPress = jest.fn();
    const { getByRole, getByText, rerender } = render(
      <Button onPress={onPress}>Practice now</Button>,
    );

    fireEvent.press(getByRole("button"));
    expect(onPress).toHaveBeenCalledTimes(1);

    rerender(
      <Button loading loadingLabel="Preparing quiz" onPress={onPress}>
        Practice now
      </Button>,
    );
    expect(getByText("Preparing quiz")).toBeTruthy();
    expect(getByRole("button").props.accessibilityState).toEqual(
      expect.objectContaining({ busy: true, disabled: true }),
    );

    rerender(<Button disabled>Practice now</Button>);
    expect(getByRole("button").props.accessibilityState.disabled).toBe(true);
  });

  it("keeps an announced text busy state when reduced motion hides the spinner", () => {
    mockReducedMotion = true;
    const { getByRole, getByText, queryByTestId } = render(
      <Button loading loadingLabel="Preparing quiz">
        Practice now
      </Button>,
    );

    expect(queryByTestId("button-loading-indicator")).toBeNull();
    expect(getByText("Preparing quiz")).toBeTruthy();
    expect(getByRole("button").props.accessibilityState.busy).toBe(true);

    const fallback = render(<Button loading>Practice now</Button>);
    expect(
      fallback.getByText("…", { includeHiddenElements: true }),
    ).toBeTruthy();
  });

  it("associates input errors with the field and announces the message", () => {
    const { getByLabelText, getByText } = render(
      <Input
        label="Daily goal"
        value="0 minutes"
        error="Choose at least 5 minutes."
      />,
    );

    const input = getByLabelText("Daily goal");
    expect(input.props["aria-invalid"]).toBe(true);
    expect(input.props["aria-describedby"]).toBeTruthy();
    expect(
      getByText("Choose at least 5 minutes.").props.accessibilityLiveRegion,
    ).toBe("polite");
    expect(AccessibilityInfo.announceForAccessibility).toHaveBeenCalledWith(
      "Choose at least 5 minutes.",
    );

    const disabled = render(<Input label="Daily goal" editable={false} />);
    expect(
      disabled.getByLabelText("Daily goal").props.accessibilityState,
    ).toEqual({ disabled: true });
  });

  it("renders card and typography variants from the shared theme", () => {
    const { getByText, getByTestId } = render(
      <Card testID="card" variant="elevated">
        <UiText variant="title">Spanish essentials</UiText>
      </Card>,
    );

    expect(getByTestId("card")).toBeTruthy();
    expect(getByText("Spanish essentials")).toBeTruthy();
  });

  it("mounts sheets only while needed and delegates dismiss requests", () => {
    const onClose = jest.fn();
    const { getByTestId, queryByTestId, rerender } = render(
      <Sheet visible={false} onClose={onClose} testID="practice-sheet">
        <UiText>Choose a round</UiText>
      </Sheet>,
    );

    expect(queryByTestId("practice-sheet")).toBeNull();

    rerender(
      <Sheet visible onClose={onClose} testID="practice-sheet">
        <UiText>Choose a round</UiText>
      </Sheet>,
    );
    expect(getByTestId("practice-sheet")).toBeTruthy();
    fireEvent(getByTestId("practice-sheet"), "requestClose");
    expect(onClose).toHaveBeenCalledTimes(1);
    fireEvent.press(getByTestId("sheet-backdrop"));
    expect(onClose).toHaveBeenCalledTimes(2);

    rerender(
      <Sheet visible={false} onClose={onClose} testID="practice-sheet">
        <UiText>Choose a round</UiText>
      </Sheet>,
    );
    expect(queryByTestId("practice-sheet")).toBeNull();
  });

  it("keeps an interrupted sheet mounted and handles reduced-motion close", () => {
    const { getByTestId, queryByTestId, rerender } = render(
      <Sheet visible onClose={jest.fn()} testID="practice-sheet">
        <UiText>Choose a round</UiText>
      </Sheet>,
    );

    mockTimingFinished = false;
    rerender(
      <Sheet visible={false} onClose={jest.fn()} testID="practice-sheet">
        <UiText>Choose a round</UiText>
      </Sheet>,
    );
    expect(getByTestId("practice-sheet")).toBeTruthy();

    rerender(
      <Sheet visible onClose={jest.fn()} testID="practice-sheet">
        <UiText>Choose a round</UiText>
      </Sheet>,
    );
    expect(getByTestId("practice-sheet")).toBeTruthy();

    mockReducedMotion = true;
    rerender(
      <Sheet visible={false} onClose={jest.fn()} testID="practice-sheet">
        <UiText>Choose a round</UiText>
      </Sheet>,
    );
    expect(queryByTestId("practice-sheet")).toBeNull();
  });

  it("uses safe-area padding and restores focus after the sheet unmounts", async () => {
    (useSafeAreaInsets as jest.Mock).mockReturnValue({
      top: 0,
      right: 0,
      bottom: 37,
      left: 0,
    });
    const triggerRef = { current: {} as View };
    const animationFrame = jest
      .spyOn(global, "requestAnimationFrame")
      .mockImplementation((callback) => {
        callback(0);
        return 1;
      });
    const { getByTestId, rerender } = render(
      <Sheet
        visible
        onClose={jest.fn()}
        testID="practice-sheet"
        returnFocusRef={triggerRef}
      >
        <UiText>Choose a round</UiText>
      </Sheet>,
    );

    const style = getByTestId("practice-sheet-panel").props.style;
    expect(style[0].paddingBottom).toBe(37);

    await act(async () => {
      rerender(
        <Sheet
          visible={false}
          onClose={jest.fn()}
          testID="practice-sheet"
          returnFocusRef={triggerRef}
        >
          <UiText>Choose a round</UiText>
        </Sheet>,
      );
    });

    await waitFor(() =>
      expect(AccessibilityInfo.setAccessibilityFocus).toHaveBeenCalledWith(
        findNodeHandle(triggerRef.current),
      ),
    );
    animationFrame.mockRestore();
  });
});
