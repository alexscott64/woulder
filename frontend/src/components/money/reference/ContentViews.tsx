import { useEffect, useMemo, useRef, useState } from 'react';
import { Camera, FileText, Plus, RotateCcw, Route, Upload } from 'lucide-react';
import { MoneyCragNode, MoneyNote, MoneyTrashItem, MoneyTrailCategory, MoneyUpload, MoneyUploadBlockKind } from '../../../types/money';
import { cragChildren, cragProblems, findNode, flattenAreas, flattenBoulders, flattenProblems, pathTo, problemMeta } from './model';
import { DEV, T } from './theme';
import { PhotoLightbox, PhotoLightboxItem, UploadPhotoButton, useUploadImageUrl } from './PhotoLightbox';
import { ParsedTrailImport, parseTrailFile } from './trailImport';
import { TopoOverlay } from './topoOverlay';

type OpenPhoto = (item: PhotoLightboxItem) => void;
export type ReferenceFilterView = 'problems' | 'boulders' | 'trails' | 'photos' | 'notes';
export type PhotoFilter = 'all' | 'with' | 'without';
export type SortFilter = 'title' | 'area' | 'status' | 'newest' | 'oldest' | 'grade' | 'category' | 'destination';
export type TargetFilter = 'all' | 'project' | 'area' | 'boulder' | 'problem' | 'trail' | 'none';

export interface ReferenceFilters {
  search: string;
  areaId: string;
  subAreaId: string;
  photo: PhotoFilter;
  status: string;
  trailCategory: 'all' | MoneyTrailCategory;
  destination: string;
  uploadKind: 'all' | MoneyUploadBlockKind;
  target: TargetFilter;
  grade: string;
  boulderId: string;
  sort: SortFilter;
}

export const defaultReferenceFilters: ReferenceFilters = {
  search: '',
  areaId: 'all',
  subAreaId: 'all',
  photo: 'all',
  status: 'all',
  trailCategory: 'all',
  destination: 'all',
  uploadKind: 'all',
  target: 'all',
  grade: 'all',
  boulderId: 'all',
  sort: 'title',
};

interface ContentViewProps {
  view: string;
  root: MoneyCragNode | null;
  trails: MoneyCragNode[];
  notes: MoneyNote[];
  uploads: MoneyUpload[];
  trash: MoneyTrashItem[];
  canWrite: boolean;
  mobile: boolean;
  filters?: ReferenceFilters;
  setFilters?: (filters: ReferenceFilters) => void;
  currentAreaId?: string | null;
  openBoulder: (id: string) => void;
  selectTrail: (id: string) => void;
  onAddBoulder?: (parentId?: string | null) => void;
  onOpenComposer: (kind?: 'photo' | 'sketch' | 'file' | null) => void;
  onEditNote: (note: MoneyNote) => void;
  onDeleteNote: (note: MoneyNote) => void;
  onDeleteUpload: (upload: MoneyUpload) => void;
  onUpdateUploadMetadata?: (upload: MoneyUpload, metadata: { title: string | null; comments: string | null }) => Promise<MoneyUpload | void> | MoneyUpload | void;
  onRestore: (id: string) => void;
  onCreateTrail: (trail: ParsedTrailImport & { filename: string }) => void;
  onUpdateTrail: (trail: MoneyCragNode, updates: { title: string; category: MoneyTrailCategory; destinationFeatureId?: string | null; destinationLabel?: string | null }) => void;
  onSaveTopo?: (problem: MoneyCragNode, overlay: TopoOverlay) => void;
}

export function isReferenceFilterView(view: string): view is ReferenceFilterView {
  return view === 'problems' || view === 'boulders' || view === 'trails' || view === 'photos' || view === 'notes';
}

export function ContentView({ view, root, trails, notes, uploads, trash, canWrite, mobile, filters = defaultReferenceFilters, setFilters, currentAreaId, openBoulder, selectTrail, onAddBoulder, onOpenComposer, onEditNote, onDeleteNote, onDeleteUpload, onUpdateUploadMetadata, onRestore, onCreateTrail, onUpdateTrail, onSaveTopo }: ContentViewProps) {
  const [selectedPhoto, setSelectedPhoto] = useState<PhotoLightboxItem | null>(null);
  return <main style={{ flex: 1, minWidth: 0, overflowY: 'auto', background: T.app }}>
    <div style={{ maxWidth: 1080, margin: '0 auto', padding: mobile ? '20px 18px 90px' : '28px 32px 60px' }}>
      {mobile && isReferenceFilterView(view) && setFilters && <div style={{ ...card, padding: 12, marginBottom: 18 }}><ReferenceFilterRail view={view} root={root} trails={trails} notes={notes} uploads={uploads} filters={filters} setFilters={setFilters} currentAreaId={currentAreaId} /></div>}
      {view === 'problems' && <Problems root={root} notes={notes} uploads={uploads} filters={filters} openBoulder={openBoulder} onOpenPhoto={setSelectedPhoto} />}
      {view === 'boulders' && <Boulders root={root} notes={notes} uploads={uploads} filters={filters} currentAreaId={currentAreaId} canWrite={canWrite} openBoulder={openBoulder} onAddBoulder={onAddBoulder} onOpenPhoto={setSelectedPhoto} />}
      {view === 'trails' && <Trails root={root} trails={trails} filters={filters} canWrite={canWrite} selectTrail={selectTrail} onCreateTrail={onCreateTrail} onUpdateTrail={onUpdateTrail} />}
      {view === 'photos' && <Photos root={root} notes={notes} uploads={uploads} filters={filters} canWrite={canWrite} onOpenComposer={onOpenComposer} onOpenPhoto={setSelectedPhoto} onDeleteUpload={onDeleteUpload} />}
      {view === 'notes' && <Notes root={root} notes={notes} uploads={uploads} filters={filters} canWrite={canWrite} onOpenComposer={onOpenComposer} onOpenPhoto={setSelectedPhoto} onEditNote={onEditNote} onDeleteNote={onDeleteNote} onDeleteUpload={onDeleteUpload} />}
      {view === 'trash' && <Trash items={trash} canWrite={canWrite} onRestore={onRestore} />}
    </div>
    {selectedPhoto && <PhotoLightbox item={selectedPhoto} root={root} canWrite={canWrite} canDelete={canWrite} onUpdateMetadata={onUpdateUploadMetadata} onSaveTopo={onSaveTopo} onDelete={upload => { onDeleteUpload(upload); setSelectedPhoto(null); }} onClose={() => setSelectedPhoto(null)} />}
  </main>;
}

