"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Mail } from "lucide-react";
import Link from "next/link";
import { API_URL as API } from "@/lib/config";

export default function Home() {
  const router = useRouter();

  useEffect(() => {
    if (localStorage.getItem("relay_token")) {
      router.replace("/dashboard/aliases");
    }
  }, [router]);
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

          <Link
            href={`${API}/api/auth/google`}
            className="flex w-full items-center justify-center gap-3 rounded-md border border-input bg-background px-4 py-2.5 text-sm font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground"
          >
            <GoogleIcon />
            Continue with Google
          </Link>

          <p className="text-center text-xs text-muted-foreground">
            By signing in you agree to keep your aliases tidy.
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

function GoogleIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 18 18" aria-hidden="true">
      <path fill="#4285F4"
        d="M17.64 9.2c0-.637-.057-1.251-.164-1.84H9v3.481h4.844a4.14 4.14 0 0 1-1.796 2.716v2.259h2.908c1.702-1.567 2.684-3.875 2.684-6.615Z" />
      <path fill="#34A853"
        d="M9 18c2.43 0 4.467-.806 5.956-2.18l-2.908-2.259c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332A8.997 8.997 0 0 0 9 18Z" />
      <path fill="#FBBC05"
        d="M3.964 10.71A5.41 5.41 0 0 1 3.682 9c0-.593.102-1.17.282-1.71V4.958H.957A8.996 8.996 0 0 0 0 9c0 1.452.348 2.827.957 4.042l3.007-2.332Z" />
      <path fill="#EA4335"
        d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0A8.997 8.997 0 0 0 .957 4.958L3.964 7.29C4.672 5.163 6.656 3.58 9 3.58Z" />
    </svg>
  );
}