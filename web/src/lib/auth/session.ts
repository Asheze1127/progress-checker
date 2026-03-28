/**
 * clearSession calls the server-side logout route to remove the httpOnly session cookie.
 * Client-side JS cannot manipulate httpOnly cookies via document.cookie.
 */
export async function clearSession(): Promise<void> {
  await fetch("/api/auth/logout", { method: "POST" });
}
