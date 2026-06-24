import { ChangeEvent, FormEvent, ReactNode, useEffect, useMemo, useState } from 'react';
import {
  AlertTriangle,
  BookOpen,
  Camera,
  ChevronDown,
  CircleDot,
  ClipboardCheck,
  FileImage,
  Image as ImageIcon,
  Layers,
  Loader2,
  LogOut,
  Map as MapIcon,
  MapPin,
  Maximize2,
  MessageSquarePlus,
  Mountain,
  Navigation,
  Pencil,
  PencilRuler,
  Plus,
  RefreshCw,
  Route,
  Search,
  Sparkles,
  Upload,
  Users,
  X,
} from 'lucide-react';
import { UseQueryResult } from '@tanstack/react-query';
import { MoneyFeature, MoneyFeatureDetail, MoneyFeatureFilters, MoneyFeatureType, MoneyProject, MoneyUpload, MoneyCurrentUser, MoneyPosition } from '../../types/money';
import { moneyApi } from '../../services/money';
import { featureTypeLabel, formatBytes, getFeatureCoordinates } from './geometry';
import { getPOIIconOption } from './POIIcons';
import { MoneyMap } from './MoneyMap';
import { DrawingToolbar } from './DrawingToolbar';
import { FeatureEditor } from './FeatureEditor';
import { ImageUploader } from './ImageUploader';
import { NotesPanel } from './NotesPanel';

export type FieldLens = 'all' | 'recent' | 'linked' | 'projects' | 'approach' | 'photos' | 'loops';
export type FieldPageKey = 'scratch' | 'approach' | 'topos' | 'photos' | 'projects';
export type FieldBlockKind = 'note' | 'photo' | 'sketch' | 'map' | 'poi' | 'trail' | 'task';

type FieldBlock = {
  id: string;
  kind: FieldBlockKind;
  title: string;
  body: string;
  date?: string;
  feature?: MoneyFeature;
  upload?: MoneyUpload;
  accent: string;
  rotation: string;
  wide?: boolean;
};

type ScratchNote = {
  id: string;
  body: string;
  createdAt: string;
};

interface FieldPagesProps {
  user?: MoneyCurrentUser | null;
  project?: MoneyProject;
  features: MoneyFeature[];
  visibleMapFeatures: MoneyFeature[];
  selectedFeature: MoneyFeature | null;
  selectedFeatureId: string | null;
  detailQuery: UseQueryResult<MoneyFeatureDetail>;
  projectId?: string;
  canWrite: boolean;
  isFetching: boolean;
  isOnline: boolean;
  loadError: boolean;
  savingFeature: boolean;
  filters: MoneyFeatureFilters;
  noteCounts: Record<string, number>;
  primaryUploads: Record<string, MoneyUpload>;
  lens: FieldLens;
  page: FieldPageKey;
  drawingType: MoneyFeatureType | null;
  draftPoints: MoneyPosition[];
  onLensChange: (lens: FieldLens) => void;
  onPageChange: (page: FieldPageKey) => void;
  onFiltersChange: (filters: MoneyFeatureFilters) => void;
  onSelectFeature: (feature: MoneyFeature) => void;
  onClearSelection: () => void;
  onRefresh: () => void;
  onLogout: () => void;
  onChanged: () => void;
  onFeatureSaved: (feature: MoneyFeature) => void;
  onFeatureArchived: () => void;
  onStartDrawing: (type: MoneyFeatureType) => void;
  onAddDraftPoint: (point: MoneyPosition) => void;
  onUndoDraftPoint: () => void;
  onCancelDrawing: () => void;
  onFinishDrawing: () => void;
}

const lensOptions: Array<{ id: FieldLens; label: string; hint: string }> = [
  { id: 'all', label: 'All notes', hint: 'Everything active' },
  { id: 'recent', label: 'Recent', hint: 'Latest crag updates' },
  { id: 'linked', label: 'Pinned to map', hint: 'Has a map anchor' },
  { id: 'projects', label: 'Ideas & tasks', hint: 'Drafts and open work' },
  { id: 'approach', label: 'Trails', hint: 'Approach and connector notes' },
  { id: 'photos', label: 'Photos & topos', hint: 'Visual references' },
  { id: 'loops', label: 'Needs follow-up', hint: 'Open tasks and reminders' },
];

const pages: Array<{ id: FieldPageKey; label: string; short: string; icon: typeof BookOpen; copy: string }> = [
  { id: 'scratch', label: 'Scratch', short: 'Scratch', icon: Pencil, copy: 'Local quick jots and loose beta. Select a map item to persist linked notes and photos.' },
  { id: 'approach', label: 'Trails', short: 'Trails', icon: Route, copy: 'Approaches, connectors, parking notes, and navigation clues.' },
  { id: 'topos', label: 'Topos & sketches', short: 'Topos', icon: Layers, copy: 'Topo ideas, drawings, and map-backed crag sketches.' },
  { id: 'photos', label: 'Photos', short: 'Photos', icon: Camera, copy: 'Uploaded reference photos tied to boulders, trails, and topo ideas.' },
  { id: 'projects', label: 'Ideas & tasks', short: 'Tasks', icon: ClipboardCheck, copy: 'Open cleanup, development ideas, and next-pass reminders.' },
];

const addTools: Array<{ type: 'note' | 'photo' | 'sketch' | 'map' | 'poi' | 'trail' | 'task'; label: string; icon: typeof Pencil }> = [
  { type: 'note', label: 'Note', icon: MessageSquarePlus },
  { type: 'photo', label: 'Photo', icon: Camera },
  { type: 'sketch', label: 'Sketch', icon: PencilRuler },
  { type: 'map', label: 'Topo', icon: MapIcon },
  { type: 'poi', label: 'Map pin', icon: MapPin },
  { type: 'trail', label: 'Trail', icon: Route },
  { type: 'task', label: 'Task', icon: ClipboardCheck },
];

function formatDate(value?: string) {
  if (!value) return 'Not saved';
  return new Date(value).toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}

