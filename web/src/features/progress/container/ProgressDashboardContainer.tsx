"use client";

import { useCallback, useEffect, useState } from "react";
import type { TeamProgress } from "../model/progress";
import { ProgressList } from "../element/ProgressList";
import { api } from "@/lib/api/client";

type LoadingState = "loading" | "error" | "success";

/**
 * Container component that fetches team progress data from the API
 * and delegates rendering to ProgressList.
 */
export function ProgressDashboardContainer() {
  const [teams, setTeams] = useState<TeamProgress[]>([]);
  const [state, setState] = useState<LoadingState>("loading");
  const [errorMessage, setErrorMessage] = useState("");

  const fetchProgress = useCallback(async () => {
    setState("loading");
    setErrorMessage("");

    try {
      const { data, error } = await api.GET("/api/v1/progress");
      if (error) {
        setErrorMessage(error.error || "Unknown error occurred");
        setState("error");
        return;
      }
      setTeams(data.data as TeamProgress[]);
      setState("success");
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Unknown error occurred";
      setErrorMessage(message);
      setState("error");
    }
  }, []);

  useEffect(() => {
    fetchProgress();
  }, [fetchProgress]);

  if (state === "loading") {
    return (
      <div className="flex items-center justify-center p-12">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
        <span className="ml-3 text-muted-foreground">読み込み中...</span>
      </div>
    );
  }

  if (state === "error") {
    return (
      <div className="rounded-lg border border-red-200 bg-red-50 p-6 text-center">
        <p className="text-red-700 font-medium mb-2">
          データの取得に失敗しました
        </p>
        <p className="text-sm text-red-600 mb-4">{errorMessage}</p>
        <button
          type="button"
          onClick={fetchProgress}
          className="rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
        >
          再試行
        </button>
      </div>
    );
  }

  return <ProgressList teams={teams} />;
}
