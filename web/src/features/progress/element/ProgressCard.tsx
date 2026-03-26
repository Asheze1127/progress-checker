"use client";

import type { TeamProgress } from "../model/progress";
import { getPhaseColor, getPhaseIcon, getPhaseLabel } from "../utils/phase";

interface ProgressCardProps {
  team: TeamProgress;
}

/** Format an ISO timestamp to a localized Japanese date-time string */
function formatTimestamp(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleString("ja-JP", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

/**
 * Presentational card displaying a single team's latest progress.
 * Shows team name, current phase badge, SOS indicator, and latest comment.
 */
export function ProgressCard({ team }: ProgressCardProps) {
  const latestBody = team.latest_progress?.progress_bodies[0] ?? null;
  const hasSos = team.latest_progress?.progress_bodies.some((b) => b.sos) ?? false;

  return (
    <div className="rounded-lg border bg-card p-4 shadow-sm">
      {/* Header: team name + SOS indicator */}
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-lg font-semibold truncate">{team.team_name}</h3>
        {hasSos && (
          <span
            className="ml-2 flex items-center gap-1 rounded-full bg-red-100 px-2 py-0.5 text-xs font-bold text-red-700"
            role="alert"
            aria-label="SOS"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 20 20"
              fill="currentColor"
              className="h-4 w-4"
              aria-hidden="true"
            >
              <path
                fillRule="evenodd"
                d="M8.485 2.495c.673-1.167 2.357-1.167 3.03 0l6.28 10.875c.673 1.167-.168 2.625-1.516 2.625H3.72c-1.347 0-2.189-1.458-1.515-2.625L8.485 2.495zM10 6a.75.75 0 01.75.75v3.5a.75.75 0 01-1.5 0v-3.5A.75.75 0 0110 6zm0 9a1 1 0 100-2 1 1 0 000 2z"
                clipRule="evenodd"
              />
            </svg>
            SOS
          </span>
        )}
      </div>

      {/* Phase badge */}
      {latestBody ? (
        <div className="mb-2">
          <span
            className={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-sm font-medium ${getPhaseColor(latestBody.phase)}`}
          >
            <span aria-hidden="true">{getPhaseIcon(latestBody.phase)}</span>
            {getPhaseLabel(latestBody.phase)}
          </span>
        </div>
      ) : (
        <div className="mb-2">
          <span className="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-sm text-gray-500">
            未報告
          </span>
        </div>
      )}

      {/* Latest comment */}
      <p className="text-sm text-muted-foreground line-clamp-2 min-h-[2.5rem]">
        {latestBody?.comment || "コメントなし"}
      </p>

      {/* Timestamp */}
      {latestBody && (
        <p className="mt-2 text-xs text-muted-foreground">
          最終更新: {formatTimestamp(latestBody.submitted_at)}
        </p>
      )}
    </div>
  );
}
