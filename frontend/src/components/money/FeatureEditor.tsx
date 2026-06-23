import { FormEvent, useState } from 'react';
import { Archive, Loader2, Save } from 'lucide-react';
import { moneyApi } from '../../services/money';
import { MoneyFeature, MoneyFeatureRequest, MoneyFeatureStatus } from '../../types/money';
import { featureTypeLabel } from './geometry';

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
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const payload = (): MoneyFeatureRequest => ({
    feature_type: feature.feature_type,
    title: title.trim(),
    description: description.trim() || null,
    status,
    geojson: feature.geojson,
    style: feature.style ?? {},
    properties: feature.properties ?? {},
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
      setError('Could not save feature.');
    } finally {
      setSaving(false);
    }
  };

  const handleArchive = async () => {
    setSaving(true);
    setError(null);
    try {
      await moneyApi.archiveFeature(feature.id);
      onArchived();
    } catch {
      setError('Could not archive feature.');
    } finally {
      setSaving(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-xs font-bold uppercase tracking-widest text-emerald-200">{featureTypeLabel(feature.feature_type)}</p>
          <h3 className="text-lg font-black text-white">Feature details</h3>
        </div>
        <span className="rounded-full bg-white/10 px-3 py-1 text-xs font-bold text-slate-300">{feature.status}</span>
      </div>

      <label className="block">
        <span className="mb-1 block text-xs font-semibold text-slate-300">Title</span>
        <input value={title} onChange={event => setTitle(event.target.value)} disabled={!canWrite} maxLength={200} className="w-full rounded-2xl border border-white/10 bg-white/10 px-3 py-2.5 text-sm text-white outline-none ring-emerald-300/40 focus:ring-2 disabled:opacity-70" />
      </label>

      <label className="block">
        <span className="mb-1 block text-xs font-semibold text-slate-300">Description</span>
        <textarea value={description} onChange={event => setDescription(event.target.value)} disabled={!canWrite} maxLength={2000} className="min-h-24 w-full resize-none rounded-2xl border border-white/10 bg-white/10 px-3 py-2.5 text-sm text-white outline-none ring-emerald-300/40 focus:ring-2 disabled:opacity-70" />
      </label>

      <label className="block">
        <span className="mb-1 block text-xs font-semibold text-slate-300">Status</span>
        <select value={status} onChange={event => setStatus(event.target.value as MoneyFeatureStatus)} disabled={!canWrite} className="w-full rounded-2xl border border-white/10 bg-slate-900 px-3 py-2.5 text-sm text-white outline-none disabled:opacity-70">
          <option value="draft">Draft</option>
          <option value="active">Active</option>
          <option value="archived">Archived</option>
        </select>
      </label>

      {error && <div className="rounded-2xl border border-red-300/20 bg-red-500/10 px-3 py-2 text-xs text-red-100">{error}</div>}

      {canWrite && (
        <div className="flex gap-2">
          <button type="submit" disabled={saving || !title.trim()} className="flex flex-1 items-center justify-center gap-2 rounded-2xl bg-emerald-300 px-4 py-2.5 text-sm font-bold text-slate-950 disabled:opacity-40">
            {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            Save
          </button>
          <button type="button" onClick={handleArchive} disabled={saving || feature.status === 'archived'} className="rounded-2xl bg-white/10 px-4 py-2.5 text-sm font-bold text-slate-100 disabled:opacity-40">
            <Archive className="h-4 w-4" />
          </button>
        </div>
      )}
    </form>
  );
}
