import { MoneyCragNode, MoneyFeature, MoneyNoteBlock, MoneyUpload } from '../../../types/money';
import { flattenProblems } from './model';

export type NormalizedPoint = [number, number];

export interface TopoOverlayPath {
  id: string;
  points: NormalizedPoint[];
}

export type TopoStartMarkerType = 'generic' | 'left' | 'right';

export interface TopoStartMarker {
  id: string;
  point: NormalizedPoint;
  label: string;
  type?: TopoStartMarkerType;
}

export interface TopoOverlay {
  id: string;
  upload_id: string;
  photo_id?: string;
  problem_id: string;
  label?: string;
  color: string;
  width: number;
  order: number;
  paths: TopoOverlayPath[];
  starts?: TopoStartMarker[];
  updated_at?: string;
}

export interface SketchBlockData {
  width: number;
  height: number;
  background?: string;
  strokes: Array<{ id: string; points: NormalizedPoint[]; color: string; width: number }>;
}

export function clamp01(value: number): number {
  if (!Number.isFinite(value)) return 0;
  return Math.min(1, Math.max(0, value));
}

export function pointFromClient(clientX: number, clientY: number, rect: Pick<DOMRect, 'left' | 'top' | 'width' | 'height'>): NormalizedPoint {
  return [clamp01((clientX - rect.left) / rect.width), clamp01((clientY - rect.top) / rect.height)];
}

export function overlayPathD(points: NormalizedPoint[], width = 1000, height = 1000): string {
  return points.map(([x, y], index) => `${index === 0 ? 'M' : 'L'} ${(clamp01(x) * width).toFixed(1)} ${(clamp01(y) * height).toFixed(1)}`).join(' ');
}

export function scalePoint(point: NormalizedPoint, width: number, height: number): [number, number] {
  return [clamp01(point[0]) * width, clamp01(point[1]) * height];
}

export function startMarkerSegments(point: NormalizedPoint, size = 34, width = 1000, height = 1000): Array<[number, number, number, number]> {
  const [x, y] = scalePoint(point, width, height);
  const half = size / 2;
  return [[x - half, y - half, x + half, y + half], [x + half, y - half, x - half, y + half]];
}

export function labelForStartMarkerType(type: TopoStartMarkerType): string {
  if (type === 'left') return 'L';
  if (type === 'right') return 'R';
  return 'X';
}

export function displayStartMarkerLabel(marker: TopoStartMarker): string {
  if (marker.type) return labelForStartMarkerType(marker.type);
  const normalized = marker.label.trim().toLowerCase();
  if (normalized === 'start l' || normalized === 'left' || normalized === 'l') return 'L';
  if (normalized === 'start r' || normalized === 'right' || normalized === 'r') return 'R';
  if (normalized === 'x' || normalized === 'generic') return 'X';
  return marker.label || 'X';
}

export function topoOverlaysForFeature(feature: MoneyFeature): TopoOverlay[] {
  const value = feature.properties?.topo_overlays;
  if (!Array.isArray(value)) return [];
  return value.filter(isTopoOverlay);
}

export function overlaysForUpload(root: MoneyCragNode | null, uploadId: string): Array<{ problem: MoneyCragNode; overlay: TopoOverlay }> {
  return flattenProblems(root).flatMap(problem => topoOverlaysForFeature(problem.feature).filter(overlay => overlay.upload_id === uploadId || overlay.photo_id === uploadId).map(overlay => ({ problem, overlay })));
}

export function firstTopoForProblem(problem: MoneyCragNode, uploads: MoneyUpload[]): { upload: MoneyUpload; overlay: TopoOverlay } | null {
  const overlay = topoOverlaysForFeature(problem.feature).find(item => uploads.some(upload => upload.id === item.upload_id || upload.id === item.photo_id));
  if (!overlay) return null;
  const upload = uploads.find(item => item.id === overlay.upload_id || item.id === overlay.photo_id);
  return upload ? { upload, overlay } : null;
}

export function upsertTopoOverlay(feature: MoneyFeature, overlay: TopoOverlay): Record<string, unknown> {
  const existing = topoOverlaysForFeature(feature).filter(item => item.id !== overlay.id && !(item.upload_id === overlay.upload_id && item.problem_id === overlay.problem_id));
  return { ...feature.properties, topo_overlays: [...existing, overlay] };
}

export function sketchBlock(name: string, data: SketchBlockData): MoneyNoteBlock {
  return { kind: 'sketch', name, metadata: { sketchpad: data, vector_schema: 'money-sketch-v1' } };
}

function isTopoOverlay(value: unknown): value is TopoOverlay {
  if (!value || typeof value !== 'object') return false;
  const item = value as Partial<TopoOverlay>;
  return typeof item.id === 'string'
    && typeof item.upload_id === 'string'
    && typeof item.problem_id === 'string'
    && typeof item.color === 'string'
    && typeof item.width === 'number'
    && Array.isArray(item.paths)
    && (item.starts === undefined || Array.isArray(item.starts));
}
