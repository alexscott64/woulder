import { useState, useCallback, useRef, useEffect } from 'react';
import { ChevronDown, ChevronUp, RotateCcw } from 'lucide-react';
import {
  getGradeScalesForTypes,
  type GradeScale,
  type GradeRangeSelection,
} from '../../utils/grades';

interface GradeRangeFilterProps {
  selectedTypes: string[];
  selections: GradeRangeSelection;
  onChange: (selections: GradeRangeSelection) => void;
}

export function GradeRangeFilter({ selectedTypes, selections, onChange }: GradeRangeFilterProps) {
  const [collapsed, setCollapsed] = useState(false);
  const scales = getGradeScalesForTypes(selectedTypes);

  if (scales.length === 0) return null;

  // Check if any scale has a non-default range
  const hasActiveFilter = scales.some((scale) => {
    const sel = selections[scale.key];
    return sel && (sel[0] !== 0 || sel[1] !== scale.grades.length - 1);
  });

  const handleReset = () => {
    const reset: GradeRangeSelection = {};
    for (const scale of scales) {
      reset[scale.key] = [0, scale.grades.length - 1];
    }
    onChange(reset);
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="flex items-center gap-1.5 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white transition-colors"
        >
          Grade Range
          {collapsed ? (
            <ChevronDown className="w-4 h-4" />
          ) : (
            <ChevronUp className="w-4 h-4" />
          )}
          {hasActiveFilter && collapsed && (
            <span className="ml-1 w-2 h-2 rounded-full bg-blue-500 inline-block" />
          )}
        </button>
        {hasActiveFilter && !collapsed && (
          <button
            onClick={handleReset}
            className="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 hover:underline"
          >
            <RotateCcw className="w-3 h-3" />
            Reset
          </button>
        )}
      </div>

      {!collapsed && (
        <div className="space-y-3">
          {scales.map((scale) => (
            <ScaleRangeSlider
              key={scale.key}
              scale={scale}
              value={selections[scale.key] ?? [0, scale.grades.length - 1]}
              onChange={(range) => {
                onChange({ ...selections, [scale.key]: range });
              }}
            />
          ))}
        </div>
      )}
    </div>
  );
}

// --- Dual-range slider for a single grade scale ---

interface ScaleRangeSliderProps {
  scale: GradeScale;
  value: [number, number];
  onChange: (range: [number, number]) => void;
}

