import { MousePointer2, MapPin, Pencil, Route, Square, Undo2, X } from 'lucide-react';
import { MoneyFeatureType, MoneyPosition } from '../../types/money';
import { featureTypeLabel, minimumPointCount } from './geometry';

interface DrawingToolbarProps {
  drawingType: MoneyFeatureType | null;
  draftPoints: MoneyPosition[];
  canWrite: boolean;
  onStart: (type: MoneyFeatureType) => void;
  onUndo: () => void;
  onCancel: () => void;
  onFinish: () => void;
}

const tools: Array<{ type: MoneyFeatureType; icon: typeof Route }> = [
  { type: 'trail', icon: Route },
  { type: 'topo', icon: Square },
  { type: 'poi', icon: MapPin },
  { type: 'drawing', icon: Pencil },
];

export function DrawingToolbar({ drawingType, draftPoints, canWrite, onStart, onUndo, onCancel, onFinish }: DrawingToolbarProps) {
  if (!canWrite) {
    return null;
  }

  const canFinish = drawingType ? draftPoints.length >= minimumPointCount(drawingType) : false;

  return (
    <div className="absolute left-3 top-24 z-20 rounded-3xl border border-white/20 bg-slate-950/90 p-2 text-white shadow-xl backdrop-blur md:left-4 md:top-28">
      {!drawingType ? (
        <div className="grid gap-2">
          <div className="px-2 py-1 text-[0.65rem] font-bold uppercase tracking-widest text-slate-400">Draw</div>
          {tools.map(tool => {
            const Icon = tool.icon;
            return (
              <button
                key={tool.type}
                onClick={() => onStart(tool.type)}
                className="flex h-11 w-11 items-center justify-center rounded-2xl bg-white/10 text-slate-100 transition hover:bg-emerald-300 hover:text-slate-950"
                title={`Create ${featureTypeLabel(tool.type)}`}
              >
                <Icon className="h-5 w-5" />
              </button>
            );
          })}
        </div>
      ) : (
        <div className="w-52 space-y-2 p-1">
          <div className="flex items-center gap-2 rounded-2xl bg-emerald-300/15 px-3 py-2">
            <MousePointer2 className="h-4 w-4 text-emerald-200" />
            <div>
              <p className="text-sm font-bold">{featureTypeLabel(drawingType)}</p>
              <p className="text-xs text-slate-300">Tap map to add points ({draftPoints.length})</p>
            </div>
          </div>
          <div className="grid grid-cols-3 gap-2">
            <button onClick={onUndo} disabled={draftPoints.length === 0} className="rounded-2xl bg-white/10 px-3 py-2 text-sm font-semibold disabled:opacity-40">
              <Undo2 className="mx-auto h-4 w-4" />
            </button>
            <button onClick={onCancel} className="rounded-2xl bg-white/10 px-3 py-2 text-sm font-semibold">
              <X className="mx-auto h-4 w-4" />
            </button>
            <button onClick={onFinish} disabled={!canFinish} className="rounded-2xl bg-emerald-300 px-3 py-2 text-sm font-bold text-slate-950 disabled:opacity-40">
              Save
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