function featurePage(feature: MoneyFeature): FieldPageKey {
  if (feature.feature_type === 'trail') return 'approach';
  if (feature.feature_type === 'topo' || feature.feature_type === 'drawing') return 'topos';
  if (feature.status === 'draft') return 'projects';
  return 'scratch';
}

function blockKind(feature: MoneyFeature, upload?: MoneyUpload): FieldBlockKind {
  if (upload) return 'photo';
  if (feature.feature_type === 'trail') return 'trail';
  if (feature.feature_type === 'drawing') return 'sketch';
  if (feature.status === 'draft') return 'task';
  if (feature.feature_type === 'topo') return 'map';
  return 'poi';
}

function buildBlocks(features: MoneyFeature[], noteCounts: Record<string, number>, uploads: Record<string, MoneyUpload>, scratchNotes: ScratchNote[], selectedDetail?: MoneyFeatureDetail): FieldBlock[] {
  const selectedNotes = selectedDetail?.notes ?? [];
  const selectedUploads = selectedDetail?.uploads ?? [];

  const featureBlocks = features.map((feature, index): FieldBlock => {
    const upload = uploads[feature.id];
    const notes = noteCounts[feature.id] ?? 0;
    const coordinates = getFeatureCoordinates(feature);
    const position = coordinates[0];
    const poi = feature.feature_type === 'poi' ? getPOIIconOption(feature.properties?.poi_category).label : featureTypeLabel(feature.feature_type);
    return {
      id: `feature-${feature.id}`,
      kind: blockKind(feature, upload),
      title: feature.title,
      body: [feature.description, notes ? `${notes} linked note${notes === 1 ? '' : 's'}` : null, position ? `${position[1].toFixed(5)}, ${position[0].toFixed(5)}` : null, poi].filter(Boolean).join(' · ') || 'Pinned crag note.',
      date: feature.updated_at,
      feature,
      upload,
      accent: feature.feature_type === 'trail' ? '#7EA16B' : feature.feature_type === 'topo' ? '#6F9FB5' : feature.feature_type === 'drawing' ? '#C7D38A' : '#C88A3D',
      rotation: index % 5 === 0 ? '-rotate-1' : index % 5 === 2 ? 'rotate-1' : '',
      wide: feature.feature_type === 'trail' || feature.feature_type === 'topo',
    };
  });

  const scratchBlocks = scratchNotes.map((note, index): FieldBlock => ({
    id: `scratch-${note.id}`,
    kind: 'note',
    title: 'Scratch note',
    body: note.body,
    date: note.createdAt,
    accent: '#C88A3D',
    rotation: index % 2 === 0 ? 'rotate-1' : '-rotate-1',
  }));

  const selectedNoteBlocks = selectedDetail ? selectedNotes.map((note, index): FieldBlock => ({
    id: `note-${note.id}`,
    kind: 'note',
    title: selectedDetail.feature.title,
    body: note.body,
    date: note.created_at,
    feature: selectedDetail.feature,
    accent: note.visibility === 'private' ? '#6F9FB5' : '#7EA16B',
    rotation: index % 2 === 0 ? '' : '-rotate-1',
  })) : [];

  const selectedUploadBlocks = selectedDetail ? selectedUploads.map((upload, index): FieldBlock => ({
    id: `upload-${upload.id}`,
    kind: 'photo',
    title: upload.original_filename,
    body: `${selectedDetail.feature.title} · ${formatBytes(upload.byte_size)}`,
    date: upload.created_at,
    feature: selectedDetail.feature,
    upload,
    accent: '#6F9FB5',
    rotation: index % 2 === 0 ? 'rotate-1' : '',
    wide: true,
  })) : [];

  return [...scratchBlocks, ...selectedNoteBlocks, ...selectedUploadBlocks, ...featureBlocks];
}

function filterBlocks(blocks: FieldBlock[], page: FieldPageKey, lens: FieldLens) {
  return blocks.filter(block => {
    const pageMatch = block.id.startsWith('scratch-') ? page === 'scratch' : block.feature ? featurePage(block.feature) === page || (page === 'photos' && block.upload) : page === 'scratch';
    if (!pageMatch) return false;
    if (lens === 'all') return block.feature?.status !== 'archived';
    if (lens === 'recent') {
      const date = block.date ? new Date(block.date).getTime() : 0;
      return Date.now() - date < 1000 * 60 * 60 * 24 * 14;
    }
    if (lens === 'linked') return Boolean(block.feature);
    if (lens === 'projects') return block.feature?.status === 'draft' || block.kind === 'task';
    if (lens === 'approach') return block.kind === 'trail';
    if (lens === 'photos') return block.kind === 'photo' || block.kind === 'sketch' || block.kind === 'map';
    if (lens === 'loops') return block.kind === 'task' || block.feature?.status === 'draft' || block.body.toLowerCase().includes('todo');
    return true;
  });
}

function NotebookButton({ active, children, onClick }: { active?: boolean; children: ReactNode; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`w-full rounded-2xl border px-3 py-3 text-left transition ${active ? 'border-[#7EA16B] bg-[#1B2925] text-[#F2F0E7] shadow-[0_0_0_1px_rgba(126,161,107,0.22)]' : 'border-transparent bg-transparent text-[#AAB8AD] hover:border-[#2B403A] hover:bg-[#172522]/80 hover:text-[#F2F0E7]'}`}
    >
      {children}
    </button>
  );
}

