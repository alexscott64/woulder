import { MapPin, MousePointer2, Pencil, Route, Square, Undo2, X } from 'lucide-react';
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

const tools: Array<{ type: MoneyFeatureType; icon: typeof Route; short: string; hint: string }> = [
  { type: 'poi', icon: MapPin, short: 'Map pin', hint: 'Boulder, hazard, water, or parking note' },
  { type: 'trail', icon: Route, short: 'Trail', hint: 'Approach or connector' },
  { type: 'topo', icon: Square, short: 'Topo', hint: 'Boulder or sector outline' },
  { type: 'drawing', icon: Pencil, short: 'Sketch', hint: 'Free map sketch' },
];

export function DrawingToolbar({ drawingType, draftPoints, canWrite, onStart, onUndo, onCancel, onFinish }: DrawingToolbarProps) {
  if (!canWrite) return null;

  const canFinish = drawingType ? draftPoints.length >= minimumPointCount(drawingType) : false;

  return (
    <div className="absolute bottom-3 left-3 right-3 z-40 rounded-2xl border border-[#2B403A] bg-[#111D1B]/95 p-2 text-[#F2F0E7] shadow-xl backdrop-blur md:bottom-4 md:left-4 md:right-auto md:w-auto">
      {!drawingType ? (
        <div className="flex items-center gap-2 overflow-x-auto">
          {tools.map(tool => {
            const Icon = tool.icon;
            return (
              <button
                key={tool.type}
                type="button"
                onClick={() => onStart(tool.type)}
                className="group flex shrink-0 items-center gap-2 rounded-xl border border-[#2B403A] bg-[#172522] px-3 py-2 text-left text-sm font-semibold text-[#F2F0E7] transition hover:border-[#7EA16B] hover:bg-[#1B2925]"
                title={`Create ${featureTypeLabel(tool.type)} · ${tool.hint}`}
              >
                <Icon className="h-4 w-4 text-[#7EA16B]" />
                <span>{tool.short}</span>
              </button>
            );
          })}
        </div>
      ) : (
        <div className="flex flex-wrap items-center gap-2">
          <div className="flex flex-1 items-center gap-2 rounded-xl bg-[#172522] px-3 py-2 sm:min-w-56">
            <MousePointer2 className="h-4 w-4 text-[#7EA16B]" />
            <div>
              <p className="text-sm font-semibold">{featureTypeLabel(drawingType)} capture</p>
              <p className="text-xs text-[#AAB8AD]">Tap map to add points: {draftPoints.length}</p>
            </div>
          </div>
          <button type="button" onClick={onUndo} disabled={draftPoints.length === 0} className="rounded-xl border border-[#2B403A] bg-[#0B1714]/80 px-3 py-3 text-sm font-semibold disabled:opacity-40" title="Undo point"><Undo2 className="h-4 w-4" /></button>
          <button type="button" onClick={onCancel} className="rounded-xl border border-[#2B403A] bg-[#0B1714]/80 px-3 py-3 text-sm font-semibold" title="Cancel drawing"><X className="h-4 w-4" /></button>
          <button type="button" onClick={onFinish} disabled={!canFinish} className="rounded-xl bg-[#7EA16B] px-4 py-3 text-sm font-semibold text-[#07110F] disabled:bg-[#2B403A] disabled:text-[#74847B]">Save</button>
        </div>
      )}
    </div>
  );
}
