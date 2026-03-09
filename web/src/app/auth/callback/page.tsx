"use client";

import { useEffect, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";

function CallbackHandler() {
  const router = useRouter();
  const params = useSearchParams();

  useEffect(() => {
    const token = params.get("token");
    if (token) {
      localStorage.setItem("relay_token", token);
      router.replace("/dashboard/aliases");
    } else {
      router.replace("/");
    }
  }, [params, router]);

  return <p className="text-m text-muted-foreground">Signing you in…</p>;
}

export default function AuthCallback() {
  return (
    <div className="flex h-screen items-center justify-center">
      <Suspense fallback={<p className="text-m text-muted-foreground">Signing you in…</p>}>
        <CallbackHandler />
      </Suspense>
    </div>
  );
}