import type { MoneyPosition } from '../../../types/money';

export const LINE_MAP_BOUNDS = [-121.66, 47.63, -121.36, 47.815] as const;
export const LINE_MAP_CATEGORIES = ['road', 'creek', 'reservoir', 'contour', 'index-contour'] as const;
export const LINE_MAP_LABEL_KINDS = ['road', 'creek', 'reservoir', 'context'] as const;
export const LINE_MAP_IMPORTANCE = ['major', 'medium', 'minor'] as const;
export const LINE_MAP_LABEL_PRIORITIES = ['high', 'medium', 'low'] as const;

export type LineMapCategory = typeof LINE_MAP_CATEGORIES[number];
export type LineMapLabelKind = typeof LINE_MAP_LABEL_KINDS[number];
export type LineMapImportance = typeof LINE_MAP_IMPORTANCE[number];
export type LineMapLabelPriority = typeof LINE_MAP_LABEL_PRIORITIES[number];
export type LineMapPath = {
  id: string;
  category: LineMapCategory;
  name: string;
  path: MoneyPosition[];
  labelPosition?: MoneyPosition;
  labelMinZoom?: number;
  elevationM?: number;
  intervalM?: number;
  sourceKind?: string;
  surface?: string;
  seasonal?: boolean;
  importance?: LineMapImportance;
  visible?: boolean;
  source?: string;
  accuracy?: string;
};
export type LineMapLabel = {
  id: string;
  kind: LineMapLabelKind;
  name: string;
  position: MoneyPosition;
  minZoom: number;
  labelType?: string;
  priority?: LineMapLabelPriority;
  source?: string;
  accuracy?: string;
};

type RawFeatureCollection = {
  type?: string;
  features?: RawFeature[];
};

type RawFeature = {
  id?: string | number;
  type?: string;
  properties?: Record<string, unknown>;
  geometry?: {
    type?: string;
    coordinates?: unknown;
  } | null;
};

export type MoneyCreekLineMapData = {
  paths: LineMapPath[];
  byCategory: Record<LineMapCategory, LineMapPath[]>;
  labels: LineMapLabel[];
};

type RawLineMapModule = { default: string };

const categorySet = new Set<string>(LINE_MAP_CATEGORIES);
let lineMapLoadPromise: Promise<MoneyCreekLineMapData> | null = null;

export function loadMoneyCreekLineMap(): Promise<MoneyCreekLineMapData> {
  lineMapLoadPromise ??= import('./fixtures/money-creek-line-map.geojson?raw')
    .then((module: RawLineMapModule) => {
      const fixture = JSON.parse(module.default) as RawFeatureCollection;
      const paths = normalizeLineMap(fixture);
      const byCategory = groupLineMapByCategory(paths);
      const labels = normalizeLineMapLabels(fixture, paths);
      return { paths, byCategory, labels };
    });
  return lineMapLoadPromise;
}

export function __resetMoneyCreekLineMapCacheForTests(): void {
  lineMapLoadPromise = null;
}

export function normalizeLineMap(collection: RawFeatureCollection): LineMapPath[] {
  if (collection.type !== 'FeatureCollection' || !Array.isArray(collection.features)) return [];

  return collection.features.flatMap((feature, featureIndex) => {
    const category = normalizeCategory(feature.properties?.category);
    if (!category || feature.type !== 'Feature' || !feature.geometry) return [];

    const lines = extractLineStrings(feature.geometry.type, feature.geometry.coordinates);
    return lines.flatMap((path, lineIndex) => {
      const validPath = normalizePath(path);
      if (!validPath) return [];
      return [{
        id: `${String(feature.id ?? `feature-${featureIndex}`)}:${lineIndex}`,
        category,
        name: typeof feature.properties?.name === 'string' ? feature.properties.name : category,
        path: validPath,
        labelPosition: normalizePosition(feature.properties?.label_position),
        labelMinZoom: typeof feature.properties?.label_min_zoom === 'number' ? feature.properties.label_min_zoom : undefined,
        elevationM: typeof feature.properties?.elevation_m === 'number' ? feature.properties.elevation_m : undefined,
        intervalM: typeof feature.properties?.interval_m === 'number' ? feature.properties.interval_m : undefined,
        sourceKind: typeof feature.properties?.source_kind === 'string' ? feature.properties.source_kind : undefined,
        surface: typeof feature.properties?.surface === 'string' ? feature.properties.surface : undefined,
        seasonal: typeof feature.properties?.seasonal === 'boolean' ? feature.properties.seasonal : undefined,
        importance: normalizeImportance(feature.properties?.importance),
        visible: typeof feature.properties?.visible === 'boolean' ? feature.properties.visible : undefined,
        source: typeof feature.properties?.source === 'string' ? feature.properties.source : undefined,
        accuracy: typeof feature.properties?.accuracy === 'string' ? feature.properties.accuracy : undefined,
      }];
    });
  });
}

