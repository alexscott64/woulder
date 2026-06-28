import { useRef, useState } from 'react';
import { RotateCcw, Save, Trash2, X } from 'lucide-react';
import { T } from './theme';
import { NormalizedPoint, SketchBlockData, overlayPathD, pointFromClient, sketchBlock } from './topoOverlay';
import { MoneyNoteBlock } from '../../../types/money';

type Stroke = { id: string; points: NormalizedPoint[]; color: string; width: number };

interface SketchPadDialogProps {
  onClose: () => void;
  onSave: (block: MoneyNoteBlock) => void;
}

const colors = ['#1F2937', '#0F766E', '#2563EB', '#B45309', '#BE123C'];

export function SketchPadDialog({ onClose, onSave }: SketchPadDialogProps) {
  const padRef = useRef<HTMLDivElement | null>(null);
  const [strokes, setStrokes] = useState<Stroke[]>([]);
  const [active, setActive] = useState<Stroke | null>(null);
  const [color, setColor] = useState(colors[0]);
  const [width, setWidth] = useState(4);

  const eventPoint = (event: React.PointerEvent<HTMLDivElement>) => {
    const rect = padRef.current?.getBoundingClientRect();
    return rect ? pointFromClient(event.clientX, event.clientY, rect) : null;
  };
  const down = (event: React.PointerEvent<HTMLDivElement>) => {
    const point = eventPoint(event);
    if (!point) return;
    event.currentTarget.setPointerCapture(event.pointerId);
    setActive({ id: crypto.randomUUID(), points: [point], color, width });
  };
  const move = (event: React.PointerEvent<HTMLDivElement>) => {
    if (!active) return;
    const point = eventPoint(event);
    if (!point) return;
    setActive(stroke => stroke ? { ...stroke, points: [...stroke.points, point] } : stroke);
  };
  const finish = () => setActive(stroke => { if (stroke && stroke.points.length > 1) setStrokes(current => [...current, stroke]); return null; });
  const save = () => {
    const all = [...strokes, ...(active ? [active] : [])].filter(stroke => stroke.points.length > 1);
    if (all.length === 0) return;
    const data: SketchBlockData = { width: 1000, height: 700, background: 'field-sheet', strokes: all };
    onSave(sketchBlock(`Sketch ${new Date().toLocaleString()}`, data));
  };

  return <div role="presentation" onClick={onClose} style={{ position: 'fixed', inset: 0, zIndex: 85, background: 'rgba(8,5,4,0.72)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 18 }}>
    <div role="dialog" aria-modal="true" aria-label="Sketch pad" onClick={event => event.stopPropagation()} style={{ width: 'min(920px,100%)', maxHeight: '92vh', background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 16, boxShadow: T.shadow, overflow: 'hidden', display: 'grid', gridTemplateRows: 'auto minmax(0,1fr) auto' }}>
      <header style={{ display: 'flex', alignItems: 'center', gap: 10, padding: 14, borderBottom: `1px solid ${T.line}` }}><b style={{ color: T.ink }}>Sketch pad</b><span style={{ color: T.mut, fontSize: 12 }}>Draw with mouse, pen, or touch. Saved as vector note metadata.</span><button aria-label="Close sketch pad" onClick={onClose} style={{ marginLeft: 'auto', border: `1px solid ${T.line2}`, background: T.inset, color: T.ink, borderRadius: 8, width: 34, height: 34 }}><X size={16} /></button></header>
      <div style={{ minHeight: 360, padding: 14, background: T.inset }}>
        <div ref={padRef} data-testid="sketch-drawing-pad" onPointerDown={down} onPointerMove={move} onPointerUp={finish} onPointerCancel={finish} style={{ position: 'relative', height: 'min(58vh,560px)', minHeight: 320, borderRadius: 14, overflow: 'hidden', background: '#fffaf0', border: `1px solid ${T.line2}`, touchAction: 'none', cursor: 'crosshair' }}>
          <div style={{ position: 'absolute', inset: 0, backgroundImage: 'linear-gradient(#d8c7a355 1px, transparent 1px), linear-gradient(90deg, #d8c7a344 1px, transparent 1px)', backgroundSize: '32px 32px' }} />
          <svg viewBox="0 0 1000 1000" preserveAspectRatio="none" style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}>{[...strokes, ...(active ? [active] : [])].map(stroke => <path key={stroke.id} d={overlayPathD(stroke.points)} fill="none" stroke={stroke.color} strokeWidth={stroke.width * 2} strokeLinecap="round" strokeLinejoin="round" />)}</svg>
          {strokes.length === 0 && !active && <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#6B5E50', fontWeight: 800 }}>Draw a sketch</div>}
        </div>
      </div>
      <footer style={{ display: 'flex', alignItems: 'center', gap: 10, padding: 14, borderTop: `1px solid ${T.line}` }}>
        {colors.map(option => <button key={option} aria-label={`Sketch color ${option}`} onClick={() => setColor(option)} style={{ width: 30, height: 30, borderRadius: 8, border: color === option ? `2px solid ${T.ink}` : `1px solid ${T.line2}`, background: option }} />)}
        <label style={{ color: T.mut, fontSize: 12 }}>Width {width}<input aria-label="Sketch stroke width" type="range" min="2" max="12" value={width} onChange={event => setWidth(Number(event.target.value))} /></label>
        <button onClick={() => setStrokes(items => items.slice(0, -1))} disabled={strokes.length === 0} style={btn}><RotateCcw size={14} />Undo</button>
        <button onClick={() => setStrokes([])} disabled={strokes.length === 0} style={btn}><Trash2 size={14} />Clear</button>
        <button onClick={save} disabled={strokes.length === 0 && !active} style={{ ...btn, marginLeft: 'auto', background: strokes.length || active ? T.accent : T.line2, color: strokes.length || active ? T.onAccent : T.faint }}><Save size={14} />Save sketch</button>
      </footer>
    </div>
  </div>;
}

const btn: React.CSSProperties = { border: `1px solid ${T.line2}`, background: T.inset, color: T.ink, borderRadius: 8, padding: '8px 10px', display: 'inline-flex', alignItems: 'center', gap: 6, fontWeight: 800, cursor: 'pointer' };
