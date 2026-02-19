import { X, Footprints, ExternalLink, Calendar, User, MapPin, List, FolderTree, Search, Mountain } from 'lucide-react';
import { format } from 'date-fns';
import { useState, useEffect } from 'react';
import { ClimbHistoryEntry, KayaAscentEntry } from '../types/weather';
import { formatDaysAgo } from '../utils/weather/formatters';
import { AreaDrillDownView } from './AreaDrillDownView';
import { climbActivityApi } from '../services/api';

interface RecentActivityModalProps {
  locationId: number;
  locationName: string;
  climbHistory: ClimbHistoryEntry[];
  onClose: () => void;
}

// Unified climb entry type for display
interface UnifiedClimbEntry {
  id: string; // Unique ID for React key
  route_name: string;
  route_rating: string;
  area_name: string;
  climbed_at: string;
  climbed_by: string;
  comment?: string;
  days_since_climb: number;
  source: 'mp' | 'kaya';
  mp_route_id?: number;
  mp_area_id?: number;
  kaya_climb_slug?: string;
}

type View = 'recent' | 'areas';

// Helper function to clean comments
function cleanComment(comment: string | undefined): string | null {
  if (!comment) return null;

  const normalized = comment.replace(/[\s\t]+/g, '');
  const patternsToFilter = [
    '·Send.',
    '·Attempt.',
    '·Flash.',
    'Send.',
    'Attempt.',
    'Flash.'
  ];

  for (const pattern of patternsToFilter) {
    const normalizedPattern = pattern.replace(/[\s\t]+/g, '');
    if (normalized === normalizedPattern || normalized === `&middot;${normalizedPattern}`) {
      return null;
    }
  }

  let cleaned = comment
    .replace(/^[\s·•]+/, '')
    .replace(/^&middot;[\s]+/, '')
    .trim();

  return cleaned || null;
}

// Convert MP climb history to unified format
function convertMPToUnified(mpClimb: ClimbHistoryEntry): UnifiedClimbEntry {
  return {
    id: `mp-${mpClimb.mp_route_id}-${mpClimb.climbed_at}`,
    route_name: mpClimb.route_name,
    route_rating: mpClimb.route_rating,
    area_name: mpClimb.area_name,
    climbed_at: mpClimb.climbed_at,
    climbed_by: mpClimb.climbed_by,
    comment: mpClimb.comment,
    days_since_climb: mpClimb.days_since_climb,
    source: 'mp',
    mp_route_id: mpClimb.mp_route_id,
    mp_area_id: mpClimb.mp_area_id,
  };
}

// Convert Kaya ascent to unified format
function convertKayaToUnified(kayaAscent: KayaAscentEntry): UnifiedClimbEntry {
  return {
    id: `kaya-${kayaAscent.kaya_ascent_id}`,
    route_name: kayaAscent.route_name,
    route_rating: kayaAscent.route_grade,
    area_name: kayaAscent.area_name,
    climbed_at: kayaAscent.climbed_at,
    climbed_by: kayaAscent.climbed_by,
    comment: kayaAscent.comment,
    days_since_climb: kayaAscent.days_since_climb,
    source: 'kaya',
    kaya_climb_slug: kayaAscent.kaya_climb_slug,
  };
}

