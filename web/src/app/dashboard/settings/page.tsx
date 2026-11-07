"use client";

import { useEffect, useState } from "react";
import { useClerk } from "@clerk/nextjs";
import toast, { Toaster } from "react-hot-toast";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { api, type User } from "@/lib/api";

export default function SettingsPage() {
  const clerk = useClerk();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    api.getMe()
      .then(setUser)
      .catch(() => toast.error("Failed to load account info"))
      .finally(() => setLoading(false));
  }, []);

  async function handleDeleteAccount() {
    setDeleting(true);
    try {
      await api.deleteMe();
      toast.success("Account deleted");
      await clerk.signOut({ redirectUrl: "/" });
    } catch {
      toast.error("Failed to delete account");
      setDeleting(false);
    }
  }

  function signOut() {
    clerk.signOut({ redirectUrl: "/" });
  }

  return (
    <div className="max-w-2xl space-y-6">
      <Toaster position="top-right" />

      <div>
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Manage your account.
        </p>
      </div>

      <Card>
        {/* <CardHeader>
          <CardTitle>Account</CardTitle>
          <CardDescription>Your sign-in details.</CardDescription>
        </CardHeader> */}
        <CardContent className="space-y-4">
          {loading ? (
            <p className="text-sm text-muted-foreground">Loading…</p>
          ) : user ? (
            <>
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium">Email</p>
                  <p className="text-sm text-muted-foreground">{user.email}</p>
                </div>
              </div>
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium">Sign-in method</p>
                </div>
                <Badge variant="outline" className="capitalize">
                  {user.provider}
                </Badge>
              </div>
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium">Member since</p>
                  <p className="text-sm text-muted-foreground">
                    {new Date(user.created_at).toLocaleDateString()}
                  </p>
                </div>
              </div>
            </>
          ) : null}

          <div className="pt-2">
            <Button variant="outline" onClick={signOut}>
              Sign out
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card className="border-destructive/50">
        <CardContent>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium">Delete account</p>
              <p className="text-sm text-muted-foreground">
                Removes your account and all aliases permanently.
              </p>
            </div>

            <Dialog>
              <DialogTrigger asChild>
                <Button variant="destructive" size="sm">
                  Delete account
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Delete your account?</DialogTitle>
                  <DialogDescription>
                    This will permanently delete your account and all{" "}
                    <strong>aliases</strong>. Emails sent to those aliases will
                    no longer be forwarded. This cannot be undone.
                  </DialogDescription>
                </DialogHeader>
                <DialogFooter>
                  <Button
                    variant="destructive"
                    onClick={handleDeleteAccount}
                    disabled={deleting}
                  >
                    {deleting ? "Deleting…" : "Yes, delete everything"}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}