import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { Layers, Plus, RotateCcw, X } from 'lucide-react';
import { MoneyCragNode, MoneyPosition } from '../../../types/money';
import { bbox, centroid, cragBoulders, cragChildren, geometryPoints, H, poly, smoothOpen, W } from './model';
import { DEV, T } from './theme';
import { Terrain } from './Terrain';

type LayerState = { base: string; contours: boolean; trails: boolean; areas: Record<string, boolean>; dev: Record<string, boolean> };

interface Props {
  area: MoneyCragNode;
  trails: MoneyCragNode[];
  selectedBoulderId: string | null;
  selectedTrailId: string | null;
  mode: 'view' | 'create-area' | 'create-boulder';
  layers: LayerState;
  mobile: boolean;
  onEnter: (id: string) => void;
  onSelectBoulder: (id: string | null) => void;
  onSelectTrail: (id: string | null) => void;
  onCreateDone: (points: MoneyPosition[]) => void;
  onCreateCancel: () => void;
  setLayers: (layers: LayerState) => void;
}

export function CragMap({ area, trails, selectedBoulderId, selectedTrailId, mode, layers, mobile, onEnter, onSelectBoulder, onSelectTrail, onCreateDone, onCreateCancel, setLayers }: Props) {
  const wrapRef = useRef<HTMLDivElement | null>(null);
  const [tf, setTf] = useState({ x: 0, y: 0, k: 1 });
  const tfRef = useRef(tf);
  const [draft, setDraft] = useState<MoneyPosition[]>([]);
  const [layersOpen, setLayersOpen] = useState(!mobile);
  const creating = mode !== 'view';
  const setTransform = (next: typeof tf) => { tfRef.current = next; setTf(next); };

  const fit = (box: [number, number, number, number]) => {
    const el = wrapRef.current;
    if (!el) return;
    const r = el.getBoundingClientRect();
    const bw = Math.max(40, (box[2] - box[0]) * 1.32), bh = Math.max(40, (box[3] - box[1]) * 1.32);
    const k = Math.max(0.3, Math.min(4, Math.min(r.width / bw, r.height / bh)));
    const cx = (box[0] + box[2]) / 2, cy = (box[1] + box[3]) / 2;
    setTransform({ k, x: r.width / 2 - cx * k, y: r.height / 2 - cy * k });
  };
  useLayoutEffect(() => fit(bbox(area)), [area.feature.id]);
  useEffect(() => { if (creating) setDraft([]); }, [creating]);

  const toWorld = (cx: number, cy: number): MoneyPosition => {
    const r = wrapRef.current!.getBoundingClientRect();
    const t = tfRef.current;
    return [Math.round((cx - r.left - t.x) / t.k), Math.round((cy - r.top - t.y) / t.k)];
  };
  const zoomAt = (cx: number, cy: number, factor: number) => {
    const r = wrapRef.current!.getBoundingClientRect();
    const t = tfRef.current;
    const nk = Math.min(4, Math.max(0.3, t.k * factor));
    const px = cx - r.left, py = cy - r.top, k = nk / t.k;
    setTransform({ k: nk, x: px - (px - t.x) * k, y: py - (py - t.y) * k });
  };

  const pointer = useRef<{ x: number; y: number; moved: number } | null>(null);
  const onPointerDown = (e: React.PointerEvent) => { pointer.current = { x: e.clientX, y: e.clientY, moved: 0 }; e.currentTarget.setPointerCapture(e.pointerId); };
  const onPointerMove = (e: React.PointerEvent) => {
    const p = pointer.current;
    if (!p) return;
    const dx = e.clientX - p.x, dy = e.clientY - p.y;
    p.moved += Math.abs(dx) + Math.abs(dy);
    p.x = e.clientX; p.y = e.clientY;
    if (!creating) setTransform({ ...tfRef.current, x: tfRef.current.x + dx, y: tfRef.current.y + dy });
  };
  const onPointerUp = (e: React.PointerEvent) => {
    const p = pointer.current;
    pointer.current = null;
    if (!p) return;
    if (creating && p.moved <= 6) setDraft(d => [...d, toWorld(e.clientX, e.clientY)]);
    if (!creating && p.moved <= 6) { onSelectBoulder(null); onSelectTrail(null); }
  };

  const finish = () => { if (draft.length >= 3) onCreateDone(draft); };
  const inv = `scale(${1 / tf.k})`;

  return <div ref={wrapRef} onWheel={e => { e.preventDefault(); zoomAt(e.clientX, e.clientY, Math.exp(-e.deltaY * 0.0016)); }} onPointerDown={onPointerDown} onPointerMove={onPointerMove} onPointerUp={onPointerUp} style={{ position: 'absolute', inset: 0, overflow: 'hidden', background: T.map.bg2, touchAction: 'none', cursor: creating ? 'crosshair' : 'grab' }}>
    <div style={{ position: 'absolute', top: 0, left: 0, width: W, height: H, transformOrigin: '0 0', transform: `translate(${tf.x}px,${tf.y}px) scale(${tf.k})` }}>
      <Terrain base={layers.base} contours={layers.contours} />
      {layers.base === 'satellite' && <div style={{ position: 'absolute', inset: 0, background: 'rgba(15,20,12,0.18)' }} />}
      <svg viewBox={`0 0 ${W} ${H}`} width={W} height={H} style={{ position: 'absolute', inset: 0, overflow: 'visible', pointerEvents: 'none' }}>
        <polygon points={poly(geometryPoints(area.feature.geojson))} fill="none" stroke={T.line2} strokeWidth="1.4" strokeDasharray="6 6" vectorEffect="non-scaling-stroke" opacity="0.7" />
        {layers.trails && trails.map(tr => <g key={tr.feature.id} style={{ pointerEvents: creating ? 'none' : 'auto', cursor: 'pointer' }} onPointerDown={e => e.stopPropagation()} onClick={e => { e.stopPropagation(); onSelectTrail(tr.feature.id); }}><path d={smoothOpen(geometryPoints(tr.feature.geojson))} fill="none" stroke="transparent" strokeWidth="15" vectorEffect="non-scaling-stroke" strokeLinecap="round" /><path d={smoothOpen(geometryPoints(tr.feature.geojson))} fill="none" stroke={selectedTrailId === tr.feature.id ? T.accent : T.map.trail} strokeWidth={selectedTrailId === tr.feature.id ? 3.2 : 2} strokeDasharray="2 6" vectorEffect="non-scaling-stroke" strokeLinecap="round" /></g>)}
        {cragChildren(area).map(child => layers.areas[child.feature.id] !== false && <polygon key={child.feature.id} points={poly(geometryPoints(child.feature.geojson))} fill={T.accentSoft} stroke={T.accent} strokeWidth="1.7" vectorEffect="non-scaling-stroke" style={{ pointerEvents: creating ? 'none' : 'auto', cursor: 'pointer' }} onPointerDown={e => e.stopPropagation()} onClick={e => { e.stopPropagation(); onEnter(child.feature.id); }} />)}
        {cragBoulders(area).map(b => { const dev = String(b.feature.status); if (layers.dev[dev] === false) return null; const m = DEV.meta[(dev in DEV.meta ? dev : 'scouted') as keyof typeof DEV.meta]; return <polygon key={b.feature.id} points={poly(geometryPoints(b.feature.geojson))} fill={m.bg} stroke={m.c} strokeWidth={selectedBoulderId === b.feature.id ? 2.8 : 1.6} vectorEffect="non-scaling-stroke" style={{ pointerEvents: creating ? 'none' : 'auto', cursor: 'pointer' }} onPointerDown={e => e.stopPropagation()} onClick={e => { e.stopPropagation(); onSelectBoulder(b.feature.id); }} />; })}
        {creating && draft.length > 0 && <polygon points={poly(draft)} fill={T.accentSoft} stroke={T.accent} strokeWidth="1.8" strokeDasharray="5 4" vectorEffect="non-scaling-stroke" />}
      </svg>
      {!creating && cragChildren(area).map(child => { const c = centroid(geometryPoints(child.feature.geojson)); return <div key={child.feature.id} onPointerDown={e => e.stopPropagation()} onClick={e => { e.stopPropagation(); onEnter(child.feature.id); }} style={{ position: 'absolute', left: c[0], top: c[1], transform: `translate(-50%,-50%) ${inv}`, cursor: 'pointer', zIndex: 3, textAlign: 'center' }}><div style={{ display: 'inline-flex', flexDirection: 'column', alignItems: 'center', gap: 2, background: T.mapLabelBg, border: `1px solid ${T.line2}`, borderRadius: 8, padding: '5px 10px', whiteSpace: 'nowrap' }}><span style={{ fontSize: 13, fontWeight: 700, color: T.ink }}>{child.feature.title}</span><span style={{ fontFamily: T.mono, fontSize: 9.5, color: T.mut }}>{cragBoulders(child).length} boulders · {cragChildren(child).length} sub</span></div></div>; })}
      {!creating && cragBoulders(area).map(b => { const c = centroid(geometryPoints(b.feature.geojson)); const dev = String(b.feature.status); const m = DEV.meta[(dev in DEV.meta ? dev : 'scouted') as keyof typeof DEV.meta]; return <div key={b.feature.id} onPointerDown={e => e.stopPropagation()} onClick={e => { e.stopPropagation(); onSelectBoulder(b.feature.id); }} style={{ position: 'absolute', left: c[0], top: c[1], transform: `translate(-50%,-50%) ${inv}`, cursor: 'pointer', zIndex: 3 }}><span style={{ display: 'inline-flex', alignItems: 'center', gap: 5, fontFamily: T.mono, fontSize: 10.5, fontWeight: 700, color: m.c, background: T.mapLabelBg, padding: '2px 7px', borderRadius: 4, whiteSpace: 'nowrap' }}><span style={{ width: 6, height: 6, borderRadius: '50%', background: m.c }} />{b.feature.title}</span></div>; })}
    </div>
    {!creating && <div style={{ position: 'absolute', top: mobile ? 70 : 'auto', bottom: mobile ? 'auto' : 16, right: 14, display: 'flex', flexDirection: 'column', background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 10, overflow: 'hidden', boxShadow: T.shadow }}>{[['+', 1.3], ['−', 0.77]].map(([s, f]) => <button key={s as string} onClick={() => { const r = wrapRef.current!.getBoundingClientRect(); zoomAt(r.left + r.width / 2, r.top + r.height / 2, f as number); }} style={{ width: 40, height: 38, border: 'none', background: 'transparent', color: T.ink, cursor: 'pointer', fontSize: 18 }}>{s}</button>)}</div>}
    {!creating && (layersOpen ? <LayersPanel layers={layers} setLayers={setLayers} onClose={() => setLayersOpen(false)} area={area} /> : <button onClick={() => setLayersOpen(true)} style={{ position: 'absolute', top: 14, right: 14, display: 'flex', gap: 7, background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 10, padding: '9px 13px', color: T.ink, fontWeight: 700, boxShadow: T.shadow }}><Layers size={16} />Layers</button>)}
    {creating && <div style={{ position: 'absolute', bottom: mobile ? 22 : 20, left: '50%', transform: 'translateX(-50%)', display: 'flex', gap: 8, background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 14, padding: 8, boxShadow: T.shadow }}><button onClick={onCreateCancel} style={ctrl(false)}><X size={16} />Cancel</button><button onClick={() => setDraft(d => d.slice(0, -1))} style={ctrl(false)}><RotateCcw size={16} />Undo</button><button disabled={draft.length < 3} onClick={finish} style={ctrl(true, draft.length < 3)}><Plus size={16} />Done</button></div>}
  </div>;
}