export function groupLineMapByCategory(paths: LineMapPath[]): Record<LineMapCategory, LineMapPath[]> {
  return LINE_MAP_CATEGORIES.reduce((grouped, category) => {
    grouped[category] = paths.filter(path => path.category === category);
    return grouped;
  }, {} as Record<LineMapCategory, LineMapPath[]>);
}

export function normalizeLineMapLabels(collection: RawFeatureCollection, paths: LineMapPath[] = normalizeLineMap(collection)): LineMapLabel[] {
  if (collection.type !== 'FeatureCollection' || !Array.isArray(collection.features)) return [];

  const lineLabels = paths.flatMap(path => path.labelPosition ? [{
    id: `${path.id}:label`,
    kind: labelKindFromCategory(path.category),
    name: path.name,
    position: path.labelPosition,
    minZoom: path.labelMinZoom ?? defaultLabelMinZoom(path.category),
    source: path.source,
    accuracy: path.accuracy,
  }] : []);

  const pointLabels = collection.features.flatMap((feature, featureIndex) => {
    if (feature.type !== 'Feature' || feature.geometry?.type !== 'Point' || feature.properties?.category !== 'label') return [];
    const position = normalizePosition(feature.geometry.coordinates);
    const name = typeof feature.properties?.name === 'string' ? feature.properties.name : null;
    if (!position || !name) return [];
    return [{
      id: `${String(feature.id ?? `label-${featureIndex}`)}:0`,
      kind: normalizeLabelKind(feature.properties?.label_kind) ?? 'context',
      name,
      position,
      minZoom: typeof feature.properties?.label_min_zoom === 'number' ? feature.properties.label_min_zoom : 14,
      labelType: typeof feature.properties?.label_type === 'string' ? feature.properties.label_type : undefined,
      priority: normalizeLabelPriority(feature.properties?.priority),
      source: typeof feature.properties?.source === 'string' ? feature.properties.source : undefined,
      accuracy: typeof feature.properties?.accuracy === 'string' ? feature.properties.accuracy : undefined,
    }];
  });

  return [...lineLabels, ...pointLabels];
}

export function isBoundedLonLat(position: MoneyPosition, bounds: readonly [number, number, number, number] = LINE_MAP_BOUNDS): boolean {
  const [lon, lat] = position;
  return Number.isFinite(lon) && Number.isFinite(lat) && lon >= bounds[0] && lon <= bounds[2] && lat >= bounds[1] && lat <= bounds[3];
}

function normalizeCategory(value: unknown): LineMapCategory | null {
  return typeof value === 'string' && categorySet.has(value) ? value as LineMapCategory : null;
}

function extractLineStrings(type: string | undefined, coordinates: unknown): unknown[][] {
  if (type === 'LineString' && Array.isArray(coordinates)) return [coordinates];
  if (type === 'MultiLineString' && Array.isArray(coordinates)) return coordinates.filter(Array.isArray);
  return [];
}

function normalizePath(path: unknown[]): MoneyPosition[] | null {
  const positions = path.flatMap(point => {
    const position = normalizePosition(point);
    return position ? [position] : [];
  });
  return positions.length >= 2 && positions.length === path.length ? positions : null;
}

function normalizePosition(value: unknown): MoneyPosition | undefined {
  if (!Array.isArray(value) || typeof value[0] !== 'number' || typeof value[1] !== 'number') return undefined;
  const position: MoneyPosition = [value[0], value[1]];
  return isBoundedLonLat(position) ? position : undefined;
}

function labelKindFromCategory(category: LineMapCategory): LineMapLabelKind {
  if (category === 'road' || category === 'creek' || category === 'reservoir') return category;
  return 'context';
}

function normalizeLabelKind(value: unknown): LineMapLabelKind | null {
  return typeof value === 'string' && (LINE_MAP_LABEL_KINDS as readonly string[]).includes(value) ? value as LineMapLabelKind : null;
}

function normalizeImportance(value: unknown): LineMapImportance | undefined {
  return typeof value === 'string' && (LINE_MAP_IMPORTANCE as readonly string[]).includes(value) ? value as LineMapImportance : undefined;
}

function normalizeLabelPriority(value: unknown): LineMapLabelPriority | undefined {
  return typeof value === 'string' && (LINE_MAP_LABEL_PRIORITIES as readonly string[]).includes(value) ? value as LineMapLabelPriority : undefined;
}

function defaultLabelMinZoom(category: LineMapCategory): number {
  if (category === 'road' || category === 'reservoir') return 13.4;
  if (category === 'creek') return 12.8;
  return 15;
}
