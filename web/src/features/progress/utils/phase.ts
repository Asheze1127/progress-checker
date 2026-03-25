import type { Phase } from "../model/progress";

/** Japanese display labels for each phase */
export const PHASE_LABELS: Record<Phase, string> = {
  idea: "アイデア",
  design: "設計",
  coding: "開発",
  testing: "テスト",
  demo: "デモ",
} as const;

/** Tailwind CSS class names for phase badge background/text colors */
export const PHASE_COLORS: Record<Phase, string> = {
  idea: "bg-purple-100 text-purple-800",
  design: "bg-blue-100 text-blue-800",
  coding: "bg-green-100 text-green-800",
  testing: "bg-yellow-100 text-yellow-800",
  demo: "bg-pink-100 text-pink-800",
} as const;

/** Emoji icons representing each phase */
export const PHASE_ICONS: Record<Phase, string> = {
  idea: "\u{1F4A1}",    // light bulb
  design: "\u{1F3A8}",  // palette
  coding: "\u{1F4BB}",  // laptop
  testing: "\u{1F9EA}", // test tube
  demo: "\u{1F3AC}",    // clapper board
} as const;

/** Get the display label for a phase */
export function getPhaseLabel(phase: Phase): string {
  return PHASE_LABELS[phase];
}

/** Get the Tailwind color classes for a phase badge */
export function getPhaseColor(phase: Phase): string {
  return PHASE_COLORS[phase];
}

/** Get the icon for a phase */
export function getPhaseIcon(phase: Phase): string {
  return PHASE_ICONS[phase];
}
