"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  getStaffSession,
  clearStaffSession,
} from "@/lib/auth/staff-session";

interface StaffAuthLayoutProps {
  children: React.ReactNode;
}

export default function StaffAuthenticatedLayout({
  children,
}: StaffAuthLayoutProps) {
  const router = useRouter();
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const token = getStaffSession();

    if (!token) {
      setIsLoading(false);
      router.replace("/staff/login");
      return;
    }

    try {
      const payload = JSON.parse(atob(token.split(".")[1]));
      const isExpired = payload.exp * 1000 < Date.now();
      const isStaff = payload.role === "staff";

      if (isExpired || !isStaff) {
        clearStaffSession();
        router.replace("/staff/login");
        return;
      }

      setIsAuthenticated(true);
    } catch {
      clearStaffSession();
      router.replace("/staff/login");
    } finally {
      setIsLoading(false);
    }
  }, [router]);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-gray-500">Loading...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return <>{children}</>;
}
