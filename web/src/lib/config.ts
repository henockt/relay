// The Go server reverse-proxies all non-API routes to Next.js, so the API is
// always same-origin in a deployed container. Leave this empty to use relative
// URLs. only set NEXT_PUBLIC_API_URL when the backend is hosted separately.
export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "";
