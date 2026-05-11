import type { RockCondition } from '../../types/weather';
import { ROCK_CONDITION_COLORS } from '../weather/weatherDisplay';

interface RockTempIconProps {
  /** Rock condition that drives the boulder fill color. If omitted, uses currentColor. */
  condition?: RockCondition;
  /** Pixel size for both width and height. Defaults to 16. */
  size?: number;
  /** Additional Tailwind / CSS classes for the wrapper SVG. */
  className?: string;
  /** Accessible title; defaults to "Rock surface temperature". */
  title?: string;
}

/**
 * Custom inline SVG icon that visually communicates "rock surface
 * temperature": a stylized boulder silhouette filled with the current
 * rock-condition color, topped with two faint horizontal "heat shimmer"
 * waves and a subtle upper-left highlight for dimensionality.
 *
 * Designed as a drop-in replacement for `lucide-react`'s `<Thermometer />`
 * — same 24x24 viewBox so existing `size`/`className` usage applies cleanly.
 *
 * When `condition` is omitted the boulder fill falls back to `currentColor`,
 * which is useful for headings where the icon should inherit text color.
 */
export function RockTempIcon({
  condition,
  size = 16,
  className,
  title = 'Rock surface temperature',
}: RockTempIconProps) {
  const fillColor = condition ? ROCK_CONDITION_COLORS[condition] : 'currentColor';

  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      width={size}
      height={size}
      className={className}
      role="img"
      aria-label={title}
    >
      <title>{title}</title>

      {/* Heat shimmer waves floating above the boulder */}
      <path
        d="M6 4 q1.5 -1.5 3 0 t3 0 t3 0"
        stroke="currentColor"
        strokeWidth="1"
        fill="none"
        opacity="0.55"
        strokeLinecap="round"
      />
      <path
        d="M7 7 q1.5 -1.5 3 0 t3 0 t3 0"
        stroke="currentColor"
        strokeWidth="1"
        fill="none"
        opacity="0.4"
        strokeLinecap="round"
      />

      {/*
        Boulder silhouette — intentionally asymmetric so it reads as a
        rock rather than a pebble or oval. Quadratic curves keep the
        shape soft at small sizes (12px+) while the stroke gives it
        a crisp outline against light/dark backgrounds.
      */}
      <path
        d="M3.5 19
           Q2.5 13.5 6 11.5
           Q7 8.5 11 9.5
           Q14 7.5 17 10.5
           Q21.5 11 20.5 16
           Q21.5 19.5 18 20
           L6 20
           Q2.5 20 3.5 19 Z"
        fill={fillColor}
        stroke="currentColor"
        strokeWidth="1"
        strokeLinejoin="round"
      />

      {/* Upper-left highlight stroke for dimension */}
      <path
        d="M5.5 14 Q6.5 11.5 9.5 11.5"
        stroke="white"
        strokeOpacity="0.35"
        strokeWidth="1.25"
        fill="none"
        strokeLinecap="round"
      />
    </svg>
  );
}
