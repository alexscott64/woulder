import type { ReactNode } from 'react';

/**
 * Variant kinds for the adaptive 3rd-column tile rendered by WeatherCard.
 *
 * - `hot`  : Rock surface is hot enough to warrant a friction warning.
 * - `rock` : Normal rock surface temperature reading.
 * - `snow` : Snow on the ground.
 * - `wet`  : Wet / damp rock surface.
 * - `unknown` : No data available (renders as a muted placeholder).
 *
 * `hot` and `rock` tiles are clickable (rendered as `<button>`) and deep-link
 * into the ConditionsModal's "Surface temperature & friction" section. The
 * other variants are passive (`<div>`).
 */
export type AdaptiveTileVariant = 'hot' | 'rock' | 'snow' | 'wet' | 'unknown';

export interface AdaptiveMetricTileProps {
  /** Discriminator that decides whether the tile is interactive. */
  kind: AdaptiveTileVariant;
  /** Resolved icon element (parent decides the icon + its styling). */
  icon: ReactNode;
  /** Small label above the value (e.g. "Rock Temp", "On Ground"). */
  label: string;
  /** Primary value text or formatted JSX. */
  value: ReactNode;
  /** Optional secondary line shown beneath the value. */
  sub?: ReactNode;
  /** Optional Tailwind color class applied to the value text. */
  valueClassName?: string;
  /** Click handler — only invoked for `hot` / `rock` variants. */
  onClick?: () => void;
  /** Hover tooltip (only meaningful when the tile is a button). */
  title?: string;
  /** Accessible label (only meaningful when the tile is a button). */
  ariaLabel?: string;
}

function isClickableKind(kind: AdaptiveTileVariant): boolean {
  return kind === 'hot' || kind === 'rock';
}

/**
 * Adaptive metric tile used as the 3rd column of WeatherCard's metrics grid.
 *
 * Renders the same visual layout for every variant; only the wrapper element
 * differs: `hot` / `rock` variants render as a `<button>` with hover and
 * focus affordances so users can drill into the conditions modal; all other
 * variants render as a plain `<div>`.
 */
export function AdaptiveMetricTile({
  kind,
  icon,
  label,
  value,
  sub,
  valueClassName = 'text-gray-900 dark:text-white',
  onClick,
  title,
  ariaLabel,
}: AdaptiveMetricTileProps) {
  const content = (
    <>
      {icon}
      <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">{label}</div>
      <div className={`text-sm font-semibold ${valueClassName}`}>{value}</div>
      {sub && (
        <div className="text-[10px] text-gray-500 dark:text-gray-400 mt-0.5 leading-tight">
          {sub}
        </div>
      )}
    </>
  );

  if (isClickableKind(kind)) {
    return (
      <button
        type="button"
        onClick={onClick}
        aria-label={ariaLabel}
        title={title}
        className="flex flex-col items-center text-center rounded-md p-1 -m-1 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700/60 focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-400 dark:focus-visible:ring-blue-500 transition-colors"
      >
        {content}
      </button>
    );
  }

  return <div className="flex flex-col items-center text-center">{content}</div>;
}
