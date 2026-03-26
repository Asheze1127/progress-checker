import { TeamTimelineContainer } from "@/features/progress/container/TeamTimelineContainer";

/**
 * Team progress detail page (S-05).
 * Displays chronological progress posts for a specific team.
 */
export default async function TeamProgressPage({ params }: { params: Promise<{ teamId: string }> }) {
  const { teamId } = await params;

  return (
    <div className="container mx-auto p-6">
      <TeamTimelineContainer teamId={teamId} />
    </div>
  );
}
