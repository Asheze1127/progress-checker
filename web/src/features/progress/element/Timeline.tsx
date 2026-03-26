import type { TimelineEntry } from "../model/timeline";
import { TimelineItem } from "./TimelineItem";

interface TimelineProps {
  entries: TimelineEntry[];
}

/**
 * Formats an ISO datetime string to a localized date label (e.g. "2026/03/25").
 */
function formatDateLabel(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleDateString("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  });
}

/**
 * Groups timeline entries by date for display with date separators.
 */
function groupEntriesByDate(entries: TimelineEntry[]): Map<string, TimelineEntry[]> {
  const groups = new Map<string, TimelineEntry[]>();

  for (const entry of entries) {
    const dateKey = formatDateLabel(entry.created_at);
    const existing = groups.get(dateKey);
    if (existing) {
      existing.push(entry);
    } else {
      groups.set(dateKey, [entry]);
    }
  }

  return groups;
}

const EMPTY_STATE_MESSAGE = "このチームの進捗データはまだありません";

/**
 * Chronological timeline display of progress entries, sorted newest first.
 * Groups entries by date with visual separators.
 */
export function Timeline({ entries }: TimelineProps) {
  if (entries.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <p className="text-muted-foreground">{EMPTY_STATE_MESSAGE}</p>
      </div>
    );
  }

  // Sort entries newest first
  const sortedEntries = [...entries].sort(
    (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
  );

  const groupedEntries = groupEntriesByDate(sortedEntries);
  const flatEntries = sortedEntries;

  return (
    <div className="space-y-2">
      {Array.from(groupedEntries.entries()).map(([dateLabel, groupEntries]) => (
        <div key={dateLabel}>
          {/* Date separator */}
          <div className="sticky top-0 z-10 mb-4 flex items-center gap-3 bg-background py-2">
            <div className="h-px flex-1 bg-border" />
            <span className="shrink-0 rounded-full bg-muted px-3 py-1 text-xs font-medium text-muted-foreground">
              {dateLabel}
            </span>
            <div className="h-px flex-1 bg-border" />
          </div>

          {/* Timeline entries for this date */}
          {groupEntries.map((entry) => (
            <TimelineItem
              key={entry.id}
              entry={entry}
              isLast={entry.id === flatEntries[flatEntries.length - 1]?.id}
            />
          ))}
        </div>
      ))}
    </div>
  );
}