export function ReferenceFilterRail({ view, root, trails, notes, uploads, filters, setFilters, currentAreaId }: { view: ReferenceFilterView; root: MoneyCragNode | null; trails: MoneyCragNode[]; notes: MoneyNote[]; uploads: MoneyUpload[]; filters: ReferenceFilters; setFilters: (filters: ReferenceFilters) => void; currentAreaId?: string | null }) {
  const areas = useMemo(() => flattenAreas(root), [root]);
  const boulders = useMemo(() => flattenBoulders(root), [root]);
  const problems = useMemo(() => flattenProblems(root), [root]);
  const grades = useMemo(() => [...new Set(problems.map(p => problemMeta(p.feature).grade).filter(Boolean))].sort(gradeSort), [problems]);
  const subAreaParentId = subAreaParentFilterId(filters, currentAreaId ?? root?.feature.id);
  const subAreas = useMemo(() => { const parent = findNode(root, subAreaParentId); return parent?.feature.feature_type === 'area' ? cragChildren(parent) : []; }, [root, subAreaParentId]);
  const patch = (updates: Partial<ReferenceFilters>) => setFilters({ ...filters, ...updates });
  const reset = () => setFilters({ ...defaultReferenceFilters, sort: defaultSortForView(view) });
  const title = `${view[0].toUpperCase()}${view.slice(1)} filters`;

  useEffect(() => {
    if (filters.sort === defaultReferenceFilters.sort && defaultSortForView(view) !== defaultReferenceFilters.sort) patch({ sort: defaultSortForView(view) });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [view]);

  useEffect(() => {
    if (filters.subAreaId !== 'all' && !subAreas.some(area => area.feature.id === filters.subAreaId)) patch({ subAreaId: 'all' });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filters.subAreaId, subAreas]);

  return <div data-testid="reference-filter-rail">
    <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '4px 6px 10px' }}>
      <div style={{ flex: 1, fontFamily: T.mono, fontSize: 10.5, letterSpacing: 1, color: T.faint, textTransform: 'uppercase' }}>{title}</div>
      <button type="button" onClick={reset} style={tinyBtn}>Reset</button>
    </div>
    <label style={filterLabel}>Search<input aria-label={`${title} search`} value={filters.search} onChange={e => patch({ search: e.target.value })} placeholder="Search list" style={filterField} /></label>
    {usesAreaScope(view) && <label style={filterLabel}>Area<select aria-label={`${title} area`} value={filters.areaId} onChange={e => patch({ areaId: e.target.value, subAreaId: 'all' })} style={filterField}><option value="all">{view === 'boulders' && currentAreaId ? 'Current area' : 'All areas'}</option>{areas.map(area => <option key={area.feature.id} value={area.feature.id}>{area.feature.title}</option>)}</select></label>}
    {usesAreaScope(view) && subAreas.length > 0 && <label style={filterLabel}>Sub-area<select aria-label={`${title} sub-area`} value={filters.subAreaId} onChange={e => patch({ subAreaId: e.target.value })} style={filterField}><option value="all">All sub-areas</option>{subAreas.map(area => <option key={area.feature.id} value={area.feature.id}>{area.feature.title}</option>)}</select></label>}
    {view === 'boulders' && <>
      <label style={filterLabel}>Status<select aria-label="Boulder status filter" value={filters.status} onChange={e => patch({ status: e.target.value })} style={filterField}><option value="all">Any status</option>{Object.entries(DEV.meta).map(([value, meta]) => <option key={value} value={value}>{meta.label}</option>)}</select></label>
      <PhotoSelect value={filters.photo} onChange={photo => patch({ photo })} label="Boulder photo filter" />
      <SortSelect value={filters.sort} onChange={sort => patch({ sort })} options={[['title', 'Name'], ['area', 'Area'], ['status', 'Status']]} />
    </>}
    {view === 'problems' && <>
      <label style={filterLabel}>Boulder<select aria-label="Problem boulder filter" value={filters.boulderId} onChange={e => patch({ boulderId: e.target.value })} style={filterField}><option value="all">All boulders</option>{boulders.map(boulder => <option key={boulder.feature.id} value={boulder.feature.id}>{boulder.feature.title}</option>)}</select></label>
      <label style={filterLabel}>Status<select aria-label="Problem status filter" value={filters.status} onChange={e => patch({ status: e.target.value })} style={filterField}><option value="all">Any status</option><option value="project">Project</option><option value="sent">Sent</option><option value="established">Established</option></select></label>
      <label style={filterLabel}>Grade<select aria-label="Problem grade filter" value={filters.grade} onChange={e => patch({ grade: e.target.value })} style={filterField}><option value="all">Any grade</option>{grades.map(grade => <option key={grade} value={grade}>{grade}</option>)}</select></label>
      <PhotoSelect value={filters.photo} onChange={photo => patch({ photo })} label="Problem photo filter" />
      <SortSelect value={filters.sort} onChange={sort => patch({ sort })} options={[['grade', 'Grade'], ['title', 'Name'], ['area', 'Area'], ['status', 'Status']]} />
    </>}
    {view === 'trails' && <>
      <label style={filterLabel}>Category<select aria-label="Trail category filter" value={filters.trailCategory} onChange={e => patch({ trailCategory: e.target.value as ReferenceFilters['trailCategory'] })} style={filterField}><option value="all">Any category</option>{trailCategoryOptions.map(option => <option key={option.value} value={option.value}>{option.label}</option>)}</select></label>
      <label style={filterLabel}>Destination<select aria-label="Trail destination filter" value={filters.destination} onChange={e => patch({ destination: e.target.value })} style={filterField}><option value="all">Any destination</option><option value="set">Has destination</option><option value="unset">No destination</option>{areas.map(area => <option key={area.feature.id} value={area.feature.id}>{area.feature.title}</option>)}</select></label>
      <SortSelect value={filters.sort} onChange={sort => patch({ sort })} options={[['title', 'Name'], ['category', 'Category'], ['destination', 'Destination']]} />
    </>}
    {view === 'photos' && <>
      <label style={filterLabel}>Kind<select aria-label="Photo kind filter" value={filters.uploadKind} onChange={e => patch({ uploadKind: e.target.value as ReferenceFilters['uploadKind'] })} style={filterField}><option value="all">Any kind</option><option value="photo">Photo</option><option value="sketch">Sketch</option><option value="topo">Topo</option><option value="file">File</option></select></label>
      <label style={filterLabel}>Target<select aria-label="Photo target filter" value={filters.target} onChange={e => patch({ target: e.target.value as TargetFilter })} style={filterField}><option value="all">Any target</option><option value="area">Area</option><option value="boulder">Boulder</option><option value="problem">Problem</option><option value="trail">Trail</option><option value="project">Project</option><option value="none">No target</option></select></label>
      <SortSelect value={filters.sort} onChange={sort => patch({ sort })} options={[['newest', 'Newest'], ['oldest', 'Oldest'], ['title', 'Filename']]} />
    </>}
    {view === 'notes' && <>
      <PhotoSelect value={filters.photo} onChange={photo => patch({ photo })} label="Note photo filter" />
      <label style={filterLabel}>Target<select aria-label="Note target filter" value={filters.target} onChange={e => patch({ target: e.target.value as TargetFilter })} style={filterField}><option value="all">Any target</option><option value="area">Area</option><option value="boulder">Boulder</option><option value="problem">Problem</option><option value="trail">Trail</option><option value="project">Project</option><option value="none">No target</option></select></label>
      <SortSelect value={filters.sort} onChange={sort => patch({ sort })} options={[['newest', 'Newest'], ['oldest', 'Oldest']]} />
    </>}
    <div style={{ padding: '6px 8px', color: T.faint, fontFamily: T.mono, fontSize: 10 }}>{summaryForView(view, root, trails, notes, uploads)}</div>
  </div>;
}

