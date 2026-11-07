// Only big consumer providers can sign up, to keep out throwaway addresses.
// The backend enforces the same list: keep in sync with internal/auth/domains.go
const ALLOWED_DOMAINS = new Set([
  "gmail.com",
  "googlemail.com",
  "icloud.com",
  "me.com",
  "mac.com",
  "outlook.com",
  "hotmail.com",
  "live.com",
  "msn.com",
  "yahoo.com",
  "proton.me",
  "protonmail.com",
]);

export function isAllowedEmail(email: string): boolean {
  const normalized = email.trim().toLowerCase();
  const at = normalized.lastIndexOf("@");
  if (at <= 0) return false;
  return ALLOWED_DOMAINS.has(normalized.slice(at + 1));
}
