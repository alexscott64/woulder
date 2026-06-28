export type AreaViewportFitArgs = {
  previousFocusBBoxKey: string | null;
  nextFocusBBoxKey: string;
  editing: boolean;
  hasViewportInteraction: boolean;
};

export function shouldFitAreaViewport({ previousFocusBBoxKey, nextFocusBBoxKey, editing, hasViewportInteraction }: AreaViewportFitArgs): boolean {
  if (editing) return false;
  if (previousFocusBBoxKey === nextFocusBBoxKey) return false;
  return !hasViewportInteraction;
}
