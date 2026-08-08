import { fireEvent, render, waitFor } from "@testing-library/react-native";
import { Text, TouchableOpacity } from "react-native";
import { act, useState } from "react";

import { AuthenticationSessionBoundary } from "@/components/auth/AuthenticationSessionBoundary";
import {
  getAuthenticationSessionEpoch,
  isAuthenticationSessionReady,
  subscribeAuthenticationSessionChanges,
} from "@/api/users";

jest.mock("@/api/users", () => ({
  getAuthenticationSessionEpoch: jest.fn(() => "identity-a"),
  isAuthenticationSessionReady: jest.fn(() => true),
  subscribeAuthenticationSessionChanges: jest.fn(),
}));

function AuthenticatedScreen() {
  const [selectedWordlist, setSelectedWordlist] = useState<string | null>(null);
  return (
    <>
      <TouchableOpacity onPress={() => setSelectedWordlist("A private list")}>
        <Text>Open detail</Text>
      </TouchableOpacity>
      {selectedWordlist ? <Text>{selectedWordlist}</Text> : null}
    </>
  );
}

describe("AuthenticationSessionBoundary", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.mocked(getAuthenticationSessionEpoch).mockReturnValue("identity-a");
    jest.mocked(isAuthenticationSessionReady).mockReturnValue(true);
  });

  it("unmounts identity A local UI state before mounting identity B", () => {
    let notifySessionChange!: (epoch: string | null) => void;
    jest
      .mocked(subscribeAuthenticationSessionChanges)
      .mockImplementation((listener) => {
        notifySessionChange = listener;
        return jest.fn();
      });
    const view = render(
      <AuthenticationSessionBoundary>
        <AuthenticatedScreen />
      </AuthenticationSessionBoundary>,
    );
    fireEvent.press(view.getByText("Open detail"));
    expect(view.getByText("A private list")).toBeTruthy();

    jest.mocked(isAuthenticationSessionReady).mockReturnValue(false);
    jest.mocked(getAuthenticationSessionEpoch).mockReturnValue(null);
    act(() => notifySessionChange(null));
    expect(view.queryByText("A private list")).toBeNull();
    expect(view.queryByText("Open detail")).toBeNull();

    jest.mocked(isAuthenticationSessionReady).mockReturnValue(true);
    jest.mocked(getAuthenticationSessionEpoch).mockReturnValue("identity-b");
    act(() => notifySessionChange("identity-b"));
    expect(view.getByText("Open detail")).toBeTruthy();
    expect(view.queryByText("A private list")).toBeNull();
  });

  it("reconciles a complete handoff missed before effect subscription", async () => {
    let mountCount = 0;
    function StatefulScreen() {
      const [instance] = useState(() => ++mountCount);
      return <Text>{`Session instance ${instance}`}</Text>;
    }
    jest
      .mocked(getAuthenticationSessionEpoch)
      .mockReturnValueOnce("identity-a")
      .mockReturnValue("identity-b");
    jest
      .mocked(subscribeAuthenticationSessionChanges)
      .mockImplementation(() => jest.fn());

    const view = render(
      <AuthenticationSessionBoundary>
        <StatefulScreen />
      </AuthenticationSessionBoundary>,
    );

    await waitFor(() =>
      expect(view.getByText("Session instance 2")).toBeTruthy(),
    );
    expect(mountCount).toBe(2);
  });
});
