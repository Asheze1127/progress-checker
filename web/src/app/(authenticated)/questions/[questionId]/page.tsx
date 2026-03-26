/**
 * Question detail page placeholder.
 * Displays a specific question and its AI-generated answer.
 */
export default function QuestionDetailPage({
  params,
}: {
  params: Promise<{ questionId: string }>;
}) {
  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Question Detail</h1>
      <p className="text-muted-foreground">Question details and AI response will be displayed here.</p>
    </div>
  );
}
