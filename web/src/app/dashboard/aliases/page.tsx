"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import toast, { Toaster } from "react-hot-toast";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api, type Alias } from "@/lib/api";
import AliasCard from "@/components/AliasCard";

export default function AliasesPage() {
  const router = useRouter();
  const [aliases, setAliases] = useState<Alias[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [label, setLabel] = useState("");

  useEffect(() => {
    if (!localStorage.getItem("relay_token")) {
      router.replace("/");
      return;
    }
    load();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  async function load() {
    try {
      setAliases(await api.listAliases());
    } catch {
      toast.error("Failed to load aliases");
    } finally {
      setLoading(false);
    }
  }

  async function handleCreate() {
    setCreating(true);
    try {
      const created = await api.createAlias(label.trim() || undefined);
      setAliases((prev) => [created, ...prev]);
      setLabel("");
      toast.success("Alias created");
    } catch {
      toast.error("Failed to create alias");
    } finally {
      setCreating(false);
    }
  }

  async function handleToggle(alias: Alias) {
    try {
      const updated = await api.updateAlias(alias.id, { enabled: !alias.enabled });
      setAliases((prev) => prev.map((a) => (a.id === updated.id ? updated : a)));
    } catch {
      toast.error("Failed to update alias");
    }
  }

  async function handleDelete(id: string) {
    try {
      await api.deleteAlias(id);
      setAliases((prev) => prev.filter((a) => a.id !== id));
      toast.success("Alias deleted");
    } catch {
      toast.error("Failed to delete alias");
    }
  }

  return (
    <div className="max-w-2xl space-y-6">
      <Toaster position="top-right" />

      <div>
        <h1 className="text-2xl font-bold">Aliases</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Each alias forwards to your real inbox. Share the alias, not your email.
        </p>
      </div>

      <div className="flex gap-2">
        <Input
          placeholder="Label (optional)"
          value={label}
          onChange={(e) => setLabel(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && !creating && handleCreate()}
          className="max-w-xs"
        />
        <Button onClick={handleCreate} disabled={creating}>
          {creating ? "Creating…" : "New alias"}
        </Button>
      </div>

      {loading ? (
        <p className="text-sm text-muted-foreground">Loading…</p>
      ) : aliases.length === 0 ? (
        <div className="rounded-xl border border-dashed p-12 text-center text-muted-foreground">
          <p className="font-medium">No aliases yet</p>
          <p className="text-sm mt-1">Create one above to get started.</p>
        </div>
      ) : (
        <ul className="flex flex-col gap-3">
          {aliases.map((alias) => (
            <AliasCard
              key={alias.id}
              alias={alias}
              onToggle={handleToggle}
              onDelete={handleDelete}
            />
          ))}
        </ul>
      )}
    </div>
  );
}