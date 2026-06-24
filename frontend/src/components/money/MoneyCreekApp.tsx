import { useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Camera, FileText, Loader2, LogOut, Map, Menu, Mountain, Plus, Route, Search } from 'lucide-react';
import { AuthProvider, useAuth } from '../../contexts/AuthContext';
import { moneyApi } from '../../services/money';
import { MoneyCragNode, MoneyDevStatus, MoneyNoteBlock, MoneyPosition, MoneyProblemStatus } from '../../types/money';
import { LoginScreen } from './LoginScreen';
import { ContentView } from './reference/ContentViews';
import { CragMap } from './reference/CragMap';
import { DetailPanel } from './reference/DetailPanel';
import { cragBoulders, cragChildren, cragProblems, findNode, flattenAreas, parentArea, pathTo, polygonGeoJSON } from './reference/model';
import { NoteComposer } from './reference/NoteComposer';
import { T } from './reference/theme';

const PROJECT_SLUG = 'money-creek';
type View = 'map' | 'problems' | 'boulders' | 'trails' | 'photos' | 'notes';
type Mode = 'view' | 'create-area' | 'create-boulder';
type LayersState = { base: string; contours: boolean; trails: boolean; areas: Record<string, boolean>; dev: Record<string, boolean> };

function useMobile() { const [m, setM] = useState(() => window.innerWidth < 760); useEffect(() => { const r = () => setM(window.innerWidth < 760); window.addEventListener('resize', r); return () => window.removeEventListener('resize', r); }, []); return m; }

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
  const [pending, setPending] = useState<{ kind: Mode; points: MoneyPosition[] } | null>(null);
  const [drawer, setDrawer] = useState(false);
  const [addOpen, setAddOpen] = useState(false);
  const [sheetExpanded, setSheetExpanded] = useState(false);
  const [composer, setComposer] = useState<{ kind?: 'photo' | 'sketch' | 'file' | null } | null>(null);
  const [layers, setLayers] = useState<LayersState>({ base: 'stylized', contours: true, trails: true, areas: {}, dev: { scouted: true, 'needs-work': true, cleaning: true, established: true } });

  const projectQuery = useQuery({ queryKey: ['money-project', PROJECT_SLUG], queryFn: () => moneyApi.getProject(PROJECT_SLUG), enabled: isAuthenticated, staleTime: 10 * 60 * 1000 });
  const projectId = projectQuery.data?.project.id;
  const cragQuery = useQuery({ queryKey: ['money-crag', projectId], queryFn: () => moneyApi.getCragSnapshot(projectId!), enabled: Boolean(projectId), staleTime: 30 * 1000 });
  const root = cragQuery.data?.root ?? null;
  useEffect(() => { if (root && !currentId) setCurrentId(root.feature.id); }, [root, currentId]);
  const area = findNode(root, currentId) ?? root;
  const selectedBoulder = findNode(root, selectedBoulderId);
  const selectedTrail = cragQuery.data?.trails?.find(t => t.feature.id === selectedTrailId) ?? null;
  const notes = cragQuery.data?.notes ?? [];

  const invalidate = () => void queryClient.invalidateQueries({ queryKey: ['money-crag', projectId] });
  const createArea = useMutation({ mutationFn: ({ name, parentId, points }: { name: string; parentId?: string | null; points: MoneyPosition[] }) => moneyApi.createArea(projectId!, { parent_feature_id: parentId, title: name, description: `Freshly outlined — add sub-areas or boulders to fill it in.`, geojson: polygonGeoJSON(points), properties: { kind: 'Boulders', aspect: 'newly mapped' } }), onSuccess: f => { invalidate(); setCurrentId(f.id); } });
  const createBoulder = useMutation({ mutationFn: ({ name, parentId, points }: { name: string; parentId: string; points: MoneyPosition[] }) => moneyApi.createBoulder(projectId!, { parent_feature_id: parentId, title: name, description: null, dev_status: 'scouted', geojson: polygonGeoJSON(points), properties: {} }), onSuccess: f => { invalidate(); setSelectedBoulderId(f.id); } });
  const updateDev = useMutation({ mutationFn: ({ id, dev }: { id: string; dev: MoneyDevStatus }) => moneyApi.updateBoulderStatus(id, { dev_status: dev }), onSuccess: invalidate });
  const createProblem = useMutation({ mutationFn: ({ boulderId, p }: { boulderId: string; p: { name: string; grade: string; status: MoneyProblemStatus; stars: number; fa?: string | null; types: string[] } }) => moneyApi.createProblem(projectId!, { boulder_id: boulderId, ...p }), onSuccess: invalidate });
  const createNote = useMutation({ mutationFn: async (payload: { body: string; tags: string[]; target_type: string; target_ref?: string; blocks: MoneyNoteBlock[]; files: Array<{ file: File; kind: 'photo' | 'file' }> }) => {
    const uploaded: MoneyNoteBlock[] = [];
    for (const item of payload.files) { const u = await moneyApi.uploadImage(projectId!, item.file, { blockKind: item.kind }); uploaded.push({ kind: item.kind, upload_id: u.id, name: u.original_filename }); }
    return moneyApi.createProjectNote(projectId!, { body: payload.body, visibility: 'team', target_type: payload.target_type as never, target_ref: payload.target_ref, tags: payload.tags, blocks: [...payload.blocks, ...uploaded] });
  }, onSuccess: () => { invalidate(); setComposer(null); if (mobile) setSheetExpanded(true); } });

  if (isBootstrapping) return <div className="flex min-h-screen items-center justify-center bg-[#17110F] text-[#EEE1D3]"><Loader2 className="mr-3 h-6 w-6 animate-spin text-[#AEB974]" />Restoring Money Creek session</div>;
  if (!isAuthenticated) return <LoginScreen />;
  if (projectQuery.isLoading || cragQuery.isLoading || !area) return <div className="flex min-h-screen items-center justify-center bg-[#17110F] text-[#EEE1D3]"><Loader2 className="mr-3 h-6 w-6 animate-spin text-[#AEB974]" />Loading Money Creek</div>;

  const enter = (id: string) => { setCurrentId(id); setSelectedBoulderId(null); setSelectedTrailId(null); setView('map'); setTab('overview'); if (mobile) setSheetExpanded(false); };
  const selectBoulder = (id: string | null) => { setSelectedBoulderId(id); setSelectedTrailId(null); if (mobile) setSheetExpanded(Boolean(id)); };
  const selectTrail = (id: string | null) => { setSelectedTrailId(id); setSelectedBoulderId(null); setView('map'); if (mobile) setSheetExpanded(Boolean(id)); };
  const openBoulder = (id: string) => { const parent = parentArea(root, id); if (parent) setCurrentId(parent.feature.id); setSelectedBoulderId(id); setSelectedTrailId(null); setView('map'); if (mobile) setSheetExpanded(true); };
  const startCreate = (kind: Mode) => { setMode(kind); setView('map'); setAddOpen(false); setDrawer(false); setSelectedBoulderId(null); setSelectedTrailId(null); setLayers(l => ({ ...l, base: 'satellite' })); };
  const saveCreate = (name: string) => { if (!pending || !area) return; if (pending.kind === 'create-area') createArea.mutate({ name, parentId: area.feature.id, points: pending.points }); if (pending.kind === 'create-boulder') createBoulder.mutate({ name, parentId: area.feature.id, points: pending.points }); setPending(null); setMode('view'); setLayers(l => ({ ...l, base: 'stylized' })); };
  const breadcrumbs = pathTo(root, area.feature.id);

  return <div style={{ position: 'fixed', inset: 0, display: 'flex', flexDirection: 'column', background: T.app, color: T.ink, fontFamily: T.font }}>
    <header style={{ flexShrink: 0, height: 54, background: T.surf, borderBottom: `1px solid ${T.line}`, display: 'flex', alignItems: 'center', padding: '0 12px', gap: 10, zIndex: 40 }}>
      {mobile ? <button onClick={() => setDrawer(true)} style={iconBtn}><Menu size={20} /></button> : <Logo />}
      <div style={{ display: 'flex', alignItems: 'center', gap: 5, flex: 1, minWidth: 0, overflow: 'hidden' }}>{breadcrumbs.slice(mobile ? -2 : 0).map((n, i, arr) => <span key={n.feature.id} style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>{i > 0 && <span style={{ color: T.faint, fontFamily: T.mono }}>/</span>}<button onClick={() => i < arr.length - 1 && enter(n.feature.id)} style={{ border: 'none', background: 'transparent', color: i === arr.length - 1 ? T.ink : T.mut, fontSize: i === arr.length - 1 ? 15 : 13.5, fontWeight: i === arr.length - 1 ? 800 : 600, cursor: i === arr.length - 1 ? 'default' : 'pointer', whiteSpace: 'nowrap' }}>{n.feature.title}</button></span>)}</div>
      {!mobile && <div style={{ display: 'flex', alignItems: 'center', gap: 8, background: T.inset, border: `1px solid ${T.line}`, borderRadius: 9, padding: '7px 11px', width: 190, color: T.mut }}><Search size={16} /><input placeholder="Search" style={{ flex: 1, minWidth: 0, background: 'transparent', border: 'none', outline: 'none', color: T.ink, fontFamily: T.mono, fontSize: 12 }} /></div>}
      <button onClick={() => void logout()} title={user?.email} style={iconBtn}><LogOut size={17} /></button>
      {canWrite && <div style={{ position: 'relative' }}><button onClick={() => setAddOpen(!addOpen)} style={{ display: 'flex', alignItems: 'center', gap: 6, fontWeight: 700, fontSize: 13, color: T.onAccent, background: T.accent, border: 'none', borderRadius: 9, padding: mobile ? '9px 12px' : '9px 14px', minHeight: 40, cursor: 'pointer' }}><Plus size={16} />{!mobile && ' Add'}</button>{addOpen && <AddMenu startCreate={startCreate} openComposer={kind => setComposer({ kind })} />}</div>}
    </header>
    <div style={{ flex: 1, display: 'flex', minHeight: 0, position: 'relative' }}>{!mobile && <nav style={{ width: 210, flexShrink: 0, background: T.surf, borderRight: `1px solid ${T.line}`, padding: '12px 10px', overflowY: 'auto' }}><WorkspaceNav view={view} setView={v => setView(v)} root={root} trails={cragQuery.data?.trails ?? []} notes={notes} /><div style={{ borderTop: `1px solid ${T.line}`, margin: '4px 4px 10px' }} /><TreeRail root={root} currentId={area.feature.id} enter={enter} /></nav>}{view === 'map' ? <><div style={{ flex: 1, position: 'relative', minWidth: 0 }}><CragMap root={root as MoneyCragNode} area={area} trails={cragQuery.data?.trails ?? []} selectedBoulderId={selectedBoulderId} selectedTrailId={selectedTrailId} mode={mode} layers={layers} setLayers={setLayers} mobile={mobile} onEnter={enter} onSelectBoulder={selectBoulder} onSelectTrail={selectTrail} onCreateDone={points => setPending({ kind: mode, points })} onCreateCancel={() => { setMode('view'); setLayers(l => ({ ...l, base: 'stylized' })); }} /></div>{mode === 'view' && <DetailPanel area={area} selectedBoulder={selectedBoulder as MoneyCragNode | null} selectedTrail={selectedTrail} notes={notes} tab={tab} setTab={setTab} mobile={mobile} expanded={sheetExpanded} setExpanded={setSheetExpanded} canWrite={canWrite} onEnter={enter} onSelectBoulder={id => selectBoulder(id)} onNewArea={() => startCreate('create-area')} onNewBoulder={() => startCreate('create-boulder')} onSetDev={(id, dev) => updateDev.mutate({ id, dev })} onAddProblem={(boulderId, p) => createProblem.mutate({ boulderId, p })} onOpenComposer={kind => setComposer({ kind })} />}</> : <ContentView view={view} root={root} trails={cragQuery.data?.trails ?? []} notes={notes} mobile={mobile} openBoulder={openBoulder} selectTrail={id => selectTrail(id)} onOpenComposer={kind => setComposer({ kind })} />}</div>
    {mobile && drawer && <div onClick={() => setDrawer(false)} style={{ position: 'fixed', inset: 0, zIndex: 50, background: 'rgba(8,5,4,0.55)' }}><div onClick={e => e.stopPropagation()} style={{ position: 'absolute', top: 0, left: 0, bottom: 0, width: 270, background: T.surf, borderRight: `1px solid ${T.line2}`, padding: '14px 10px', overflowY: 'auto' }}><div style={{ display: 'flex', alignItems: 'center', gap: 9, padding: '4px 8px 14px' }}><Logo /><b>Money Creek</b></div><WorkspaceNav view={view} setView={v => { setView(v); setDrawer(false); }} root={root} trails={cragQuery.data?.trails ?? []} notes={notes} /><div style={{ borderTop: `1px solid ${T.line}`, margin: '4px 4px 10px' }} /><TreeRail root={root} currentId={area.feature.id} enter={id => { enter(id); setDrawer(false); }} /></div></div>}
    {pending && <NameSheet kind={pending.kind} parentName={area.feature.title} onSave={saveCreate} onCancel={() => setPending(null)} />}
    {composer && <NoteComposer root={root} area={area} boulder={selectedBoulder as MoneyCragNode | null} initialBlock={composer.kind} mobile={mobile} onClose={() => setComposer(null)} onSubmit={payload => createNote.mutate(payload)} />}
  </div>;
}

