import { NextResponse } from "next/server";
import { STAFF_SESSION_COOKIE_NAME } from "@/lib/auth/constants";

export async function POST() {
  const response = NextResponse.json({ message: "logged out" });
  response.cookies.set(STAFF_SESSION_COOKIE_NAME, "", {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "strict",
    path: "/",
    maxAge: 0,
  });
  return response;
}
