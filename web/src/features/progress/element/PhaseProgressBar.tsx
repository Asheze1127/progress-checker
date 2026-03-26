import { cn } from "@/lib/utils/cn";
import { type Phase, PHASES } from "../model/timeline";

/**
 * Phase label mapping for display purposes.
 */
const PHASE_LABELS: Record<Phase, string> = {
  idea: "Idea",
  design: "Design",
  coding: "Coding",
  testing: "Testing",
  demo: "Demo",
};

interface PhaseProgressBarProps {
  currentPhase: Phase;
}

/**
 * Visual indicator showing progress through the hackathon phases.
 * Highlights the current phase and marks completed phases.
 */
export function PhaseProgressBar({ currentPhase }: PhaseProgressBarProps) {
  const currentIndex = PHASES.indexOf(currentPhase);

  return (
    <div className="flex items-center gap-1">
      {PHASES.map((phase, index) => {
        const isCompleted = index < currentIndex;
        const isCurrent = index === currentIndex;

        return (
          <div key={phase} className="flex items-center gap-1">
            {index > 0 && (
              <div
                className={cn("h-0.5 w-3", isCompleted ? "bg-primary" : "bg-border")}
                aria-hidden="true"
              />
            )}
            <span
              className={cn(
                "rounded-full px-2 py-0.5 text-xs font-medium transition-colors",
                isCurrent && "bg-primary text-primary-foreground",
                isCompleted && "bg-primary/20 text-primary",
                !isCurrent && !isCompleted && "bg-muted text-muted-foreground",
              )}
            >
              {PHASE_LABELS[phase]}
            </span>
          </div>
        );
      })}
    </div>
  );
}
