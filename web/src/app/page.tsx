"use client";

import { Suspense, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuth } from "@clerk/nextjs";
import { Mail } from "lucide-react";
import { EmailCodeSignIn } from "@/components/email-code-sign-in";

// Set by the API client when it signs a user out after a rejected request.
const LOGIN_ERRORS: Record<string, string> = {
  email_not_allowed: "relay only accepts Gmail, iCloud, Outlook, Yahoo and Proton addresses.",
};

function LoginError() {
  const error = useSearchParams().get("error");
  if (!error) return null;

  return (
    <p
      role="alert"
      className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
    >
      {LOGIN_ERRORS[error] ?? "We couldn't sign you in. Please try again."}
    </p>
  );
}

export default function Home() {
  const router = useRouter();
  const { isSignedIn } = useAuth();

  useEffect(() => {
    if (isSignedIn) router.replace("/dashboard/aliases");
  }, [isSignedIn, router]);
  return (
    <div className="grid min-h-svh lg:grid-cols-2">
      {/* Left — branding panel */}
      <div className="relative hidden bg-zinc-900 lg:flex flex-col justify-between p-10 text-white">
        <div className="flex items-center gap-2 text-lg font-semibold">
          <Mail className="size-5 shrink-0" />
          relay
        </div>

        <div className="space-y-4">
          <h1 className="text-4xl font-bold leading-tight">
            Your inbox,<br />your rules.
          </h1>
          <p className="text-zinc-400 max-w-sm">
            Generate disposable email aliases that forward to your real address.
            Block senders. Stay private. Delete anytime.
          </p>
          <ul className="space-y-2 text-zinc-400 text-sm">
            {[
              "Multiple aliases for different use cases",
              "Forwarding with attachment support",
              "Per-alias enable / disable toggle",
            ].map((f) => (
              <li key={f} className="flex items-center gap-2">
                <CheckIcon className="size-4 text-zinc-500" />
                {f}
              </li>
            ))}
          </ul>
        </div>

        <p className="text-xs text-zinc-600">
          Open source · No tracking
        </p>
      </div>

      {/* Right — sign-in panel */}
      <div className="flex flex-col items-center justify-center gap-8 p-8">
        {/* Mobile logo */}
        <div className="flex items-center gap-2 text-lg font-semibold lg:hidden">
          <Mail className="size-5" />
          relay
        </div>

        <div className="w-full max-w-sm space-y-6">
          <div className="space-y-2 text-center">
            <h2 className="text-2xl font-bold">Welcome back</h2>
            <p className="text-sm text-muted-foreground">
              Sign in to manage your aliases
            </p>
          </div>

          <Suspense fallback={null}>
            <LoginError />
          </Suspense>

          <EmailCodeSignIn />

          <p className="text-center text-xs text-muted-foreground">
            By signing in, you agree to keep your aliases tidy.
          </p>
        </div>
      </div>
    </div>
  );
}

function CheckIcon({ className }: { className?: string }) {
  return (
    <svg className={className} xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"
      fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M20 6 9 17l-5-5" />
    </svg>
  );
}
