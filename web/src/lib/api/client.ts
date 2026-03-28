import createClient from "openapi-fetch";
import type { paths } from "./v1";
import { clearSession } from "@/lib/auth/session";
import { clearStaffSession } from "@/lib/auth/staff-session";
import { env } from "@/lib/env";

export const api = createClient<paths>({
  baseUrl: env.NEXT_PUBLIC_API_URL,
  credentials: "include",
});

api.use({
  async onResponse({ response }) {
    if (response.status === 401) {
      const url = new URL(response.url);
      const isStaffEndpoint = url.pathname.startsWith("/api/v1/staff");
      if (isStaffEndpoint) {
        await clearStaffSession();
        if (typeof window !== "undefined") {
          window.location.href = "/staff/login";
        }
      } else {
        await clearSession();
        if (typeof window !== "undefined") {
          window.location.href = "/login";
        }
      }
    }
    return response;
  },
});
