import { useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, Camera, FileText, Loader2, LogOut, Map, Menu, Mountain, Plus, Route, Search, Trash2 } from 'lucide-react';
import { AuthProvider, useAuth } from '../../contexts/AuthContext';
import { moneyApi } from '../../services/money';
import { MoneyArchiveMode, MoneyCragNode, MoneyDevStatus, MoneyFeature, MoneyNote, MoneyNoteBlock, MoneyNoteTargetType, MoneyPosition, MoneyProblemStatus, MoneyTrailCategory, MoneyUpload } from '../../types/money';
import { LoginScreen } from './LoginScreen';
import { ContentView, ReferenceFilterRail, defaultReferenceFilters, isReferenceFilterView } from './reference/ContentViews';
import { CragMap } from './reference/CragMap';
import { DetailPanel } from './reference/DetailPanel';
import { AddBoulderModal, AddBoulderPayload } from './reference/AddBoulderModal';
import { parseCoordinateSearch } from './reference/coordinates';
import { cragBoulders, cragChildren, cragProblems, findNode, flattenAreas, parentArea, pathTo, pointGeoJSON, polygonGeoJSON } from './reference/model';
import { NoteComposer, NoteComposerPayload } from './reference/NoteComposer';
import { T } from './reference/theme';
import { ParsedTrailImport } from './reference/trailImport';
import { TopoOverlay, upsertTopoOverlay } from './reference/topoOverlay';

const PROJECT_SLUG = 'money-creek';
type View = 'map' | 'problems' | 'boulders' | 'trails' | 'photos' | 'notes' | 'trash';
type Mode = 'view' | 'create-area' | 'create-boulder' | 'edit-area';
type LayersState = { base: string; roads: boolean; water: boolean; contours: boolean; trails: boolean; areas: Record<string, boolean>; dev: Record<string, boolean> };
type TrailUpdate = { title: string; category: MoneyTrailCategory; destinationFeatureId?: string | null; destinationLabel?: string | null };

function useMobile() {
  const [m, setM] = useState(() => window.innerWidth < 760);
  useEffect(() => {
    const r = () => setM(window.innerWidth < 760);
    window.addEventListener('resize', r);
    return () => window.removeEventListener('resize', r);
  }, []);
  return m;
}

