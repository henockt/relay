"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth, useClerk } from "@clerk/nextjs";
import { configureAuth } from "@/lib/api";

// Renders the dashboard only once Clerk has loaded a signed-in session,
// so pages never call the API without a token.
export function AuthGate({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const { isLoaded, isSignedIn, getToken } = useAuth();
  const { signOut } = useClerk();

  // set during render: child effects (the first API calls) run before ours
  configureAuth({ getToken: () => getToken(), signOut: (redirectUrl) => signOut({ redirectUrl }) });

  useEffect(() => {
    if (isLoaded && !isSignedIn) router.replace("/");
  }, [isLoaded, isSignedIn, router]);

  if (!isLoaded || !isSignedIn) return null;
  return <>{children}</>;
}
