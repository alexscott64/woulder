
interface RouteTypeFilterProps {
  selectedTypes: string[];
  onChange: (types: string[]) => void;
}

const ROUTE_TYPES = [
  { value: 'Boulder', label: 'Boulder', emoji: '🪨' },
  { value: 'Sport', label: 'Sport', emoji: '🧗' },
  { value: 'Trad', label: 'Trad', emoji: '⚙️' },
  { value: 'Ice', label: 'Ice', emoji: '🧊' },
];

export function RouteTypeFilter({ selectedTypes, onChange }: RouteTypeFilterProps) {
  const toggleType = (type: string) => {
    if (selectedTypes.includes(type)) {
      onChange(selectedTypes.filter(t => t !== type));
    } else {
      onChange([...selectedTypes, type]);
    }
  };

  const selectAll = () => {
    onChange(ROUTE_TYPES.map(rt => rt.value));
  };

  const clearAll = () => {
    onChange([]);
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
          Route Types
        </label>
        <div className="flex gap-2">
          <button
            onClick={selectAll}
            className="text-xs text-blue-600 dark:text-blue-400 hover:underline"
          >
            All
          </button>
          <button
            onClick={clearAll}
            className="text-xs text-gray-600 dark:text-gray-400 hover:underline"
          >
            Clear
          </button>
        </div>
      </div>
      <div className="flex gap-1.5 sm:gap-2 overflow-x-auto pb-1">
        {ROUTE_TYPES.map((routeType) => {
          const isSelected = selectedTypes.includes(routeType.value);
          return (
            <button
              key={routeType.value}
              onClick={() => toggleType(routeType.value)}
              className={`
                flex-shrink-0 px-2 sm:px-3 py-1.5 rounded-lg text-xs sm:text-sm font-medium transition-all whitespace-nowrap
                ${
                  isSelected
                    ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300 border-2 border-blue-500'
                    : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 border-2 border-transparent hover:border-gray-400'
                }
              `}
            >
              <span className="mr-1">{routeType.emoji}</span>
              {routeType.label}
            </button>
          );
        })}
      </div>
    </div>
  );
}
