import { Camera, Circle, FileText, MapPin, Route, Search, Square } from 'lucide-react';
import { MoneyFeature, MoneyFeatureFilters, MoneyFeatureType, MoneyUpload } from '../../types/money';
import { featureTypeLabel } from './geometry';

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
  topo: Square,
  poi: MapPin,
  drawing: Circle,
};

const typeOptions: Array<MoneyFeatureType | 'all'> = ['all', 'trail', 'topo', 'poi', 'drawing'];

export function FeatureList({ features, selectedFeatureId, filters, noteCounts, primaryUploads, onSelect, onFiltersChange }: FeatureListProps) {
  const search = filters.search?.toLowerCase().trim() ?? '';
  const filteredFeatures = features.filter(feature => {
    if (filters.type && filters.type !== 'all' && feature.feature_type !== filters.type) return false;
    if (filters.status && filters.status !== 'all' && feature.status !== filters.status) return false;
    if (search && !`${feature.title} ${feature.description ?? ''}`.toLowerCase().includes(search)) return false;
    return true;
  });

  return (
    <section className="space-y-4">
      <div className="relative">
        <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
        <input
          value={filters.search ?? ''}
          onChange={event => onFiltersChange({ ...filters, search: event.target.value })}
          className="w-full rounded-2xl border border-white/10 bg-white/10 py-3 pl-10 pr-4 text-sm text-white outline-none ring-emerald-300/40 placeholder:text-slate-500 focus:ring-2"
          placeholder="Search features"
        />
      </div>

      <div className="flex gap-2 overflow-x-auto pb-1 scrollbar-hide">
        {typeOptions.map(type => (
          <button
            key={type}
            onClick={() => onFiltersChange({ ...filters, type })}
            className={`shrink-0 rounded-full px-3 py-1.5 text-xs font-bold transition ${filters.type === type || (!filters.type && type === 'all') ? 'bg-emerald-300 text-slate-950' : 'bg-white/10 text-slate-300 hover:bg-white/15'}`}
          >
            {type === 'all' ? 'All' : featureTypeLabel(type)}
          </button>
        ))}
      </div>

      <div className="flex gap-2">
        {(['all', 'draft', 'active', 'archived'] as const).map(status => (
          <button
            key={status}
            onClick={() => onFiltersChange({ ...filters, status })}
            className={`rounded-full px-3 py-1.5 text-xs font-semibold capitalize transition ${filters.status === status || (!filters.status && status === 'all') ? 'bg-sky-300 text-slate-950' : 'bg-white/10 text-slate-300 hover:bg-white/15'}`}
          >
            {status}
          </button>
        ))}
      </div>

      <div className="space-y-2">
        {filteredFeatures.map(feature => {
          const Icon = typeIcon[feature.feature_type];
          const selected = feature.id === selectedFeatureId;
          const upload = primaryUploads[feature.id];
          return (
            <button
              key={feature.id}
              onClick={() => onSelect(feature)}
              className={`w-full rounded-3xl border p-3 text-left transition ${selected ? 'border-emerald-300 bg-emerald-300/15' : 'border-white/10 bg-white/8 hover:bg-white/12'}`}
            >
              <div className="flex gap-3">
                <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-slate-900 text-emerald-200">
                  <Icon className="h-5 w-5" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="flex items-center justify-between gap-2">
                    <h3 className="truncate font-bold text-white">{feature.title}</h3>
                    <span className={`rounded-full px-2 py-0.5 text-[0.65rem] font-bold uppercase ${feature.status === 'active' ? 'bg-emerald-300/20 text-emerald-100' : feature.status === 'archived' ? 'bg-slate-500/20 text-slate-300' : 'bg-amber-300/20 text-amber-100'}`}>{feature.status}</span>
                  </div>
                  <p className="mt-1 line-clamp-2 text-xs text-slate-400">{feature.description || featureTypeLabel(feature.feature_type)}</p>
                  <div className="mt-3 flex items-center gap-3 text-xs text-slate-400">
                    <span className="flex items-center gap-1"><FileText className="h-3.5 w-3.5" />{noteCounts[feature.id] ?? 0}</span>
                    {upload && <span className="flex items-center gap-1"><Camera className="h-3.5 w-3.5" />Image</span>}
                    <span>{new Date(feature.updated_at).toLocaleDateString()}</span>
                  </div>
                </div>
              </div>
            </button>
          );
        })}
        {filteredFeatures.length === 0 && <div className="rounded-3xl border border-dashed border-white/15 p-6 text-center text-sm text-slate-400">No matching features.</div>}
      </div>
    </section>
  );
}
