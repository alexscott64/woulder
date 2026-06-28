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
  { id: 'boulder', label: 'Boulder / wall', description: 'Bloc, wall, or developed stone', icon: Mountain, tone: 'bg-stone-100 text-slate-900 border-stone-300' },
  { id: 'project', label: 'Project line', description: 'Line to clean, equip, or revisit', icon: Pickaxe, tone: 'bg-amber-50 text-amber-950 border-amber-200' },
  { id: 'landing', label: 'Landing', description: 'Pad zone, staging, or landing work', icon: Flag, tone: 'bg-moss-50 text-emerald-950 border-emerald-200' },
  { id: 'cleanup', label: 'Cleanup', description: 'Brush, trash, roots, or drainage task', icon: Brush, tone: 'bg-emerald-50 text-emerald-950 border-emerald-200' },
  { id: 'hazard', label: 'Hazard', description: 'Loose rock, bad landing, or access issue', icon: ShieldAlert, tone: 'bg-red-50 text-red-900 border-red-200' },
  { id: 'parking', label: 'Parking', description: 'Pullout, gate, or shuttle note', icon: Car, tone: 'bg-sky-50 text-sky-950 border-sky-200' },
  { id: 'trailhead', label: 'Trailhead', description: 'Start of approach or spur', icon: Hammer, tone: 'bg-orange-50 text-orange-950 border-orange-200' },
  { id: 'water', label: 'Water', description: 'Creek crossing, seep, or drainage', icon: Droplets, tone: 'bg-cyan-50 text-cyan-950 border-cyan-200' },
  { id: 'camp', label: 'Camp / base', description: 'Bivy, hangout, or base area', icon: Tent, tone: 'bg-violet-50 text-violet-950 border-violet-200' },
  { id: 'viewpoint', label: 'Viewpoint', description: 'Scouting view or landmark', icon: Binoculars, tone: 'bg-indigo-50 text-indigo-950 border-indigo-200' },
  { id: 'photo', label: 'Photo angle', description: 'Useful camera angle or topo shot', icon: Camera, tone: 'bg-rose-50 text-rose-950 border-rose-200' },
  { id: 'forest', label: 'Forest / shade', description: 'Tree, moss, shade, or access note', icon: Trees, tone: 'bg-green-50 text-green-950 border-green-200' },
  { id: 'shelter', label: 'Structure', description: 'Cabin, bridge, culvert, or structure', icon: Home, tone: 'bg-slate-50 text-slate-950 border-slate-200' },
  { id: 'general', label: 'General note', description: 'General point of interest', icon: MapPin, tone: 'bg-stone-50 text-slate-900 border-stone-300' },
];

export function normalizePOICategory(value: unknown): MoneyPOICategory {
  if (typeof value === 'string' && POI_ICON_OPTIONS.some(option => option.id === value)) return value as MoneyPOICategory;
  return 'general';
}

export function getPOIIconOption(value: unknown): POIIconOption {
  return POI_ICON_OPTIONS.find(option => option.id === normalizePOICategory(value)) ?? POI_ICON_OPTIONS[POI_ICON_OPTIONS.length - 1];
}