function Head({ title, sub, action }: { title: string; sub?: string; action?: React.ReactNode }) { return <div style={{ display: 'flex', alignItems: 'flex-end', marginBottom: 20, gap: 12, flexWrap: 'wrap' }}><div><h1 style={{ margin: 0, fontSize: 26, fontWeight: 800, letterSpacing: -0.6, color: T.ink }}>{title}</h1>{sub && <div style={{ fontSize: 13, color: T.mut, marginTop: 4 }}>{sub}</div>}</div>{action && <div style={{ marginLeft: 'auto' }}>{action}</div>}</div>; }
const card: React.CSSProperties = { background: T.surf, border: `1px solid ${T.line}`, borderRadius: 14 };
const miniBtn: React.CSSProperties = { border: `1px solid ${T.line2}`, background: T.inset, color: T.mut, borderRadius: 6, cursor: 'pointer', fontSize: 10.5, padding: '3px 6px', fontWeight: 800 };
const tinyBtn: React.CSSProperties = { border: `1px solid ${T.line2}`, background: T.inset, color: T.mut, borderRadius: 6, cursor: 'pointer', fontSize: 10, padding: '3px 6px', fontWeight: 800 };
const filterLabel: React.CSSProperties = { display: 'grid', gap: 5, padding: '0 6px 10px', color: T.mut, fontFamily: T.mono, fontSize: 10.5, textTransform: 'uppercase', letterSpacing: 0.6 };
const filterField: React.CSSProperties = { width: '100%', background: T.surf2, border: `1px solid ${T.line}`, borderRadius: 8, padding: '8px 9px', color: T.ink, fontFamily: T.font, fontSize: 12.5, outline: 'none', minHeight: 36, textTransform: 'none', letterSpacing: 0 };

function Problems({ root, notes, uploads, filters, openBoulder, onOpenPhoto }: { root: MoneyCragNode | null; notes: MoneyNote[]; uploads: MoneyUpload[]; filters: ReferenceFilters; openBoulder: (id: string) => void; onOpenPhoto: OpenPhoto }) {
  const all = useMemo(() => filterProblems(root, notes, uploads, filters), [root, notes, uploads, filters]);
  return <><Head title="Problems" sub={`${all.length} matching lines across every area.`} /><div style={{ ...card, overflow: 'hidden' }}>{all.map((p, i) => { const meta = problemMeta(p.feature); const upload = primaryUploadForFeature(p.feature.id, uploads, notes); return <div key={p.feature.id} onClick={() => openBoulder(p.boulder.feature.id)} style={{ display: 'grid', gridTemplateColumns: '56px 64px minmax(140px,1fr) minmax(110px,150px) minmax(100px,150px) 104px', gap: 12, alignItems: 'center', padding: '11px 16px', borderTop: i ? `1px solid ${T.line}` : 'none', cursor: 'pointer' }}><FeatureThumb upload={upload} label={p.feature.title} ratio="1 / 1" onOpen={item => onOpenPhoto(item)} notes={notes} root={root} /><span style={{ fontFamily: T.mono, fontSize: 13, fontWeight: 700, color: T.accent }}>{meta.grade}</span><span style={{ fontSize: 14, fontWeight: 600, color: T.ink }}>{p.feature.title}</span><span style={{ fontSize: 12.5, color: T.mut }}>{p.boulder.feature.title}</span><span style={{ fontSize: 12.5, color: T.mut }}>{p.area.feature.title}</span><Badge status={meta.status} /></div>; })}{all.length === 0 && <EmptyState label="No problems match these filters." />}</div></>;
}

function Boulders({ root, notes, uploads, filters, currentAreaId, canWrite, openBoulder, onAddBoulder, onOpenPhoto }: { root: MoneyCragNode | null; notes: MoneyNote[]; uploads: MoneyUpload[]; filters: ReferenceFilters; currentAreaId?: string | null; canWrite: boolean; openBoulder: (id: string) => void; onAddBoulder?: (parentId?: string | null) => void; onOpenPhoto: OpenPhoto }) {
  const list = useMemo(() => filterBoulders(root, notes, uploads, filters, currentAreaId), [root, notes, uploads, filters, currentAreaId]);
  const parentId = filters.subAreaId !== 'all' ? filters.subAreaId : filters.areaId !== 'all' ? filters.areaId : currentAreaId;
  return <><Head title="Boulders" sub={`${list.length} matching blocks — tracked from scouted to established.`} action={canWrite && onAddBoulder ? <button onClick={() => onAddBoulder(parentId)} style={btnA}><Plus size={15} />Add boulder</button> : undefined} /><div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill,minmax(280px,1fr))', gap: 14 }}>{list.map(b => { const dev = b.feature.status as keyof typeof DEV.meta; const m = DEV.meta[dev] ?? DEV.meta.scouted; const upload = primaryUploadForFeature(b.feature.id, uploads, notes); return <div key={b.feature.id} onClick={() => openBoulder(b.feature.id)} style={{ ...card, overflow: 'hidden', cursor: 'pointer' }}>{upload ? <FeatureThumb upload={upload} label={b.feature.title} ratio="16 / 8" onOpen={item => onOpenPhoto(item)} notes={notes} root={root} /> : <Stripe label={`${b.feature.title.toLowerCase()} · no photo`} ratio="16 / 8" />}<div style={{ padding: 14 }}><div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}><span style={{ fontFamily: T.mono, fontSize: 10, fontWeight: 700, color: m.c, background: m.bg, padding: '2px 8px', borderRadius: 5 }}>{m.short}</span><span style={{ marginLeft: 'auto', fontFamily: T.mono, fontSize: 11, color: T.mut }}>{cragProblems(b).length} problems</span></div><div style={{ fontSize: 16, fontWeight: 800, color: T.ink }}>{b.feature.title}</div><div style={{ marginTop: 4, fontSize: 12, color: T.mut }}>{areaLabelForFeature(root, b.feature.id)}</div></div></div>; })}{list.length === 0 && <EmptyState label="No boulders match these filters." />}</div></>;
}

