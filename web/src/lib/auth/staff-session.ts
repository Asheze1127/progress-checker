const STAFF_SESSION_COOKIE_NAME = "progress_checker_staff_session";

export function setStaffSession(token: string): void {
  const maxAgeSeconds = 24 * 60 * 60;
  const isSecure = window.location.protocol === "https:";
  const secureFlag = isSecure ? "; Secure" : "";

  document.cookie = `${STAFF_SESSION_COOKIE_NAME}=${encodeURIComponent(token)}; Path=/; Max-Age=${maxAgeSeconds}; SameSite=Strict${secureFlag}`;
}

export function getStaffSession(): string | null {
  const cookies = document.cookie.split(";");

  for (const cookie of cookies) {
    const [name, ...valueParts] = cookie.trim().split("=");
    if (name === STAFF_SESSION_COOKIE_NAME) {
      const value = valueParts.join("=");
      return value ? decodeURIComponent(value) : null;
    }
  }

  return null;
}

export function clearStaffSession(): void {
  document.cookie = `${STAFF_SESSION_COOKIE_NAME}=; Path=/; Max-Age=0; SameSite=Strict`;
}
