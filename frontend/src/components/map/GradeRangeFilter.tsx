import { useState, useMemo } from 'react';
import { ChevronDown } from 'lucide-react';
import {
  getGradeScalesForTypes,
  type GradeRangeSelection,
} from '../../utils/grades';

interface GradeRangeFilterProps {
  selectedTypes: string[];
  selections: GradeRangeSelection;
  onChange: (selections: GradeRangeSelection) => void;
  /** Hint from parent to auto-switch to a specific tab (e.g. when a route type is toggled on) */
  activeTabHint?: string | null;
}

export function GradeRangeFilter({ selectedTypes, selections, onChange, activeTabHint }: GradeRangeFilterProps) {
  const scales = getGradeScalesForTypes(selectedTypes);
  const [activeTab, setActiveTab] = useState<string | null>(null);

  // When parent sends a tab hint, switch to it
  const [lastHint, setLastHint] = useState<string | null>(null);
  if (activeTabHint && activeTabHint !== lastHint) {
    setLastHint(activeTabHint);
    if (scales.some((s) => s.key === activeTabHint)) {
      setActiveTab(activeTabHint);
    }
  }

  // Auto-select first tab if current tab doesn't exist in scales
  const resolvedTab = useMemo(() => {
    if (scales.length === 0) return null;
    if (activeTab && scales.some((s) => s.key === activeTab)) return activeTab;
    return scales[0].key;
  }, [scales, activeTab]);

  if (scales.length === 0) return null;

  const activeScale = scales.find((s) => s.key === resolvedTab) ?? scales[0];
  const sel = selections[activeScale.key] ?? [0, activeScale.grades.length - 1];
  const [minIdx, maxIdx] = sel;
  const isFullRange = minIdx === 0 && maxIdx === activeScale.grades.length - 1;

  // Check if ANY scale has a non-default range (for the active indicator)
  const hasActiveFilter = scales.some((scale) => {
    const s = selections[scale.key];
    return s && (s[0] !== 0 || s[1] !== scale.grades.length - 1);
  });

  const handleMinChange = (newMin: number) => {
    onChange({
      ...selections,
      [activeScale.key]: [Math.min(newMin, maxIdx), maxIdx],
    });
  };

  const handleMaxChange = (newMax: number) => {
    onChange({
      ...selections,
      [activeScale.key]: [minIdx, Math.max(newMax, minIdx)],
    });
  };

  const handleReset = () => {
    onChange({
      ...selections,
      [activeScale.key]: [0, activeScale.grades.length - 1],
    });
  };

  return (
    <div className="space-y-1.5">
      {/* Row 1: Label + tabs */}
      <div className="flex items-center gap-2 flex-wrap">
        <label className="text-sm font-medium text-gray-700 dark:text-gray-300 flex-shrink-0">
          Grades
          {hasActiveFilter && (
            <span className="ml-1.5 w-1.5 h-1.5 rounded-full bg-blue-500 inline-block align-middle" />
          )}
        </label>

        {/* Tabs for each grade family */}
        {scales.length > 1 && (
          <div className="flex gap-1 flex-wrap">
            {scales.map((scale) => {
              const isActive = scale.key === resolvedTab;
              const scaleHasFilter = (() => {
                const s = selections[scale.key];
                return s && (s[0] !== 0 || s[1] !== scale.grades.length - 1);
              })();
              return (
                <button
                  key={scale.key}
                  onClick={() => setActiveTab(scale.key)}
                  className={`
                    px-2 py-0.5 rounded text-xs font-medium transition-colors relative
                    ${isActive
                      ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300'
                      : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                    }
                  `}
                >
                  <span className="mr-0.5">{scale.emoji}</span>
                  {scale.label}
                  {scaleHasFilter && !isActive && (
                    <span className="absolute -top-0.5 -right-0.5 w-1.5 h-1.5 rounded-full bg-blue-500" />
                  )}
                </button>
              );
            })}
          </div>
        )}

        {/* Single-tab label when only one scale */}
        {scales.length === 1 && (
          <span className="text-xs text-gray-500 dark:text-gray-400">
            <span className="mr-0.5">{activeScale.emoji}</span>
            {activeScale.label}
          </span>
        )}
      </div>

      {/* Row 2: Min / Max dropdowns */}
      <div className="flex items-center gap-1.5 pl-0 sm:pl-0">
        <GradeSelect
          grades={activeScale.grades}
          value={minIdx}
          onChange={handleMinChange}
          label="Min grade"
        />
        <span className="text-xs text-gray-400 dark:text-gray-500">–</span>
        <GradeSelect
          grades={activeScale.grades}
          value={maxIdx}
          onChange={handleMaxChange}
          label="Max grade"
        />
        {!isFullRange && (
          <button
            onClick={handleReset}
            className="text-xs text-blue-600 dark:text-blue-400 hover:underline ml-1 flex-shrink-0"
          >
            Reset
          </button>
        )}
      </div>
    </div>
  );
}

// Compact grade select dropdown
function GradeSelect({
  grades,
  value,
  onChange,
  label,
}: {
  grades: readonly string[];
  value: number;
  onChange: (idx: number) => void;
  label: string;
}) {
  return (
    <div className="relative">
      <select
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        aria-label={label}
        className="
          appearance-none bg-gray-100 dark:bg-gray-700 border border-gray-200 dark:border-gray-600
          rounded-md pl-2 pr-6 py-1 text-xs font-medium
          text-gray-900 dark:text-white
          focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500
          cursor-pointer min-w-[4.5rem]
        "
      >
        {grades.map((grade, idx) => (
          <option key={grade} value={idx}>
            {grade}
          </option>
        ))}
      </select>
      <ChevronDown className="absolute right-1.5 top-1/2 -translate-y-1/2 w-3 h-3 text-gray-400 dark:text-gray-500 pointer-events-none" />
    </div>
  );
}
