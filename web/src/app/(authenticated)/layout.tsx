"use client";

import { QueryProvider } from "@/lib/providers/QueryProvider";
import { Sidebar } from "@/components/layout/Sidebar";

interface AuthLayoutProps {
  children: React.ReactNode;
}

export default function AuthenticatedLayout({ children }: AuthLayoutProps) {
  return (
    <QueryProvider>
      <div className="flex min-h-screen bg-gray-50">
        <Sidebar variant="mentor" />
        <main className="flex-1 p-6">{children}</main>
      </div>
    </QueryProvider>
  );
}
