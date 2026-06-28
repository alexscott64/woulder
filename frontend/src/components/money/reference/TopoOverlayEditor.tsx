import { useMemo, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { Edit3, RotateCcw, Save, Trash2 } from 'lucide-react';
import { MoneyCragNode, MoneyUpload } from '../../../types/money';
import { cragProblems } from './model';
import { T } from './theme';
import { NormalizedPoint, TopoOverlay, TopoStartMarker, TopoStartMarkerType, displayStartMarkerLabel, labelForStartMarkerType, overlayPathD, overlaysForUpload, pointFromClient, startMarkerSegments, topoOverlaysForFeature } from './topoOverlay';

interface TopoOverlayEditorProps {
  root: MoneyCragNode | null;
  boulder?: MoneyCragNode | null;
  upload: MoneyUpload;
  canWrite: boolean;
  activeProblemId?: string | null;
  imageSrc?: string | null;
  imageAlt?: string;
  controlsPortalTarget?: HTMLElement | null;
  onSave: (problem: MoneyCragNode, overlay: TopoOverlay) => void;
}

type DraftPath = { id: string; points: NormalizedPoint[] };
type TopoTool = 'line' | 'start';

const colors = ['#F97316', '#38BDF8', '#A3E635', '#F472B6', '#FDE047'];

export function TopoOverlayEditor({ root, boulder, upload, canWrite, activeProblemId, imageSrc, imageAlt, controlsPortalTarget, onSave }: TopoOverlayEditorProps) {
  const boxRef = useRef<HTMLDivElement | null>(null);
  const availableProblems = useMemo(() => boulder ? cragProblems(boulder) : overlaysForUpload(root, upload.id).map(item => item.problem), [boulder, root, upload.id]);
  const overlays = useMemo(() => overlaysForUpload(root, upload.id), [root, upload.id]);
  const initialProblemId = activeProblemId ?? overlays[0]?.problem.feature.id ?? availableProblems[0]?.feature.id ?? '';
  const [problemId, setProblemId] = useState(initialProblemId);
  const selectedProblem = availableProblems.find(problem => problem.feature.id === problemId) ?? overlays.find(item => item.problem.feature.id === problemId)?.problem ?? availableProblems[0] ?? null;
  const selectedOverlay = selectedProblem ? topoOverlaysForFeature(selectedProblem.feature).find(item => item.upload_id === upload.id || item.photo_id === upload.id) : undefined;
  const [tool, setTool] = useState<TopoTool>('line');
  const [draftPaths, setDraftPaths] = useState<DraftPath[] | null>(null);
  const [draftStarts, setDraftStarts] = useState<TopoStartMarker[] | null>(null);
  const [activePath, setActivePath] = useState<DraftPath | null>(null);
  const [startMarkerType, setStartMarkerType] = useState<TopoStartMarkerType>('generic');
  const [color, setColor] = useState(selectedOverlay?.color ?? colors[0]);
  const [width, setWidth] = useState(selectedOverlay?.width ?? 5);
  const workingPaths = draftPaths ?? selectedOverlay?.paths ?? [];
  const workingStarts = draftStarts ?? selectedOverlay?.starts ?? [];
  const hasDraftChanges = draftPaths !== null || draftStarts !== null;
  const canSave = Boolean(canWrite && selectedProblem && hasDraftChanges);

  const pointForEvent = (event: React.PointerEvent<HTMLDivElement>): NormalizedPoint | null => {
    const rect = boxRef.current?.getBoundingClientRect();
    return rect ? pointFromClient(event.clientX, event.clientY, rect) : null;
  };

  const resetDrafts = (nextProblemId: string) => {
    setProblemId(nextProblemId);
    setDraftPaths(null);
    setDraftStarts(null);
    setActivePath(null);
  };

  const nextStartLabel = (type: TopoStartMarkerType) => labelForStartMarkerType(type);

  const start = (event: React.PointerEvent<HTMLDivElement>) => {
    if (!canWrite || !selectedProblem) return;
    const point = pointForEvent(event);
    if (!point) return;
    event.preventDefault();
    event.stopPropagation();
    if (tool === 'start') {
      setDraftStarts(starts => {
        const base = starts ?? selectedOverlay?.starts ?? [];
        return [...base, { id: crypto.randomUUID(), point, type: startMarkerType, label: nextStartLabel(startMarkerType) }];
      });
      return;
    }
    event.currentTarget.setPointerCapture?.(event.pointerId);
    setActivePath({ id: crypto.randomUUID(), points: [point] });
  };

  const move = (event: React.PointerEvent<HTMLDivElement>) => {
    if (!activePath || tool !== 'line') return;
    const point = pointForEvent(event);
    if (!point) return;
    event.preventDefault();
    event.stopPropagation();
    setActivePath(path => path ? { ...path, points: [...path.points, point] } : path);
  };

  const finish = (event?: React.PointerEvent<HTMLDivElement>) => {
    event?.stopPropagation();
    setActivePath(path => {
      if (path && path.points.length > 1) setDraftPaths(paths => [...(paths ?? selectedOverlay?.paths ?? []), path]);
      return null;
    });
  };

  const undo = () => {
    if (tool === 'start') setDraftStarts(starts => (starts ?? selectedOverlay?.starts ?? []).slice(0, -1));
    else setDraftPaths(paths => (paths ?? selectedOverlay?.paths ?? []).slice(0, -1));
  };

  const clear = () => {
    setDraftPaths([]);
    setDraftStarts([]);
    setActivePath(null);
  };

  const save = () => {
    if (!selectedProblem || !canSave) return;
    onSave(selectedProblem, {
      id: selectedOverlay?.id ?? crypto.randomUUID(),
      upload_id: upload.id,
      photo_id: upload.id,
      problem_id: selectedProblem.feature.id,
      label: selectedProblem.feature.title,
      color,
      width,
      order: selectedOverlay?.order ?? topoOverlaysForFeature(selectedProblem.feature).length,
      paths: workingPaths,
      starts: workingStarts,
      updated_at: new Date().toISOString(),
    });
    setDraftPaths(null);
    setDraftStarts(null);
  };

  const undoDisabled = !canWrite || (tool === 'start' ? workingStarts.length === 0 : workingPaths.length === 0);
  const cursor = !canWrite ? 'default' : tool === 'start' ? 'copy' : 'crosshair';
  const toolbar = <TopoToolbar
    availableProblems={availableProblems}
    selectedProblemId={selectedProblem?.feature.id ?? ''}
    tool={tool}
    startMarkerType={startMarkerType}
    color={color}
    width={width}
    canWrite={canWrite}
    canSave={canSave}
    undoDisabled={undoDisabled}
    hasTopo={workingPaths.length > 0 || workingStarts.length > 0}
    linkedCount={overlays.length}
    onProblemChange={resetDrafts}
    onToolChange={setTool}
    onStartMarkerTypeChange={setStartMarkerType}
    onColorChange={setColor}
    onWidthChange={setWidth}
    onUndo={undo}
    onClear={clear}
    onSave={save}
  />;

  return <section aria-label="Photo topo editor" className="money-topo-editor" style={{ width: '100%', height: '100%', minHeight: 0, minWidth: 0, display: 'grid', gridTemplateRows: controlsPortalTarget ? 'minmax(0,1fr)' : 'minmax(0,1fr) auto', gap: 12, alignItems: 'stretch', justifyItems: 'stretch', overflow: 'hidden' }}>
    <div className="money-topo-stage" style={{ position: 'relative', width: '100%', height: '100%', maxWidth: '100%', maxHeight: '100%', minHeight: 0, minWidth: 0, overflow: 'auto', display: 'flex', alignItems: 'center', justifyContent: 'center', lineHeight: 0 }}>
      <div
        ref={boxRef}
        data-testid="topo-photo-canvas"
        onPointerDown={start}
        onPointerMove={move}
        onPointerUp={finish}
        onPointerCancel={finish}
        style={{ position: 'relative', display: 'inline-block', flex: '0 0 auto', maxWidth: '100%', maxHeight: '100%', borderRadius: 10, overflow: 'hidden', touchAction: canWrite ? 'none' : 'auto', cursor, background: 'rgba(0,0,0,0.28)', boxShadow: '0 0 0 1px rgba(238,225,211,0.12)' }}
      >
        {imageSrc && <img src={imageSrc} alt={imageAlt ?? upload.original_filename} draggable={false} style={{ display: 'block', width: 'auto', height: 'auto', maxWidth: '100%', maxHeight: '100%', userSelect: 'none', pointerEvents: 'none', objectFit: 'contain' }} />}
        {!imageSrc && <div style={{ width: 'min(760px, 78vw)', height: 'min(520px, 62vh)' }} />}
        <svg viewBox="0 0 1000 1000" preserveAspectRatio="none" style={{ position: 'absolute', inset: 0, width: '100%', height: '100%', pointerEvents: 'none' }}>
          {overlays.map(({ overlay, problem }) => <g key={overlay.id} opacity={!activeProblemId || problem.feature.id === activeProblemId ? 1 : 0.35}><TopoOverlayMarks overlay={overlay} /></g>)}
          {[...workingPaths, ...(activePath ? [activePath] : [])].map(path => <path key={path.id} d={overlayPathD(path.points)} fill="none" stroke={color} strokeWidth={width * 2.2} strokeLinecap="round" strokeLinejoin="round" />)}
          {workingStarts.map(marker => <StartMarker key={marker.id} marker={marker} color={color} />)}
        </svg>
        {workingPaths.length === 0 && workingStarts.length === 0 && overlays.length === 0 && <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', color: T.ink, fontSize: 13, fontWeight: 800, textAlign: 'center', padding: 18, pointerEvents: 'none', textShadow: '0 1px 4px rgba(0,0,0,0.9)' }}>Draw on the enlarged photo. Coordinates are saved in normalized image space.</div>}
      </div>
    </div>
    {controlsPortalTarget ? createPortal(toolbar, controlsPortalTarget) : toolbar}
  </section>;
}

function TopoToolbar({ availableProblems, selectedProblemId, tool, startMarkerType, color, width, canWrite, canSave, undoDisabled, hasTopo, linkedCount, onProblemChange, onToolChange, onStartMarkerTypeChange, onColorChange, onWidthChange, onUndo, onClear, onSave }: {
  availableProblems: MoneyCragNode[];
  selectedProblemId: string;
  tool: TopoTool;
  startMarkerType: TopoStartMarkerType;
  color: string;
  width: number;
  canWrite: boolean;
  canSave: boolean;
  undoDisabled: boolean;
  hasTopo: boolean;
  linkedCount: number;
  onProblemChange: (problemId: string) => void;
  onToolChange: (tool: TopoTool) => void;
  onStartMarkerTypeChange: (type: TopoStartMarkerType) => void;
  onColorChange: (color: string) => void;
  onWidthChange: (width: number) => void;
  onUndo: () => void;
  onClear: () => void;
  onSave: () => void;
}) {
  const stop = (event: React.SyntheticEvent) => event.stopPropagation();

  return <div className="money-topo-toolbar" onClick={stop} onPointerDown={stop} style={toolbarStyle}>
    <style>{topoEditorCss}</style>
    <div className="money-topo-toolbar-title" style={{ display: 'flex', alignItems: 'center', gap: 8, minWidth: 0 }}>
      <Edit3 size={15} color={T.accent} />
      <div style={{ fontWeight: 900, color: T.ink, whiteSpace: 'nowrap' }}>Draw topo</div>
      {linkedCount > 0 && <span style={{ fontFamily: T.mono, fontSize: 10.5, color: T.mut, whiteSpace: 'nowrap' }}>{linkedCount} linked</span>}
    </div>
    <label className="money-topo-problem" style={{ ...label, minWidth: 160, flex: '1 1 190px' }}>Problem
      <select aria-label="Topo problem" value={selectedProblemId} onChange={event => onProblemChange(event.target.value)} style={field}>
        {availableProblems.length === 0 && <option value="">No problems on this boulder</option>}
        {availableProblems.map(problem => <option key={problem.feature.id} value={problem.feature.id}>{problem.feature.title}</option>)}
      </select>
    </label>
    <div className="money-topo-tools" style={{ display: 'flex', gap: 8 }}>
      <button type="button" aria-pressed={tool === 'line'} onClick={() => onToolChange('line')} style={toolButton(tool === 'line')}>Line</button>
      <button type="button" aria-pressed={tool === 'start'} onClick={() => onToolChange('start')} style={toolButton(tool === 'start')}>Start</button>
    </div>
    {tool === 'start' && <div className="money-topo-start-types" role="group" aria-label="Start marker type" style={{ display: 'flex', gap: 6 }}>
      <button type="button" aria-pressed={startMarkerType === 'generic'} onClick={() => onStartMarkerTypeChange('generic')} style={toolButton(startMarkerType === 'generic')}>X</button>
      <button type="button" aria-pressed={startMarkerType === 'left'} onClick={() => onStartMarkerTypeChange('left')} style={toolButton(startMarkerType === 'left')}>Left</button>
      <button type="button" aria-pressed={startMarkerType === 'right'} onClick={() => onStartMarkerTypeChange('right')} style={toolButton(startMarkerType === 'right')}>Right</button>
    </div>}
    <div className="money-topo-colors" style={{ display: 'flex', gap: 5 }}>{colors.map(option => <button key={option} type="button" aria-label={`Topo color ${option}`} onClick={() => onColorChange(option)} style={{ width: 30, height: 30, borderRadius: 9, border: color === option ? `2px solid ${T.ink}` : `1px solid ${T.line2}`, background: option, cursor: 'pointer', flex: '0 0 auto' }} />)}</div>
    <label className="money-topo-width" style={{ ...label, minWidth: 116, flex: '0 1 150px' }}>Width {width}<input aria-label="Topo stroke width" type="range" min="2" max="12" value={width} onChange={event => onWidthChange(Number(event.target.value))} style={{ width: '100%' }} /></label>
    <div className="money-topo-actions" style={{ display: 'flex', gap: 8, marginLeft: 'auto' }}>
      <button type="button" onClick={onUndo} disabled={undoDisabled} style={smallBtn}><RotateCcw size={14} />Undo</button>
      <button type="button" onClick={onClear} disabled={!canWrite || !hasTopo} style={smallBtn}><Trash2 size={14} />Clear</button>
      <button type="button" onClick={onSave} disabled={!canSave} style={{ ...smallBtn, background: canSave ? T.accent : T.inset, color: canSave ? T.onAccent : T.faint }}><Save size={14} />Save</button>
    </div>
  </div>;
}

export function TopoOverlaySvg({ overlays, activeProblemId }: { overlays: Array<{ problem: MoneyCragNode; overlay: TopoOverlay }>; activeProblemId?: string | null }) {
  if (overlays.length === 0) return null;
  return <svg data-testid="photo-topo-overlay" viewBox="0 0 1000 1000" preserveAspectRatio="none" style={{ position: 'absolute', inset: 0, width: '100%', height: '100%', pointerEvents: 'none' }}>
    {overlays.map(({ problem, overlay }) => <g key={overlay.id} opacity={!activeProblemId || activeProblemId === problem.feature.id ? 1 : 0.25}><TopoOverlayMarks overlay={overlay} /><text x={(overlay.paths[0]?.points[0]?.[0] ?? overlay.starts?.[0]?.point[0] ?? 0.05) * 1000} y={(overlay.paths[0]?.points[0]?.[1] ?? overlay.starts?.[0]?.point[1] ?? 0.08) * 1000 - 12} fill={overlay.color} fontSize="36" fontWeight="800">{overlay.label ?? problem.feature.title}</text></g>)}
  </svg>;
}

function TopoOverlayMarks({ overlay }: { overlay: TopoOverlay }) {
  return <>
    {overlay.paths.map(path => <path key={path.id} d={overlayPathD(path.points)} fill="none" stroke={overlay.color} strokeWidth={overlay.width * 2.2} strokeLinecap="round" strokeLinejoin="round" />)}
    {(overlay.starts ?? []).map(marker => <StartMarker key={marker.id} marker={marker} color={overlay.color} />)}
  </>;
}

function StartMarker({ marker, color }: { marker: TopoStartMarker; color: string }) {
  const [x, y] = [marker.point[0] * 1000, marker.point[1] * 1000];
  const label = displayStartMarkerLabel(marker);
  const showLabel = label !== 'X';
  return <g data-testid="topo-start-marker">
    {startMarkerSegments(marker.point, 34).map(([x1, y1, x2, y2], index) => <line key={index} x1={x1} y1={y1} x2={x2} y2={y2} stroke={color} strokeWidth="8" strokeLinecap="round" />)}
    {showLabel && <text x={x + 28} y={y - 24} fill={color} fontSize="44" fontWeight="900">{label}</text>}
  </g>;
}

const label: React.CSSProperties = { display: 'grid', gap: 5, color: T.mut, fontFamily: T.mono, fontSize: 10.5, textTransform: 'uppercase', letterSpacing: 0.6 };
const field: React.CSSProperties = { width: '100%', background: T.surf2, border: `1px solid ${T.line}`, borderRadius: 8, padding: '8px 9px', color: T.ink, fontFamily: T.font, fontSize: 12.5, outline: 'none', minHeight: 38, textTransform: 'none', letterSpacing: 0 };
const smallBtn: React.CSSProperties = { border: `1px solid ${T.line2}`, background: T.inset, color: T.mut, borderRadius: 9, cursor: 'pointer', fontSize: 11.5, padding: '8px 10px', minHeight: 38, fontWeight: 900, display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: 5, whiteSpace: 'nowrap' };
const toolbarStyle: React.CSSProperties = { display: 'flex', alignItems: 'end', gap: 10, flexWrap: 'wrap', padding: 10, border: `1px solid rgba(238,225,211,0.18)`, borderRadius: 14, background: 'rgba(27,20,17,0.88)', boxShadow: T.shadowSm };
function toolButton(active: boolean): React.CSSProperties { return { ...smallBtn, minWidth: 70, background: active ? 'rgba(249,115,22,0.22)' : T.inset, color: active ? T.ink : T.mut, borderColor: active ? T.accent : T.line2 }; }

const topoEditorCss = `
@media (max-width: 760px) {
  .money-topo-editor { align-items: stretch !important; padding-bottom: 0 !important; }
  .money-topo-stage { width: 100% !important; height: auto !important; min-height: 42dvh !important; max-height: none !important; overflow: visible !important; display: flex !important; align-items: center !important; justify-content: center !important; }
  .money-topo-stage img { max-height: none !important; object-fit: contain !important; }
  .money-topo-toolbar { max-height: none !important; overflow-y: visible !important; border-radius: 14px !important; padding: 10px !important; align-items: stretch !important; gap: 8px !important; }
  .money-topo-toolbar-title { flex: 1 0 100% !important; }
  .money-topo-problem { flex: 1 0 100% !important; min-width: 0 !important; }
  .money-topo-tools { flex: 1 0 auto !important; }
  .money-topo-tools button, .money-topo-start-types button { min-width: 76px !important; min-height: 44px !important; }
  .money-topo-colors { flex: 1 1 auto !important; justify-content: flex-end !important; }
  .money-topo-colors button { width: 34px !important; height: 34px !important; }
  .money-topo-width { flex: 1 0 100% !important; min-width: 0 !important; }
  .money-topo-actions { width: 100% !important; margin-left: 0 !important; display: grid !important; grid-template-columns: 1fr 1fr 1.2fr !important; }
  .money-topo-actions button { min-height: 44px !important; padding: 8px 6px !important; }
}
`;