const trailCategoryOptions: Array<{ value: MoneyTrailCategory; label: string }> = [
  { value: 'connector', label: 'Connector' },
  { value: 'approach', label: 'Approach' },
  { value: 'trail_to_area', label: 'Trail to area' },
  { value: 'trail_to_destination', label: 'Trail to destination' },
];

function Trails({ root, trails, filters, canWrite, selectTrail, onCreateTrail, onUpdateTrail }: { root: MoneyCragNode | null; trails: MoneyCragNode[]; filters: ReferenceFilters; canWrite: boolean; selectTrail: (id: string) => void; onCreateTrail: (trail: ParsedTrailImport & { filename: string }) => void; onUpdateTrail: (trail: MoneyCragNode, updates: { title: string; category: MoneyTrailCategory; destinationFeatureId?: string | null; destinationLabel?: string | null }) => void }) {
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);
  const visibleTrails = useMemo(() => filterTrails(root, trails, filters), [root, trails, filters]);
  const onFile = async (file: File | undefined) => { if (!file) return; setUploading(true); setUploadError(null); try { const parsed = await parseTrailFile(file); onCreateTrail({ ...parsed, filename: file.name }); } catch (err) { setUploadError(err instanceof Error ? err.message : 'Unable to import trail.'); } finally { setUploading(false); if (inputRef.current) inputRef.current.value = ''; } };
  return <><Head title="Trails" sub={`${visibleTrails.length} matching approaches and connectors.`} action={canWrite ? <><input ref={inputRef} type="file" accept=".gpx,.geojson,.json,application/gpx+xml,application/geo+json,application/json" onChange={e => void onFile(e.target.files?.[0])} style={{ display: 'none' }} aria-label="Trail file" /><button onClick={() => inputRef.current?.click()} disabled={uploading} style={{ ...btnA, opacity: uploading ? 0.6 : 1 }}><Upload size={15} />{uploading ? 'Importing…' : 'Upload trail'}</button></> : undefined} />{uploadError && <div role="alert" style={{ ...card, padding: 12, marginBottom: 14, color: '#E6A299', borderColor: '#8F4E45' }}>{uploadError}</div>}<div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill,minmax(300px,1fr))', gap: 14 }}>{visibleTrails.map(tr => <TrailCard key={tr.feature.id} root={root} trail={tr} canWrite={canWrite} selectTrail={selectTrail} onUpdateTrail={onUpdateTrail} />)}{visibleTrails.length === 0 && <div style={{ ...card, padding: 18, color: T.mut, fontSize: 13 }}>No trails match these filters.</div>}</div></>;
}

function TrailCard({ root, trail, canWrite, selectTrail, onUpdateTrail }: { root: MoneyCragNode | null; trail: MoneyCragNode; canWrite: boolean; selectTrail: (id: string) => void; onUpdateTrail: (trail: MoneyCragNode, updates: { title: string; category: MoneyTrailCategory; destinationFeatureId?: string | null; destinationLabel?: string | null }) => void }) { const [editing, setEditing] = useState(false); const category = normalizeTrailCategory(trail.feature.properties.trail_category); const destinationFeatureId = typeof trail.feature.properties.trail_destination_feature_id === 'string' ? trail.feature.properties.trail_destination_feature_id : ''; const destinationLabel = typeof trail.feature.properties.trail_destination_label === 'string' ? trail.feature.properties.trail_destination_label : ''; const areas = useMemo(() => flattenAreas(root), [root]); const destinationName = destinationFeatureId ? areas.find(node => node.feature.id === destinationFeatureId)?.feature.title : destinationLabel; return <div style={{ ...card, padding: 16 }}><div onClick={() => selectTrail(trail.feature.id)} style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8, cursor: 'pointer' }}><Route color={T.accent} /><div style={{ flex: 1 }}><div style={{ fontSize: 15, fontWeight: 700, color: T.ink }}>{trail.feature.title}</div><div style={{ fontFamily: T.mono, fontSize: 11, color: T.mut }}>{trailCategoryLabel(category)} · {destinationName || 'No destination'} · {String(trail.feature.properties.dist ?? '—')} · {String(trail.feature.properties.gain ?? '—')}</div></div></div>{canWrite && <button aria-label={`Edit trail ${trail.feature.title}`} onClick={() => setEditing(!editing)} style={miniBtn}>{editing ? 'Close' : 'Edit info'}</button>}{editing && <TrailInfoEditor root={root} trail={trail} onSave={updates => { onUpdateTrail(trail, updates); setEditing(false); }} />}</div>; }
function TrailInfoEditor({ root, trail, onSave }: { root: MoneyCragNode | null; trail: MoneyCragNode; onSave: (updates: { title: string; category: MoneyTrailCategory; destinationFeatureId?: string | null; destinationLabel?: string | null }) => void }) { const category = normalizeTrailCategory(trail.feature.properties.trail_category); const destinationFeatureId = typeof trail.feature.properties.trail_destination_feature_id === 'string' ? trail.feature.properties.trail_destination_feature_id : ''; const destinationLabel = typeof trail.feature.properties.trail_destination_label === 'string' ? trail.feature.properties.trail_destination_label : ''; const [title, setTitle] = useState(trail.feature.title); const [draftCategory, setDraftCategory] = useState<MoneyTrailCategory>(category); const [draftDestinationId, setDraftDestinationId] = useState(destinationFeatureId); const [draftDestinationLabel, setDraftDestinationLabel] = useState(destinationLabel); const areas = useMemo(() => flattenAreas(root), [root]); useEffect(() => { setTitle(trail.feature.title); setDraftCategory(category); setDraftDestinationId(destinationFeatureId); setDraftDestinationLabel(destinationLabel); }, [category, destinationFeatureId, destinationLabel, trail.feature.id, trail.feature.title]); const dirty = title.trim() !== trail.feature.title || draftCategory !== category || draftDestinationId !== destinationFeatureId || draftDestinationLabel.trim() !== destinationLabel; const needsDestination = draftCategory === 'trail_to_area' || draftCategory === 'trail_to_destination'; const invalid = !dirty || !title.trim() || Boolean(needsDestination && !draftDestinationId && !draftDestinationLabel.trim()); return <div style={{ background: T.inset, border: `1px solid ${T.line}`, borderRadius: 10, padding: 12, marginTop: 12 }}><input aria-label="Trail label" value={title} onChange={e => setTitle(e.target.value)} maxLength={200} style={field} /><select aria-label="Trail category" value={draftCategory} onChange={e => setDraftCategory(e.target.value as MoneyTrailCategory)} style={{ ...field, marginTop: 8 }}>{trailCategoryOptions.map(option => <option key={option.value} value={option.value}>{option.label}</option>)}</select>{needsDestination && <><select aria-label="Destination area" value={draftDestinationId} onChange={e => setDraftDestinationId(e.target.value)} style={{ ...field, marginTop: 8 }}><option value="">Pick mapped area…</option>{areas.map(area => <option key={area.feature.id} value={area.feature.id}>{area.feature.title}</option>)}</select><input aria-label="Destination label" placeholder="Or custom destination label" value={draftDestinationLabel} onChange={e => setDraftDestinationLabel(e.target.value)} maxLength={200} style={{ ...field, marginTop: 8 }} /></>}<button disabled={invalid} onClick={() => onSave({ title: title.trim(), category: draftCategory, destinationFeatureId: draftDestinationId || null, destinationLabel: draftDestinationLabel.trim() || null })} style={{ ...btnA, marginTop: 10, opacity: invalid ? 0.5 : 1 }}>Save trail info</button></div>; }

