import type { TeamProgress } from "../model/progress";
import { ProgressCard } from "./ProgressCard";

interface ProgressListProps {
  teams: TeamProgress[];
}

/**
 * Responsive grid list of ProgressCards.
 * Displays an empty state message when no teams have progress data.
 */
export function ProgressList({ teams }: ProgressListProps) {
  if (teams.length === 0) {
    return (
      <div className="flex items-center justify-center rounded-lg border border-dashed p-12">
        <p className="text-muted-foreground">進捗データがありません</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {teams.map((team) => (
        <ProgressCard key={team.team_id} team={team} />
      ))}
    </div>
  );
}
