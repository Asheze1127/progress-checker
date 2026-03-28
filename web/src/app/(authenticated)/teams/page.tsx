"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api/client";

interface Team {
  id: string;
  name: string;
}

export default function TeamsPage() {
  const [teams, setTeams] = useState<Team[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const fetchTeams = useCallback(async () => {
    try {
      const { data } = await api.GET("/api/v1/teams");
      if (data) {
        setTeams(data.teams.map((t) => ({ id: t.id, name: t.name })));
      }
    } catch {
      // fallback
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTeams();
  }, [fetchTeams]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-gray-500">読み込み中...</p>
      </div>
    );
  }

  return (
    <div className="max-w-3xl space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">チーム一覧</h1>
      <p className="text-sm text-gray-500">
        チームを選択してメンバーの確認・登録を行います。
      </p>

      <div className="bg-white rounded-lg shadow border divide-y">
        {teams.length === 0 ? (
          <div className="px-4 py-8 text-center text-gray-500">
            チームがありません
          </div>
        ) : (
          teams.map((team) => (
            <Link
              key={team.id}
              href={`/teams/${team.id}`}
              className="block px-4 py-4 hover:bg-gray-50 transition-colors"
            >
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-gray-900">
                  {team.name}
                </span>
                <span className="text-xs text-gray-400">&rarr;</span>
              </div>
            </Link>
          ))
        )}
      </div>
    </div>
  );
}
