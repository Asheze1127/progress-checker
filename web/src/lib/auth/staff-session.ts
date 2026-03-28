/**
 * clearStaffSession calls the server-side logout route to remove the httpOnly staff session cookie.
 * Client-side JS cannot manipulate httpOnly cookies via document.cookie.
 */
export async function clearStaffSession(): Promise<void> {
  await fetch("/api/auth/staff/logout", { method: "POST" });
}