function NotebookRail({ page, features, noteCounts, uploads, onPageChange }: { page: FieldPageKey; features: MoneyFeature[]; noteCounts: Record<string, number>; uploads: Record<string, MoneyUpload>; onPageChange: (page: FieldPageKey) => void }) {
  const pageCounts = useMemo(() => pages.reduce<Record<FieldPageKey, number>>((acc, item) => {
    acc[item.id] = features.filter(feature => featurePage(feature) === item.id || (item.id === 'photos' && uploads[feature.id])).length;
    return acc;
  }, { scratch: 0, approach: 0, topos: 0, photos: 0, projects: 0 }), [features, uploads]);
  const totalNotes = Object.values(noteCounts).reduce((sum, count) => sum + count, 0);

  return (
    <aside className="hidden min-h-0 w-72 shrink-0 border-r border-[#2B403A] bg-[#0B1714] p-4 text-[#F2F0E7] lg:block">
      <div className="rounded-[1.5rem] border border-[#2B403A] bg-[#111D1B] p-4 shadow-[0_18px_60px_rgba(0,0,0,0.28)]">
        <p className="text-xs font-bold uppercase tracking-[0.22em] text-[#7EA16B]">Money Creek notebook</p>
        <h2 className="mt-2 text-3xl font-black leading-none tracking-tight">Crag notebook</h2>
        <p className="mt-3 text-sm leading-6 text-[#AAB8AD]">Notes, trails, photos, topos, sketches, and follow-ups for Money Creek.</p>
      </div>
      <div className="mt-4 space-y-2">
        {pages.map(item => {
          const Icon = item.icon;
          return (
            <NotebookButton key={item.id} active={page === item.id} onClick={() => onPageChange(item.id)}>
              <div className="flex items-center justify-between gap-3">
                <span className="flex min-w-0 items-center gap-3"><Icon className="h-5 w-5 shrink-0 text-[#7EA16B]" /><span className="truncate font-bold">{item.label}</span></span>
                <span className="rounded-full border border-[#2B403A] bg-[#0B1714]/70 px-2 py-0.5 text-xs font-bold text-[#C7D38A]">{pageCounts[item.id]}</span>
              </div>
              <p className="mt-1 line-clamp-2 text-xs leading-5 text-[#AAB8AD]">{item.copy}</p>
            </NotebookButton>
          );
        })}
      </div>
      <div className="mt-4 rounded-[1.25rem] border border-[#2B403A] bg-[#172522]/75 p-3 text-xs leading-5 text-[#AAB8AD]">
        <p><strong className="text-[#F2F0E7]">Notebook count:</strong> {features.length} map notes, {totalNotes} written notes, {Object.keys(uploads).length} photo references.</p>
      </div>
    </aside>
  );
}

function LensBar({ lens, filters, onLensChange, onFiltersChange }: { lens: FieldLens; filters: MoneyFeatureFilters; onLensChange: (lens: FieldLens) => void; onFiltersChange: (filters: MoneyFeatureFilters) => void }) {
  return (
    <div className="flex flex-col gap-3 border-b border-[#2B403A] bg-[#0B1714]/92 p-3 text-[#F2F0E7] backdrop-blur md:flex-row md:items-center md:justify-between">
      <div className="flex min-w-0 items-center gap-2 overflow-x-auto pb-1 md:pb-0">
        {lensOptions.map(option => (
          <button
            key={option.id}
            type="button"
            onClick={() => onLensChange(option.id)}
            className={`shrink-0 rounded-full border px-3 py-2 text-xs font-bold transition ${lens === option.id ? 'border-[#7EA16B] bg-[#7EA16B] text-[#07110F]' : 'border-[#2B403A] bg-[#111D1B]/80 text-[#AAB8AD] hover:border-[#6F9FB5] hover:text-[#F2F0E7]'}`}
            title={option.hint}
          >
            {option.label}
          </button>
        ))}
      </div>
      <label className="flex min-w-0 items-center gap-2 rounded-full border border-[#2B403A] bg-[#111D1B]/85 px-3 py-2 text-sm text-[#AAB8AD] md:w-72">
        <Search className="h-4 w-4 shrink-0" />
        <input
          value={filters.search ?? ''}
          onChange={event => onFiltersChange({ ...filters, search: event.target.value })}
          placeholder="Search the crag notebook"
          className="min-w-0 flex-1 bg-transparent text-[#F2F0E7] outline-none placeholder:text-[#74847B]"
        />
      </label>
    </div>
  );
}

function AddBlockStrip({ canWrite, selectedFeature, onQuickNote, onPhotoPick, onStartDrawing, onCreateScratchTask }: { canWrite: boolean; selectedFeature: MoneyFeature | null; onQuickNote: () => void; onPhotoPick: () => void; onStartDrawing: (type: MoneyFeatureType) => void; onCreateScratchTask: () => void }) {
  if (!canWrite) return null;
  return (
    <div className="flex gap-2 overflow-x-auto rounded-[1.5rem] border border-[#2B403A] bg-[#111D1B]/90 p-2 shadow-[0_18px_50px_rgba(0,0,0,0.24)]">
      {addTools.map(tool => {
        const Icon = tool.icon;
        const handleClick = () => {
          if (tool.type === 'note') onQuickNote();
          if (tool.type === 'photo') onPhotoPick();
          if (tool.type === 'sketch') onStartDrawing('drawing');
          if (tool.type === 'map') onStartDrawing('topo');
          if (tool.type === 'poi') onStartDrawing('poi');
          if (tool.type === 'trail') onStartDrawing('trail');
          if (tool.type === 'task') onCreateScratchTask();
        };
        return (
          <button key={tool.type} type="button" onClick={handleClick} className="flex shrink-0 items-center gap-2 rounded-2xl border border-[#2B403A] bg-[#172522] px-4 py-3 text-sm font-black text-[#F2F0E7] transition hover:-translate-y-0.5 hover:border-[#7EA16B] hover:bg-[#1B2925]">
            <Icon className="h-4 w-4 text-[#7EA16B]" />
            {tool.label}
          </button>
        );
      })}
      {!selectedFeature && <span className="flex shrink-0 items-center rounded-2xl px-3 text-xs font-semibold text-[#AAB8AD]">Pick a map item before adding photos or linked notes.</span>}
    </div>
  );
}

