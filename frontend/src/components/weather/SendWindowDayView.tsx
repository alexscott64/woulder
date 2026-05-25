/**
 * SendWindowDayView — single-day surface-temperature & friction "send window"
 * view, plus the underlying reusable building blocks used by both the
 * multi-day grid in `ConditionsModal.RockTempTabContent` and the per-day
 * "Send Windows" section in `ConditionDetailsModal`.
 *
 * Components exported:
 *   - `SendWindowGanttRow` — one Gantt row (24h heatmap + send-window
 *     overlays) for a single day. Stateless / presentational.
 *   - `DayCard` — the compact per-day card showing the day's send-windows
 *     as a list (time range, duration, peak temp, "damp early" flag).
 *   - `SendWindowDayView` — wrapper that takes a `RockTemperatureStatus`
 *     and a target `date` (YYYY-MM-DD) and renders the single-day Gantt
 *     strip + DayCard for that date. Used by the Condition Details modal.
 *
 * Filtering rules used by the wrapper:
 *   - hourly samples whose local YYYY-MM-DD equals `date` populate the
 *     Gantt heatmap.
 *   - send_windows whose placement against `date` produces a non-zero
 *     width (i.e. they overlap the day) are kept. This matches the
 *     overlap semantics used by the 7-day grid.
 */
import { useMemo } from 'react';
import {
  CONDENSATION_OVERLAY_CLASS,
  ROCK_CONDITION_COLORS,
  ROCK_CONDITION_LABELS,
  computeWindowGanttPlacement,
  formatCompactDuration,
  formatCompactTimeRange,
  formatSendWindowDetail,
  formatTimeAxisLabel,
  formatWeekdayLong,
} from './weatherDisplay';
import type {
  DailyRockTemp,
  RockCondition,
  RockTempHour,
  RockTemperatureStatus,
  SendWindow,
} from '../../types/weather';

// ---------------------------------------------------------------------------
// Helpers (kept local so this module is self-contained).
// ---------------------------------------------------------------------------

/**
 * Pick the hourly sample for a given local hour-of-day from a day's hourly
 * array (already filtered to that local date). Linear scan — at most 24
 * entries per day.
 */
export function pickHourSample(
  dayHours: RockTempHour[],
  hour: number,
): RockTempHour | undefined {
  for (const h of dayHours) {
    const d = new Date(h.time);
    if (d.getHours() === hour) return h;
  }
  return undefined;
}

// Rank used to summarize best/worst conditions in the Gantt row tooltip.
const ROCK_CONDITION_RANK: Record<RockCondition, number> = {
  prime: 0,
  good: 1,
  marginal: 2,
  too_cold: 3,
  poor: 4,
  very_poor: 5,
};

function rankWorst(a: RockCondition, b: RockCondition): RockCondition {
  return ROCK_CONDITION_RANK[a] >= ROCK_CONDITION_RANK[b] ? a : b;
}
function rankBest(a: RockCondition, b: RockCondition): RockCondition {
  return ROCK_CONDITION_RANK[a] <= ROCK_CONDITION_RANK[b] ? a : b;
}

/** Build the row tooltip summarizing best/worst condition + windows. */
export function buildDaySummary(
  day: DailyRockTemp,
  hours: RockTempHour[],
  windows: SendWindow[],
): string {
  let best: RockCondition | undefined;
  let worst: RockCondition | undefined;
  for (const h of hours) {
    best = best ? rankBest(best, h.condition) : h.condition;
    worst = worst ? rankWorst(worst, h.condition) : h.condition;
  }
  if (!best) best = day.peak_condition;
  if (!worst) worst = day.overall_condition;

  const parts: string[] = [];
  parts.push(`${formatWeekdayLong(day.local_date)}`);
  if (best && worst) {
    parts.push(
      best === worst
        ? `${ROCK_CONDITION_LABELS[best]}`
        : `${ROCK_CONDITION_LABELS[best]} → ${ROCK_CONDITION_LABELS[worst]}`,
    );
  }
  if (windows.length === 0) {
    parts.push('no send windows');
  } else if (windows.length === 1) {
    parts.push('1 send window');
  } else {
    parts.push(`${windows.length} send windows`);
  }
  return parts.join(' • ');
}

// ---------------------------------------------------------------------------
// SendWindowGanttRow — one Gantt row (heatmap + window overlays).
// ---------------------------------------------------------------------------

export interface SendWindowGanttRowProps {
  /** Daily roll-up (provides local_date, fallback condition, etc.). */
  day: DailyRockTemp;
  /** Hourly samples already filtered to this day's local date. */
  hourlyForDay: RockTempHour[];
  /** Send-windows already filtered to overlap this day. */
  windowsForDay: SendWindow[];
  /** Optional left-column label. If null, the label column is omitted. */
  label?: string | null;
  /** Whether this row should render with the "today" highlight. */
  isToday?: boolean;
  /** Optional data-testid for the label cell (defaults to send-window-day-axis-label). */
  labelTestId?: string;
}

