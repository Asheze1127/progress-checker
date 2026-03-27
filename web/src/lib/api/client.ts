import createClient from "openapi-fetch";
import type { paths } from "./v1";
import { getSession } from "@/lib/auth/session";
import { getStaffSession } from "@/lib/auth/staff-session";
import { env } from "@/lib/env";

export const api = createClient<paths>({
  baseUrl: env.NEXT_PUBLIC_API_URL,
});

// Attach auth token to every request based on the API path
api.use({
  onRequest({ request }) {
    const url = new URL(request.url);
    const isStaffEndpoint = url.pathname.startsWith("/api/v1/staff");
    const token = isStaffEndpoint ? getStaffSession() : getSession();
    if (token) {
      request.headers.set("Authorization", `Bearer ${token}`);
    }
    return request;
  },
});
