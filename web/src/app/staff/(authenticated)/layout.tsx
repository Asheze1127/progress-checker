"use client";

import { QueryProvider } from "@/lib/providers/QueryProvider";
import { Sidebar } from "@/components/layout/Sidebar";

interface StaffAuthLayoutProps {
  children: React.ReactNode;
}

export default function StaffAuthenticatedLayout({
  children,
}: StaffAuthLayoutProps) {
  return (
    <QueryProvider>
      <div className="flex min-h-screen bg-gray-50">
        <Sidebar variant="staff" />
        <main className="flex-1 p-6">{children}</main>
      </div>
    </QueryProvider>
  );
}
