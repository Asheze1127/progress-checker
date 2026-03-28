"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { clearSession } from "@/lib/auth/session";
import { clearStaffSession } from "@/lib/auth/staff-session";

interface NavItem {
  href: string;
  label: string;
}

const mentorNav: NavItem[] = [
  { href: "/progress", label: "進捗一覧" },
  { href: "/teams", label: "チーム" },
];

const staffNav: NavItem[] = [
  { href: "/staff/dashboard", label: "チーム管理" },
];

export function Sidebar({ variant }: { variant: "mentor" | "staff" }) {
  const pathname = usePathname();
  const router = useRouter();
  const nav = variant === "staff" ? staffNav : mentorNav;

  const handleLogout = async () => {
    if (variant === "staff") {
      await clearStaffSession();
      router.push("/staff/login");
    } else {
      await clearSession();
      router.push("/login");
    }
  };

  return (
    <aside className="w-56 bg-white border-r border-gray-200 min-h-screen flex flex-col">
      <div className="px-4 py-5 border-b border-gray-200">
        <h1 className="text-lg font-bold text-gray-900">
          {variant === "staff" ? "Staff" : "Mentor"}
        </h1>
      </div>

      <nav className="flex-1 px-2 py-4 space-y-1">
        {nav.map((item) => {
          const isActive =
            pathname === item.href || pathname.startsWith(`${item.href}/`);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`block px-3 py-2 rounded-md text-sm font-medium ${
                isActive
                  ? "bg-indigo-50 text-indigo-700"
                  : "text-gray-700 hover:bg-gray-100"
              }`}
            >
              {item.label}
            </Link>
          );
        })}
      </nav>

      <div className="px-2 py-4 border-t border-gray-200">
        <button
          type="button"
          onClick={handleLogout}
          className="w-full px-3 py-2 text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-md text-left"
        >
          ログアウト
        </button>
      </div>
    </aside>
  );
}