function ctrl(accent: boolean, disabled = false): React.CSSProperties { return { display: 'flex', alignItems: 'center', gap: 7, border: accent ? 'none' : `1px solid ${T.line2}`, borderRadius: 9, padding: '11px 18px', background: disabled ? T.line : accent ? T.accent : 'transparent', color: disabled ? T.faint : accent ? T.onAccent : T.ink, fontFamily: T.font, fontWeight: 700, cursor: disabled ? 'default' : 'pointer' }; }
function LayersPanel({ layers, setLayers, onClose, area }: { layers: LayerState; setLayers: (l: LayerState) => void; onClose: () => void; area: MoneyCragNode }) { const bases = ['stylized', 'topo', 'satellite', 'slope']; return <div style={{ position: 'absolute', top: 14, right: 14, width: 224, background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 12, boxShadow: T.shadow, overflow: 'hidden' }}><div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '11px 13px', borderBottom: `1px solid ${T.line}`, color: T.ink }}><Layers size={16} /><b style={{ fontSize: 13.5 }}>Layers</b><button onClick={onClose} style={{ marginLeft: 'auto', border: 'none', background: 'transparent', color: T.mut, cursor: 'pointer' }}>×</button></div><div style={{ padding: '11px 13px', maxHeight: 380, overflowY: 'auto' }}><Label>Base map</Label>{bases.map(base => <Row key={base} onClick={() => setLayers({ ...layers, base })} on={layers.base === base} label={base} />)}<Label>Areas in view</Label>{cragChildren(area).map(c => <Row key={c.feature.id} onClick={() => setLayers({ ...layers, areas: { ...layers.areas, [c.feature.id]: layers.areas[c.feature.id] === false } })} on={layers.areas[c.feature.id] !== false} label={c.feature.title} />)}<Label>Development</Label>{DEV.order.map(k => <Row key={k} onClick={() => setLayers({ ...layers, dev: { ...layers.dev, [k]: layers.dev[k] === false } })} on={layers.dev[k] !== false} label={DEV.meta[k].short} color={DEV.meta[k].c} />)}<div style={{ borderTop: `1px solid ${T.line}`, margin: '10px 0 8px' }} />{(['contours', 'trails'] as const).map(k => <Row key={k} onClick={() => setLayers({ ...layers, [k]: !layers[k] })} on={layers[k]} label={k} />)}</div></div>; }
function Label({ children }: { children: React.ReactNode }) { return <div style={{ fontFamily: T.mono, fontSize: 10, letterSpacing: 0.6, color: T.faint, textTransform: 'uppercase', margin: '14px 0 7px' }}>{children}</div>; }
function Row({ on, label, onClick, color }: { on: boolean; label: string; onClick: () => void; color?: string }) { return <div onClick={onClick} style={{ display: 'flex', alignItems: 'center', gap: 9, padding: '5px 0', cursor: 'pointer' }}><span style={{ width: 9, height: 9, borderRadius: color ? '50%' : 2, background: color ?? T.accent, transform: color ? 'none' : 'rotate(45deg)' }} /><span style={{ flex: 1, fontSize: 12.5, color: on ? T.ink : T.mut, textTransform: label.length < 9 ? 'capitalize' : undefined }}>{label}</span><span style={{ width: 28, height: 16, borderRadius: 9, background: on ? (color ?? T.accent) : T.line2, position: 'relative' }}><span style={{ position: 'absolute', top: 2, left: on ? 14 : 2, width: 12, height: 12, borderRadius: '50%', background: '#fff' }} /></span></div>; }
