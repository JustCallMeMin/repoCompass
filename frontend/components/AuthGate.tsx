"use client";

import { ReactNode, useEffect, useState } from "react";
import { getSession, SessionInfo } from "../lib/api";
import { ErrorState, LoadingState, UnauthorizedState } from "./ui";

export function AuthGate({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<SessionInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    getSession()
      .then(setSession)
      .catch((err) => setError(err instanceof Error ? err.message : "Session check failed"))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <LoadingState label="Checking session" />;
  }
  if (error) {
    return <UnauthorizedState />;
  }
  if (!session?.user_id) {
    return <ErrorState message="Session is missing a user." />;
  }
  return children;
}