function AuthThumb({ upload }: { upload: MoneyUpload }) {
  const [src, setSrc] = useState<string | null>(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    let active = true;
    let url: string | null = null;
    moneyApi.getUploadBlobUrl(upload.id)
      .then(blobUrl => {
        if (!active) return;
        url = blobUrl;
        setSrc(blobUrl);
      })
      .catch(() => {
        if (active) setFailed(true);
      });
    return () => {
      active = false;
      if (url) URL.revokeObjectURL(url);
    };
  }, [upload.id]);

  if (failed) return <div className="flex h-36 items-center justify-center rounded-2xl border border-[#2B403A] bg-[#111D1B] text-xs font-bold text-[#AAB8AD]">Image unavailable</div>;
  if (!src) return <div className="flex h-36 items-center justify-center rounded-2xl border border-[#2B403A] bg-[#111D1B]"><Loader2 className="h-5 w-5 animate-spin text-[#7EA16B]" /></div>;
  return <img src={src} alt={upload.original_filename} width={upload.width} height={upload.height} loading="lazy" className="h-40 w-full rounded-2xl border border-[#2B403A] object-cover" />;
}

function BlockGlyph({ block }: { block: FieldBlock }) {
  if (block.feature?.feature_type === 'trail') return <Route className="h-5 w-5" />;
  if (block.feature?.feature_type === 'topo') return <Layers className="h-5 w-5" />;
  if (block.feature?.feature_type === 'drawing') return <PencilRuler className="h-5 w-5" />;
  if (block.kind === 'note') return <MessageSquarePlus className="h-5 w-5" />;
  if (block.kind === 'photo') return <Camera className="h-5 w-5" />;
  if (block.kind === 'task') return <ClipboardCheck className="h-5 w-5" />;
  return <MapPin className="h-5 w-5" />;
}

function FieldBlockRenderer({ block, selected, onSelect }: { block: FieldBlock; selected: boolean; onSelect: (block: FieldBlock) => void }) {
  return (
    <button
      type="button"
      onClick={() => onSelect(block)}
      className={`group min-h-44 rounded-[1.4rem] border bg-[#111D1B] p-4 text-left text-[#F2F0E7] shadow-[0_18px_45px_rgba(0,0,0,0.22)] transition hover:-translate-y-1 hover:border-[#7EA16B] hover:bg-[#172522] hover:shadow-[0_24px_60px_rgba(0,0,0,0.34)] ${block.rotation} ${block.wide ? 'md:col-span-2' : ''} ${selected ? 'border-[#C7D38A] ring-4 ring-[#C7D38A]/15' : 'border-[#2B403A]'}`}
      style={{ borderTopColor: block.accent, borderTopWidth: 5 }}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-center gap-2">
          <span className="flex h-10 w-10 items-center justify-center rounded-2xl border border-[#2B403A] bg-[#0B1714]/70" style={{ color: block.accent }}><BlockGlyph block={block} /></span>
          <div>
            <p className="text-[0.68rem] font-black uppercase tracking-[0.18em] text-[#AAB8AD]">{block.kind.replace('-', ' ')}</p>
            <h3 className="line-clamp-2 text-lg font-black leading-tight">{block.title}</h3>
          </div>
        </div>
        {block.feature && <MapPin className="h-4 w-4 shrink-0 text-[#7EA16B]" />}
      </div>
      {block.upload && <div className="mt-3"><AuthThumb upload={block.upload} /></div>}
      <p className="mt-3 whitespace-pre-wrap text-sm leading-6 text-[#AAB8AD]">{block.body}</p>
      <div className="mt-4 flex flex-wrap items-center gap-2 text-xs font-bold text-[#AAB8AD]">
        <span className="rounded-full border border-[#2B403A] bg-[#0B1714]/70 px-2 py-1">{formatDate(block.date)}</span>
        {block.feature && <span className="rounded-full border border-[#2B403A] bg-[#0B1714]/70 px-2 py-1">{featureTypeLabel(block.feature.feature_type)}</span>}
        {block.feature?.status === 'draft' && <span className="rounded-full border border-[#C88A3D]/50 bg-[#C88A3D]/12 px-2 py-1 text-[#E0B36F]">needs follow-up</span>}
      </div>
    </button>
  );
}

function QuickNoteSheet({ open, selectedFeature, project, canWrite, onClose, onScratchNote, onSaved }: { open: boolean; selectedFeature: MoneyFeature | null; project?: MoneyProject; canWrite: boolean; onClose: () => void; onScratchNote: (note: ScratchNote) => void; onSaved: () => void }) {
  const [body, setBody] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!open || !canWrite) return null;

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (!body.trim()) return;
    setSaving(true);
    setError(null);
    try {
      if (selectedFeature) {
        await moneyApi.createNote(selectedFeature.id, { body: body.trim(), visibility: 'team' });
        onSaved();
      } else {
        onScratchNote({ id: crypto.randomUUID(), body: body.trim(), createdAt: new Date().toISOString() });
      }
      setBody('');
      onClose();
    } catch {
      setError('Could not save this note.');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-end bg-[#07110F]/70 p-3 backdrop-blur-sm sm:items-center sm:justify-center">
      <form onSubmit={handleSubmit} className="w-full max-w-xl rounded-[1.75rem] border border-[#2B403A] bg-[#111D1B] p-4 text-[#F2F0E7] shadow-2xl">
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-xs font-black uppercase tracking-[0.18em] text-[#7EA16B]">Quick note</p>
            <h2 className="text-2xl font-black">{selectedFeature ? `Attach to ${selectedFeature.title}` : 'Scratch note'}</h2>
            <p className="mt-1 text-sm leading-6 text-[#AAB8AD]">{selectedFeature ? 'This note is saved to the selected map item.' : 'No map item selected — this remains local scratch until you select or create a map item.'}</p>
          </div>
          <button type="button" onClick={onClose} className="rounded-full border border-[#2B403A] bg-[#0B1714]/70 p-2"><X className="h-4 w-4" /></button>
        </div>
        <textarea value={body} onChange={event => setBody(event.target.value)} maxLength={5000} placeholder="Beta, access note, cleanup task, photo reminder..." className="mt-4 min-h-40 w-full resize-none rounded-2xl border border-[#2B403A] bg-[#0B1714]/80 p-3 text-sm leading-6 text-[#F2F0E7] outline-none placeholder:text-[#74847B] focus:border-[#7EA16B]" />
        {error && <p className="mt-2 rounded-xl border border-red-900/60 bg-red-950/40 px-3 py-2 text-sm font-bold text-red-200">{error}</p>}
        <div className="mt-3 flex items-center justify-between gap-3">
          <p className="text-xs font-semibold text-[#AAB8AD]">{project ? 'Money Creek notebook' : 'Notebook loading'}</p>
          <button type="submit" disabled={saving || !body.trim()} className="inline-flex items-center gap-2 rounded-2xl bg-[#7EA16B] px-5 py-3 text-sm font-black text-[#07110F] disabled:bg-[#2B403A] disabled:text-[#74847B]">
            {saving && <Loader2 className="h-4 w-4 animate-spin" />}
            Save note
          </button>
        </div>
      </form>
    </div>
  );
}

function PhotoCaptureInput({ selectedFeature, projectId, onChanged, onPrompt }: { selectedFeature: MoneyFeature | null; projectId?: string; onChanged: () => void; onPrompt: () => void }) {
  const [uploading, setUploading] = useState(false);
  const handleChange = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file) return;
    if (!selectedFeature || !projectId) {
      onPrompt();
      return;
    }
    setUploading(true);
    try {
      await moneyApi.uploadImage(projectId, file, { featureId: selectedFeature.id });
      onChanged();
    } finally {
      setUploading(false);
    }
  };

  return (
    <label className="fixed bottom-20 right-4 z-40 flex cursor-pointer items-center gap-2 rounded-full bg-[#7EA16B] px-4 py-3 text-sm font-black text-[#07110F] shadow-xl md:hidden">
      {uploading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Upload className="h-4 w-4" />}
      Photo
      <input type="file" accept="image/jpeg,image/png,image/webp" capture="environment" onChange={handleChange} disabled={uploading} className="sr-only" />
    </label>
  );
}

