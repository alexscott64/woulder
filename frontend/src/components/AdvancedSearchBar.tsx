import React, { useState } from 'react';
import { Search, X, Filter, ChevronDown } from 'lucide-react';

interface SearchFilters {
  query: string;
  gradeMin?: string;
  gradeMax?: string;
  routeType?: 'boulder' | 'sport' | 'trad' | 'all';
  dryStatus?: 'dry' | 'drying' | 'wet' | 'all';
}

interface AdvancedSearchBarProps {
  onSearch: (filters: SearchFilters) => void;
  placeholder?: string;
  className?: string;
}

const boulderGrades = ['V0', 'V1', 'V2', 'V3', 'V4', 'V5', 'V6', 'V7', 'V8', 'V9', 'V10', 'V11', 'V12', 'V13', 'V14', 'V15', 'V16', 'V17'];

export const AdvancedSearchBar: React.FC<AdvancedSearchBarProps> = ({
  onSearch,
  placeholder = 'Search routes and areas...',
  className = '',
}) => {
  const [query, setQuery] = useState('');
  const [showFilters, setShowFilters] = useState(false);
  const [filters, setFilters] = useState<SearchFilters>({
    query: '',
    routeType: 'all',
    dryStatus: 'all',
  });

  const handleSearch = () => {
    onSearch({ ...filters, query });
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  };

  const clearSearch = () => {
    setQuery('');
    setFilters({
      query: '',
      routeType: 'all',
      dryStatus: 'all',
    });
    onSearch({
      query: '',
      routeType: 'all',
      dryStatus: 'all',
    });
  };

  const hasActiveFilters = filters.gradeMin || filters.gradeMax || filters.routeType !== 'all' || filters.dryStatus !== 'all';

  return (
    <div className={`space-y-3 ${className}`}>
      {/* Main search bar */}
      <div className="relative flex items-center gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder={placeholder}
            className="w-full pl-10 pr-10 py-2.5 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900 dark:text-white"
          />
          {query && (
            <button
              onClick={clearSearch}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
            >
              <X className="w-4 h-4" />
            </button>
          )}
        </div>

        <button
          onClick={() => setShowFilters(!showFilters)}
          className={`flex items-center gap-2 px-4 py-2.5 rounded-lg border transition-colors ${
            hasActiveFilters
              ? 'bg-blue-50 dark:bg-blue-900/20 border-blue-500 text-blue-600 dark:text-blue-400'
              : 'bg-white dark:bg-gray-800 border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'
          }`}
        >
          <Filter className="w-4 h-4" />
          <span className="hidden sm:inline">Filters</span>
          {hasActiveFilters && (
            <span className="w-2 h-2 bg-blue-500 rounded-full"></span>
          )}
          <ChevronDown className={`w-4 h-4 transition-transform ${showFilters ? 'rotate-180' : ''}`} />
        </button>

        <button
          onClick={handleSearch}
          className="px-6 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors"
        >
          Search
        </button>
      </div>

      {/* Advanced filters */}
      {showFilters && (
        <div className="bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg p-4 space-y-4 animate-in slide-in-from-top-2">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* Grade range */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Grade Range
              </label>
              <div className="flex items-center gap-2">
                <select
                  value={filters.gradeMin || ''}
                  onChange={(e) => setFilters({ ...filters, gradeMin: e.target.value || undefined })}
                  className="flex-1 px-3 py-1.5 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-sm text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">Min</option>
                  {boulderGrades.map((grade) => (
                    <option key={grade} value={grade}>{grade}</option>
                  ))}
                </select>
                <span className="text-gray-500">-</span>
                <select
                  value={filters.gradeMax || ''}
                  onChange={(e) => setFilters({ ...filters, gradeMax: e.target.value || undefined })}
                  className="flex-1 px-3 py-1.5 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-sm text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">Max</option>
                  {boulderGrades.map((grade) => (
                    <option key={grade} value={grade}>{grade}</option>
                  ))}
                </select>
              </div>
            </div>

            {/* Route type */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Route Type
              </label>
              <select
                value={filters.routeType}
                onChange={(e) => setFilters({ ...filters, routeType: e.target.value as any })}
                className="w-full px-3 py-1.5 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-sm text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="all">All Types</option>
                <option value="boulder">Boulder</option>
                <option value="sport">Sport</option>
                <option value="trad">Trad</option>
              </select>
            </div>

            {/* Dry status */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Dry Status
              </label>
              <select
                value={filters.dryStatus}
                onChange={(e) => setFilters({ ...filters, dryStatus: e.target.value as any })}
                className="w-full px-3 py-1.5 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-sm text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="all">All Conditions</option>
                <option value="dry">Dry Only</option>
                <option value="drying">Drying (&lt;24h)</option>
                <option value="wet">Wet (&gt;24h)</option>
              </select>
            </div>

            {/* Clear filters button */}
            <div className="flex items-end">
              <button
                onClick={() => {
                  setFilters({
                    query: '',
                    routeType: 'all',
                    dryStatus: 'all',
                  });
                }}
                disabled={!hasActiveFilters}
                className="w-full px-3 py-1.5 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Clear Filters
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
