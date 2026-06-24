import { Camera, Circle, FileText, Layers, MapPin, Route, Search } from 'lucide-react';
import { MoneyFeature, MoneyFeatureFilters, MoneyFeatureStatus, MoneyFeatureType, MoneyUpload } from '../../types/money';
import { featureTypeLabel } from './geometry';
import { getPOIIconOption } from './POIIcons';

interface FeatureListProps {
  features: MoneyFeature[];
  selectedFeatureId: string | null;
  filters: MoneyFeatureFilters;
  noteCounts: Record<string, number>;
  primaryUploads: Record<string, MoneyUpload>;
  onSelect: (feature: MoneyFeature) => void;
  onFiltersChange: (filters: MoneyFeatureFilters) => void;
}

const typeIcon: Record<MoneyFeatureType, typeof Route> = {
  trail: Route,
  topo: Layers,
  poi: MapPin,
  drawing: Circle,
};

const typeOptions: Array<{ value: MoneyFeatureType | 'all'; label: string }> = [
  { value: 'all', label: 'All' },
  { value: 'trail', label: 'Trails' },
  { value: 'topo', label: 'Topos' },
  { value: 'poi', label: 'Map pins' },
  { value: 'drawing', label: 'Sketches' },
];

const statusOptions: Array<{ value: MoneyFeatureStatus | 'all'; label: string }> = [
  { value: 'all', label: 'All status' },
  { value: 'draft', label: 'Draft' },
  { value: 'active', label: 'Active' },
  { value: 'archived', label: 'Archived' },
];

function statusClass(status: MoneyFeature['status']) {
  if (status === 'active') return 'bg-emerald-50 text-emerald-800 border-emerald-200';
  if (status === 'archived') return 'bg-stone-100 text-stone-600 border-stone-200';
  return 'bg-amber-50 text-amber-800 border-amber-200';
}

export function FeatureList({ features, selectedFeatureId, filters, noteCounts, primaryUploads, onSelect, onFiltersChange }: FeatureListProps) {
  const search = filters.search?.toLowerCase().trim() ?? '';
  const filteredFeatures = features.filter(feature => {
    if (filters.type && filters.type !== 'all' && feature.feature_type !== filters.type) return false;
    if (filters.status && filters.status !== 'all' && feature.status !== filters.status) return false;
    if (search && !`${feature.title} ${feature.description ?? ''} ${feature.properties?.poi_label ?? ''}`.toLowerCase().includes(search)) return false;
    return true;
  });

  const updateType = (type: MoneyFeatureType | 'all') => onFiltersChange({ ...filters, type });
  const updateStatus = (status: MoneyFeatureStatus | 'all') => onFiltersChange({ ...filters, status });

  return (
    <section className="space-y-3">
      <div className="relative">
        <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
        <input
          value={filters.search ?? ''}
          onChange={event => onFiltersChange({ ...filters, search: event.target.value })}
          className="w-full rounded-2xl border border-stone-300 bg-white py-3 pl-10 pr-4 text-sm text-slate-950 outline-none placeholder:text-slate-400 focus:border-teal-700 focus:ring-2 focus:ring-teal-700/10"
          placeholder="Search trails, map pins, notes, photos"
        />
      </div>

      <div className="space-y-2 rounded-2xl border border-stone-200 bg-stone-50 p-2">
        <div className="flex gap-1 overflow-x-auto pb-1">
          {typeOptions.map(type => (
            <button
              key={type.value}
              type="button"
              onClick={() => updateType(type.value)}
              className={`shrink-0 rounded-full px-3 py-1.5 text-xs font-semibold transition ${filters.type === type.value || (!filters.type && type.value === 'all') ? 'bg-teal-900 text-white' : 'bg-white text-slate-600 hover:text-slate-950'}`}
            >
              {type.label}
            </button>
          ))}
        </div>
        <div className="flex gap-1 overflow-x-auto">
          {statusOptions.map(status => (
            <button
              key={status.value}
              type="button"
              onClick={() => updateStatus(status.value)}
              className={`shrink-0 rounded-full px-3 py-1.5 text-xs font-semibold transition ${filters.status === status.value || (!filters.status && status.value === 'all') ? 'bg-slate-800 text-white' : 'bg-white text-slate-600 hover:text-slate-950'}`}
            >
              {status.label}
            </button>
          ))}
        </div>
      </div>

      <div className="flex items-center justify-between text-xs text-slate-500">
        <span>{filteredFeatures.length} visible</span>
        {(filters.search || filters.type !== 'all' || filters.status !== 'all') && (
          <button type="button" onClick={() => onFiltersChange({ type: 'all', status: 'all', search: '' })} className="font-semibold text-teal-800 hover:text-teal-950">Clear filters</button>
        )}
      </div>

      <div className="grid gap-2">
        {filteredFeatures.map(feature => {
          const Icon = feature.feature_type === 'poi' ? getPOIIconOption(feature.properties?.poi_category).icon : typeIcon[feature.feature_type];
          const selected = feature.id === selectedFeatureId;
          const upload = primaryUploads[feature.id];
          const poiOption = feature.feature_type === 'poi' ? getPOIIconOption(feature.properties?.poi_category) : null;
          return (
            <button
              key={feature.id}
              type="button"
              onClick={() => onSelect(feature)}
              className={`w-full rounded-2xl border p-3 text-left transition hover:border-teal-700 ${selected ? 'border-teal-800 bg-teal-50 shadow-sm' : 'border-stone-200 bg-white hover:bg-stone-50'}`}
            >
              <div className="flex gap-3">
                <div className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border ${poiOption ? poiOption.tone : 'border-teal-200 bg-teal-50 text-teal-900'}`}>
                  <Icon className="h-5 w-5" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="flex items-start justify-between gap-2">
                    <h3 className="line-clamp-2 font-semibold leading-tight text-slate-950">{feature.title}</h3>
                    <span className={`shrink-0 rounded-full border px-2 py-0.5 text-[0.68rem] font-semibold uppercase ${statusClass(feature.status)}`}>{feature.status}</span>
                  </div>
                  <p className="mt-1 line-clamp-2 text-xs leading-5 text-slate-600">{feature.description || poiOption?.description || featureTypeLabel(feature.feature_type)}</p>
                  <div className="mt-3 flex flex-wrap items-center gap-3 text-xs text-slate-500">
                    {poiOption && <span>{poiOption.label}</span>}
                    <span className="flex items-center gap-1"><FileText className="h-3.5 w-3.5" />{noteCounts[feature.id] ?? 0}</span>
                    {upload && <span className="flex items-center gap-1"><Camera className="h-3.5 w-3.5" />Photo</span>}
                    <span>{new Date(feature.updated_at).toLocaleDateString()}</span>
                  </div>
                </div>
              </div>
            </button>
          );
        })}
        {filteredFeatures.length === 0 && <div className="rounded-2xl border border-dashed border-stone-300 bg-white/75 p-6 text-center text-sm text-slate-600">No records match the current filters.</div>}
      </div>
    </section>
  );
}