function FieldPageCanvas({ page, lens, blocks, selectedFeatureId, canWrite, selectedFeature, onSelectBlock, onQuickNote, onPhotoPick, onStartDrawing, onScratchTask }: { page: FieldPageKey; lens: FieldLens; blocks: FieldBlock[]; selectedFeatureId: string | null; canWrite: boolean; selectedFeature: MoneyFeature | null; onSelectBlock: (block: FieldBlock) => void; onQuickNote: () => void; onPhotoPick: () => void; onStartDrawing: (type: MoneyFeatureType) => void; onScratchTask: () => void }) {
  const pageMeta = pages.find(item => item.id === page) ?? pages[0];
  const visibleBlocks = filterBlocks(blocks, page, lens);

  return (
    <section className="min-w-0 flex-1 overflow-y-auto bg-[#07110F] p-3 custom-scrollbar lg:max-h-[calc(100vh-2rem)]">
      <div className="mx-auto max-w-5xl">
        <div className="rounded-[2rem] border border-[#2B403A] bg-[#111D1B] p-4 shadow-[0_20px_70px_rgba(0,0,0,0.3)] sm:p-6">
          <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
            <div>
              <p className="text-xs font-black uppercase tracking-[0.22em] text-[#7EA16B]">Crag notebook</p>
              <h1 className="mt-1 text-4xl font-black tracking-tight text-[#F2F0E7] sm:text-6xl">{pageMeta.label}</h1>
              <p className="mt-3 max-w-2xl text-sm leading-7 text-[#AAB8AD]">{pageMeta.copy}</p>
            </div>
            <div className="rounded-[1.25rem] border border-[#2B403A] bg-[#172522]/70 p-3 text-sm text-[#AAB8AD] md:w-64">
              <p className="font-black text-[#F2F0E7]">Selected map context</p>
              <p className="mt-1 leading-6">Pick a map item to add notes, photos, or cleanup details to it.</p>
            </div>
          </div>
          <div className="mt-5">
            <AddBlockStrip canWrite={canWrite} selectedFeature={selectedFeature} onQuickNote={onQuickNote} onPhotoPick={onPhotoPick} onStartDrawing={onStartDrawing} onCreateScratchTask={onScratchTask} />
          </div>
        </div>

        <div className="mt-4 grid auto-rows-fr gap-4 md:grid-cols-2 xl:grid-cols-3">
          {visibleBlocks.map(block => <FieldBlockRenderer key={block.id} block={block} selected={Boolean(block.feature && block.feature.id === selectedFeatureId)} onSelect={onSelectBlock} />)}
        </div>

        {visibleBlocks.length === 0 && (
          <div className="mt-4 rounded-[2rem] border border-dashed border-[#2B403A] bg-[#111D1B]/75 p-8 text-center text-[#AAB8AD]">
            <Sparkles className="mx-auto h-10 w-10 text-[#C7D38A]" />
            <h2 className="mt-3 text-2xl font-black text-[#F2F0E7]">Nothing here yet</h2>
            <p className="mt-2 text-sm leading-6">Switch lenses, jot a note, or draw a trail/topo to start this notebook page.</p>
          </div>
        )}
      </div>
    </section>
  );
}