function MoneyCreekWorkspace() {
  const { user, isAuthenticated, isBootstrapping, canWrite, logout } = useAuth();
  const queryClient = useQueryClient();
  const mobile = useMobile();
  const [currentId, setCurrentId] = useState<string | null>(null);
  const [selectedBoulderId, setSelectedBoulderId] = useState<string | null>(null);
  const [selectedTrailId, setSelectedTrailId] = useState<string | null>(null);
  const [view, setView] = useState<View>('map');
  const [tab, setTab] = useState('overview');
  const [mode, setMode] = useState<Mode>('view');
  const [pending, setPending] = useState<{ kind: Extract<Mode, 'create-area' | 'create-boulder'>; points: MoneyPosition[] } | null>(null);
  const [drawer, setDrawer] = useState(false);
  const [addOpen, setAddOpen] = useState(false);
  const [sheetExpanded, setSheetExpanded] = useState(false);
  const [composer, setComposer] = useState<{ kind?: 'photo' | 'sketch' | 'file' | null; note?: MoneyNote | null } | null>(null);
  const [deleteChoice, setDeleteChoice] = useState<MoneyCragNode | null>(null);
  const [layers, setLayers] = useState<LayersState>({ base: 'stylized', roads: true, water: true, contours: true, trails: true, areas: {}, dev: { scouted: true, 'needs-work': true, cleaning: true, established: true } });
  const [referenceFilters, setReferenceFilters] = useState(defaultReferenceFilters);
  const [mapSearch, setMapSearch] = useState('');
  const [mapSearchError, setMapSearchError] = useState<string | null>(null);
  const [goToPoint, setGoToPoint] = useState<{ position: MoneyPosition; nonce: number } | null>(null);
  const [addBoulder, setAddBoulder] = useState<{ position?: MoneyPosition | null; parentId?: string | null } | null>(null);

  const projectQuery = useQuery({ queryKey: ['money-project', PROJECT_SLUG], queryFn: () => moneyApi.getProject(PROJECT_SLUG), enabled: isAuthenticated, staleTime: 10 * 60 * 1000 });
  const projectId = projectQuery.data?.project.id;
  const cragQuery = useQuery({ queryKey: ['money-crag', projectId], queryFn: () => moneyApi.getCragSnapshot(projectId!), enabled: Boolean(projectId), staleTime: 30 * 1000 });
  const trashQuery = useQuery({ queryKey: ['money-trash', projectId], queryFn: () => moneyApi.listTrash(projectId!), enabled: Boolean(projectId), staleTime: 30 * 1000 });
  const root = cragQuery.data?.root ?? null;

  useEffect(() => {
    if (root && (!currentId || !findNode(root, currentId))) {
      setCurrentId(root.feature.id);
      setSelectedBoulderId(null);
      setSelectedTrailId(null);
    }
  }, [root, currentId]);

  const area = findNode(root, currentId) ?? root;
  const selectedBoulder = findNode(root, selectedBoulderId);
  const selectedTrail = cragQuery.data?.trails?.find(t => t.feature.id === selectedTrailId) ?? null;
  const notes = cragQuery.data?.notes ?? [];
  const uploads = cragQuery.data?.uploads ?? [];

  const invalidate = () => {
    void queryClient.invalidateQueries({ queryKey: ['money-crag', projectId] });
    void queryClient.invalidateQueries({ queryKey: ['money-trash', projectId] });
  };
  const createArea = useMutation({ mutationFn: ({ name, parentId, points }: { name: string; parentId?: string | null; points: MoneyPosition[] }) => moneyApi.createArea(projectId!, { parent_feature_id: parentId, title: name, description: `Freshly outlined — add sub-areas or boulders to fill it in.`, geojson: polygonGeoJSON(points), properties: { kind: 'Boulders', aspect: 'newly mapped' } }), onSuccess: f => { invalidate(); setCurrentId(f.id); } });
  const createBoulder = useMutation({ mutationFn: ({ name, parentId, points, position, devStatus }: { name: string; parentId: string; points?: MoneyPosition[]; position?: MoneyPosition; devStatus?: MoneyDevStatus }) => moneyApi.createBoulder(projectId!, { parent_feature_id: parentId, title: name, description: null, dev_status: devStatus ?? 'scouted', geojson: position ? pointGeoJSON(position) : polygonGeoJSON(points ?? []), properties: position ? { center: position, source: 'coordinate-add' } : {} }), onSuccess: f => { invalidate(); setSelectedBoulderId(f.id); setCurrentId(f.parent_feature_id ?? currentId); setView('map'); setAddBoulder(null); setGoToPoint(null); } });
  const updateDev = useMutation({ mutationFn: ({ id, dev }: { id: string; dev: MoneyDevStatus }) => moneyApi.updateBoulderStatus(id, { dev_status: dev }), onSuccess: invalidate });
  const renameBoulder = useMutation({ mutationFn: ({ boulder, title }: { boulder: MoneyCragNode; title: string }) => moneyApi.updateFeature(boulder.feature.id, { parent_feature_id: boulder.feature.parent_feature_id ?? null, feature_type: 'boulder', title, description: boulder.feature.description ?? null, status: boulder.feature.status, geojson: boulder.feature.geojson, style: boulder.feature.style ?? {}, properties: boulder.feature.properties ?? {}, sort_order: boulder.feature.sort_order }), onSuccess: feature => { invalidate(); setSelectedBoulderId(feature.id); } });
  const updateAreaGeometry = useMutation({ mutationFn: ({ id, points }: { id: string; points: MoneyPosition[] }) => moneyApi.updateAreaGeometry(id, { geojson: polygonGeoJSON(points) }), onSuccess: () => { invalidate(); setMode('view'); setLayers(l => ({ ...l, base: 'stylized' })); } });
  const createProblem = useMutation({ mutationFn: ({ boulderId, p }: { boulderId: string; p: { name: string; grade: string; status: MoneyProblemStatus; stars: number; fa?: string | null; types: string[] } }) => moneyApi.createProblem(projectId!, { boulder_id: boulderId, ...p }), onSuccess: invalidate });
  const createNote = useMutation({ mutationFn: async (payload: NoteComposerPayload) => {
    const featureId = isFeatureNoteTarget(payload.target_type) ? payload.target_ref : undefined;
    const uploaded = new globalThis.Map<string, MoneyNoteBlock>();
    for (const item of payload.files) {
      const u = await moneyApi.uploadImage(projectId!, item.file, { featureId, blockKind: item.kind });
      uploaded.set(item.blockKey, { kind: item.kind, upload_id: u.id, name: u.original_filename });
    }
    return moneyApi.createProjectNote(projectId!, { body: payload.body, visibility: 'team', target_type: payload.target_type, target_ref: payload.target_ref, tags: payload.tags, blocks: resolveSubmittedBlocks(payload.blocks, uploaded) });
  }, onSuccess: () => { invalidate(); setComposer(null); if (mobile) setSheetExpanded(true); } });
  const updateNote = useMutation({ mutationFn: async ({ note, payload }: { note: MoneyNote; payload: NoteComposerPayload }) => {
    const featureId = isFeatureNoteTarget(payload.target_type) ? payload.target_ref : undefined;
    const uploaded = new globalThis.Map<string, MoneyNoteBlock>();
    for (const item of payload.files) {
      const u = await moneyApi.uploadImage(projectId!, item.file, { featureId, noteId: note.id, blockKind: item.kind });
      uploaded.set(item.blockKey, { kind: item.kind, upload_id: u.id, name: u.original_filename });
    }
    return moneyApi.updateNote(note.id, { body: payload.body, visibility: note.visibility ?? 'team', target_type: note.target_type, target_ref: note.target_ref ?? note.feature_id, tags: payload.tags, blocks: resolveSubmittedBlocks(payload.blocks, uploaded) });
  }, onSuccess: () => { invalidate(); setComposer(null); } });
  const deleteNote = useMutation({ mutationFn: (noteId: string) => moneyApi.deleteNote(noteId), onSuccess: invalidate });
  const deleteUpload = useMutation({ mutationFn: (uploadId: string) => moneyApi.deleteUpload(uploadId), onSuccess: invalidate });
  const updateUploadMetadata = useMutation({ mutationFn: ({ upload, title, comments }: { upload: MoneyUpload; title: string | null; comments: string | null }) => moneyApi.updateUploadMetadata(upload.id, { title, comments }), onSuccess: invalidate });
  const updateNoteAndDeleteRemovedUploads = useMutation({ mutationFn: async ({ note, payload }: { note: MoneyNote; payload: NoteComposerPayload }) => {
    const before = uploadIdsFromBlocks(note.blocks ?? []);
    const next = await updateNote.mutateAsync({ note, payload });
    const after = uploadIdsFromBlocks(payload.blocks);
    const removed = [...before].filter(id => !after.has(id));
    await Promise.all(removed.map(id => moneyApi.deleteUpload(id)));
    return next;
  }, onSuccess: invalidate });
  const deleteArea = useMutation({ mutationFn: ({ id, mode }: { id: string; mode: MoneyArchiveMode }) => moneyApi.archiveFeature(id, mode), onSuccess: (_r, vars) => { const parent = parentArea(root, vars.id); invalidate(); setCurrentId(parent?.feature.id ?? root?.feature.id ?? null); setSelectedBoulderId(null); setSelectedTrailId(null); setDeleteChoice(null); setView('map'); } });
  const moveArea = useMutation({ mutationFn: ({ id, parentId, sortOrder }: { id: string; parentId: string | null; sortOrder?: number }) => moneyApi.moveFeatureParent(id, { parent_feature_id: parentId, sort_order: sortOrder }), onSuccess: (_f, vars) => { invalidate(); setCurrentId(vars.id); setSelectedBoulderId(null); setSelectedTrailId(null); setView('map'); } });
  const restoreArea = useMutation({ mutationFn: (id: string) => moneyApi.restoreFeature(id), onSuccess: invalidate });
  const updateTrail = useMutation({ mutationFn: ({ trail, updates }: { trail: MoneyCragNode; updates: TrailUpdate }) => moneyApi.updateFeature(trail.feature.id, { feature_type: 'trail', title: updates.title, description: trail.feature.description ?? null, status: trail.feature.status, geojson: trail.feature.geojson, style: trail.feature.style ?? {}, properties: { ...trail.feature.properties, trail_category: updates.category, trail_destination_feature_id: updates.destinationFeatureId || undefined, trail_destination_label: updates.destinationLabel || undefined }, sort_order: trail.feature.sort_order }), onSuccess: feature => { invalidate(); setSelectedTrailId(feature.id); setView('map'); } });
  const deleteTrail = useMutation({ mutationFn: (trail: MoneyCragNode) => moneyApi.archiveFeature(trail.feature.id), onSuccess: () => { invalidate(); setSelectedTrailId(null); setView('map'); } });
  const updateTopo = useMutation({ mutationFn: ({ problem, overlay }: { problem: MoneyFeature; overlay: TopoOverlay }) => moneyApi.updateFeature(problem.id, { parent_feature_id: problem.parent_feature_id ?? null, feature_type: 'problem', title: problem.title, description: problem.description ?? null, status: problem.status, geojson: problem.geojson, style: problem.style ?? {}, properties: upsertTopoOverlay(problem, overlay), sort_order: problem.sort_order }), onSuccess: invalidate });
  const createTrail = useMutation({ mutationFn: (trail: ParsedTrailImport & { filename: string }) => moneyApi.createFeature(projectId!, { feature_type: 'trail', title: trail.title, description: null, status: 'active', geojson: trail.geojson, style: {}, properties: { trail_category: 'connector', source: trail.sourceFormat, source_filename: trail.filename, imported_point_count: trail.pointCount }, sort_order: (cragQuery.data?.trails?.length ?? 0) + 1 }), onSuccess: feature => { invalidate(); setSelectedTrailId(feature.id); setView('map'); } });

  if (isBootstrapping) return <div className="flex min-h-screen items-center justify-center bg-[#17110F] text-[#EEE1D3]"><Loader2 className="mr-3 h-6 w-6 animate-spin text-[#AEB974]" />Restoring Money Creek session</div>;
  if (!isAuthenticated) return <LoginScreen />;
  if (projectQuery.isLoading || cragQuery.isLoading || !area) return <div className="flex min-h-screen items-center justify-center bg-[#17110F] text-[#EEE1D3]"><Loader2 className="mr-3 h-6 w-6 animate-spin text-[#AEB974]" />Loading Money Creek</div>;

  const canEditCurrentArea = canWrite && Boolean(area.feature.parent_feature_id);
  const enter = (id: string) => { setMode('view'); setCurrentId(id); setSelectedBoulderId(null); setSelectedTrailId(null); setView('map'); setTab('overview'); if (mobile) setSheetExpanded(false); };
  const selectBoulder = (id: string | null) => { const parent = parentArea(root, id); if (parent) setCurrentId(parent.feature.id); setSelectedBoulderId(id); setSelectedTrailId(null); if (mobile) setSheetExpanded(Boolean(id)); };
  const selectTrail = (id: string | null) => { setSelectedTrailId(id); setSelectedBoulderId(null); setView('map'); if (mobile) setSheetExpanded(Boolean(id)); };
  const openBoulder = (id: string) => { const parent = parentArea(root, id); if (parent) setCurrentId(parent.feature.id); setSelectedBoulderId(id); setSelectedTrailId(null); setView('map'); if (mobile) setSheetExpanded(true); };
  const startCreate = (kind: Extract<Mode, 'create-area' | 'create-boulder'>) => { if (kind === 'create-boulder') { openAddBoulderModal(null, area.feature.id); return; } setMode(kind); setView('map'); setAddOpen(false); setDrawer(false); setSelectedBoulderId(null); setSelectedTrailId(null); setLayers(l => ({ ...l, base: 'satellite' })); };
  const saveCreate = (name: string) => { if (!pending || !area) return; if (pending.kind === 'create-area') createArea.mutate({ name, parentId: area.feature.id, points: pending.points }); if (pending.kind === 'create-boulder') createBoulder.mutate({ name, parentId: area.feature.id, points: pending.points }); setPending(null); setMode('view'); setLayers(l => ({ ...l, base: 'stylized' })); };
  const startEditArea = () => { if (!canEditCurrentArea) return; setMode('edit-area'); setView('map'); setSheetExpanded(false); setSelectedBoulderId(null); setSelectedTrailId(null); setLayers(l => ({ ...l, base: 'satellite', areas: { ...l.areas, [area.feature.id]: true } })); };
  const saveAreaEdit = (points: MoneyPosition[]) => updateAreaGeometry.mutate({ id: area.feature.id, points });
  const cancelAreaEdit = () => { setMode('view'); setLayers(l => ({ ...l, base: 'stylized' })); };
  const confirmDeleteArea = () => { if (!canEditCurrentArea) return; setDeleteChoice(area); };
  const moveAreaToParent = (id: string, parentId: string | null, sortOrder?: number) => moveArea.mutate({ id, parentId, sortOrder });
  const goBack = () => {
    if (selectedTrailId) { selectTrail(null); return; }
    if (selectedBoulderId) { selectBoulder(null); return; }
    const parent = parentArea(root, area.feature.id);
    if (parent) enter(parent.feature.id);
  };
  const openAddBoulderModal = (position?: MoneyPosition | null, parentId?: string | null) => { setAddOpen(false); setDrawer(false); setMode('view'); setSelectedBoulderId(null); setSelectedTrailId(null); setAddBoulder({ position, parentId: parentId ?? area.feature.id }); };
  const saveBoulderFromModal = (payload: AddBoulderPayload) => createBoulder.mutate({ name: payload.name, parentId: payload.parentId, position: payload.position, devStatus: payload.devStatus });
  const submitMapSearch = (event: React.FormEvent) => {
    event.preventDefault();
    const parsed = parseCoordinateSearch(mapSearch);
    if (!parsed) { setMapSearchError('Enter coordinates like 47.6997, -121.4703.'); return; }
    setMapSearchError(null);
    setView('map');
    setGoToPoint({ position: parsed.position, nonce: Date.now() });
  };
  const breadcrumbs = pathTo(root, area.feature.id);
  const canGoBack = Boolean(selectedTrailId || selectedBoulderId || parentArea(root, area.feature.id));

  return <div style={{ position: 'fixed', inset: 0, display: 'flex', flexDirection: 'column', background: T.app, color: T.ink, fontFamily: T.font }}>
    <header style={{ flexShrink: 0, height: 54, background: T.surf, borderBottom: `1px solid ${T.line}`, display: 'flex', alignItems: 'center', padding: '0 12px', gap: 10, zIndex: 40 }}>
      {mobile ? <button onClick={() => setDrawer(true)} style={iconBtn}><Menu size={20} /></button> : <Logo />}
      {canGoBack && <button onClick={goBack} aria-label="Back to parent" style={{ ...iconBtn, border: `1px solid ${T.line2}`, borderRadius: 8, background: T.inset }}><ArrowLeft size={17} /></button>}
      <div style={{ display: 'flex', alignItems: 'center', gap: 5, flex: 1, minWidth: 0, overflow: 'hidden' }}>{breadcrumbs.slice(mobile ? -2 : 0).map((n, i, arr) => { const isCurrentArea = !selectedBoulder && !selectedTrail && i === arr.length - 1; return <span key={n.feature.id} style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>{i > 0 && <span style={{ color: T.faint, fontFamily: T.mono }}>/</span>}<button aria-label={`Go to ${n.feature.title}`} onClick={() => enter(n.feature.id)} style={{ border: 'none', background: 'transparent', color: isCurrentArea ? T.ink : T.mut, fontSize: isCurrentArea ? 15 : 13.5, fontWeight: isCurrentArea ? 800 : 600, cursor: 'pointer', whiteSpace: 'nowrap' }}>{n.feature.title}</button></span>; })}{selectedBoulder && <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}><span style={{ color: T.faint, fontFamily: T.mono }}>/</span><button aria-label={`Go to ${selectedBoulder.feature.title}`} onClick={() => selectBoulder(selectedBoulder.feature.id)} style={{ border: 'none', background: 'transparent', color: T.ink, fontSize: 15, fontWeight: 800, cursor: 'pointer', whiteSpace: 'nowrap' }}>{selectedBoulder.feature.title}</button></span>}{selectedTrail && <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}><span style={{ color: T.faint, fontFamily: T.mono }}>/</span><button aria-label={`Go to ${selectedTrail.feature.title}`} onClick={() => selectTrail(selectedTrail.feature.id)} style={{ border: 'none', background: 'transparent', color: T.ink, fontSize: 15, fontWeight: 800, cursor: 'pointer', whiteSpace: 'nowrap' }}>{selectedTrail.feature.title}</button></span>}</div>
      {!mobile && <form onSubmit={submitMapSearch} title={mapSearchError ?? 'Search by coordinates'} style={{ display: 'flex', alignItems: 'center', gap: 8, background: T.inset, border: `1px solid ${mapSearchError ? '#B65B4D' : T.line}`, borderRadius: 9, padding: '7px 11px', width: 230, color: T.mut }}><Search size={16} /><input aria-label="Map search" value={mapSearch} onChange={e => { setMapSearch(e.target.value); setMapSearchError(null); }} placeholder="lat, lng" style={{ flex: 1, minWidth: 0, background: 'transparent', border: 'none', outline: 'none', color: T.ink, fontFamily: T.mono, fontSize: 12 }} /></form>}
      <button onClick={() => void logout()} title={user?.email} style={iconBtn}><LogOut size={17} /></button>
      {canWrite && <div style={{ position: 'relative' }}><button onClick={() => setAddOpen(!addOpen)} style={{ display: 'flex', alignItems: 'center', gap: 6, fontWeight: 700, fontSize: 13, color: T.onAccent, background: T.accent, border: 'none', borderRadius: 9, padding: mobile ? '9px 12px' : '9px 14px', minHeight: 40, cursor: 'pointer' }}><Plus size={16} />{!mobile && ' Add'}</button>{addOpen && <AddMenu startCreate={startCreate} openComposer={kind => setComposer({ kind })} />}</div>}
    </header>
    <div style={{ flex: 1, display: 'flex', minHeight: 0, position: 'relative' }}>
      {!mobile && <nav style={{ width: 238, flexShrink: 0, background: T.surf, borderRight: `1px solid ${T.line}`, padding: '12px 10px', overflowY: 'auto' }}><WorkspaceNav view={view} setView={v => setView(v)} root={root} trails={cragQuery.data?.trails ?? []} notes={notes} trashCount={trashQuery.data?.items.length ?? 0} /><div style={{ borderTop: `1px solid ${T.line}`, margin: '4px 4px 10px' }} />{isReferenceFilterView(view) ? <ReferenceFilterRail view={view} root={root} trails={cragQuery.data?.trails ?? []} notes={notes} uploads={uploads} filters={referenceFilters} setFilters={setReferenceFilters} currentAreaId={area.feature.id} /> : <TreeRail root={root} currentId={area.feature.id} canWrite={canWrite} enter={enter} onMove={moveAreaToParent} />}</nav>}
      {view === 'map' ? <>
        <div style={{ flex: 1, position: 'relative', minWidth: 0 }}><CragMap root={root as MoneyCragNode} area={area} trails={cragQuery.data?.trails ?? []} selectedBoulderId={selectedBoulderId} selectedTrailId={selectedTrailId} mode={mode} layers={layers} setLayers={setLayers} mobile={mobile} goToPoint={goToPoint} onEnter={enter} onSelectBoulder={selectBoulder} onSelectTrail={selectTrail} onAddBoulderAt={(position, parentId) => openAddBoulderModal(position, parentId)} onCreateDone={points => setPending({ kind: mode === 'create-boulder' ? 'create-boulder' : 'create-area', points })} onCreateCancel={() => { setMode('view'); setLayers(l => ({ ...l, base: 'stylized' })); }} onEditSave={saveAreaEdit} onEditCancel={cancelAreaEdit} /></div>
        {mode === 'view' && <DetailPanel root={root} area={area} selectedBoulder={selectedBoulder as MoneyCragNode | null} selectedTrail={selectedTrail} notes={notes} uploads={uploads} tab={tab} setTab={setTab} mobile={mobile} expanded={sheetExpanded} setExpanded={setSheetExpanded} canWrite={canWrite} canEditArea={canEditCurrentArea} onEditArea={startEditArea} onDeleteArea={confirmDeleteArea} onEnter={enter} onSelectBoulder={id => selectBoulder(id)} onNewArea={() => startCreate('create-area')} onNewBoulder={() => startCreate('create-boulder')} onSetDev={(id, dev) => updateDev.mutate({ id, dev })} onRenameBoulder={(boulder, title) => renameBoulder.mutate({ boulder, title })} onAddProblem={(boulderId, p) => createProblem.mutate({ boulderId, p })} onOpenComposer={kind => setComposer({ kind })} onEditNote={note => setComposer({ note })} onDeleteNote={note => deleteNote.mutate(note.id)} onDeleteUpload={upload => deleteUpload.mutate(upload.id)} onUpdateUploadMetadata={(upload, metadata) => updateUploadMetadata.mutateAsync({ upload, ...metadata })} onSaveTopo={(problem, overlay) => updateTopo.mutate({ problem: problem.feature, overlay })} onUpdateTrail={(trail, updates) => updateTrail.mutate({ trail, updates })} onDeleteTrail={trail => deleteTrail.mutate(trail)} />}
      </> : <ContentView view={view} root={root} trails={cragQuery.data?.trails ?? []} notes={notes} uploads={uploads} trash={trashQuery.data?.items ?? []} canWrite={canWrite} mobile={mobile} filters={referenceFilters} setFilters={setReferenceFilters} currentAreaId={area.feature.id} openBoulder={openBoulder} selectTrail={selectTrail} onAddBoulder={parentId => openAddBoulderModal(null, parentId ?? area.feature.id)} onOpenComposer={kind => setComposer({ kind })} onEditNote={note => setComposer({ note })} onDeleteNote={note => deleteNote.mutate(note.id)} onDeleteUpload={upload => deleteUpload.mutate(upload.id)} onUpdateUploadMetadata={(upload, metadata) => updateUploadMetadata.mutateAsync({ upload, ...metadata })} onRestore={id => restoreArea.mutate(id)} onCreateTrail={trail => createTrail.mutate(trail)} onUpdateTrail={(trail, updates) => updateTrail.mutate({ trail, updates })} onSaveTopo={(problem, overlay) => updateTopo.mutate({ problem: problem.feature, overlay })} />}
    </div>
    {mobile && drawer && <div onClick={() => setDrawer(false)} style={{ position: 'fixed', inset: 0, zIndex: 50, background: 'rgba(8,5,4,0.55)' }}><div onClick={e => e.stopPropagation()} style={{ position: 'absolute', top: 0, left: 0, bottom: 0, width: 290, background: T.surf, borderRight: `1px solid ${T.line2}`, padding: '14px 10px', overflowY: 'auto' }}><div style={{ display: 'flex', alignItems: 'center', gap: 9, padding: '4px 8px 14px' }}><Logo /><b>Money Creek</b></div><WorkspaceNav view={view} setView={v => { setView(v); setDrawer(false); }} root={root} trails={cragQuery.data?.trails ?? []} notes={notes} trashCount={trashQuery.data?.items.length ?? 0} /><div style={{ borderTop: `1px solid ${T.line}`, margin: '4px 4px 10px' }} />{isReferenceFilterView(view) ? <ReferenceFilterRail view={view} root={root} trails={cragQuery.data?.trails ?? []} notes={notes} uploads={uploads} filters={referenceFilters} setFilters={setReferenceFilters} currentAreaId={area.feature.id} /> : <TreeRail root={root} currentId={area.feature.id} canWrite={canWrite} enter={id => { enter(id); setDrawer(false); }} onMove={moveAreaToParent} />}</div></div>}
    {deleteChoice && <DeleteAreaDialog area={deleteChoice} onCancel={() => setDeleteChoice(null)} onDelete={mode => deleteArea.mutate({ id: deleteChoice.feature.id, mode })} />}
    {pending && <NameSheet kind={pending.kind} parentName={area.feature.title} onSave={saveCreate} onCancel={() => setPending(null)} />}
    {addBoulder && <AddBoulderModal root={root} initialPosition={addBoulder.position} initialParentId={addBoulder.parentId} fallbackParentId={area.feature.id} saving={createBoulder.isPending} onClose={() => setAddBoulder(null)} onSave={saveBoulderFromModal} />}
    {composer && <NoteComposer root={root} area={area} boulder={selectedBoulder as MoneyCragNode | null} initialBlock={composer.kind} initialNote={composer.note} uploads={uploads} mobile={mobile} onClose={() => setComposer(null)} onSubmit={payload => composer.note ? updateNoteAndDeleteRemovedUploads.mutate({ note: composer.note, payload }) : createNote.mutate(payload)} />}
  </div>;
}