function ScaleRangeSlider({ scale, value, onChange }: ScaleRangeSliderProps) {
  const [minIdx, maxIdx] = value;
  const max = scale.grades.length - 1;
  const isFullRange = minIdx === 0 && maxIdx === max;
  const trackRef = useRef<HTMLDivElement>(null);
  const [dragging, setDragging] = useState<'min' | 'max' | null>(null);

  // Calculate percentages
  const minPct = max > 0 ? (minIdx / max) * 100 : 0;
  const maxPct = max > 0 ? (maxIdx / max) * 100 : 100;

  const getIndexFromPosition = useCallback(
    (clientX: number): number => {
      if (!trackRef.current || max === 0) return 0;
      const rect = trackRef.current.getBoundingClientRect();
      const pct = Math.max(0, Math.min(1, (clientX - rect.left) / rect.width));
      return Math.round(pct * max);
    },
    [max],
  );

  // Mouse/touch event handlers
  const handlePointerDown = useCallback(
    (handle: 'min' | 'max') => (e: React.PointerEvent) => {
      e.preventDefault();
      (e.target as HTMLElement).setPointerCapture(e.pointerId);
      setDragging(handle);
    },
    [],
  );

  const handlePointerMove = useCallback(
    (e: React.PointerEvent) => {
      if (!dragging) return;
      const idx = getIndexFromPosition(e.clientX);
      if (dragging === 'min') {
        onChange([Math.min(idx, maxIdx), maxIdx]);
      } else {
        onChange([minIdx, Math.max(idx, minIdx)]);
      }
    },
    [dragging, getIndexFromPosition, minIdx, maxIdx, onChange],
  );

  const handlePointerUp = useCallback(() => {
    setDragging(null);
  }, []);

  // Track click to jump the nearest handle
  const handleTrackClick = useCallback(
    (e: React.MouseEvent) => {
      const idx = getIndexFromPosition(e.clientX);
      const distToMin = Math.abs(idx - minIdx);
      const distToMax = Math.abs(idx - maxIdx);
      if (distToMin <= distToMax) {
        onChange([Math.min(idx, maxIdx), maxIdx]);
      } else {
        onChange([minIdx, Math.max(idx, minIdx)]);
      }
    },
    [getIndexFromPosition, minIdx, maxIdx, onChange],
  );

  // Keyboard support for handles
  const handleKeyDown = useCallback(
    (handle: 'min' | 'max') => (e: React.KeyboardEvent) => {
      let newIdx: number;
      if (handle === 'min') {
        if (e.key === 'ArrowLeft' || e.key === 'ArrowDown') {
          newIdx = Math.max(0, minIdx - 1);
          onChange([newIdx, maxIdx]);
          e.preventDefault();
        } else if (e.key === 'ArrowRight' || e.key === 'ArrowUp') {
          newIdx = Math.min(maxIdx, minIdx + 1);
          onChange([newIdx, maxIdx]);
          e.preventDefault();
        }
      } else {
        if (e.key === 'ArrowLeft' || e.key === 'ArrowDown') {
          newIdx = Math.max(minIdx, maxIdx - 1);
          onChange([minIdx, newIdx]);
          e.preventDefault();
        } else if (e.key === 'ArrowRight' || e.key === 'ArrowUp') {
          newIdx = Math.min(max, maxIdx + 1);
          onChange([minIdx, newIdx]);
          e.preventDefault();
        }
      }
    },
    [minIdx, maxIdx, max, onChange],
  );

  // Compute tick mark positions — show a subset for readability
  const ticks = useTickMarks(scale, max);

  return (
    <div className="px-1">
      {/* Header row */}
      <div className="flex items-center justify-between mb-1">
        <div className="flex items-center gap-1.5">
          <span className="text-xs">{scale.emoji}</span>
          <span className="text-xs font-medium text-gray-600 dark:text-gray-400">
            {scale.label}
          </span>
        </div>
        <span className="text-xs font-semibold text-gray-900 dark:text-white">
          {isFullRange ? 'All grades' : `${scale.grades[minIdx]} – ${scale.grades[maxIdx]}`}
        </span>
      </div>

      {/* Slider track */}
      <div
        ref={trackRef}
        className="relative h-8 flex items-center cursor-pointer select-none touch-none"
        onClick={handleTrackClick}
        onPointerMove={handlePointerMove}
        onPointerUp={handlePointerUp}
      >
        {/* Background track */}
        <div className="absolute inset-x-0 h-1.5 rounded-full bg-gray-200 dark:bg-gray-700" />

        {/* Active range */}
        <div
          className="absolute h-1.5 rounded-full bg-blue-500 dark:bg-blue-400"
          style={{ left: `${minPct}%`, right: `${100 - maxPct}%` }}
        />

        {/* Min handle */}
        <div
          role="slider"
          tabIndex={0}
          aria-label={`Minimum grade: ${scale.grades[minIdx]}`}
          aria-valuemin={0}
          aria-valuemax={max}
          aria-valuenow={minIdx}
          aria-valuetext={scale.grades[minIdx]}
          className={`absolute w-5 h-5 rounded-full border-2 border-blue-500 bg-white dark:bg-gray-900 shadow-md -translate-x-1/2 z-10 cursor-grab transition-shadow ${
            dragging === 'min' ? 'ring-2 ring-blue-300 dark:ring-blue-600 cursor-grabbing scale-110' : 'hover:ring-2 hover:ring-blue-200 dark:hover:ring-blue-700'
          }`}
          style={{ left: `${minPct}%` }}
          onPointerDown={handlePointerDown('min')}
          onKeyDown={handleKeyDown('min')}
        />

        {/* Max handle */}
        <div
          role="slider"
          tabIndex={0}
          aria-label={`Maximum grade: ${scale.grades[maxIdx]}`}
          aria-valuemin={0}
          aria-valuemax={max}
          aria-valuenow={maxIdx}
          aria-valuetext={scale.grades[maxIdx]}
          className={`absolute w-5 h-5 rounded-full border-2 border-blue-500 bg-white dark:bg-gray-900 shadow-md -translate-x-1/2 z-10 cursor-grab transition-shadow ${
            dragging === 'max' ? 'ring-2 ring-blue-300 dark:ring-blue-600 cursor-grabbing scale-110' : 'hover:ring-2 hover:ring-blue-200 dark:hover:ring-blue-700'
          }`}
          style={{ left: `${maxPct}%` }}
          onPointerDown={handlePointerDown('max')}
          onKeyDown={handleKeyDown('max')}
        />
      </div>

      {/* Tick labels below the slider */}
      <div className="relative h-4 text-[10px] text-gray-500 dark:text-gray-500 select-none">
        {ticks.map((tick) => (
          <span
            key={tick.index}
            className="absolute -translate-x-1/2 whitespace-nowrap"
            style={{ left: `${max > 0 ? (tick.index / max) * 100 : 0}%` }}
          >
            {tick.label}
          </span>
        ))}
      </div>
    </div>
  );
}

// Compute sensible tick positions for a grade scale
function useTickMarks(scale: GradeScale, max: number): { index: number; label: string }[] {
  const [containerWidth, setContainerWidth] = useState(300);

  useEffect(() => {
    // Approximate width — we don't need precision, just enough to decide density
    const handleResize = () => setContainerWidth(window.innerWidth);
    handleResize();
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  if (max === 0) return [{ index: 0, label: scale.grades[0] }];

  // Target roughly 40-60px per tick
  const maxTicks = Math.max(3, Math.floor(containerWidth / 55));
  const step = Math.max(1, Math.ceil(scale.grades.length / maxTicks));

  const ticks: { index: number; label: string }[] = [];
  for (let i = 0; i < scale.grades.length; i += step) {
    ticks.push({ index: i, label: scale.grades[i] });
  }

  // Always include the last grade
  const lastIdx = scale.grades.length - 1;
  if (ticks[ticks.length - 1].index !== lastIdx) {
    ticks.push({ index: lastIdx, label: scale.grades[lastIdx] });
  }

  return ticks;
}