function MapContextDrawer({ project, features, selectedFeature, selectedFeatureId, drawingType, draftPoints, canWrite, projectId, detailQuery, filters, onSelectFeature, onAddDraftPoint, onStartDrawing, onUndoDraftPoint, onCancelDrawing, onFinishDrawing, onFeatureSaved, onFeatureArchived, onChanged, onClearSelection, onFiltersChange, onExpandMap }: { project?: MoneyProject; features: MoneyFeature[]; selectedFeature: MoneyFeature | null; selectedFeatureId: string | null; drawingType: MoneyFeatureType | null; draftPoints: MoneyPosition[]; canWrite: boolean; projectId?: string; detailQuery: UseQueryResult<MoneyFeatureDetail>; filters: MoneyFeatureFilters; onSelectFeature: (feature: MoneyFeature) => void; onAddDraftPoint: (point: MoneyPosition) => void; onStartDrawing: (type: MoneyFeatureType) => void; onUndoDraftPoint: () => void; onCancelDrawing: () => void; onFinishDrawing: () => void; onFeatureSaved: (feature: MoneyFeature) => void; onFeatureArchived: () => void; onChanged: () => void; onClearSelection: () => void; onFiltersChange: (filters: MoneyFeatureFilters) => void; onExpandMap: () => void }) {
  const detailNotes = detailQuery.data?.notes ?? [];
  const detailUploads = detailQuery.data?.uploads ?? [];
  const selectedCoordinates = selectedFeature ? getFeatureCoordinates(selectedFeature)[0] : null;

  return (
    <aside className="hidden w-[34rem] shrink-0 overflow-y-auto border-l border-[#2B403A] bg-[#0B1714] p-4 text-[#F2F0E7] custom-scrollbar xl:block xl:max-h-[calc(100vh-2rem)] 2xl:w-[40rem]">
      <div className="rounded-[1.75rem] border border-[#2B403A] bg-[#111D1B] p-4 shadow-[0_18px_60px_rgba(0,0,0,0.24)]">
        <div className="mb-4 flex items-start justify-between gap-3">
          <div>
            <p className="text-xs font-black uppercase tracking-[0.18em] text-[#7EA16B]">Map context</p>
            <h2 className="text-3xl font-black">Money Creek map</h2>
            <p className="mt-1 text-sm leading-6 text-[#AAB8AD]">Pan, zoom, draw, and refocus around the Money Creek drainage.</p>
          </div>
          <button type="button" onClick={onExpandMap} className="inline-flex shrink-0 items-center gap-2 rounded-full border border-[#2B403A] bg-[#172522] px-3 py-2 text-xs font-black text-[#F2F0E7] hover:border-[#7EA16B]">
            <Maximize2 className="h-4 w-4 text-[#7EA16B]" /> Expand
          </button>
        </div>
        <div className="relative h-[32rem] overflow-hidden rounded-[1.5rem] border border-[#2B403A] bg-[#07110F] 2xl:h-[38rem]">
          {project ? <MoneyMap project={project} features={features} selectedFeatureId={selectedFeatureId} drawingType={drawingType} draftPoints={draftPoints} focusMode onSelectFeature={onSelectFeature} onAddDraftPoint={onAddDraftPoint} /> : <div className="flex h-full items-center justify-center text-[#F2F0E7]"><Loader2 className="h-5 w-5 animate-spin" /></div>}
          <DrawingToolbar drawingType={drawingType} draftPoints={draftPoints} canWrite={canWrite} onStart={onStartDrawing} onUndo={onUndoDraftPoint} onCancel={onCancelDrawing} onFinish={onFinishDrawing} />
        </div>
        <div className="mt-3 grid gap-2 text-xs font-bold text-[#AAB8AD] sm:grid-cols-3">
          <div className="rounded-2xl border border-[#2B403A] bg-[#0B1714]/70 p-3"><span className="block text-lg font-black text-[#F2F0E7]">{features.length}</span>visible features</div>
          <div className="rounded-2xl border border-[#2B403A] bg-[#0B1714]/70 p-3"><span className="block text-lg font-black text-[#F2F0E7]">{detailNotes.length}</span>selected notes</div>
          <div className="rounded-2xl border border-[#2B403A] bg-[#0B1714]/70 p-3"><span className="block text-lg font-black text-[#F2F0E7]">{detailUploads.length}</span>selected photos</div>
        </div>
        {selectedFeature && (
          <div className="mt-3 rounded-[1.25rem] border border-[#2B403A] bg-[#172522]/75 p-3 text-sm leading-6 text-[#AAB8AD]">
            <p className="font-black text-[#F2F0E7]">Selected feature summary</p>
            <p className="mt-1"><span className="font-bold text-[#C7D38A]">{selectedFeature.title}</span> · {featureTypeLabel(selectedFeature.feature_type)} · {selectedFeature.status}</p>
            {selectedCoordinates && <p className="text-xs font-semibold">{selectedCoordinates[1].toFixed(5)}, {selectedCoordinates[0].toFixed(5)}</p>}
          </div>
        )}
        <div className="mt-3 grid grid-cols-2 gap-2">
          <select value={filters.type ?? 'all'} onChange={event => onFiltersChange({ ...filters, type: event.target.value as MoneyFeatureType | 'all' })} className="rounded-xl border border-[#2B403A] bg-[#0B1714]/80 px-3 py-2 text-sm font-bold text-[#F2F0E7] outline-none">
            <option value="all">All types</option>
            <option value="trail">Trails</option>
            <option value="topo">Topos</option>
            <option value="poi">Map pins</option>
            <option value="drawing">Sketches</option>
          </select>
          <select value={filters.status ?? 'all'} onChange={event => onFiltersChange({ ...filters, status: event.target.value as MoneyFeatureFilters['status'] })} className="rounded-xl border border-[#2B403A] bg-[#0B1714]/80 px-3 py-2 text-sm font-bold text-[#F2F0E7] outline-none">
            <option value="all">All status</option>
            <option value="draft">Draft</option>
            <option value="active">Active</option>
            <option value="archived">Archived</option>
          </select>
        </div>
      </div>

      <div className="mt-3 rounded-[1.5rem] border border-[#2B403A] bg-[#111D1B] p-3 shadow-[0_18px_60px_rgba(0,0,0,0.24)]">
        {selectedFeature ? (
          <>
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="text-xs font-black uppercase tracking-[0.18em] text-[#7EA16B]">Selected note</p>
                <h3 className="text-2xl font-black leading-tight">{selectedFeature.title}</h3>
                <p className="mt-1 text-sm font-semibold text-[#AAB8AD]">{featureTypeLabel(selectedFeature.feature_type)} · {selectedFeature.status}</p>
              </div>
              <button type="button" onClick={onClearSelection} className="rounded-full border border-[#2B403A] bg-[#0B1714]/70 p-2"><X className="h-4 w-4" /></button>
            </div>
            {detailQuery.isLoading ? <div className="mt-3 rounded-2xl border border-[#2B403A] bg-[#0B1714]/70 p-4 text-sm font-bold text-[#AAB8AD]"><Loader2 className="mr-2 inline h-4 w-4 animate-spin" />Loading selected context</div> : (
              <div className="mt-3 space-y-3">
                <FeatureEditor feature={selectedFeature} canWrite={canWrite} onSaved={onFeatureSaved} onArchived={onFeatureArchived} />
                <NotesPanel featureId={selectedFeature.id} notes={detailNotes} canWrite={canWrite} onChanged={onChanged} />
                {projectId && <ImageUploader projectId={projectId} featureId={selectedFeature.id} uploads={detailUploads} canWrite={canWrite} onUploaded={onChanged} onDeleted={onChanged} />}
              </div>
            )}
          </>
        ) : (
          <div className="rounded-[1.25rem] border border-dashed border-[#2B403A] bg-[#0B1714]/50 p-5 text-center">
            <CircleDot className="mx-auto h-10 w-10 text-[#7EA16B]" />
            <h3 className="mt-3 text-xl font-black">Select a note or map item</h3>
            <p className="mt-2 text-sm leading-6 text-[#AAB8AD]">Notes and photos attach to selected map items. Use Scratch only for local loose jots.</p>
          </div>
        )}
      </div>
    </aside>
  );
}

