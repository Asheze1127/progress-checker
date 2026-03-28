import { NextRequest, NextResponse } from "next/server";
import { SESSION_COOKIE_NAME, STAFF_SESSION_COOKIE_NAME } from "@/lib/auth/constants";

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Protect mentor authenticated routes
  if (
    pathname.startsWith("/progress") ||
    pathname.startsWith("/teams") ||
    pathname.startsWith("/questions")
  ) {
    const token = request.cookies.get(SESSION_COOKIE_NAME)?.value;
    if (!token) {
      return NextResponse.redirect(new URL("/login", request.url));
    }
  }

  // Protect staff authenticated routes
  if (pathname.startsWith("/staff/dashboard")) {
    const token = request.cookies.get(STAFF_SESSION_COOKIE_NAME)?.value;
    if (!token) {
      return NextResponse.redirect(new URL("/staff/login", request.url));
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/progress/:path*",
    "/teams/:path*",
    "/questions/:path*",
    "/staff/dashboard/:path*",
  ],
};
