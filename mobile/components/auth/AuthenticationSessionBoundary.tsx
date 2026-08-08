import {
  getAuthenticationSessionEpoch,
  isAuthenticationSessionReady,
  subscribeAuthenticationSessionChanges,
} from "@/api/users";
import { Fragment, type ReactNode, useEffect, useState } from "react";

type AuthenticationSessionBoundaryProps = {
  children: ReactNode;
};

export function AuthenticationSessionBoundary({
  children,
}: AuthenticationSessionBoundaryProps) {
  const [session, setSession] = useState(() => ({
    epoch: getAuthenticationSessionEpoch(),
    ready: isAuthenticationSessionReady(),
  }));

  useEffect(() => {
    const reconcileSession = (epoch: string | null) => {
      setSession({
        epoch,
        ready: isAuthenticationSessionReady(),
      });
    };
    const unsubscribe = subscribeAuthenticationSessionChanges(reconcileSession);
    // Close the render-to-effect window: the complete A -> pending -> B
    // transition may have occurred before the listener was installed.
    reconcileSession(getAuthenticationSessionEpoch());
    return unsubscribe;
  }, []);

  if (!session.ready) return null;

  return (
    <Fragment key={session.epoch ?? "anonymous-session"}>{children}</Fragment>
  );
}
