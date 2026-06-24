import { FormEvent, useState } from 'react';
import { CheckCircle2, Loader2, MessageSquarePlus } from 'lucide-react';
import { moneyApi } from '../../services/money';
import { MoneyFeature, MoneyProject } from '../../types/money';

interface QuickJotComposerProps {
  project?: MoneyProject;
  selectedFeature: MoneyFeature | null;
  canWrite: boolean;
  onCreatedFeature: (feature: MoneyFeature) => void;
  onChanged: () => void;
}

function titleFromBody(body: string) {
  const firstLine = body.trim().split('\n')[0]?.trim() || 'Inbox note';
  return firstLine.length > 72 ? `${firstLine.slice(0, 69)}...` : firstLine;
}

export function QuickJotComposer({ project, selectedFeature, canWrite, onCreatedFeature, onChanged }: QuickJotComposerProps) {
  const [body, setBody] = useState('');
  const [target, setTarget] = useState<'selected' | 'inbox'>('selected');
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const attachToSelected = Boolean(selectedFeature && target === 'selected');
  const targetLabel = attachToSelected ? `Attach to ${selectedFeature?.title}` : 'Create inbox note';

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (!body.trim() || !project) return;
    setSaving(true);
    setSaved(false);
    setError(null);
    try {
      if (attachToSelected && selectedFeature) {
        await moneyApi.createNote(selectedFeature.id, { body: body.trim(), visibility: 'team' });
      } else {
        const feature = await moneyApi.createFeature(project.id, {
          feature_type: 'poi',
          title: titleFromBody(body),
          description: body.trim(),
          status: 'draft',
          geojson: { type: 'Point', coordinates: [project.center_lon, project.center_lat] },
          style: {},
          properties: { poi_category: 'general', poi_label: 'General note', source: 'quick-jot-inbox' },
        });
        await moneyApi.createNote(feature.id, { body: body.trim(), visibility: 'team' });
        onCreatedFeature(feature);
      }
      setBody('');
      setSaved(true);
      onChanged();
      window.setTimeout(() => setSaved(false), 2200);
    } catch {
      setError('Could not save note.');
    } finally {
      setSaving(false);
    }
  };

  if (!canWrite) return null;

  return (
    <form onSubmit={handleSubmit} className="rounded-[1.5rem] border border-stone-200 bg-white p-3 text-slate-950 shadow-sm">
      <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-teal-900 text-white"><MessageSquarePlus className="h-4 w-4" /></span>
          <div>
            <p className="text-sm font-semibold">Quick jot</p>
            <p className="text-xs text-slate-500">Capture an idea now; organize it later.</p>
          </div>
        </div>
        <select value={target} onChange={event => setTarget(event.target.value as 'selected' | 'inbox')} className="rounded-xl border border-stone-300 bg-stone-50 px-3 py-2 text-xs font-semibold text-slate-700 outline-none focus:border-teal-700" aria-label="Note destination">
          <option value="selected" disabled={!selectedFeature}>Selected item</option>
          <option value="inbox">Inbox note</option>
        </select>
      </div>
      <textarea
        value={body}
        onChange={event => setBody(event.target.value)}
        className="min-h-24 w-full resize-none rounded-2xl border border-stone-300 bg-stone-50 p-3 text-sm leading-6 text-slate-900 outline-none placeholder:text-slate-400 focus:border-teal-700 focus:ring-2 focus:ring-teal-700/10"
        placeholder="Route idea, cleanup task, trail import issue, photo reminder..."
        maxLength={5000}
      />
      <div className="mt-3 flex items-center justify-between gap-2">
        <p className="min-w-0 truncate text-xs text-slate-500">{targetLabel}</p>
        <button type="submit" disabled={saving || !body.trim() || !project} className="inline-flex items-center gap-2 rounded-xl bg-teal-900 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-teal-800 disabled:bg-stone-300">
          {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : saved ? <CheckCircle2 className="h-4 w-4" /> : null}
          {saved ? 'Saved' : 'Save note'}
        </button>
      </div>
      {error && <div className="mt-2 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-xs font-semibold text-red-700">{error}</div>}
    </form>
  );
}
