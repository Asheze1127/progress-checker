import { cn } from "@/lib/utils/cn";
import type { TimelineEntry } from "../model/timeline";
import { PhaseProgressBar } from "./PhaseProgressBar";

interface TimelineItemProps {
  entry: TimelineEntry;
  /** Whether this is the last item (hides the connecting line). */
  isLast: boolean;
}

/**
 * Formats an ISO datetime string to a localized time string (HH:MM).
 */
function formatTime(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleTimeString("ja-JP", {
    hour: "2-digit",
    minute: "2-digit",
  });
}

/**
 * Single entry in the timeline, showing participant info, phase, SOS status, and comment.
 * Connected to adjacent items by a vertical line.
 */
export function TimelineItem({ entry, isLast }: TimelineItemProps) {
  const latestBody = entry.progress_bodies[entry.progress_bodies.length - 1];

  if (!latestBody) {
    return null;
  }

  return (
    <div className="flex gap-4">
      {/* Left column: timestamp and vertical connector */}
      <div className="flex flex-col items-center">
        <span className="w-16 shrink-0 text-xs text-muted-foreground text-right pt-1">
          {formatTime(entry.created_at)}
        </span>
      </div>

      {/* Timeline dot and line */}
      <div className="flex flex-col items-center">
        <div
          className={cn(
            "mt-2 h-3 w-3 shrink-0 rounded-full border-2",
            latestBody.sos ? "border-destructive bg-destructive/20" : "border-primary bg-primary/20",
          )}
        />
        {!isLast && <div className="w-0.5 grow bg-border" />}
      </div>

      {/* Right column: content card */}
      <div className={cn("mb-6 flex-1 rounded-lg border p-4", latestBody.sos && "border-destructive/50")}>
        <div className="mb-2 flex flex-wrap items-center gap-2">
          <span className="font-medium text-sm">{entry.participant_name}</span>
          {latestBody.sos && (
            <span className="rounded-md bg-destructive/10 px-2 py-0.5 text-xs font-semibold text-destructive">
              SOS
            </span>
          )}
        </div>

        <div className="mb-3">
          <PhaseProgressBar currentPhase={latestBody.phase} />
        </div>

        {latestBody.comment && <p className="text-sm text-foreground/80 whitespace-pre-wrap">{latestBody.comment}</p>}

        {/* Show all progress bodies if there are multiple */}
        {entry.progress_bodies.length > 1 && (
          <details className="mt-3">
            <summary className="cursor-pointer text-xs text-muted-foreground hover:text-foreground">
              {entry.progress_bodies.length} updates in this entry
            </summary>
            <div className="mt-2 space-y-2 border-l-2 border-muted pl-3">
              {entry.progress_bodies.map((body) => (
                <div key={body.submitted_at} className="text-xs">
                  <span className="font-medium capitalize">{body.phase}</span>
                  {body.sos && <span className="ml-1 text-destructive font-semibold">SOS</span>}
                  {body.comment && <p className="mt-0.5 text-muted-foreground">{body.comment}</p>}
                  <span className="text-muted-foreground">{formatTime(body.submitted_at)}</span>
                </div>
              ))}
            </div>
          </details>
        )}
      </div>
    </div>
  );
}