function Photos({ root, notes, uploads, filters, canWrite, onOpenComposer, onOpenPhoto, onDeleteUpload }: { root: MoneyCragNode | null; notes: MoneyNote[]; uploads: MoneyUpload[]; filters: ReferenceFilters; canWrite: boolean; onOpenComposer: (kind?: 'photo' | 'sketch' | 'file' | null) => void; onOpenPhoto: OpenPhoto; onDeleteUpload: (upload: MoneyUpload) => void }) {
  const photoItems = useMemo(() => filterPhotoItems(root, notes, uploads, filters), [uploads, notes, root, filters]);
  return <><Head title="Photos" sub={`${photoItems.length} matching uploads · topos, sketches, conditions`} action={<button onClick={() => onOpenComposer('photo')} style={btnA}><Camera size={15} />Add photo</button>} /><div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill,minmax(200px,1fr))', gap: 12 }}>{photoItems.map(item => <UploadPhotoButton key={item.upload.id} upload={item.upload} ratio="4 / 3" title={item.title} canDelete={canWrite} onDelete={onDeleteUpload} onOpen={() => onOpenPhoto(item)} />)}{photoItems.length === 0 && uploads.length === 0 && ['overview topo', 'approach', 'conditions', 'panorama'].map(l => <Stripe key={l} label={l} />)}{photoItems.length === 0 && uploads.length > 0 && <EmptyState label="No photos match these filters." />}</div></>;
}

