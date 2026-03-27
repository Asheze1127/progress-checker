"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  getStaffSession,
  clearStaffSession,
} from "@/lib/auth/staff-session";

interface Team {
  id: string;
  name: string;
}

const createTeamSchema = z.object({
  name: z.string().min(1, "チーム名を入力してください"),
});

type CreateTeamData = z.infer<typeof createTeamSchema>;

function staffFetch(path: string, options?: RequestInit) {
  const token = getStaffSession();
  return fetch(path, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(options?.headers as Record<string, string>),
    },
  });
}

export default function StaffDashboardPage() {
  const router = useRouter();
  const [teams, setTeams] = useState<Team[]>([]);
  const [teamError, setTeamError] = useState<string | null>(null);

  const teamForm = useForm<CreateTeamData>({
    resolver: zodResolver(createTeamSchema),
  });

  const fetchTeams = useCallback(async () => {
    try {
      const res = await staffFetch("/api/v1/staff/teams");
      if (res.status === 401 || res.status === 403) {
        clearStaffSession();
        router.push("/staff/login");
        return;
      }
      if (res.ok) {
        const data = await res.json();
        setTeams(data.teams);
      }
    } catch {
      // network error
    }
  }, [router]);

  useEffect(() => {
    fetchTeams();
  }, [fetchTeams]);

  const handleCreateTeam = async (data: CreateTeamData) => {
    setTeamError(null);
    try {
      const res = await staffFetch("/api/v1/staff/teams", {
        method: "POST",
        body: JSON.stringify(data),
      });

      if (!res.ok) {
        const err = await res.json();
        setTeamError(err.error || "チーム作成に失敗しました");
        return;
      }

      teamForm.reset();
      await fetchTeams();
    } catch {
      setTeamError("ネットワークエラーが発生しました");
    }
  };

  const handleLogout = () => {
    clearStaffSession();
    router.push("/staff/login");
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
          <h1 className="text-xl font-bold text-gray-900">
            Staff Dashboard
          </h1>
          <button
            onClick={handleLogout}
            className="text-sm text-gray-600 hover:text-gray-900"
          >
            ログアウト
          </button>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-8 space-y-8">
        <section className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            チーム管理
          </h2>
          <p className="text-sm text-gray-500 mb-4">
            メンターの作成はSlackコマンド <code className="bg-gray-100 px-1 rounded">/create-mentor @user TeamName</code> で行います。
          </p>

          <form
            onSubmit={teamForm.handleSubmit(handleCreateTeam)}
            className="flex gap-3 mb-4"
          >
            <input
              {...teamForm.register("name")}
              placeholder="新しいチーム名"
              className="flex-1 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            />
            <button
              type="submit"
              disabled={teamForm.formState.isSubmitting}
              className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50 text-sm font-medium"
            >
              作成
            </button>
          </form>
          {teamForm.formState.errors.name && (
            <p className="text-sm text-red-600 mb-2">
              {teamForm.formState.errors.name.message}
            </p>
          )}
          {teamError && (
            <p className="text-sm text-red-600 mb-2">{teamError}</p>
          )}

          <div className="border rounded-md divide-y">
            {teams.length === 0 ? (
              <p className="px-4 py-3 text-sm text-gray-500">
                チームがまだありません
              </p>
            ) : (
              teams.map((team) => (
                <div key={team.id} className="px-4 py-3 flex justify-between items-center">
                  <span className="text-sm font-medium text-gray-900">
                    {team.name}
                  </span>
                  <span className="text-xs text-gray-400">{team.id}</span>
                </div>
              ))
            )}
          </div>
        </section>
      </main>
    </div>
  );
}
