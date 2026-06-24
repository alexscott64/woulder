import { useMemo, useRef, useState } from 'react';
import { Check, Eraser, Pencil, RotateCcw, Save, SquareDashedMousePointer, Trash2 } from 'lucide-react';
import { MoneyPosition } from '../../types/money';

interface SketchStroke {
  id: string;
  points: Array<[number, number]>;
  color: string;
  width: number;
}

interface SketchbookViewProps {
  canWrite: boolean;
  center: MoneyPosition;
  onSaveDrawing: (points: MoneyPosition[], title: string, properties?: Record<string, unknown>) => void;
  saving?: boolean;
}

const paperToGeoScale = 0.000018;

function strokePath(points: Array<[number, number]>): string {
  return points.map((point, index) => `${index === 0 ? 'M' : 'L'} ${point[0].toFixed(1)} ${point[1].toFixed(1)}`).join(' ');
}

function simplifyStroke(points: Array<[number, number]>, every = 3): Array<[number, number]> {
  if (points.length <= 2) return points;
  return points.filter((_, index) => index === 0 || index === points.length - 1 || index % every === 0);
}

export function SketchbookView({ canWrite, center, onSaveDrawing, saving = false }: SketchbookViewProps) {
  const paperRef = useRef<HTMLDivElement | null>(null);
  const [strokes, setStrokes] = useState<SketchStroke[]>([]);
  const [activeStroke, setActiveStroke] = useState<SketchStroke | null>(null);
  const [color, setColor] = useState('#0f766e');
  const [width, setWidth] = useState(4);

  const allPoints = useMemo(() => strokes.flatMap(stroke => simplifyStroke(stroke.points)), [strokes]);
  const canSave = canWrite && allPoints.length >= 2 && !saving;

  const getPaperPoint = (event: React.PointerEvent<HTMLDivElement>): [number, number] | null => {
    const rect = paperRef.current?.getBoundingClientRect();
    if (!rect) return null;
    return [
      Math.min(1000, Math.max(0, ((event.clientX - rect.left) / rect.width) * 1000)),
      Math.min(700, Math.max(0, ((event.clientY - rect.top) / rect.height) * 700)),
    ];
  };

  const handlePointerDown = (event: React.PointerEvent<HTMLDivElement>) => {
    if (!canWrite) return;
    const point = getPaperPoint(event);
    if (!point) return;
    event.currentTarget.setPointerCapture(event.pointerId);
    setActiveStroke({ id: crypto.randomUUID(), points: [point], color, width });
  };

  const handlePointerMove = (event: React.PointerEvent<HTMLDivElement>) => {
    if (!activeStroke) return;
    const point = getPaperPoint(event);
    if (!point) return;
    setActiveStroke(stroke => stroke ? { ...stroke, points: [...stroke.points, point] } : stroke);
  };

  const finishStroke = () => {
    setActiveStroke(stroke => {
      if (stroke && stroke.points.length > 1) setStrokes(existing => [...existing, stroke]);
      return null;
    });
  };

  const handleSave = () => {
    if (!canSave) return;
    const title = window.prompt('Name this blank field sheet');
    if (!title?.trim()) return;
    const geoPoints = allPoints.map(([x, y]) => [
      center[0] + (x - 500) * paperToGeoScale,
      center[1] - (y - 350) * paperToGeoScale,
    ] as MoneyPosition);
    onSaveDrawing(geoPoints, title.trim(), {
      source: 'sketchbook-canvas',
      sketch: { width: 1000, height: 700, stroke_count: strokes.length, color },
    });
    setStrokes([]);
  };

  return (
    <section className="grid min-h-0 gap-4 lg:grid-cols-[18rem_minmax(0,1fr)]">
      <aside className="rounded-[2rem] border-2 border-stone-950 bg-[#f7ecd2] p-4 text-stone-950 shadow-[6px_6px_0_rgba(28,25,23,0.25)]">
        <div className="flex items-center gap-3">
          <div className="flex h-12 w-12 rotate-[-5deg] items-center justify-center rounded-2xl border-2 border-stone-950 bg-violet-200 shadow-[3px_3px_0_rgba(28,25,23,0.22)]">
            <Pencil className="h-6 w-6" />
          </div>
          <div>
            <p className="text-xs font-black uppercase tracking-[0.25em] text-violet-900">Blank sheet</p>
            <h2 className="text-xl font-black">Sketchbook</h2>
          </div>
        </div>
        <p className="mt-4 text-sm font-semibold leading-6 text-stone-700">Draw topo ideas, cleanup plans, staging zones, or not-quite-explainable stump sightings. Saving creates a drawing record tied back to Money Creek.</p>

        <div className="mt-5 space-y-4">
          <label className="block text-sm font-black text-stone-900">
            Ink color
            <div className="mt-2 grid grid-cols-5 gap-2">
              {['#0f766e', '#0369a1', '#b45309', '#be123c', '#1f2937'].map(option => (
                <button key={option} type="button" onClick={() => setColor(option)} className={`h-9 rounded-xl border-2 shadow-[2px_2px_0_rgba(28,25,23,0.18)] ${color === option ? 'border-stone-950 ring-2 ring-cyan-700' : 'border-stone-300'}`} style={{ backgroundColor: option }} aria-label={`Use ink color ${option}`} />
              ))}
            </div>
          </label>

          <label className="block text-sm font-black text-stone-900">
            Line weight: {width}px
            <input type="range" min="2" max="10" value={width} onChange={event => setWidth(Number(event.target.value))} className="mt-2 w-full accent-cyan-800" />
          </label>

          <div className="grid grid-cols-2 gap-2">
            <button type="button" onClick={() => setStrokes(strokes.slice(0, -1))} disabled={!canWrite || strokes.length === 0} className="inline-flex items-center justify-center gap-2 rounded-2xl border-2 border-stone-950 bg-white px-3 py-2 text-sm font-black text-stone-950 shadow-[2px_2px_0_rgba(28,25,23,0.18)] disabled:opacity-40"><RotateCcw className="h-4 w-4" /> Undo</button>
            <button type="button" onClick={() => setStrokes([])} disabled={!canWrite || strokes.length === 0} className="inline-flex items-center justify-center gap-2 rounded-2xl border-2 border-stone-950 bg-white px-3 py-2 text-sm font-black text-stone-950 shadow-[2px_2px_0_rgba(28,25,23,0.18)] disabled:opacity-40"><Trash2 className="h-4 w-4" /> Clear</button>
          </div>

          <button type="button" onClick={handleSave} disabled={!canSave} className="flex w-full items-center justify-center gap-2 rounded-2xl border-2 border-stone-950 bg-emerald-800 px-4 py-3 text-sm font-black text-white shadow-[3px_3px_0_rgba(28,25,23,0.25)] disabled:bg-stone-400">
            {saving ? <Eraser className="h-4 w-4 animate-pulse" /> : <Save className="h-4 w-4" />}
            Save as drawing record
          </button>
          {!canWrite && <p className="rounded-2xl border border-stone-300 bg-white/70 p-3 text-xs font-bold text-stone-700">Viewer access is read-only.</p>}
        </div>
      </aside>

      <div className="min-h-[58vh] rounded-[2rem] border-2 border-stone-950 bg-stone-950 p-3 shadow-[8px_8px_0_rgba(28,25,23,0.28)]">
        <div ref={paperRef} onPointerDown={handlePointerDown} onPointerMove={handlePointerMove} onPointerUp={finishStroke} onPointerCancel={finishStroke} className="relative h-full min-h-[58vh] touch-none overflow-hidden rounded-[1.5rem] bg-[#fffaf0] text-stone-950">
          <div className="absolute inset-0 opacity-70" style={{ backgroundImage: 'linear-gradient(#d8c7a355 1px, transparent 1px), linear-gradient(90deg, #d8c7a344 1px, transparent 1px)', backgroundSize: '32px 32px' }} />
          <div className="absolute inset-0 opacity-35" style={{ backgroundImage: 'radial-gradient(ellipse at 22% 30%, transparent 30%, #0e749022 31%, transparent 32%), radial-gradient(ellipse at 78% 65%, transparent 28%, #65a30d24 29%, transparent 30%)' }} />
          <div className="absolute left-5 top-5 rotate-[-2deg] rounded-2xl border-2 border-stone-950 bg-amber-100 px-4 py-2 text-xs font-black uppercase tracking-[0.25em] text-stone-900 shadow-md">Money Creek field sheet</div>
          {strokes.length === 0 && !activeStroke && (
            <div className="absolute inset-0 flex items-center justify-center p-6 text-center">
              <div className="rounded-[2rem] border-2 border-dashed border-stone-500 bg-white/85 p-6 shadow-sm">
                <SquareDashedMousePointer className="mx-auto h-10 w-10 text-cyan-800" />
                <p className="mt-3 text-lg font-black text-stone-950">Draw on the blank field sheet</p>
                <p className="mt-1 text-sm font-semibold text-stone-600">Mouse, pen, or finger strokes become a saveable crag-development drawing.</p>
              </div>
            </div>
          )}
          <svg viewBox="0 0 1000 700" className="absolute inset-0 h-full w-full">
            {[...strokes, ...(activeStroke ? [activeStroke] : [])].map(stroke => <path key={stroke.id} d={strokePath(stroke.points)} fill="none" stroke={stroke.color} strokeLinecap="round" strokeLinejoin="round" strokeWidth={stroke.width} />)}
          </svg>
          <div className="absolute bottom-4 right-4 flex items-center gap-2 rounded-full border-2 border-stone-950 bg-white px-4 py-2 text-xs font-black text-stone-900 shadow-md"><Check className="h-4 w-4 text-emerald-800" /> {strokes.length} strokes · {allPoints.length} points</div>
        </div>
      </div>
    </section>
  );
}
