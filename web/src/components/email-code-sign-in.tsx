"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useSignIn, useSignUp } from "@clerk/nextjs";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { isAllowedEmail } from "@/lib/allowed-emails";

const NOT_ALLOWED =
  "Please use a Gmail, iCloud, Outlook, Yahoo or Proton address.";

type ClerkErr = { code?: string; message?: string; errors?: { code?: string; longMessage?: string; message?: string }[] };

function errorCode(err: unknown) {
  const e = err as ClerkErr;
  return e?.errors?.[0]?.code ?? e?.code;
}

function errorMessage(err: unknown) {
  const e = err as ClerkErr;
  const first = e?.errors?.[0];
  return first?.longMessage ?? first?.message ?? e?.message ?? "Something went wrong. Please try again.";
}

// Two steps: email, then the 6-digit code Clerk emails. The domain is
// checked before anything is sent; existing users sign in, new ones sign up.
export function EmailCodeSignIn() {
  const router = useRouter();
  const { signIn } = useSignIn();
  const { signUp } = useSignUp();

  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [mode, setMode] = useState<"signIn" | "signUp" | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const goToDashboard = () => router.replace("/dashboard/aliases");

  async function sendCode(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    const address = email.trim();
    if (!isAllowedEmail(address)) {
      setError(NOT_ALLOWED);
      return;
    }
    if (!signIn || !signUp) return;

    setBusy(true);
    try {
      const { error: signInErr } = await signIn.emailCode.sendCode({ emailAddress: address });
      if (!signInErr) {
        setMode("signIn");
        return;
      }
      if (errorCode(signInErr) !== "form_identifier_not_found") {
        setError(errorMessage(signInErr));
        return;
      }

      // no account yet: sign up instead, same email-code flow
      const { error: createErr } = await signUp.create({ emailAddress: address });
      if (createErr) {
        setError(errorMessage(createErr));
        return;
      }
      const { error: sendErr } = await signUp.verifications.sendEmailCode();
      if (sendErr) {
        setError(errorMessage(sendErr));
        return;
      }
      setMode("signUp");
    } finally {
      setBusy(false);
    }
  }

  async function verifyCode(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    if (!signIn || !signUp) return;

    setBusy(true);
    try {
      if (mode === "signIn") {
        const { error: verifyErr } = await signIn.emailCode.verifyCode({ code: code.trim() });
        if (verifyErr) return setError(errorMessage(verifyErr));
        if (signIn.status !== "complete") return setError("Couldn't finish signing in. Please try again.");
        await signIn.finalize({ navigate: goToDashboard });
      } else {
        const { error: verifyErr } = await signUp.verifications.verifyEmailCode({ code: code.trim() });
        if (verifyErr) return setError(errorMessage(verifyErr));
        if (signUp.status !== "complete") return setError("Couldn't finish signing up. Please try again.");
        await signUp.finalize({ navigate: goToDashboard });
      }
    } finally {
      setBusy(false);
    }
  }

  function startOver() {
    setMode(null);
    setCode("");
    setError(null);
  }

  const errorBox = error && (
    <p
      role="alert"
      className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
    >
      {error}
    </p>
  );

  if (!mode) {
    return (
      <form onSubmit={sendCode} className="space-y-4">
        {errorBox}
        <div className="space-y-2">
          <Label htmlFor="email">Email</Label>
          <Input
            id="email"
            type="email"
            autoComplete="email"
            placeholder="you@gmail.com"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </div>
        <Button type="submit" className="w-full" disabled={busy || !signIn}>
          {busy ? "Sending code…" : "Continue with email"}
        </Button>
        {/* Clerk's bot protection renders here when it's enabled */}
        <div id="clerk-captcha" />
      </form>
    );
  }

  return (
    <form onSubmit={verifyCode} className="space-y-4">
      {errorBox}
      <p className="text-sm text-muted-foreground">
        We sent a code to <span className="font-medium text-foreground">{email.trim()}</span>.
      </p>
      <div className="space-y-2">
        <Label htmlFor="code">Verification code</Label>
        <Input
          id="code"
          inputMode="numeric"
          autoComplete="one-time-code"
          maxLength={6}
          required
          autoFocus
          value={code}
          onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))}
        />
      </div>
      <Button type="submit" className="w-full" disabled={busy || code.length < 6}>
        {busy ? "Verifying…" : "Sign in"}
      </Button>
      <Button type="button" variant="ghost" className="w-full" onClick={startOver} disabled={busy}>
        Use a different email
      </Button>
    </form>
  );
}
