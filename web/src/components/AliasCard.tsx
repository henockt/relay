"use client";

import { useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { Alias } from "@/lib/api";

interface Props {
  alias: Alias;
  onToggle: (alias: Alias) => void;
  onDelete: (id: string) => void;
}

export default function AliasCard({ alias, onToggle, onDelete }: Props) {
  const [copied, setCopied] = useState(false);

  async function copy() {
    await navigator.clipboard.writeText(alias.address);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <Card className={alias.enabled ? "" : "opacity-60"}>
      <CardContent className="flex items-start justify-between gap-4 py-4">
        {/* Left: address + stats */}
        <div className="flex-1 min-w-0 space-y-1">
          <div className="flex items-center gap-2 flex-wrap">
            <code className="text-m font-mono truncate">{alias.address}</code>
            <Button
              variant="ghost"
              size="sm"
              className="h-6 px-2 text-xs text-muted-foreground"
              onClick={copy}
            >
              {copied ? "Copied!" : "Copy"}
            </Button>
          </div>

          {alias.label && (
            <p className="text-sm text-muted-foreground truncate">{alias.label}</p>
          )}

          <p className="text-xs text-muted-foreground">
            {alias.emails_forwarded} forwarded &middot; {alias.emails_blocked} blocked
          </p>
        </div>

        {/* Right: status badge + controls */}
        <div className="flex items-center gap-3 shrink-0">
          <Badge variant={alias.enabled ? "default" : "secondary"}>
            {alias.enabled ? "Active" : "Paused"}
          </Badge>

          {/* Toggle switch */}
          <button
            onClick={() => onToggle(alias)}
            title={alias.enabled ? "Pause alias" : "Activate alias"}
            className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring ${
              alias.enabled ? "bg-primary" : "bg-input"
            }`}
          >
            <span
              className={`pointer-events-none inline-block h-4 w-4 rounded-full bg-background shadow-lg ring-0 transition-transform ${
                alias.enabled ? "translate-x-4" : "translate-x-0"
              }`}
            />
          </button>

          <Button
            variant="ghost"
            size="sm"
            className="h-6 px-2 text-xs text-destructive hover:text-destructive"
            onClick={() => onDelete(alias.id)}
          >
            Delete
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}