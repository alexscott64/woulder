import { Camera, ChevronDown, FileText, MapPin, PencilRuler, Plus, Route, Square } from 'lucide-react';
import { MoneyFeatureType } from '../../types/money';

interface CaptureMenuProps {
  canWrite: boolean;
  isOpen: boolean;
  onToggle: () => void;
  onStartDrawing: (type: MoneyFeatureType) => void;
  onOpenSketch: () => void;
  onOpenNote: () => void;
}

const captureItems: Array<{ label: string; hint: string; icon: typeof Plus; action: 'note' | 'poi' | 'photo' | 'trail' | 'topo' | 'sketch' }> = [
  { label: 'Quick note', hint: 'Add to selected item or inbox', icon: FileText, action: 'note' },
  { label: 'Point', hint: 'Mark a boulder, hazard, photo angle, or access detail', icon: MapPin, action: 'poi' },
  { label: 'Photo', hint: 'Open the selected item photo area', icon: Camera, action: 'photo' },
  { label: 'Trail', hint: 'Draw an approach or connector', icon: Route, action: 'trail' },
  { label: 'Area', hint: 'Outline a bloc or topo zone', icon: Square, action: 'topo' },
  { label: 'Sketch', hint: 'Open the drawing sheet', icon: PencilRuler, action: 'sketch' },
];

export function CaptureMenu({ canWrite, isOpen, onToggle, onStartDrawing, onOpenSketch, onOpenNote }: CaptureMenuProps) {
  if (!canWrite) return null;

  const runAction = (action: (typeof captureItems)[number]['action']) => {
    if (action === 'note' || action === 'photo') onOpenNote();
    if (action === 'poi') onStartDrawing('poi');
    if (action === 'trail') onStartDrawing('trail');
    if (action === 'topo') onStartDrawing('topo');
    if (action === 'sketch') onOpenSketch();
    onToggle();
  };

  return (
    <div className="fixed bottom-4 right-4 z-50 sm:bottom-6 sm:right-6">
      {isOpen && (
        <div className="mb-3 w-[min(22rem,calc(100vw-2rem))] rounded-[1.5rem] border border-stone-200 bg-white p-2 text-slate-950 shadow-2xl">
          <div className="rounded-2xl bg-stone-50 p-3">
            <p className="text-sm font-semibold text-slate-950">Capture</p>
            <p className="mt-1 text-xs leading-5 text-slate-500">Add a note, draw on the map, or attach a photo without changing tools.</p>
          </div>
          <div className="mt-2 grid gap-2 sm:grid-cols-2">
            {captureItems.map(item => {
              const Icon = item.icon;
              return (
                <button key={item.action} type="button" onClick={() => runAction(item.action)} className="group rounded-2xl border border-stone-200 bg-white p-3 text-left transition hover:border-teal-700 hover:bg-teal-50">
                  <span className="flex items-center gap-2 text-sm font-semibold"><Icon className="h-4 w-4 text-teal-900" />{item.label}</span>
                  <span className="mt-1 block text-xs leading-4 text-slate-500">{item.hint}</span>
                </button>
              );
            })}
          </div>
        </div>
      )}
      <button type="button" onClick={onToggle} className="group flex items-center gap-3 rounded-full bg-teal-900 px-5 py-4 font-semibold text-white shadow-xl transition hover:-translate-y-0.5 hover:bg-teal-800 focus:outline-none focus:ring-4 focus:ring-teal-700/20">
        <span className="flex h-8 w-8 items-center justify-center rounded-full bg-white/15"><Plus className="h-5 w-5" /></span>
        Capture
        <ChevronDown className={`h-4 w-4 transition ${isOpen ? 'rotate-180' : ''}`} />
      </button>
    </div>
  );
}
