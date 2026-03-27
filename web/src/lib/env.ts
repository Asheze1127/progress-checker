export const env = {
  NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL ?? "",
} as const;

// Build-time validation — runs only on the server during startup
if (typeof window === "undefined" && !env.NEXT_PUBLIC_API_URL) {
  throw new Error("Missing required environment variable: NEXT_PUBLIC_API_URL");
}