/**
 * Render one Gantt row: 24-hour condition heatmap with send-window outline
 * overlays. Pure presentational — no data fetching, no filtering.
 */
export function SendWindowGanttRow({
  day,
  hourlyForDay,
  windowsForDay,
  label,
  isToday = false,
  labelTestId = 'send-window-day-axis-label',
}: SendWindowGanttRowProps) {
  const summary = buildDaySummary(day, hourlyForDay, windowsForDay);
  const showLabel = label !== null;
  return (
    <div className="flex items-center gap-2" title={summary}>
      {showLabel && (
        <div
          data-testid={labelTestId}
          className={`w-10 sm:w-12 flex-shrink-0 text-[10px] sm:text-xs ${
            isToday
              ? 'font-bold text-blue-600 dark:text-blue-300'
              : 'text-gray-600 dark:text-gray-400'
          }`}
        >
          {label ?? (isToday ? 'Today' : formatWeekdayLong(day.local_date).slice(0, 3))}
        </div>
      )}
      <div
        data-testid="send-window-bar-row"
        className={`relative flex-1 h-3 sm:h-5 rounded overflow-hidden ${
          isToday
            ? 'bg-gray-200 dark:bg-gray-700 ring-1 ring-blue-400/40'
            : 'bg-gray-100 dark:bg-gray-800'
        }`}
      >
        {/* Layer 1: hourly heatmap (24 absolutely-positioned cells) */}
        {Array.from({ length: 24 }, (_, hour) => {
          const sample = pickHourSample(hourlyForDay, hour);
          const left = (hour / 24) * 100;
          const width = (1 / 24) * 100;
          if (!sample) {
            const fallback = day.overall_condition;
            if (fallback) {
              return (
                <span
                  key={hour}
                  className="absolute top-0 bottom-0 opacity-40"
                  style={{
                    left: `${left}%`,
                    width: `${width}%`,
                    backgroundColor: ROCK_CONDITION_COLORS[fallback],
                  }}
                />
              );
            }
            return (
              <span
                key={hour}
                className={`absolute top-0 bottom-0 bg-gray-300/40 dark:bg-gray-600/30 ${CONDENSATION_OVERLAY_CLASS}`}
                style={{ left: `${left}%`, width: `${width}%` }}
              />
            );
          }
          return (
            <span
              key={hour}
              className={`absolute top-0 bottom-0 ${sample.condensing ? CONDENSATION_OVERLAY_CLASS : ''}`}
              style={{
                left: `${left}%`,
                width: `${width}%`,
                backgroundColor: ROCK_CONDITION_COLORS[sample.condition],
              }}
            />
          );
        })}

        {/* 6h gridlines */}
        {[6, 12, 18].map((h) => (
          <span
            key={`grid-${h}`}
            className="absolute top-0 bottom-0 w-px bg-gray-700/30 dark:bg-gray-100/20"
            style={{ left: `${(h / 24) * 100}%` }}
            aria-hidden="true"
          />
        ))}

        {/* Layer 2: send-window outline overlays */}
        {windowsForDay.map((w, i) => {
          const { leftPercent, widthPercent } = computeWindowGanttPlacement(
            w,
            day.local_date,
          );
          if (widthPercent <= 0) return null;
          const tip = `${formatSendWindowDetail(w)}${w.dry_throughout ? '' : ' (may be damp early)'}`;
          return (
            <div
              key={`win-${i}`}
              data-testid="send-window-bar"
              title={tip}
              className="absolute top-0 bottom-0 rounded-sm border-2 pointer-events-none"
              style={{
                left: `${leftPercent}%`,
                width: `${widthPercent}%`,
                borderColor: ROCK_CONDITION_COLORS[w.condition],
                backgroundColor: 'transparent',
              }}
            />
          );
        })}
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// DayCard — compact send-window list for one day.
// ---------------------------------------------------------------------------

export interface DayCardProps {
  day: DailyRockTemp;
  windows: SendWindow[];
  isToday: boolean;
}

export function DayCard({ day, windows, isToday }: DayCardProps) {
  return (
    <div
      data-testid="send-window-daycard"
      className={`rounded-md border border-gray-200 dark:border-gray-700 p-3 text-xs min-h-[80px] ${
        isToday ? 'ring-2 ring-blue-400/50 bg-blue-50/40 dark:bg-blue-900/10' : 'bg-white dark:bg-gray-900/40'
      }`}
    >
      <div className="flex items-center justify-between mb-1.5">
        <span
          className={`text-sm font-semibold truncate ${
            isToday
              ? 'text-blue-700 dark:text-blue-300'
              : 'text-gray-700 dark:text-gray-300'
          }`}
        >
          {isToday ? 'Today' : formatWeekdayLong(day.local_date).slice(0, 3)}
        </span>
        <span
          className="inline-block w-2.5 h-2.5 rounded-full flex-shrink-0"
          style={{ backgroundColor: ROCK_CONDITION_COLORS[day.overall_condition] }}
          title={ROCK_CONDITION_LABELS[day.overall_condition]}
        />
      </div>
      {windows.length === 0 ? (
        <p className="text-[11px] text-gray-400 dark:text-gray-500 italic">no windows</p>
      ) : (
        <ul className="space-y-1.5">
          {windows.map((w, i) => (
            <li key={i} className="leading-tight">
              <div className="flex items-center gap-1.5">
                <span
                  className="inline-block w-2 h-2 rounded-full flex-shrink-0"
                  style={{ backgroundColor: ROCK_CONDITION_COLORS[w.condition] }}
                />
                <span className="text-[13px] text-gray-700 dark:text-gray-300 truncate">
                  {formatCompactTimeRange(w.start_time, w.end_time)}
                  <span className="text-gray-400 dark:text-gray-500"> · </span>
                  {formatCompactDuration(w.duration_h)}
                </span>
              </div>
              <div className="text-[11px] text-gray-500 dark:text-gray-500 ml-3.5">
                peak {Math.round(w.peak_temp_f)}°F
                {!w.dry_throughout && <span className="ml-1">· damp early</span>}
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// SendWindowDayView — single-day wrapper used by ConditionDetailsModal.
// ---------------------------------------------------------------------------

export interface SendWindowDayViewProps {
  /** Source rock-temp status (provides hourly_forecast + send_windows + daily_forecast). */
  status: RockTemperatureStatus;
  /**
   * Target day in local YYYY-MM-DD form. The Gantt + DayCard are filtered
   * to this day.
   */
  date: string;
}

/** Format a YYYY-MM-DD as a local-date key (matches RockTempTabContent). */
function localDateKey(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

/**
 * Render the single-day Gantt strip + DayCard for a target date. Filters
 * the supplied `RockTemperatureStatus` arrays internally so callers stay
 * simple.
 */
export function SendWindowDayView({ status, date }: SendWindowDayViewProps) {
  // Locate (or synthesize) the day rollup for the target date.
  const day: DailyRockTemp = useMemo(() => {
    const found = status.daily_forecast?.find((d) => d.local_date === date);
    if (found) return found;
    // Fallback synthetic entry so the row still renders (rare: backend
    // didn't include the day in daily_forecast for some reason).
    return {
      local_date: date,
      peak_surface_temp_f: 0,
      min_surface_temp_f: 0,
      peak_condition: 'good',
      overall_condition: 'good',
      has_condensation: false,
      window_count: 0,
    };
  }, [status.daily_forecast, date]);

  // Hourly samples whose local date matches the target date.
  const hourlyForDay = useMemo(() => {
    const out: RockTempHour[] = [];
    for (const h of status.hourly_forecast ?? []) {
      const d = new Date(h.time);
      if (isNaN(d.getTime())) continue;
      if (localDateKey(d) === date) out.push(h);
    }
    return out;
  }, [status.hourly_forecast, date]);

  // Send-windows that visually overlap the target day, mirroring the
  // overlap rule used by the multi-day grid (non-zero placement width).
  const windowsForDay = useMemo(() => {
    return (status.send_windows ?? []).filter((w) => {
      const p = computeWindowGanttPlacement(w, date);
      return p.widthPercent > 0;
    });
  }, [status.send_windows, date]);

  // Determine if `date` is "today" in the browser's local timezone — used
  // for the highlight styling in both row + card.
  const todayKey = useMemo(() => localDateKey(new Date()), []);
  const isToday = date === todayKey;

  return (
    <div data-testid="send-window-single-day" className="space-y-2">
      {/* Hour-axis legend: 12a / 6a / 12p / 6p / 12a aligned over the row track. */}
      <div className="flex items-center gap-2">
        <div className="w-10 sm:w-12 flex-shrink-0" aria-hidden="true" />
        <div className="relative flex-1 h-3">
          {[0, 6, 12, 18, 24].map((h) => (
            <span
              key={h}
              className="absolute top-0 text-[9px] sm:text-[10px] text-gray-500 dark:text-gray-400 -translate-x-1/2"
              style={{ left: `${(h / 24) * 100}%` }}
            >
              {formatTimeAxisLabel(h)}
            </span>
          ))}
        </div>
      </div>

      {/* Single Gantt row. */}
      <SendWindowGanttRow
        day={day}
        hourlyForDay={hourlyForDay}
        windowsForDay={windowsForDay}
        isToday={isToday}
      />

      {/* Single DayCard. In the multi-day grid these are constrained to
          ~140px columns, but with only one card here we let it span the
          full available width so it doesn't read like a tiny square
          floating in an oversized modal. */}
      <div className="pt-1">
        <DayCard day={day} windows={windowsForDay} isToday={isToday} />
      </div>
    </div>
  );
}
