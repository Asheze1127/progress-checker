const SESSION_COOKIE_NAME = "progress_checker_session";

/**
 * Store the session token in a cookie.
 * Uses Secure and SameSite=Strict attributes for security.
 * The cookie expires after 24 hours to match the backend session expiry.
 */
export function setSession(token: string): void {
  const maxAgeSeconds = 24 * 60 * 60; // 24 hours
  const isSecure = window.location.protocol === "https:";
  const secureFlag = isSecure ? "; Secure" : "";

  document.cookie = `${SESSION_COOKIE_NAME}=${encodeURIComponent(token)}; Path=/; Max-Age=${maxAgeSeconds}; SameSite=Strict${secureFlag}`;
}

/**
 * Retrieve the session token from the cookie.
 * Returns null if no session token is found.
 */
export function getSession(): string | null {
  const cookies = document.cookie.split(";");

  for (const cookie of cookies) {
    const [name, ...valueParts] = cookie.trim().split("=");
    if (name === SESSION_COOKIE_NAME) {
      const value = valueParts.join("=");
      return value ? decodeURIComponent(value) : null;
    }
  }

  return null;
}

/**
 * Clear the session token by expiring the cookie.
 */
export function clearSession(): void {
  document.cookie = `${SESSION_COOKIE_NAME}=; Path=/; Max-Age=0; SameSite=Strict`;
}
