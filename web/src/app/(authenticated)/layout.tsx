"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getSession } from "@/lib/auth/session";
import { QueryProvider } from "@/lib/providers/QueryProvider";
import { Sidebar } from "@/components/layout/Sidebar";

interface AuthLayoutProps {
  children: React.ReactNode;
}

export default function AuthenticatedLayout({ children }: AuthLayoutProps) {
  const router = useRouter();
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const token = getSession();

    if (!token) {
      router.replace("/login");
    } else {
      setIsAuthenticated(true);
    }

    setIsLoading(false);
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

  return (
    <QueryProvider>
      <div className="flex min-h-screen bg-gray-50">
        <Sidebar variant="mentor" />
        <main className="flex-1 p-6">{children}</main>
      </div>
    </QueryProvider>
  );
}
