"use client";

import { useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api/client";

interface Participant {
  id: string;
  slack_user_id: string;
  name: string;
  email: string;
}

interface SlackUser {
  id: string;
  name: string;
  email: string;
}

export default function TeamDetailPage() {
  const { teamId } = useParams<{ teamId: string }>();
  const [teamName, setTeamName] = useState("");
  const [participants, setParticipants] = useState<Participant[]>([]);
  const [slackUsers, setSlackUsers] = useState<SlackUser[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingUsers, setIsLoadingUsers] = useState(true);
  const [showAddForm, setShowAddForm] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [registeringId, setRegisteringId] = useState<string | null>(null);

  const fetchParticipants = useCallback(async () => {
    try {
      const { data } = await api.GET("/api/v1/teams/{teamId}/participants", {
        params: { path: { teamId } },
      });
      if (data) {
        setParticipants(data.participants);
      }
    } catch {
      // fallback
    } finally {
      setIsLoading(false);
    }
  }, [teamId]);

  const fetchTeamName = useCallback(async () => {
    const { data } = await api.GET("/api/v1/teams");
    if (data) {
      const team = data.teams.find((t) => t.id === teamId);
      if (team) {
        setTeamName(team.name);
      }
    }
  }, [teamId]);

  const fetchSlackUsers = useCallback(async () => {
    setIsLoadingUsers(true);
    try {
      const { data } = await api.GET("/api/v1/slack/users");
      if (data) {
        setSlackUsers(data.users);
      }
    } catch {
      // fallback
    } finally {
      setIsLoadingUsers(false);
    }
  }, []);

  useEffect(() => {
    fetchParticipants();
    fetchTeamName();
  }, [fetchParticipants, fetchTeamName]);

  useEffect(() => {
    if (showAddForm) {
      fetchSlackUsers();
    }
  }, [showAddForm, fetchSlackUsers]);

  const handleRegister = async (user: SlackUser) => {
    setError(null);
    setSuccess(null);
    setRegisteringId(user.id);

    try {
      const { data, error: apiError } = await api.POST("/api/v1/participants", {
        body: { slack_user_id: user.id, team_id: teamId },
      });

      if (apiError) {
        setError(apiError.error || "登録に失敗しました");
        return;
      }

      setSuccess(`${data.user.name} を登録しました`);
      await fetchParticipants();
    } catch {
      setError("登録に失敗しました");
    } finally {
      setRegisteringId(null);
    }
  };

  const registeredSlackIds = new Set(participants.map((p) => p.slack_user_id));

  const filteredUsers = slackUsers.filter((u) => {
    if (registeredSlackIds.has(u.id)) return false;
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return u.name.toLowerCase().includes(q) || u.email.toLowerCase().includes(q);
  });

  return (
    <div className="max-w-3xl space-y-6">
      <div>
        <Link
          href="/teams"
          className="text-sm text-indigo-600 hover:text-indigo-800"
        >
          &larr; チーム一覧に戻る
        </Link>
        <h1 className="text-2xl font-bold text-gray-900 mt-2">
          {teamName || "チーム"}のメンバー
        </h1>
      </div>

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

      {/* 現在のメンバー一覧 */}
      <section className="bg-white rounded-lg shadow border">
        <div className="px-4 py-3 border-b flex items-center justify-between">
          <h2 className="text-sm font-semibold text-gray-900">
            メンバー一覧 ({participants.length})
          </h2>
          <button
            type="button"
            onClick={() => setShowAddForm(!showAddForm)}
            className="px-3 py-1.5 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-xs font-medium"
          >
            {showAddForm ? "閉じる" : "メンバーを追加"}
          </button>
        </div>

        {isLoading ? (
          <div className="px-4 py-8 text-center text-gray-500">
            読み込み中...
          </div>
        ) : participants.length === 0 ? (
          <div className="px-4 py-8 text-center text-gray-500">
            メンバーがまだいません
          </div>
        ) : (
          <div className="divide-y">
            {participants.map((p) => (
              <div key={p.id} className="px-4 py-3">
                <p className="text-sm font-medium text-gray-900">{p.name}</p>
                <p className="text-xs text-gray-500">{p.email}</p>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* Slackユーザーから追加 */}
      {showAddForm && (
        <section className="bg-white rounded-lg shadow border">
          <div className="px-4 py-3 border-b">
            <h2 className="text-sm font-semibold text-gray-900">
              Slackユーザーから追加
            </h2>
            <input
              type="text"
              placeholder="名前・メールで検索..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="mt-2 w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
            />
          </div>

          {isLoadingUsers ? (
            <div className="px-4 py-8 text-center text-gray-500">
              Slackユーザーを読み込み中...
            </div>
          ) : filteredUsers.length === 0 ? (
            <div className="px-4 py-8 text-center text-gray-500">
              追加可能なユーザーがいません
            </div>
          ) : (
            <div className="divide-y max-h-96 overflow-y-auto">
              {filteredUsers.map((user) => (
                <div
                  key={user.id}
                  className="px-4 py-3 flex items-center justify-between"
                >
                  <div>
                    <p className="text-sm font-medium text-gray-900">
                      {user.name}
                    </p>
                    <p className="text-xs text-gray-500">{user.email}</p>
                  </div>
                  <button
                    type="button"
                    onClick={() => handleRegister(user)}
                    disabled={registeringId === user.id}
                    className="px-3 py-1.5 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 text-xs font-medium"
                  >
                    {registeringId === user.id ? "登録中..." : "登録"}
                  </button>
                </div>
              ))}
            </div>
          )}
        </section>
      )}
    </div>
  );
}