function MobileCaptureSheet({ page, onPageChange, onAdd, onMap, lens, onLensChange }: { page: FieldPageKey; onPageChange: (page: FieldPageKey) => void; onAdd: () => void; onMap: () => void; lens: FieldLens; onLensChange: (lens: FieldLens) => void }) {
  return (
    <div className="fixed inset-x-2 bottom-2 z-40 rounded-[1.5rem] border border-[#2B403A] bg-[#111D1B]/95 p-2 text-[#F2F0E7] shadow-2xl backdrop-blur lg:hidden">
      <div className="grid grid-cols-5 gap-1">
        <button type="button" onClick={() => onPageChange(page === 'scratch' ? 'approach' : 'scratch')} className="rounded-2xl px-2 py-2 text-xs font-black"><BookOpen className="mx-auto h-5 w-5 text-[#7EA16B]" />Pages</button>
        <button type="button" onClick={onAdd} className="rounded-2xl bg-[#7EA16B] px-2 py-2 text-xs font-black text-[#07110F]"><Plus className="mx-auto h-5 w-5" />Add</button>
        <button type="button" onClick={onMap} className="rounded-2xl px-2 py-2 text-xs font-black"><MapIcon className="mx-auto h-5 w-5 text-[#6F9FB5]" />Map</button>
        <label className="relative rounded-2xl px-2 py-2 text-center text-xs font-black"><ChevronDown className="mx-auto h-5 w-5 text-[#C88A3D]" />Lens<select value={lens} onChange={event => onLensChange(event.target.value as FieldLens)} className="absolute inset-0 opacity-0">{lensOptions.map(option => <option key={option.id} value={option.id}>{option.label}</option>)}</select></label>
        <button type="button" onClick={() => onPageChange('photos')} className="rounded-2xl px-2 py-2 text-xs font-black"><ImageIcon className="mx-auto h-5 w-5 text-[#C7D38A]" />Photos</button>
      </div>
    </div>
  );
}

