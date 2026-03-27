"use client";

import { useCallback, useEffect, useState } from "react";
import type { TeamTimeline } from "../model/timeline";
import { Timeline } from "../element/Timeline";
import { api } from "@/lib/api/client";

interface TeamTimelineContainerProps {
  teamId: string;
}

/**
 * Container component that manages data fetching and state for the team timeline.
 * Handles loading, error, and empty states.
 */
export function TeamTimelineContainer({ teamId }: TeamTimelineContainerProps) {
  const [timeline, setTimeline] = useState<TeamTimeline | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadTimeline = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const { data, error: apiError } = await api.GET("/api/v1/progress", {
        params: { query: { team_id: teamId } },
      });
      if (apiError) {
        setError(apiError.error || "An unexpected error occurred");
        return;
      }
      setTimeline(data as unknown as TeamTimeline);
    } catch (fetchError) {
      const message = fetchError instanceof Error ? fetchError.message : "An unexpected error occurred";
      setError(message);
    } finally {
      setIsLoading(false);
    }
  }, [teamId]);

  useEffect(() => {
    loadTimeline();
  }, [loadTimeline]);

  return (
    <div>
      {/* Header with back navigation */}
      <div className="mb-6 flex items-center gap-4">
        <a
          href="/progress"
          className="inline-flex items-center gap-1 rounded-md px-3 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
        >
          <span aria-hidden="true">&larr;</span>
          Back
        </a>
        {timeline && <h1 className="text-2xl font-bold">{timeline.team_name}</h1>}
      </div>

      {/* Loading state */}
      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-muted border-t-primary" />
        </div>
      )}

      {/* Error state */}
      {error && !isLoading && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-6 text-center">
          <p className="text-sm text-destructive">{error}</p>
          <button
            type="button"
            onClick={loadTimeline}
            className="mt-3 rounded-md bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            Retry
          </button>
        </div>
      )}

      {/* Timeline content */}
      {!isLoading && !error && timeline && <Timeline entries={timeline.entries} />}
    </div>
  );
}
