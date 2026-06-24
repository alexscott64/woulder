import { Camera, Inbox, Layers, Map as MapIcon, MapPin, PencilRuler, Route, LucideIcon } from 'lucide-react';

export type BinderSection = 'inbox' | 'map' | 'trails' | 'topos' | 'pois' | 'photos' | 'sketches';

export interface BinderSectionConfig {
  id: BinderSection;
  label: string;
  kicker: string;
  icon: LucideIcon;
}

export const BINDER_SECTIONS: BinderSectionConfig[] = [
  { id: 'map', label: 'Map', kicker: 'All records', icon: MapIcon },
  { id: 'inbox', label: 'Inbox', kicker: 'Draft items', icon: Inbox },
  { id: 'trails', label: 'Trails', kicker: 'Approaches', icon: Route },
  { id: 'topos', label: 'Topos', kicker: 'Sketches', icon: Layers },
  { id: 'pois', label: 'Map pins', kicker: 'Notes', icon: MapPin },
  { id: 'photos', label: 'Photos', kicker: 'References', icon: Camera },
  { id: 'sketches', label: 'Sketches', kicker: 'Drawings', icon: PencilRuler },
];