export function FieldPages(props: FieldPagesProps) {
  const [scratchNotes, setScratchNotes] = useState<ScratchNote[]>([]);
  const [quickNoteOpen, setQuickNoteOpen] = useState(false);
  const [photoPrompt, setPhotoPrompt] = useState(false);
  const [mobileMapOpen, setMobileMapOpen] = useState(false);
  const [uploadingPhoto, setUploadingPhoto] = useState(false);

  const allBlocks = useMemo(() => buildBlocks(props.features, props.noteCounts, props.primaryUploads, scratchNotes, props.detailQuery.data), [props.features, props.noteCounts, props.primaryUploads, scratchNotes, props.detailQuery.data]);

  const handleSelectBlock = (block: FieldBlock) => {
    if (block.feature) props.onSelectFeature(block.feature);
  };

  const handlePhotoPick = () => {
    if (!props.selectedFeature) {
      setPhotoPrompt(true);
      return;
    }
    const uploadInput = document.getElementById('money-selected-photo-input') as HTMLInputElement | null;
    uploadInput?.click();
  };

  const handleHiddenPhotoChange = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file) return;
    if (!props.selectedFeature || !props.projectId) {
      setPhotoPrompt(true);
      return;
    }
    setUploadingPhoto(true);
    try {
      await moneyApi.uploadImage(props.projectId, file, { featureId: props.selectedFeature.id });
      props.onChanged();
    } finally {
      setUploadingPhoto(false);
    }
  };

  const handleScratchTask = () => {
    setScratchNotes(current => [{ id: crypto.randomUUID(), body: 'Follow up: add beta, link this to the map, or turn it into a trail/topo note.', createdAt: new Date().toISOString() }, ...current]);
    props.onPageChange('scratch');
  };

  return (
    <main className="min-h-screen bg-[#07110F] text-[#F2F0E7]">
      <header className="border-b border-[#2B403A] bg-[#07110F]/95 p-3 backdrop-blur">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div className="flex min-w-0 items-start gap-3">
            <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-[1.25rem] bg-[#111D1B] text-[#F2F0E7] ring-1 ring-[#2B403A]"><Mountain className="h-7 w-7" /></div>
            <div className="min-w-0">
              <p className="text-xs font-black uppercase tracking-[0.24em] text-[#7EA16B]">Money Creek crag notebook</p>
              <h1 className="truncate text-2xl font-black tracking-tight sm:text-4xl">Crag notes for wet basalt days</h1>
              <p className="mt-1 flex flex-wrap items-center gap-2 text-sm font-semibold text-[#AAB8AD]"><Users className="h-4 w-4" />{props.user?.display_name} · {props.user?.role}</p>
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            {!props.isOnline && <span className="rounded-full border border-[#C88A3D]/50 bg-[#C88A3D]/12 px-3 py-2 text-xs font-black text-[#E0B36F]">Offline</span>}
            {props.loadError && <span className="rounded-full border border-red-900/60 bg-red-950/40 px-3 py-2 text-xs font-black text-red-200"><AlertTriangle className="mr-1 inline h-4 w-4" />Load issue</span>}
            {(props.savingFeature || uploadingPhoto) && <span className="rounded-full bg-[#7EA16B] px-3 py-2 text-xs font-black text-[#07110F]"><Loader2 className="mr-1 inline h-4 w-4 animate-spin" />{uploadingPhoto ? 'Uploading' : 'Saving'}</span>}
            <button type="button" onClick={props.onRefresh} className="rounded-full border border-[#2B403A] bg-[#111D1B] px-4 py-2 text-sm font-black text-[#F2F0E7] hover:border-[#7EA16B]"><RefreshCw className={`mr-2 inline h-4 w-4 ${props.isFetching ? 'animate-spin' : ''}`} />Refresh</button>
            <a href="/" className="rounded-full border border-[#2B403A] bg-[#111D1B] px-4 py-2 text-sm font-black text-[#F2F0E7] hover:border-[#7EA16B]"><Navigation className="mr-2 inline h-4 w-4" />Home</a>
            <button type="button" onClick={props.onLogout} className="rounded-full border border-[#2B403A] bg-[#111D1B] px-4 py-2 text-sm font-black text-[#F2F0E7] hover:border-[#C88A3D]"><LogOut className="mr-2 inline h-4 w-4" />Sign out</button>
          </div>
        </div>
      </header>

      <section className="flex min-h-[calc(100vh-5.5rem)] overflow-hidden">
        <NotebookRail page={props.page} features={props.features} noteCounts={props.noteCounts} uploads={props.primaryUploads} onPageChange={props.onPageChange} />
        <div className="flex min-w-0 flex-1 flex-col">
          <LensBar lens={props.lens} filters={props.filters} onLensChange={props.onLensChange} onFiltersChange={props.onFiltersChange} />
          <FieldPageCanvas page={props.page} lens={props.lens} blocks={allBlocks} selectedFeatureId={props.selectedFeatureId} canWrite={props.canWrite} selectedFeature={props.selectedFeature} onSelectBlock={handleSelectBlock} onQuickNote={() => setQuickNoteOpen(true)} onPhotoPick={handlePhotoPick} onStartDrawing={props.onStartDrawing} onScratchTask={handleScratchTask} />
        </div>
        <MapContextDrawer project={props.project} features={props.visibleMapFeatures} selectedFeature={props.selectedFeature} selectedFeatureId={props.selectedFeatureId} drawingType={props.drawingType} draftPoints={props.draftPoints} canWrite={props.canWrite} projectId={props.projectId} detailQuery={props.detailQuery} filters={props.filters} onSelectFeature={props.onSelectFeature} onAddDraftPoint={props.onAddDraftPoint} onStartDrawing={props.onStartDrawing} onUndoDraftPoint={props.onUndoDraftPoint} onCancelDrawing={props.onCancelDrawing} onFinishDrawing={props.onFinishDrawing} onFeatureSaved={props.onFeatureSaved} onFeatureArchived={props.onFeatureArchived} onChanged={props.onChanged} onClearSelection={props.onClearSelection} onFiltersChange={props.onFiltersChange} onExpandMap={() => setMobileMapOpen(true)} />
      </section>

      {mobileMapOpen && (
        <div className="fixed inset-0 z-50 bg-[#07110F]/70 p-3 backdrop-blur-sm xl:hidden">
          <div className="relative h-full overflow-hidden rounded-[1.75rem] border border-[#2B403A] bg-[#07110F]">
            <button type="button" onClick={() => setMobileMapOpen(false)} className="absolute right-3 top-3 z-20 rounded-full bg-[#111D1B] p-2 text-[#F2F0E7]"><X className="h-4 w-4" /></button>
            {props.project && <MoneyMap project={props.project} features={props.visibleMapFeatures} selectedFeatureId={props.selectedFeatureId} drawingType={props.drawingType} draftPoints={props.draftPoints} focusMode onSelectFeature={feature => { props.onSelectFeature(feature); setMobileMapOpen(false); }} onAddDraftPoint={props.onAddDraftPoint} />}
            <DrawingToolbar drawingType={props.drawingType} draftPoints={props.draftPoints} canWrite={props.canWrite} onStart={props.onStartDrawing} onUndo={props.onUndoDraftPoint} onCancel={props.onCancelDrawing} onFinish={props.onFinishDrawing} />
          </div>
        </div>
      )}

      <QuickNoteSheet open={quickNoteOpen} selectedFeature={props.selectedFeature} project={props.project} canWrite={props.canWrite} onClose={() => setQuickNoteOpen(false)} onScratchNote={note => setScratchNotes(current => [note, ...current])} onSaved={props.onChanged} />
      {photoPrompt && <div className="fixed inset-x-3 bottom-24 z-50 rounded-[1.5rem] border border-[#2B403A] bg-[#111D1B] p-4 text-[#F2F0E7] shadow-xl md:left-auto md:right-4 md:w-96"><div className="flex items-start gap-3"><FileImage className="h-5 w-5 text-[#C88A3D]" /><div><p className="font-black">Select or create a map item first</p><p className="mt-1 text-sm leading-6 text-[#AAB8AD]">Photos need a selected map item. Pick a boulder, trail, topo, or point first.</p></div><button type="button" onClick={() => setPhotoPrompt(false)} className="ml-auto"><X className="h-4 w-4" /></button></div></div>}
      <input id="money-selected-photo-input" type="file" accept="image/jpeg,image/png,image/webp" capture="environment" onChange={handleHiddenPhotoChange} disabled={uploadingPhoto} className="sr-only" />
      <PhotoCaptureInput selectedFeature={props.selectedFeature} projectId={props.projectId} onChanged={props.onChanged} onPrompt={() => setPhotoPrompt(true)} />
      <MobileCaptureSheet page={props.page} onPageChange={props.onPageChange} onAdd={() => setQuickNoteOpen(true)} onMap={() => setMobileMapOpen(true)} lens={props.lens} onLensChange={props.onLensChange} />
      <div className="h-24 lg:hidden" />
    </main>
  );
}
