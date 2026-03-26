/**
 * Team detail page placeholder.
 * Displays detailed progress for a specific team.
 */
export default function TeamDetailPage({
  params,
}: {
  params: Promise<{ teamId: string }>;
}) {
  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Team Detail</h1>
      <p className="text-muted-foreground">Team progress details will be displayed here.</p>
    </div>
  );
}
