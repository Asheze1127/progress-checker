/** Phase of a team's hackathon progress */
export type Phase = "idea" | "design" | "coding" | "testing" | "demo";

/** A single progress entry body submitted by a participant */
export interface ProgressBody {
  phase: Phase;
  sos: boolean;
  comment: string;
  submitted_at: string;
}

/** Latest progress information for a team */
export interface TeamProgress {
  team_id: string;
  team_name: string;
  latest_progress: {
    id: string;
    participant_id: string;
    progress_bodies: ProgressBody[];
  } | null;
}

/** Response shape from GET /api/v1/progress */
export interface ProgressListResponse {
  data: TeamProgress[];
}