function WorkspaceNav({ view, setView, root, trails, notes, trashCount }: { view: View; setView: (v: View) => void; root: MoneyCragNode | null; trails: MoneyCragNode[]; notes: unknown[]; trashCount: number }) { const items: Array<[View, string, React.ReactNode, number | null]> = [['map', 'Map', <Map size={18} />, null], ['problems', 'Problems', <FileText size={18} />, 0], ['boulders', 'Boulders', <Mountain size={18} />, 0], ['trails', 'Trails', <Route size={18} />, trails.length], ['photos', 'Photos', <Camera size={18} />, null], ['notes', 'Notes', <FileText size={18} />, notes.length], ['trash', 'Trash', <Trash2 size={18} />, trashCount]]; return <div style={{ marginBottom: 12 }}><NavLabel>Workspace</NavLabel>{items.map(([id, label, icon, count]) => <NavRow key={id} on={view === id} onClick={() => setView(id)} icon={icon} label={label} count={id === 'boulders' && root ? flattenAreas(root).reduce((n, a) => n + cragBoulders(a).length, 0) : id === 'problems' && root ? flattenAreas(root).reduce((n, a) => n + cragBoulders(a).reduce((m, b) => m + cragProblems(b).length, 0), 0) : count} />)}</div>; }
function TreeRail({ root, currentId, canWrite, enter, onMove }: { root: MoneyCragNode | null; currentId: string; canWrite: boolean; enter: (id: string) => void; onMove: (id: string, parentId: string | null, sortOrder?: number) => void }) { const [dragId, setDragId] = useState<string | null>(null); const rows = useMemo(() => { const out: Array<{ node: MoneyCragNode; depth: number; parentId: string | null }> = []; const walk = (node: MoneyCragNode, depth: number, parentId: string | null) => { out.push({ node, depth, parentId }); cragChildren(node).forEach(c => walk(c, depth + 1, node.feature.id)); }; if (root) walk(root, 0, null); return out; }, [root]); const rowById = (id: string) => rows.find(r => r.node.feature.id === id); const canMove = (id: string) => Boolean(canWrite && root && id !== root.feature.id); const moveNear = (id: string, dir: -1 | 1) => { const row = rowById(id); if (!row) return; onMove(id, row.parentId, (row.node.feature.sort_order ?? 0) + dir); }; const moveUpLevel = (id: string) => { const row = rowById(id); if (!row?.parentId) return; const parent = rowById(row.parentId); onMove(id, parent?.parentId ?? null, (row.node.feature.sort_order ?? 0) + 1); }; return <div><NavLabel>Areas</NavLabel>{rows.map(({ node, depth }) => { const movable = canMove(node.feature.id); return <div key={node.feature.id} draggable={movable} onDragStart={e => { if (!movable) return; setDragId(node.feature.id); e.dataTransfer.setData('text/plain', node.feature.id); e.dataTransfer.effectAllowed = 'move'; }} onDragOver={e => { if (dragId && dragId !== node.feature.id) e.preventDefault(); }} onDrop={e => { e.preventDefault(); const id = e.dataTransfer.getData('text/plain') || dragId; if (id && id !== node.feature.id) onMove(id, node.feature.id, 0); setDragId(null); }} onDragEnd={() => setDragId(null)} onClick={() => enter(node.feature.id)} style={{ display: 'flex', alignItems: 'center', gap: 6, padding: '8px 8px', paddingLeft: 10 + depth * 14, borderRadius: 8, marginBottom: 1, color: node.feature.id === currentId ? T.accent : T.mut, background: node.feature.id === currentId ? T.accentSoft : 'transparent', cursor: 'pointer', fontSize: 13.5, fontWeight: node.feature.id === currentId ? 800 : 500 }}><span style={{ flex: 1, minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{node.feature.title}</span>{movable && <span onClick={e => e.stopPropagation()} style={{ display: 'inline-flex', gap: 2 }}><button title="Move up" onClick={() => moveNear(node.feature.id, -1)} style={miniBtn}>↑</button><button title="Move down" onClick={() => moveNear(node.feature.id, 1)} style={miniBtn}>↓</button><button title="Promote one level" onClick={() => moveUpLevel(node.feature.id)} style={miniBtn}>↰</button></span>}</div>; })}</div>; }
function AddMenu({ startCreate, openComposer }: { startCreate: (m: Extract<Mode, 'create-area' | 'create-boulder'>) => void; openComposer: (k?: 'photo' | 'sketch' | 'file' | null) => void }) { const items: Array<[string, () => void]> = [['New sub-area', () => startCreate('create-area')], ['New boulder', () => startCreate('create-boulder')], ['Add a note', () => openComposer(null)], ['Add a photo', () => openComposer('photo')]]; return <><div onClick={() => null} style={{ position: 'fixed', inset: 0, zIndex: 44 }} /><div style={{ position: 'absolute', top: '120%', right: 0, width: 200, background: T.raise, border: `1px solid ${T.line2}`, borderRadius: 11, boxShadow: T.shadow, padding: 6, zIndex: 45 }}>{items.map(([label, fn]) => <button key={label} onClick={fn} style={{ width: '100%', border: 'none', background: 'transparent', color: T.ink, padding: 10, borderRadius: 7, cursor: 'pointer', fontWeight: 600, textAlign: 'left' }}>{label}</button>)}</div></>; }
function NameSheet({ kind, parentName, onSave, onCancel }: { kind: Extract<Mode, 'create-area' | 'create-boulder'>; parentName: string; onSave: (name: string) => void; onCancel: () => void }) { const [name, setName] = useState(''); const area = kind === 'create-area'; return <div onClick={onCancel} style={{ position: 'fixed', inset: 0, zIndex: 60, background: 'rgba(8,5,4,0.6)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}><div onClick={e => e.stopPropagation()} style={{ background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 16, boxShadow: T.shadow, padding: 24, width: 380 }}><div style={{ fontSize: 18, fontWeight: 800 }}>{area ? 'Name this area' : 'Name this boulder'}</div><div style={{ fontSize: 12.5, color: T.mut, margin: '4px 0 16px' }}>Added to <b style={{ color: T.ink }}>{parentName}</b>.</div><input autoFocus value={name} onChange={e => setName(e.target.value)} onKeyDown={e => { if (e.key === 'Enter' && name.trim()) onSave(name.trim()); }} style={{ width: '100%', background: T.surf2, border: `1px solid ${T.line2}`, borderRadius: 10, padding: '12px 14px', color: T.ink }} /><div style={{ display: 'flex', gap: 8, marginTop: 16 }}><button disabled={!name.trim()} onClick={() => onSave(name.trim())} style={{ flex: 1, border: 'none', borderRadius: 10, padding: 13, background: name.trim() ? T.accent : T.line2, color: name.trim() ? T.onAccent : T.faint, fontWeight: 700 }}>Create</button><button onClick={onCancel} style={{ border: `1px solid ${T.line2}`, borderRadius: 10, padding: '13px 18px', background: 'transparent', color: T.ink, fontWeight: 700 }}>Cancel</button></div></div></div>; }
function DeleteAreaDialog({ area, onCancel, onDelete }: { area: MoneyCragNode; onCancel: () => void; onDelete: (mode: MoneyArchiveMode) => void }) { return <div onClick={onCancel} style={{ position: 'fixed', inset: 0, zIndex: 70, background: 'rgba(8,5,4,0.62)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 16 }}><div onClick={e => e.stopPropagation()} style={{ background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 16, boxShadow: T.shadow, padding: 22, width: 430, maxWidth: '100%' }}><div style={{ fontSize: 18, fontWeight: 800, marginBottom: 6 }}>Delete {area.feature.title}</div><p style={{ color: T.mut, fontSize: 13, lineHeight: 1.45, margin: '0 0 16px' }}>Choose whether to trash the whole area tree, or archive only this area and move its direct sub-areas, boulders, problems, and trails to the parent area. Restoring this container later will not pull promoted contents back automatically.</p><div style={{ display: 'grid', gap: 8 }}><button onClick={() => onDelete('subtree')} style={dangerBtn}>Delete area + subtree</button><button onClick={() => onDelete('promote_children')} style={secondaryBtn}>Delete only this area, promote contents to parent</button><button onClick={onCancel} style={ghostBtn}>Cancel</button></div></div></div>; }
function Logo() { return <div style={{ width: 30, height: 30, borderRadius: 8, background: T.accent, color: T.onAccent, display: 'flex', alignItems: 'center', justifyContent: 'center' }}><Mountain size={18} /></div>; }
function NavLabel({ children }: { children: React.ReactNode }) { return <div style={{ fontFamily: T.mono, fontSize: 10.5, letterSpacing: 1, color: T.faint, textTransform: 'uppercase', padding: '4px 10px 8px' }}>{children}</div>; }
function NavRow({ on, onClick, icon, label, count }: { on: boolean; onClick: () => void; icon: React.ReactNode; label: string; count: number | null }) { return <div onClick={onClick} style={{ display: 'flex', alignItems: 'center', gap: 11, padding: '9px 11px', borderRadius: 9, marginBottom: 1, background: on ? T.accentSoft : 'transparent', color: on ? T.accent : T.mut, cursor: 'pointer', fontWeight: on ? 700 : 500, fontSize: 14 }}>{icon}<span style={{ flex: 1 }}>{label}</span>{count != null && <span style={{ fontFamily: T.mono, fontSize: 11, color: on ? T.accent : T.faint }}>{count}</span>}</div>; }
const iconBtn: React.CSSProperties = { border: 'none', background: 'transparent', color: T.ink, cursor: 'pointer', padding: 6, display: 'flex' };
const miniBtn: React.CSSProperties = { border: `1px solid ${T.line2}`, background: T.inset, color: T.mut, borderRadius: 4, cursor: 'pointer', fontSize: 10, padding: '1px 4px' };
const dangerBtn: React.CSSProperties = { border: 'none', borderRadius: 10, padding: 12, background: '#B65B4D', color: '#fff', fontWeight: 800, cursor: 'pointer', textAlign: 'left' };
const secondaryBtn: React.CSSProperties = { border: `1px solid ${T.line2}`, borderRadius: 10, padding: 12, background: T.raise, color: T.ink, fontWeight: 800, cursor: 'pointer', textAlign: 'left' };
const ghostBtn: React.CSSProperties = { border: 'none', borderRadius: 10, padding: 12, background: 'transparent', color: T.mut, fontWeight: 700, cursor: 'pointer' };

function isFeatureNoteTarget(targetType: MoneyNoteTargetType): boolean {
  return targetType === 'feature' || targetType === 'area' || targetType === 'boulder' || targetType === 'trail' || targetType === 'point';
}

function uploadIdsFromBlocks(blocks: MoneyNoteBlock[]): Set<string> {
  return new Set(blocks.map(block => block.upload_id).filter((id): id is string => Boolean(id)));
}

export function resolveSubmittedBlocks(blocks: MoneyNoteBlock[], uploaded: Map<string, MoneyNoteBlock>): MoneyNoteBlock[] {
  return blocks.flatMap(block => {
    const localBlockKey = typeof block.metadata?.local_block_key === 'string' ? block.metadata.local_block_key : null;
    if (localBlockKey) return uploaded.has(localBlockKey) ? [uploaded.get(localBlockKey)!] : [];
    if (block.url?.startsWith('blob:')) return [];
    return [block];
  });
}

export default function MoneyCreekApp() { return <AuthProvider><MoneyCreekWorkspace /></AuthProvider>; }