function Notes({ root, notes, uploads, filters, canWrite, onOpenComposer, onOpenPhoto, onEditNote, onDeleteNote, onDeleteUpload }: { root: MoneyCragNode | null; notes: MoneyNote[]; uploads: MoneyUpload[]; filters: ReferenceFilters; canWrite: boolean; onOpenComposer: (kind?: 'photo' | 'sketch' | 'file' | null) => void; onOpenPhoto: OpenPhoto; onEditNote: (note: MoneyNote) => void; onDeleteNote: (note: MoneyNote) => void; onDeleteUpload: (upload: MoneyUpload) => void }) {
  const visibleNotes = useMemo(() => filterNotes(root, notes, uploads, filters), [root, notes, uploads, filters]);
  return <><Head title="Notes" sub={`${visibleNotes.length} matching notes — beta, conditions, decisions, access.`} action={<button onClick={() => onOpenComposer(null)} style={btnA}><FileText size={15} />New note</button>} /><div style={{ maxWidth: 680 }}>{visibleNotes.map(n => { const noteUploads = uploadsForNoteBlocks(uploads, n); const contextLabel = contextLabelForFeature(root, n.target_ref ?? n.feature_id); return <div key={n.id} style={{ ...card, padding: 16, marginBottom: 12 }}><div style={{ display: 'flex', alignItems: 'center', gap: 9, marginBottom: 8 }}><span style={{ width: 24, height: 24, borderRadius: '50%', background: T.av, color: '#fff', fontSize: 10, fontWeight: 700, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>MC</span><span style={{ fontSize: 13, fontWeight: 700, color: T.ink }}>Team</span><span style={{ marginLeft: 'auto', fontFamily: T.mono, fontSize: 11, color: T.mut }}>{new Date(n.created_at).toLocaleDateString()}</span>{canWrite && <><button aria-label="Edit note" onClick={() => onEditNote(n)} style={miniBtn}>Edit</button><button aria-label="Delete note" onClick={() => { if (window.confirm('Delete this note?')) onDeleteNote(n); }} style={{ ...miniBtn, color: '#E6A299' }}>Delete</button></>}</div><p style={{ margin: '0 0 8px', fontSize: 13.5, lineHeight: 1.55, color: '#D9CBBD' }}>{n.body}</p>{noteUploads.length > 0 && <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill,minmax(120px,1fr))', gap: 8, marginBottom: 10 }}>{noteUploads.map(upload => <UploadPhotoButton key={upload.id} upload={upload} canDelete={canWrite} onDelete={onDeleteUpload} onOpen={() => onOpenPhoto(lightboxItemForUpload(upload, notes, root))} />)}</div>}{contextLabel && <div style={{ fontFamily: T.mono, fontSize: 11, color: T.mut }}>{contextLabel}</div>}</div>; })}{visibleNotes.length === 0 && <div style={{ ...card, padding: 18, color: T.mut, fontSize: 13 }}>No notes match these filters.</div>}</div></>;
}

function Trash({ items, canWrite, onRestore }: { items: MoneyTrashItem[]; canWrite: boolean; onRestore: (id: string) => void }) { return <><Head title="Trash" sub={`${items.length} deleted area${items.length === 1 ? '' : 's'} ready to restore.`} /><div style={{ ...card, overflow: 'hidden' }}>{items.length === 0 && <div style={{ padding: 18, color: T.mut, fontSize: 13 }}>Trash is empty.</div>}{items.map((item, i) => <div key={item.id} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '13px 16px', borderTop: i ? `1px solid ${T.line}` : 'none' }}><div style={{ flex: 1, minWidth: 0 }}><div style={{ fontSize: 14, fontWeight: 800, color: T.ink }}>{item.title}</div><div style={{ fontFamily: T.mono, fontSize: 10.5, color: T.mut }}>{item.path.join(' / ')} · {item.descendant_count} descendant{item.descendant_count === 1 ? '' : 's'} · deleted {new Date(item.deleted_at).toLocaleDateString()}</div></div><button disabled={!canWrite} onClick={() => onRestore(item.id)} style={{ display: 'flex', alignItems: 'center', gap: 7, border: `1px solid ${T.accentDim}`, borderRadius: 9, padding: '9px 12px', background: canWrite ? T.accentSoft : 'transparent', color: canWrite ? T.accent : T.faint, fontWeight: 800, cursor: canWrite ? 'pointer' : 'default' }}><RotateCcw size={15} />Restore</button></div>)}</div></>; }
function Stripe({ label, ratio }: { label: string; ratio?: string }) { return <div style={{ width: '100%', aspectRatio: ratio || '4 / 3', backgroundColor: T.map.slot, backgroundImage: `repeating-linear-gradient(45deg, ${T.map.slotStripe} 0 1px, transparent 1px 9px)`, display: 'flex', alignItems: 'flex-end', overflow: 'hidden', borderRadius: 8 }}><span style={{ fontFamily: T.mono, fontSize: 10, color: T.map.slotText, padding: '5px 7px' }}>{label}</span></div>; }
function Badge({ status }: { status: string }) { const c = status === 'sent' ? T.blue : status === 'project' ? T.accent : T.mut; return <span style={{ fontFamily: T.mono, fontSize: 10, fontWeight: 700, textTransform: 'uppercase', color: c, background: `${c}22`, padding: '2px 7px', borderRadius: 5 }}>{status}</span>; }
function EmptyState({ label }: { label: string }) { return <div style={{ ...card, padding: 18, color: T.mut, fontSize: 13 }}>{label}</div>; }
const btnA: React.CSSProperties = { display: 'flex', alignItems: 'center', gap: 7, border: 'none', borderRadius: 9, padding: '10px 15px', background: T.accent, color: T.onAccent, fontWeight: 700, fontSize: 13, cursor: 'pointer' };
const field: React.CSSProperties = { width: '100%', background: T.surf2, border: `1px solid ${T.line}`, borderRadius: 7, padding: '9px 10px', color: T.ink, fontFamily: T.font, fontSize: 13, outline: 'none', minHeight: 40 };

function FeatureThumb({ upload, label, ratio, onOpen, notes, root }: { upload?: MoneyUpload; label: string; ratio: string; onOpen: OpenPhoto; notes: MoneyNote[]; root: MoneyCragNode | null }) {
  const { src, loading } = useUploadImageUrl(upload?.id ?? '');
  if (!upload) return <Stripe label={`${label.toLowerCase()} · no photo`} ratio={ratio} />;
  return <button type="button" aria-label={`Open photo ${label}`} onClick={event => { event.stopPropagation(); onOpen(lightboxItemForUpload(upload, notes, root)); }} style={{ width: '100%', aspectRatio: ratio, border: 'none', padding: 0, background: T.map.slot, cursor: 'zoom-in', borderRadius: 8, overflow: 'hidden', display: 'block' }}>{loading || !src ? <span style={{ display: 'flex', width: '100%', height: '100%', alignItems: 'center', justifyContent: 'center', color: T.accent, fontFamily: T.mono, fontSize: 10 }}>Loading photo</span> : <img src={src} alt={label} style={{ display: 'block', width: '100%', height: '100%', objectFit: 'cover' }} />}</button>;
}

function PhotoSelect({ value, onChange, label }: { value: PhotoFilter; onChange: (value: PhotoFilter) => void; label: string }) { return <label style={filterLabel}>Photos<select aria-label={label} value={value} onChange={e => onChange(e.target.value as PhotoFilter)} style={filterField}><option value="all">With or without</option><option value="with">Has photo</option><option value="without">No photo</option></select></label>; }
function SortSelect({ value, onChange, options }: { value: SortFilter; onChange: (value: SortFilter) => void; options: Array<[SortFilter, string]> }) { return <label style={filterLabel}>Sort<select aria-label="Sort filter" value={value} onChange={e => onChange(e.target.value as SortFilter)} style={filterField}>{options.map(([option, label]) => <option key={option} value={option}>{label}</option>)}</select></label>; }

function uploadsForNoteBlocks(uploads: MoneyUpload[], note: MoneyNote): MoneyUpload[] { const ids = new Set((note.blocks ?? []).map(block => block.upload_id).filter((id): id is string => Boolean(id))); return uploads.filter(upload => ids.has(upload.id) || upload.note_id === note.id); }
function lightboxItemForUpload(upload: MoneyUpload, notes: MoneyNote[], root: MoneyCragNode | null): PhotoLightboxItem { const note = noteForUpload(upload, notes); const featureId = upload.feature_id ?? note?.target_ref ?? note?.feature_id; return { upload, title: upload.title || upload.original_filename, contextLabel: contextLabelForFeature(root, featureId), noteBody: note?.body }; }
function noteForUpload(upload: MoneyUpload, notes: MoneyNote[]): MoneyNote | undefined { return notes.find(note => note.id === upload.note_id || (note.blocks ?? []).some(block => block.upload_id === upload.id)); }
function uploadIsVisibleInNotes(upload: MoneyUpload, notes: MoneyNote[]): boolean { return !upload.note_id || notes.some(note => note.id === upload.note_id || (note.blocks ?? []).some(block => block.upload_id === upload.id)); }
function contextLabelForFeature(root: MoneyCragNode | null, featureId?: string | null): string | undefined { if (!featureId) return undefined; const problem = flattenProblems(root).find(item => item.feature.id === featureId); if (problem) return [...pathTo(root, problem.area.feature.id).map(node => node.feature.title), problem.boulder.feature.title, problem.feature.title].join(' / '); const path = pathTo(root, featureId); if (path.length > 0) return path.map(node => node.feature.title).join(' / '); const node = findNode(root, featureId); return node?.feature.title; }
function normalizeTrailCategory(value: unknown): MoneyTrailCategory { return value === 'approach' || value === 'trail_to_area' || value === 'trail_to_destination' ? value : 'connector'; }
function trailCategoryLabel(category: MoneyTrailCategory): string { return trailCategoryOptions.find(option => option.value === category)?.label ?? 'Connector'; }
function defaultSortForView(view: ReferenceFilterView): SortFilter { return view === 'photos' || view === 'notes' ? 'newest' : view === 'problems' ? 'grade' : 'title'; }
function summaryForView(view: ReferenceFilterView, root: MoneyCragNode | null, trails: MoneyCragNode[], notes: MoneyNote[], uploads: MoneyUpload[]): string { if (view === 'boulders') return `${flattenBoulders(root).length} total boulders`; if (view === 'problems') return `${flattenProblems(root).length} total problems`; if (view === 'trails') return `${trails.length} total trails`; if (view === 'photos') return `${uploads.length} total uploads`; return `${notes.length} total notes`; }

function primaryUploadForFeature(featureId: string, uploads: MoneyUpload[], notes: MoneyNote[]): MoneyUpload | undefined {
  const featureNotes = notes.filter(note => noteTargetsFeature(note, featureId));
  const noteIds = new Set(featureNotes.map(note => note.id));
  const blockUploadIds = new Set(featureNotes.flatMap(note => note.blocks ?? []).map(block => block.upload_id).filter((id): id is string => Boolean(id)));
  return uploads.find(upload => upload.feature_id === featureId || blockUploadIds.has(upload.id) || (upload.note_id && noteIds.has(upload.note_id)));
}

function noteTargetsFeature(note: MoneyNote, featureId: string): boolean { return note.target_ref === featureId || note.feature_id === featureId; }
function searchMatch(text: string, search: string): boolean { return !search.trim() || text.toLowerCase().includes(search.trim().toLowerCase()); }
function usesAreaScope(view: ReferenceFilterView): boolean { return view === 'boulders' || view === 'problems' || view === 'trails' || view === 'photos' || view === 'notes'; }
function subAreaParentFilterId(filters: ReferenceFilters, currentAreaId?: string | null): string { return filters.areaId !== 'all' ? filters.areaId : currentAreaId ?? 'all'; }
function effectiveParentAreaId(filters: ReferenceFilters, currentAreaId?: string | null, view?: ReferenceFilterView): string { return filters.areaId !== 'all' ? filters.areaId : view === 'boulders' && currentAreaId ? currentAreaId : 'all'; }
function effectiveAreaFilterId(filters: ReferenceFilters, currentAreaId?: string | null, view?: ReferenceFilterView): string { return filters.subAreaId !== 'all' ? filters.subAreaId : effectiveParentAreaId(filters, currentAreaId, view); }
function areaIdsForFeature(root: MoneyCragNode | null, featureId?: string | null): string[] { if (!featureId) return []; const problem = flattenProblems(root).find(item => item.feature.id === featureId); if (problem) return pathTo(root, problem.area.feature.id).map(node => node.feature.id); const path = pathTo(root, featureId); return path.filter(node => node.feature.feature_type === 'area').map(node => node.feature.id); }
function matchesArea(root: MoneyCragNode | null, featureId: string | undefined | null, areaId: string): boolean { return areaId === 'all' || areaIdsForFeature(root, featureId).includes(areaId); }
function areaLabelForFeature(root: MoneyCragNode | null, featureId: string): string { const path = pathTo(root, featureId).filter(node => node.feature.feature_type === 'area'); return path.at(-1)?.feature.title ?? 'Unmapped area'; }
function uploadFeatureId(upload: MoneyUpload, notes: MoneyNote[]): string | undefined { const note = noteForUpload(upload, notes); return upload.feature_id ?? note?.target_ref ?? note?.feature_id; }
function targetForFeature(root: MoneyCragNode | null, featureId?: string | null): TargetFilter { if (!featureId) return 'none'; const problem = flattenProblems(root).find(item => item.feature.id === featureId); if (problem) return 'problem'; const node = findNode(root, featureId); return node?.feature.feature_type === 'area' || node?.feature.feature_type === 'boulder' || node?.feature.feature_type === 'trail' ? node.feature.feature_type : 'none'; }
function targetForNote(root: MoneyCragNode | null, note: MoneyNote): TargetFilter { if (note.target_type === 'project') return 'project'; if (note.target_type === 'none') return 'none'; return targetForFeature(root, note.target_ref ?? note.feature_id); }
function gradeSort(a: string, b: string): number { return gradeValue(a) - gradeValue(b) || a.localeCompare(b); }
function gradeValue(grade: string): number { const match = grade.match(/V(\d+)/i); return match ? Number(match[1]) : 999; }

function filterBoulders(root: MoneyCragNode | null, notes: MoneyNote[], uploads: MoneyUpload[], filters: ReferenceFilters, currentAreaId?: string | null): MoneyCragNode[] {
  return flattenBoulders(root).filter(boulder => {
    const hasPhoto = Boolean(primaryUploadForFeature(boulder.feature.id, uploads, notes));
    return searchMatch(`${boulder.feature.title} ${areaLabelForFeature(root, boulder.feature.id)}`, filters.search)
      && matchesArea(root, boulder.feature.id, effectiveAreaFilterId(filters, currentAreaId, 'boulders'))
      && (filters.status === 'all' || boulder.feature.status === filters.status)
      && (filters.photo === 'all' || (filters.photo === 'with' ? hasPhoto : !hasPhoto));
  }).sort((a, b) => sortNodes(a, b, filters, root));
}

function filterProblems(root: MoneyCragNode | null, notes: MoneyNote[], uploads: MoneyUpload[], filters: ReferenceFilters): Array<MoneyCragNode & { boulder: MoneyCragNode; area: MoneyCragNode }> {
  return flattenProblems(root).filter(problem => {
    const meta = problemMeta(problem.feature);
    const hasPhoto = Boolean(primaryUploadForFeature(problem.feature.id, uploads, notes));
    return searchMatch(`${problem.feature.title} ${problem.boulder.feature.title} ${problem.area.feature.title} ${meta.grade}`, filters.search)
      && matchesArea(root, problem.area.feature.id, effectiveAreaFilterId(filters, undefined, 'problems'))
      && (filters.boulderId === 'all' || problem.boulder.feature.id === filters.boulderId)
      && (filters.status === 'all' || meta.status === filters.status)
      && (filters.grade === 'all' || meta.grade === filters.grade)
      && (filters.photo === 'all' || (filters.photo === 'with' ? hasPhoto : !hasPhoto));
  }).sort((a, b) => filters.sort === 'grade' ? gradeSort(problemMeta(a.feature).grade, problemMeta(b.feature).grade) : sortProblemNodes(a, b, filters));
}

function filterTrails(root: MoneyCragNode | null, trails: MoneyCragNode[], filters: ReferenceFilters): MoneyCragNode[] {
  const areas = flattenAreas(root);
  return trails.filter(trail => {
    const category = normalizeTrailCategory(trail.feature.properties.trail_category);
    const destinationId = typeof trail.feature.properties.trail_destination_feature_id === 'string' ? trail.feature.properties.trail_destination_feature_id : '';
    const destinationLabel = typeof trail.feature.properties.trail_destination_label === 'string' ? trail.feature.properties.trail_destination_label : '';
    const destinationName = destinationId ? areas.find(area => area.feature.id === destinationId)?.feature.title ?? '' : destinationLabel;
    return searchMatch(`${trail.feature.title} ${trailCategoryLabel(category)} ${destinationName}`, filters.search)
      && (filters.trailCategory === 'all' || category === filters.trailCategory)
      && matchesArea(root, destinationId, effectiveAreaFilterId(filters, undefined, 'trails'))
      && (filters.destination === 'all' || (filters.destination === 'set' ? Boolean(destinationId || destinationLabel) : filters.destination === 'unset' ? !destinationId && !destinationLabel : destinationId === filters.destination));
  }).sort((a, b) => sortTrails(a, b, filters, root));
}

function filterPhotoItems(root: MoneyCragNode | null, notes: MoneyNote[], uploads: MoneyUpload[], filters: ReferenceFilters): PhotoLightboxItem[] {
  const visibleUploads = uploads.filter(upload => uploadIsVisibleInNotes(upload, notes));
  return visibleUploads.map(upload => lightboxItemForUpload(upload, notes, root)).filter(item => {
    const featureId = uploadFeatureId(item.upload, notes);
    const target = targetForFeature(root, featureId) || (noteForUpload(item.upload, notes)?.target_type === 'project' ? 'project' : 'none');
    return searchMatch(`${item.upload.title ?? ''} ${item.upload.original_filename} ${item.upload.comments ?? ''} ${item.contextLabel ?? ''} ${item.noteBody ?? ''}`, filters.search)
      && matchesArea(root, featureId, effectiveAreaFilterId(filters, undefined, 'photos'))
      && (filters.uploadKind === 'all' || item.upload.block_kind === filters.uploadKind)
      && (filters.target === 'all' || target === filters.target);
  }).sort((a, b) => sortUploads(a.upload, b.upload, filters));
}

function filterNotes(root: MoneyCragNode | null, notes: MoneyNote[], uploads: MoneyUpload[], filters: ReferenceFilters): MoneyNote[] {
  return notes.filter(note => {
    const featureId = note.target_ref ?? note.feature_id;
    const target = note.target_type === 'project' ? 'project' : targetForNote(root, note);
    const noteUploads = uploadsForNoteBlocks(uploads, note);
    const context = contextLabelForFeature(root, featureId) ?? '';
    return searchMatch(`${note.body} ${context} ${(note.tags ?? []).join(' ')}`, filters.search)
      && matchesArea(root, featureId, effectiveAreaFilterId(filters, undefined, 'notes'))
      && (filters.target === 'all' || target === filters.target)
      && (filters.photo === 'all' || (filters.photo === 'with' ? noteUploads.length > 0 : noteUploads.length === 0));
  }).sort((a, b) => filters.sort === 'oldest' ? Date.parse(a.created_at) - Date.parse(b.created_at) : Date.parse(b.created_at) - Date.parse(a.created_at));
}

function sortNodes(a: MoneyCragNode, b: MoneyCragNode, filters: ReferenceFilters, root: MoneyCragNode | null): number { if (filters.sort === 'area') return areaLabelForFeature(root, a.feature.id).localeCompare(areaLabelForFeature(root, b.feature.id)) || a.feature.title.localeCompare(b.feature.title); if (filters.sort === 'status') return String(a.feature.status).localeCompare(String(b.feature.status)) || a.feature.title.localeCompare(b.feature.title); return a.feature.title.localeCompare(b.feature.title); }
function sortProblemNodes(a: MoneyCragNode & { boulder: MoneyCragNode; area: MoneyCragNode }, b: MoneyCragNode & { boulder: MoneyCragNode; area: MoneyCragNode }, filters: ReferenceFilters): number { if (filters.sort === 'area') return a.area.feature.title.localeCompare(b.area.feature.title) || a.feature.title.localeCompare(b.feature.title); if (filters.sort === 'status') return String(a.feature.status).localeCompare(String(b.feature.status)) || a.feature.title.localeCompare(b.feature.title); return a.feature.title.localeCompare(b.feature.title); }
function sortTrails(a: MoneyCragNode, b: MoneyCragNode, filters: ReferenceFilters, root: MoneyCragNode | null): number { if (filters.sort === 'category') return trailCategoryLabel(normalizeTrailCategory(a.feature.properties.trail_category)).localeCompare(trailCategoryLabel(normalizeTrailCategory(b.feature.properties.trail_category))) || a.feature.title.localeCompare(b.feature.title); if (filters.sort === 'destination') return trailDestinationName(a, root).localeCompare(trailDestinationName(b, root)) || a.feature.title.localeCompare(b.feature.title); return a.feature.title.localeCompare(b.feature.title); }
function sortUploads(a: MoneyUpload, b: MoneyUpload, filters: ReferenceFilters): number { if (filters.sort === 'oldest') return Date.parse(a.created_at) - Date.parse(b.created_at); if (filters.sort === 'title') return (a.title || a.original_filename).localeCompare(b.title || b.original_filename); return Date.parse(b.created_at) - Date.parse(a.created_at); }
function trailDestinationName(trail: MoneyCragNode, root: MoneyCragNode | null): string { const destinationId = typeof trail.feature.properties.trail_destination_feature_id === 'string' ? trail.feature.properties.trail_destination_feature_id : ''; const destinationLabel = typeof trail.feature.properties.trail_destination_label === 'string' ? trail.feature.properties.trail_destination_label : ''; return destinationId ? flattenAreas(root).find(area => area.feature.id === destinationId)?.feature.title ?? '' : destinationLabel; }