// Recent Climbs View (extracted from original)
function RecentClimbsView({ climbHistory }: { climbHistory: UnifiedClimbEntry[] }) {
  return (
    <div className="space-y-3">
      {climbHistory.map((climb, index) => {
        const displayComment = cleanComment(climb.comment);
        const climbDate = new Date(climb.climbed_at);
        const dateStr = format(climbDate, 'MMM d, yyyy');
        const isKaya = climb.source === 'kaya';

        return (
          <div
            key={climb.id}
            className={`bg-white dark:bg-gray-900 border rounded-lg p-3 sm:p-4 hover:shadow-md transition-all ${
              index === 0
                ? 'border-blue-200 dark:border-blue-800 shadow-sm'
                : 'border-gray-200 dark:border-gray-700 hover:border-blue-200 dark:hover:border-blue-800'
            }`}
          >
            <div className="flex items-start justify-between gap-3 mb-2">
              {isKaya ? (
                <div className="flex items-center gap-1.5 flex-1 min-w-0">
                  <span className="text-sm sm:text-base font-bold text-gray-900 dark:text-white truncate">
                    {climb.route_name}
                  </span>
                </div>
              ) : (
                <a
                  href={`https://www.mountainproject.com/route/${climb.mp_route_id}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-1.5 group flex-1 min-w-0"
                >
                  <span className="text-sm sm:text-base font-bold text-gray-900 dark:text-white group-hover:text-blue-600 dark:group-hover:text-blue-500 transition-colors truncate">
                    {climb.route_name}
                  </span>
                  <ExternalLink className="h-3.5 w-3.5 text-gray-400 group-hover:text-blue-600 dark:group-hover:text-blue-500 flex-shrink-0 transition-colors" />
                </a>
              )}
              <span className="text-xs sm:text-sm font-bold text-blue-700 dark:text-blue-300 bg-blue-50 dark:bg-blue-900/30 px-2 py-0.5 rounded flex-shrink-0">
                {climb.route_rating}
              </span>
            </div>

            <div className="inline-flex items-center gap-1 text-xs text-gray-600 dark:text-gray-400 mb-2">
              <MapPin className="h-3 w-3" />
              <span>{climb.area_name}</span>
              {isKaya && (
                <span className="ml-1 px-1.5 py-0.5 bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 rounded text-xs font-medium flex items-center gap-0.5">
                  <Mountain className="h-3 w-3" />
                  Kaya
                </span>
              )}
              {!isKaya && climb.mp_area_id && (
                <a
                  href={`https://www.mountainproject.com/area/${climb.mp_area_id}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-0.5 hover:text-blue-600 dark:hover:text-blue-500 transition-colors group"
                >
                  <ExternalLink className="h-2.5 w-2.5 opacity-0 group-hover:opacity-100 transition-opacity" />
                </a>
              )}
            </div>

            <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-600 dark:text-gray-400">
              {climb.climbed_by && climb.climbed_by.trim() !== '' && (
                <>
                  <div className="flex items-center gap-1">
                    <User className="h-3 w-3" />
                    <span>{climb.climbed_by}</span>
                  </div>
                  <span className="text-gray-400">•</span>
                </>
              )}
              <div className="flex items-center gap-1">
                <Calendar className="h-3 w-3" />
                <span>{dateStr}</span>
              </div>
              <span className="text-gray-400">•</span>
              <span>{formatDaysAgo(climb.days_since_climb)}</span>
              {index === 0 && (
                <>
                  <span className="text-gray-400">•</span>
                  <span className="text-blue-600 dark:text-blue-400 font-semibold">Latest</span>
                </>
              )}
            </div>

            {displayComment && (
              <div className="mt-2 pt-2 border-t border-blue-100 dark:border-blue-900/50">
                <p className="text-xs text-gray-700 dark:text-gray-300 italic">
                  &quot;{displayComment}&quot;
                </p>
              </div>
            )}
          </div>
        );
      })}

      {climbHistory.length === 0 && (
        <div className="text-center py-12">
          <Footprints className="w-16 h-16 mx-auto text-gray-300 dark:text-gray-600 mb-4" />
          <p className="text-gray-500 dark:text-gray-400">No recent activity recorded</p>
        </div>
      )}
    </div>
  );
}

export function RecentActivityModal({
  locationId,
  locationName,
  climbHistory,
  onClose
}: RecentActivityModalProps) {
  const [view, setView] = useState<View>('recent');
  const [searchQuery, setSearchQuery] = useState('');
  const [kayaAscents, setKayaAscents] = useState<KayaAscentEntry[]>([]);
  const [isLoadingKaya, setIsLoadingKaya] = useState(true);

  // Fetch Kaya ascents on mount
  useEffect(() => {
    const fetchKayaAscents = async () => {
      try {
        setIsLoadingKaya(true);
        const ascents = await climbActivityApi.getKayaAscentsForLocation(locationId, 5);
        setKayaAscents(ascents);
      } catch (error) {
        console.error('Failed to fetch Kaya ascents:', error);
        setKayaAscents([]);
      } finally {
        setIsLoadingKaya(false);
      }
    };

    fetchKayaAscents();
  }, [locationId]);

  // Merge and sort all climbs (MP: 5 most recent + Kaya: 5 most recent)
  const allClimbs: UnifiedClimbEntry[] = [
    ...climbHistory.slice(0, 5).map(convertMPToUnified),
    ...kayaAscents.map(convertKayaToUnified),
  ].sort((a, b) => {
    // Sort by date descending (most recent first)
    return new Date(b.climbed_at).getTime() - new Date(a.climbed_at).getTime();
  });

  // Filter climb history based on search query
  const filteredClimbHistory = searchQuery
    ? allClimbs.filter(climb =>
        climb.route_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        climb.area_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        climb.climbed_by.toLowerCase().includes(searchQuery.toLowerCase())
      )
    : allClimbs;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-2 sm:p-4">
      <div className="bg-white dark:bg-gray-800 rounded-xl shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="bg-gradient-to-r from-gray-50 via-blue-50/30 to-gray-50 dark:from-gray-900 dark:via-blue-950/30 dark:to-gray-900 px-4 sm:px-5 py-3 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2.5 min-w-0 flex-1">
              <div className="bg-blue-100 dark:bg-blue-900/40 p-1.5 rounded-lg">
                <Footprints className="w-5 h-5 text-blue-600 dark:text-blue-400 flex-shrink-0" />
              </div>
              <div className="min-w-0 flex-1">
                <h2 className="text-base sm:text-lg font-bold text-gray-900 dark:text-white">
                  Recent Activity
                  {isLoadingKaya && <span className="ml-2 text-xs text-gray-500 dark:text-gray-400">(Loading Kaya...)</span>}
                </h2>
                <p className="text-xs text-gray-500 dark:text-gray-400 truncate">
                  {locationName}
                </p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="flex-shrink-0 p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg transition-colors ml-2"
            >
              <X className="w-5 h-5 text-gray-500 dark:text-gray-400" />
            </button>
          </div>

          {/* Tabs */}
          <div className="flex gap-1 overflow-x-auto">
            <button
              onClick={() => setView('recent')}
              className={`flex items-center gap-1.5 px-3 sm:px-4 py-2 text-xs sm:text-sm font-medium rounded-lg transition-colors whitespace-nowrap ${
                view === 'recent'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'
              }`}
            >
              <List className="w-4 h-4" />
              <span>Recent Climbs</span>
            </button>
            <button
              onClick={() => setView('areas')}
              className={`flex items-center gap-1.5 px-3 sm:px-4 py-2 text-xs sm:text-sm font-medium rounded-lg transition-colors whitespace-nowrap ${
                view === 'areas'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'
              }`}
            >
              <FolderTree className="w-4 h-4" />
              <span>By Area</span>
            </button>
          </div>

          {/* Search Bar */}
          <div className="mt-3 relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400 dark:text-gray-500 pointer-events-none" />
            <input
              type="text"
              placeholder={view === 'recent' ? "Search climbs, areas, or climbers..." : "Search areas or routes..."}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-3 py-2 text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400"
            />
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
              >
                <X className="w-4 h-4" />
              </button>
            )}
          </div>
        </div>

        {/* Content */}
        {view === 'recent' ? (
          <div className="flex-1 overflow-y-auto p-3 sm:p-4">
            <RecentClimbsView climbHistory={filteredClimbHistory} />
          </div>
        ) : (
          <div className="flex flex-col flex-1 overflow-hidden min-h-0">
            <AreaDrillDownView locationId={locationId} locationName={locationName} searchQuery={searchQuery} />
          </div>
        )}

        {/* Footer */}
        <div className="px-4 py-3 border-t border-gray-200 dark:border-gray-700">
          <button
            onClick={onClose}
            className="w-full py-2 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 font-medium rounded-lg transition-colors text-sm"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
