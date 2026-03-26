/**
 * Type definitions for the team progress timeline feature.
 * Shared types (Phase, ProgressBody, PHASES) are defined in progress.ts.
 */

export type { Phase, ProgressBody } from "./progress";
export { PHASES } from "./progress";

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
