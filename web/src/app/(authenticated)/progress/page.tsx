import { ProgressDashboardContainer } from "@/features/progress/container/ProgressDashboardContainer";

export default function ProgressPage() {
  return (
    <div className="container mx-auto p-6">
      <h1 className="text-2xl font-bold mb-6">進捗一覧</h1>
      <ProgressDashboardContainer />
    </div>
  );
}
