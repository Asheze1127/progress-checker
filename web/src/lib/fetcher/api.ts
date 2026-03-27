import { getSession } from "@/lib/auth/session";
import { env } from "@/lib/env";

/**
 * Generic fetch wrapper for the backend API.
 * Automatically attaches the session token as a Bearer token.
 * Throws on non-2xx responses with the HTTP status code.
 */
export async function fetchAPI<T>(
  path: string,
  options?: RequestInit,
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options?.headers as Record<string, string>),
  };

  const token = getSession();
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const response = await fetch(`${env.NEXT_PUBLIC_API_URL}${path}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    throw new Error(`API error: ${response.status}`);
  }

  return response.json() as Promise<T>;
}