function WorkspaceNav({ view, setView, root, trails, notes }: { view: View; setView: (v: View) => void; root: MoneyCragNode | null; trails: MoneyCragNode[]; notes: unknown[] }) { const items: Array<[View, string, React.ReactNode, number | null]> = [['map', 'Map', <Map size={18} />, null], ['problems', 'Problems', <FileText size={18} />, 0], ['boulders', 'Boulders', <Mountain size={18} />, 0], ['trails', 'Trails', <Route size={18} />, trails.length], ['photos', 'Photos', <Camera size={18} />, null], ['notes', 'Notes', <FileText size={18} />, notes.length]]; return <div style={{ marginBottom: 12 }}><NavLabel>Workspace</NavLabel>{items.map(([id, label, icon, count]) => <NavRow key={id} on={view === id} onClick={() => setView(id)} icon={icon} label={label} count={id === 'boulders' && root ? flattenAreas(root).reduce((n, a) => n + cragBoulders(a).length, 0) : id === 'problems' && root ? flattenAreas(root).reduce((n, a) => n + cragBoulders(a).reduce((m, b) => m + cragProblems(b).length, 0), 0) : count} />)}</div>; }
function TreeRail({ root, currentId, enter }: { root: MoneyCragNode | null; currentId: string; enter: (id: string) => void }) { const rows = useMemo(() => { const out: Array<{ node: MoneyCragNode; depth: number }> = []; const walk = (node: MoneyCragNode, depth: number) => { out.push({ node, depth }); cragChildren(node).forEach(c => walk(c, depth + 1)); }; if (root) walk(root, 0); return out; }, [root]); return <div><NavLabel>Areas</NavLabel>{rows.map(({ node, depth }) => <div key={node.feature.id} onClick={() => enter(node.feature.id)} style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '8px 10px', paddingLeft: 10 + depth * 14, borderRadius: 8, marginBottom: 1, background: node.feature.id === currentId ? T.accentSoft : 'transparent', color: node.feature.id === currentId ? T.accent : T.mut, cursor: 'pointer', fontWeight: node.feature.id === currentId ? 700 : 500, fontSize: 13.5 }}><span style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{node.feature.title}</span><span style={{ fontFamily: T.mono, fontSize: 10, color: T.faint }}>{cragChildren(node).length || cragBoulders(node).length}</span></div>)}</div>; }
function AddMenu({ startCreate, openComposer }: { startCreate: (m: Mode) => void; openComposer: (k?: 'photo' | 'sketch' | 'file' | null) => void }) { const items: Array<[string, () => void]> = [['New sub-area', () => startCreate('create-area')], ['New boulder', () => startCreate('create-boulder')], ['Add a note', () => openComposer(null)], ['Add a photo', () => openComposer('photo')]]; return <><div onClick={() => null} style={{ position: 'fixed', inset: 0, zIndex: 44 }} /><div style={{ position: 'absolute', top: '120%', right: 0, width: 200, background: T.raise, border: `1px solid ${T.line2}`, borderRadius: 11, boxShadow: T.shadow, padding: 6, zIndex: 45 }}>{items.map(([label, fn]) => <button key={label} onClick={fn} style={{ width: '100%', border: 'none', background: 'transparent', color: T.ink, padding: 10, borderRadius: 7, cursor: 'pointer', fontWeight: 600, textAlign: 'left' }}>{label}</button>)}</div></>; }
function NameSheet({ kind, parentName, onSave, onCancel }: { kind: Mode; parentName: string; onSave: (name: string) => void; onCancel: () => void }) { const [name, setName] = useState(''); const area = kind === 'create-area'; return <div onClick={onCancel} style={{ position: 'fixed', inset: 0, zIndex: 60, background: 'rgba(8,5,4,0.6)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}><div onClick={e => e.stopPropagation()} style={{ background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 16, boxShadow: T.shadow, padding: 24, width: 380 }}><div style={{ fontSize: 18, fontWeight: 800 }}>{area ? 'Name this area' : 'Name this boulder'}</div><div style={{ fontSize: 12.5, color: T.mut, margin: '4px 0 16px' }}>Added to <b style={{ color: T.ink }}>{parentName}</b>.</div><input autoFocus value={name} onChange={e => setName(e.target.value)} onKeyDown={e => { if (e.key === 'Enter' && name.trim()) onSave(name.trim()); }} style={{ width: '100%', background: T.surf2, border: `1px solid ${T.line2}`, borderRadius: 10, padding: '12px 14px', color: T.ink }} /><div style={{ display: 'flex', gap: 8, marginTop: 16 }}><button disabled={!name.trim()} onClick={() => onSave(name.trim())} style={{ flex: 1, border: 'none', borderRadius: 10, padding: 13, background: name.trim() ? T.accent : T.line2, color: name.trim() ? T.onAccent : T.faint, fontWeight: 700 }}>Create</button><button onClick={onCancel} style={{ border: `1px solid ${T.line2}`, borderRadius: 10, padding: '13px 18px', background: 'transparent', color: T.ink, fontWeight: 700 }}>Cancel</button></div></div></div>; }
function Logo() { return <div style={{ width: 30, height: 30, borderRadius: 8, background: T.accent, color: T.onAccent, display: 'flex', alignItems: 'center', justifyContent: 'center' }}><Mountain size={18} /></div>; }
function NavLabel({ children }: { children: React.ReactNode }) { return <div style={{ fontFamily: T.mono, fontSize: 10.5, letterSpacing: 1, color: T.faint, textTransform: 'uppercase', padding: '4px 10px 8px' }}>{children}</div>; }
function NavRow({ on, onClick, icon, label, count }: { on: boolean; onClick: () => void; icon: React.ReactNode; label: string; count: number | null }) { return <div onClick={onClick} style={{ display: 'flex', alignItems: 'center', gap: 11, padding: '9px 11px', borderRadius: 9, marginBottom: 1, background: on ? T.accentSoft : 'transparent', color: on ? T.accent : T.mut, cursor: 'pointer', fontWeight: on ? 700 : 500, fontSize: 14 }}>{icon}<span style={{ flex: 1 }}>{label}</span>{count != null && <span style={{ fontFamily: T.mono, fontSize: 11, color: on ? T.accent : T.faint }}>{count}</span>}</div>; }
const iconBtn: React.CSSProperties = { border: 'none', background: 'transparent', color: T.ink, cursor: 'pointer', padding: 6, display: 'flex' };

export default function MoneyCreekApp() { return <AuthProvider><MoneyCreekWorkspace /></AuthProvider>; }
