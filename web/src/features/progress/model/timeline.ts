/**
 * Type definitions for the team progress timeline feature.
 */

export const PHASES = ["idea", "design", "coding", "testing", "demo"] as const;

export type Phase = (typeof PHASES)[number];

export interface ProgressBody {
  phase: Phase;
  sos: boolean;
  comment: string;
  submitted_at: string;
}

export interface TimelineEntry {
  id: string;
  participant_id: string;
  participant_name: string;
  progress_bodies: ProgressBody[];
  created_at: string;
}

export interface TeamTimeline {
  team_id: string;
  team_name: string;
  entries: TimelineEntry[];
}
