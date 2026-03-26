"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getSession } from "@/lib/auth/session";
import { QueryProvider } from "@/lib/providers/QueryProvider";

const VALIDATE_SESSION_URL = "/api/v1/auth/validate";

interface AuthLayoutProps {
  children: React.ReactNode;
}

/**
 * AuthenticatedLayout wraps pages that require authentication.
 * It checks for a valid session on mount and redirects to /login if not authenticated.
 */
export default function AuthenticatedLayout({ children }: AuthLayoutProps) {
  const router = useRouter();
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const checkAuth = async () => {
      const token = getSession();

      if (!token) {
        router.replace("/login");
        return;
      }

      try {
        const response = await fetch(VALIDATE_SESSION_URL, {
          method: "GET",
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (!response.ok) {
          router.replace("/login");
          return;
        }

        setIsAuthenticated(true);
      } catch {
        router.replace("/login");
      } finally {
        setIsLoading(false);
      }
    };

    checkAuth();
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

  return <QueryProvider>{children}</QueryProvider>;
}
