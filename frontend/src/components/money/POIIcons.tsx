import {
  Binoculars,
  Brush,
  Camera,
  Car,
  Droplets,
  Flag,
  Hammer,
  Home,
  LucideIcon,
  MapPin,
  Mountain,
  Pickaxe,
  ShieldAlert,
  Tent,
  Trees,
} from 'lucide-react';

export type MoneyPOICategory =
  | 'boulder'
  | 'parking'
  | 'camp'
  | 'water'
  | 'viewpoint'
  | 'hazard'
  | 'landing'
  | 'project'
  | 'cleanup'
  | 'trailhead'
  | 'shelter'
  | 'photo'
  | 'forest'
  | 'general';

export interface POIIconOption {
  id: MoneyPOICategory;
  label: string;
  description: string;
  icon: LucideIcon;
  tone: string;
}

export const POI_ICON_OPTIONS: POIIconOption[] = [
  { id: 'boulder', label: 'Boulder', description: 'Bloc, wall, or developed stone', icon: Mountain, tone: 'bg-stone-200 text-stone-950 border-stone-400' },
  { id: 'project', label: 'Project', description: 'Line to clean, equip, or return to', icon: Pickaxe, tone: 'bg-amber-200 text-amber-950 border-amber-400' },
  { id: 'landing', label: 'Landing', description: 'Pad zone, staging, or landing work', icon: Flag, tone: 'bg-emerald-200 text-emerald-950 border-emerald-400' },
  { id: 'cleanup', label: 'Cleanup', description: 'Brush, trash, roots, or drainage task', icon: Brush, tone: 'bg-lime-200 text-lime-950 border-lime-400' },
  { id: 'hazard', label: 'Hazard', description: 'Loose rock, bad landing, access issue', icon: ShieldAlert, tone: 'bg-red-200 text-red-950 border-red-400' },
  { id: 'parking', label: 'Parking', description: 'Pullout, gate, or shuttle note', icon: Car, tone: 'bg-sky-200 text-sky-950 border-sky-400' },
  { id: 'trailhead', label: 'Trailhead', description: 'Start of approach or spur', icon: Hammer, tone: 'bg-orange-200 text-orange-950 border-orange-400' },
  { id: 'water', label: 'Water', description: 'Creek crossing, seep, or drainage', icon: Droplets, tone: 'bg-cyan-200 text-cyan-950 border-cyan-400' },
  { id: 'camp', label: 'Camp', description: 'Bivy, hangout, or base area', icon: Tent, tone: 'bg-violet-200 text-violet-950 border-violet-400' },
  { id: 'viewpoint', label: 'Lookout', description: 'Scouting view or landmark', icon: Binoculars, tone: 'bg-indigo-200 text-indigo-950 border-indigo-400' },
  { id: 'photo', label: 'Photo', description: 'Useful camera angle or topo shot', icon: Camera, tone: 'bg-pink-200 text-pink-950 border-pink-400' },
  { id: 'forest', label: 'Forest', description: 'Tree, moss, shade, or access note', icon: Trees, tone: 'bg-green-200 text-green-950 border-green-400' },
  { id: 'shelter', label: 'Shelter', description: 'Cabin, bridge, culvert, or structure', icon: Home, tone: 'bg-slate-200 text-slate-950 border-slate-400' },
  { id: 'general', label: 'Pin', description: 'General point of interest', icon: MapPin, tone: 'bg-yellow-200 text-yellow-950 border-yellow-400' },
];

export function normalizePOICategory(value: unknown): MoneyPOICategory {
  if (typeof value === 'string' && POI_ICON_OPTIONS.some(option => option.id === value)) {
    return value as MoneyPOICategory;
  }
  return 'general';
}

export function getPOIIconOption(value: unknown): POIIconOption {
  return POI_ICON_OPTIONS.find(option => option.id === normalizePOICategory(value)) ?? POI_ICON_OPTIONS[POI_ICON_OPTIONS.length - 1];
}
