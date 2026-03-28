import { NextRequest, NextResponse } from "next/server";
import { env } from "@/lib/env";
import { STAFF_SESSION_COOKIE_NAME, MAX_AGE_SECONDS } from "@/lib/auth/constants";

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();

    const res = await fetch(`${env.NEXT_PUBLIC_API_URL}/api/v1/staff/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });

    const data = await res.json();

    if (!res.ok) {
      return NextResponse.json(data, { status: res.status });
    }

    const response = NextResponse.json({ staff: data.staff });
    response.cookies.set(STAFF_SESSION_COOKIE_NAME, data.token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "strict",
      path: "/",
      maxAge: MAX_AGE_SECONDS,
    });

    return response;
  } catch {
    return NextResponse.json(
      { error: "Internal server error" },
      { status: 500 },
    );
  }
}
