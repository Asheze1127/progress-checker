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

interface User {
  id: string;
  name: string;
  role: string;
}

const createTeamSchema = z.object({
  name: z.string().min(1, "チーム名を入力してください"),
});

const createUserSchema = z.object({
  slack_user_id: z.string().min(1, "Slack User IDを入力してください"),
  name: z.string().min(1, "名前を入力してください"),
  email: z.string().email("有効なメールアドレスを入力してください"),
  role: z.enum(["participant", "mentor"]),
  team_id: z.string().min(1, "チームを選択してください"),
});

type CreateTeamData = z.infer<typeof createTeamSchema>;
type CreateUserData = z.infer<typeof createUserSchema>;

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
  const [users, setUsers] = useState<User[]>([]);
  const [setupUrl, setSetupUrl] = useState<string | null>(null);
  const [teamError, setTeamError] = useState<string | null>(null);
  const [userError, setUserError] = useState<string | null>(null);

  const teamForm = useForm<CreateTeamData>({
    resolver: zodResolver(createTeamSchema),
  });

  const userForm = useForm<CreateUserData>({
    resolver: zodResolver(createUserSchema),
    defaultValues: { role: "participant" },
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
      // network error — leave current state
    }
  }, [router]);

  const fetchUsers = useCallback(async () => {
    try {
      const res = await staffFetch("/api/v1/staff/users");
      if (res.status === 401 || res.status === 403) {
        clearStaffSession();
        router.push("/staff/login");
        return;
      }
      if (res.ok) {
        const data = await res.json();
        setUsers(data.users);
      }
    } catch {
      // network error — leave current state
    }
  }, [router]);

  useEffect(() => {
    fetchTeams();
    fetchUsers();
  }, [fetchTeams, fetchUsers]);

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

  const handleCreateUser = async (data: CreateUserData) => {
    setUserError(null);
    setSetupUrl(null);
    try {
      const res = await staffFetch("/api/v1/staff/users", {
        method: "POST",
        body: JSON.stringify(data),
      });

      if (!res.ok) {
        const err = await res.json();
        setUserError(err.error || "ユーザー作成に失敗しました");
        return;
      }

      const result = await res.json();
      setSetupUrl(result.setup_url);
      userForm.reset();
      await fetchUsers();
    } catch {
      setUserError("ネットワークエラーが発生しました");
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
        {/* Team Management */}
        <section className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            チーム管理
          </h2>

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

        {/* User Management */}
        <section className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            ユーザー管理
          </h2>

          <form
            onSubmit={userForm.handleSubmit(handleCreateUser)}
            className="space-y-3 mb-4"
          >
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Slack User ID
                </label>
                <input
                  {...userForm.register("slack_user_id")}
                  placeholder="U01ABCDEF"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
                {userForm.formState.errors.slack_user_id && (
                  <p className="mt-1 text-sm text-red-600">
                    {userForm.formState.errors.slack_user_id.message}
                  </p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  名前
                </label>
                <input
                  {...userForm.register("name")}
                  placeholder="山田太郎"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
                {userForm.formState.errors.name && (
                  <p className="mt-1 text-sm text-red-600">
                    {userForm.formState.errors.name.message}
                  </p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Email
                </label>
                <input
                  {...userForm.register("email")}
                  type="email"
                  placeholder="user@example.com"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
                {userForm.formState.errors.email && (
                  <p className="mt-1 text-sm text-red-600">
                    {userForm.formState.errors.email.message}
                  </p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  ロール
                </label>
                <select
                  {...userForm.register("role")}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                >
                  <option value="participant">Participant</option>
                  <option value="mentor">Mentor</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  チーム
                </label>
                <select
                  {...userForm.register("team_id")}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                >
                  <option value="">チームを選択</option>
                  {teams.map((team) => (
                    <option key={team.id} value={team.id}>
                      {team.name}
                    </option>
                  ))}
                </select>
                {userForm.formState.errors.team_id && (
                  <p className="mt-1 text-sm text-red-600">
                    {userForm.formState.errors.team_id.message}
                  </p>
                )}
              </div>
            </div>

            <button
              type="submit"
              disabled={userForm.formState.isSubmitting}
              className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50 text-sm font-medium"
            >
              ユーザー作成
            </button>
          </form>

          {userError && (
            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4">
              {userError}
            </div>
          )}

          {setupUrl && (
            <div className="bg-green-50 border border-green-200 text-green-800 px-4 py-3 rounded mb-4">
              <p className="font-medium mb-1">
                ユーザーを作成しました。以下のURLをユーザーに共有してください:
              </p>
              <div className="flex items-center gap-2">
                <code className="bg-white px-2 py-1 rounded text-sm break-all flex-1">
                  {setupUrl}
                </code>
                <button
                  onClick={() => navigator.clipboard.writeText(setupUrl)}
                  className="px-3 py-1 bg-green-600 text-white rounded text-sm hover:bg-green-700 shrink-0"
                >
                  コピー
                </button>
              </div>
            </div>
          )}

          <div className="border rounded-md divide-y">
            {users.length === 0 ? (
              <p className="px-4 py-3 text-sm text-gray-500">
                ユーザーがまだいません
              </p>
            ) : (
              users.map((user) => (
                <div key={user.id} className="px-4 py-3 flex justify-between items-center">
                  <div>
                    <span className="text-sm font-medium text-gray-900">
                      {user.name}
                    </span>
                    <span className="ml-2 text-xs px-2 py-0.5 rounded-full bg-gray-100 text-gray-600">
                      {user.role}
                    </span>
                  </div>
                  <span className="text-xs text-gray-400">{user.id}</span>
                </div>
              ))
            )}
          </div>
        </section>
      </main>
    </div>
  );
}
