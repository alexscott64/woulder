import { FormEvent, useState } from 'react';
import { Archive, ChevronDown, Loader2, Save, Tag } from 'lucide-react';
import { moneyApi } from '../../services/money';
import { MoneyFeature, MoneyFeatureRequest, MoneyFeatureStatus } from '../../types/money';
import { featureTypeLabel } from './geometry';
import { getPOIIconOption, normalizePOICategory, POI_ICON_OPTIONS, MoneyPOICategory } from './POIIcons';

interface FeatureEditorProps {
  feature: MoneyFeature;
  canWrite: boolean;
  onSaved: (feature: MoneyFeature) => void;
  onArchived: () => void;
}

export function FeatureEditor({ feature, canWrite, onSaved, onArchived }: FeatureEditorProps) {
  const [title, setTitle] = useState(feature.title);
  const [description, setDescription] = useState(feature.description ?? '');
  const [status, setStatus] = useState<MoneyFeatureStatus>(feature.status);
  const [poiCategory, setPoiCategory] = useState(normalizePOICategory(feature.properties?.poi_category));
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const selectedPOIOption = getPOIIconOption(poiCategory);
  const SelectedIcon = selectedPOIOption.icon;

  const payload = (): MoneyFeatureRequest => ({
    feature_type: feature.feature_type,
    title: title.trim(),
    description: description.trim() || null,
    status,
    geojson: feature.geojson,
    style: feature.style ?? {},
    properties: feature.feature_type === 'poi'
      ? { ...(feature.properties ?? {}), poi_category: poiCategory, poi_label: selectedPOIOption.label }
      : feature.properties ?? {},
  });

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (!title.trim()) return;
    setSaving(true);
    setError(null);
    try {
      const saved = await moneyApi.updateFeature(feature.id, payload());
      onSaved(saved);
    } catch {
      setError('Could not save this note.');
    } finally {
      setSaving(false);
    }
  };

  const archiveLabel = feature.feature_type === 'poi' ? 'Archive pin' : 'Archive map item';

  const handleArchive = async () => {
    setSaving(true);
    setError(null);
    try {
      await moneyApi.archiveFeature(feature.id);
      onArchived();
    } catch {
      setError('Could not archive this note.');
    } finally {
      setSaving(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-3 text-[#F2F0E7]">
      <div className="rounded-2xl border border-[#2B403A] bg-[#111D1B] p-3">
        <div className="flex items-center justify-between gap-3">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#7EA16B]">{featureTypeLabel(feature.feature_type)}</p>
            <h3 className="text-lg font-semibold text-[#F2F0E7]">Map note details</h3>
          </div>
          <span className="rounded-full bg-[#172522] px-3 py-1 text-xs font-semibold uppercase text-[#AAB8AD]">{feature.status}</span>
        </div>
      </div>

      <label className="block rounded-2xl border border-[#2B403A] bg-[#111D1B] p-3">
        <span className="mb-1 block text-xs font-semibold uppercase tracking-[0.12em] text-[#AAB8AD]">Title</span>
        <input value={title} onChange={event => setTitle(event.target.value)} disabled={!canWrite} maxLength={200} className="w-full rounded-xl border border-[#2B403A] bg-[#0B1714]/80 px-3 py-3 text-sm text-[#F2F0E7] outline-none placeholder:text-[#74847B] focus:border-[#7EA16B] focus:ring-2 focus:ring-[#7EA16B]/10 disabled:opacity-60" />
      </label>

      <label className="block rounded-2xl border border-[#2B403A] bg-[#111D1B] p-3">
        <span className="mb-1 block text-xs font-semibold uppercase tracking-[0.12em] text-[#AAB8AD]">Crag note</span>
        <textarea value={description} onChange={event => setDescription(event.target.value)} disabled={!canWrite} maxLength={2000} className="min-h-24 w-full resize-none rounded-xl border border-[#2B403A] bg-[#0B1714]/80 px-3 py-3 text-sm leading-6 text-[#F2F0E7] outline-none placeholder:text-[#74847B] focus:border-[#7EA16B] focus:ring-2 focus:ring-[#7EA16B]/10 disabled:opacity-60" placeholder="Access beta, cleanup task, landing condition, route idea, or next action." />
      </label>

      {feature.feature_type === 'poi' && (
        <section className="rounded-2xl border border-[#2B403A] bg-[#111D1B] p-3">
          <div className="flex items-center gap-2">
            <div className={`flex h-10 w-10 items-center justify-center rounded-xl border ${selectedPOIOption.tone}`}>
              <SelectedIcon className="h-5 w-5" />
            </div>
            <div className="min-w-0 flex-1">
              <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[#7EA16B]"><Tag className="mr-1 inline h-3.5 w-3.5" />Map pin type</p>
              <p className="truncate text-sm font-semibold text-[#F2F0E7]">{selectedPOIOption.label} · {selectedPOIOption.description}</p>
            </div>
          </div>
          <div className="relative mt-3">
            <select
              value={poiCategory}
              onChange={event => setPoiCategory(event.target.value as MoneyPOICategory)}
              disabled={!canWrite}
              className="w-full appearance-none rounded-xl border border-[#2B403A] bg-[#0B1714]/80 px-3 py-3 pr-10 text-sm font-semibold text-[#F2F0E7] outline-none focus:border-[#7EA16B] disabled:opacity-60"
            >
              {POI_ICON_OPTIONS.map(option => <option key={option.id} value={option.id}>{option.label} — {option.description}</option>)}
            </select>
            <ChevronDown className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[#AAB8AD]" />
          </div>
        </section>
      )}

      <label className="block rounded-2xl border border-[#2B403A] bg-[#111D1B] p-3">
        <span className="mb-1 block text-xs font-semibold uppercase tracking-[0.12em] text-[#AAB8AD]">Status</span>
        <select value={status} onChange={event => setStatus(event.target.value as MoneyFeatureStatus)} disabled={!canWrite} className="w-full rounded-xl border border-[#2B403A] bg-[#0B1714]/80 px-3 py-3 text-sm font-semibold text-[#F2F0E7] outline-none focus:border-[#7EA16B] disabled:opacity-60">
          <option value="draft">Draft</option>
          <option value="active">Active</option>
          <option value="archived">Archived</option>
        </select>
      </label>

      {error && <div className="rounded-2xl border border-red-900/60 bg-red-950/40 px-3 py-2 text-xs font-semibold text-red-200">{error}</div>}

      {canWrite && (
        <div className="flex flex-col gap-2 sm:flex-row">
          <button type="submit" disabled={saving || !title.trim()} className="flex flex-1 items-center justify-center gap-2 rounded-xl bg-[#7EA16B] px-4 py-3 text-sm font-semibold text-[#07110F] disabled:bg-[#2B403A] disabled:text-[#74847B]">
            {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            Save note
          </button>
          <button type="button" onClick={handleArchive} disabled={saving || feature.status === 'archived'} className="inline-flex items-center justify-center gap-2 rounded-xl border border-[#2B403A] bg-[#111D1B] px-4 py-3 text-sm font-semibold text-[#AAB8AD] disabled:opacity-40 hover:border-[#C88A3D] hover:text-[#E0B36F]" title={archiveLabel} aria-label={archiveLabel}>
            <Archive className="h-4 w-4" />
            {archiveLabel}
          </button>
        </div>
      )}
    </form>
  );
}
