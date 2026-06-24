import { UseQueryResult } from '@tanstack/react-query';
import { Camera, FileImage, Layers, Loader2, MapPin, PencilRuler, Route, X } from 'lucide-react';
import { MoneyFeature, MoneyFeatureDetail } from '../../types/money';
import { featureTypeLabel, formatBytes } from './geometry';
import { FeatureEditor } from './FeatureEditor';
import { ImageUploader } from './ImageUploader';
import { NotesPanel } from './NotesPanel';
import { getPOIIconOption } from './POIIcons';

interface RecordPageProps {
  selectedFeature: MoneyFeature | null;
  projectId?: string;
  canWrite: boolean;
  detailQuery: UseQueryResult<MoneyFeatureDetail>;
  onSaved: (feature: MoneyFeature) => void;
  onArchived: () => void;
  onChanged: () => void;
  onClear: () => void;
}

function formatDate(value?: string) {
  if (!value) return 'Not saved yet';
  return new Date(value).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' });
}

function statusTone(status: MoneyFeature['status']) {
  if (status === 'active') return 'bg-emerald-50 text-emerald-800 border-emerald-200';
  if (status === 'archived') return 'bg-stone-100 text-stone-600 border-stone-200';
  return 'bg-amber-50 text-amber-800 border-amber-200';
}

function FeatureGlyph({ feature, className = 'h-5 w-5' }: { feature: MoneyFeature; className?: string }) {
  if (feature.feature_type === 'poi') {
    const POIIcon = getPOIIconOption(feature.properties?.poi_category).icon;
    return <POIIcon className={className} />;
  }
  if (feature.feature_type === 'trail') return <Route className={className} />;
  if (feature.feature_type === 'topo') return <Layers className={className} />;
  return <PencilRuler className={className} />;
}

function EmptyRecordPage() {
  return (
    <aside className="min-h-[34rem] rounded-[1.5rem] border border-stone-200 bg-white p-4 text-slate-950 shadow-sm">
      <div className="rounded-2xl border border-dashed border-stone-300 bg-stone-50 p-5 text-center">
        <MapPin className="mx-auto h-10 w-10 text-teal-900" />
        <p className="mt-3 text-xs font-semibold uppercase tracking-[0.18em] text-teal-900">Detail sheet</p>
        <h2 className="mt-1 text-2xl font-semibold leading-tight">Select a map or binder record</h2>
        <p className="mt-3 text-sm leading-6 text-slate-600">Titles, status, notes, photos, and metadata stay together for the selected point, trail, area, or sketch.</p>
      </div>
      <div className="mt-4 rounded-2xl border border-stone-200 bg-white p-4">
        <FileImage className="h-8 w-8 text-teal-900" />
        <h3 className="mt-2 font-semibold">Photos stay attached</h3>
        <p className="mt-1 text-sm leading-6 text-slate-600">Choose a record to add or review photos. Uploads also appear in the Photos section.</p>
      </div>
    </aside>
  );
}

export function RecordPage({ selectedFeature, projectId, canWrite, detailQuery, onSaved, onArchived, onChanged, onClear }: RecordPageProps) {
  if (!selectedFeature || !projectId) return <EmptyRecordPage />;

  const notes = detailQuery.data?.notes ?? [];
  const uploads = detailQuery.data?.uploads ?? [];
  const upload = uploads[0];
  const poiOption = selectedFeature.feature_type === 'poi' ? getPOIIconOption(selectedFeature.properties?.poi_category) : null;

  return (
    <aside className="min-h-0 overflow-y-auto rounded-[1.5rem] border border-stone-200 bg-stone-50 p-3 text-slate-950 shadow-sm custom-scrollbar xl:max-h-[calc(100vh-13rem)]">
      <div className="rounded-2xl border border-stone-200 bg-white p-4">
        <div className="flex items-start justify-between gap-3">
          <div className="flex min-w-0 gap-3">
            <div className={`flex h-12 w-12 shrink-0 items-center justify-center rounded-xl border ${poiOption ? poiOption.tone : 'border-teal-200 bg-teal-50 text-teal-900'}`}>
              <FeatureGlyph feature={selectedFeature} className="h-6 w-6" />
            </div>
            <div className="min-w-0">
              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-teal-900">{featureTypeLabel(selectedFeature.feature_type)} record</p>
              <h2 className="break-words text-2xl font-semibold leading-tight sm:text-3xl">{selectedFeature.title}</h2>
              <p className="mt-1 text-xs text-slate-500">Updated {formatDate(selectedFeature.updated_at)}</p>
            </div>
          </div>
          <button type="button" onClick={onClear} className="rounded-full border border-stone-200 bg-white p-2 text-slate-700 hover:border-teal-700 hover:text-teal-900" title="Clear record"><X className="h-4 w-4" /></button>
        </div>
        <div className="mt-3 flex flex-wrap items-center gap-2 text-xs font-semibold">
          <span className={`rounded-full border px-3 py-1 uppercase ${statusTone(selectedFeature.status)}`}>{selectedFeature.status}</span>
          {poiOption && <span className="rounded-full border border-stone-200 bg-stone-50 px-3 py-1 text-slate-600">{poiOption.label}</span>}
          {upload && <span className="rounded-full border border-stone-200 bg-stone-50 px-3 py-1 text-slate-600"><Camera className="mr-1 inline h-3.5 w-3.5" />{upload.original_filename} · {formatBytes(upload.byte_size)}</span>}
        </div>
      </div>

      {detailQuery.isLoading ? (
        <div className="mt-3 rounded-2xl border border-stone-200 bg-white p-5 text-center text-sm font-semibold text-slate-600"><Loader2 className="mr-2 inline h-4 w-4 animate-spin" />Loading record context</div>
      ) : (
        <div className="mt-3 space-y-3">
          <FeatureEditor key={selectedFeature.id} feature={selectedFeature} canWrite={canWrite} onSaved={onSaved} onArchived={onArchived} />
          <NotesPanel featureId={selectedFeature.id} notes={notes} canWrite={canWrite} onChanged={onChanged} />
          <ImageUploader projectId={projectId} featureId={selectedFeature.id} uploads={uploads} canWrite={canWrite} onUploaded={onChanged} onDeleted={onChanged} />
        </div>
      )}
    </aside>
  );
}
