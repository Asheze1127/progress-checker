"use client";

import { useCallback, useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { fetchAPI } from "@/lib/fetcher/api";

interface Team {
  id: string;
  name: string;
}

interface UserResponse {
  id: string;
  name: string;
  role: string;
}

const registerSchema = z.object({
  slack_user_id: z.string().min(1, "Slack User IDを入力してください"),
  team_id: z.string().min(1, "チームを選択してください"),
});

type RegisterFormData = z.infer<typeof registerSchema>;

export default function MembersPage() {
  const [teams, setTeams] = useState<Team[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const form = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
  });

  const fetchTeams = useCallback(async () => {
    try {
      const data = await fetchAPI<{ data: Array<{ team_id: string; team_name: string }> }>("/api/v1/progress");
      const uniqueTeams = data.data.map((t) => ({
        id: t.team_id,
        name: t.team_name,
      }));
      setTeams(uniqueTeams);
    } catch {
      // fallback: no teams available
    }
  }, []);

  useEffect(() => {
    fetchTeams();
  }, [fetchTeams]);

  const handleRegister = async (data: RegisterFormData) => {
    setError(null);
    setSuccess(null);

    try {
      const result = await fetchAPI<{ user: UserResponse }>("/api/v1/participants", {
        method: "POST",
        body: JSON.stringify(data),
      });
      setSuccess(`${result.user.name} をチームに登録しました`);
      form.reset();
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("メンバー登録に失敗しました");
      }
    }
  };

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">
        メンバー登録
      </h1>
      <p className="text-sm text-gray-500 mb-6">
        Slackワークスペースのユーザーをチームのメンバー（participant）として登録します。
      </p>

      <form
        onSubmit={form.handleSubmit(handleRegister)}
        className="space-y-4 bg-white rounded-lg shadow p-6"
      >
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
            {error}
          </div>
        )}
        {success && (
          <div className="bg-green-50 border border-green-200 text-green-800 px-4 py-3 rounded">
            {success}
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Slack User ID
          </label>
          <input
            {...form.register("slack_user_id")}
            placeholder="U01ABCDEF"
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
          />
          {form.formState.errors.slack_user_id && (
            <p className="mt-1 text-sm text-red-600">
              {form.formState.errors.slack_user_id.message}
            </p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            チーム
          </label>
          <select
            {...form.register("team_id")}
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
          >
            <option value="">チームを選択</option>
            {teams.map((team) => (
              <option key={team.id} value={team.id}>
                {team.name}
              </option>
            ))}
          </select>
          {form.formState.errors.team_id && (
            <p className="mt-1 text-sm text-red-600">
              {form.formState.errors.team_id.message}
            </p>
          )}
        </div>

        <button
          type="submit"
          disabled={form.formState.isSubmitting}
          className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
        >
          {form.formState.isSubmitting ? "登録中..." : "メンバーを登録"}
        </button>
      </form>
    </div>
  );
}
