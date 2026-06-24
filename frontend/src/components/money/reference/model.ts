import { MoneyCragNode, MoneyDevStatus, MoneyFeature, MoneyGeometry, MoneyGeoJSON, MoneyNote, MoneyPosition, MoneyProblemStatus } from '../../../types/money';

export function cragChildren(node: MoneyCragNode): MoneyCragNode[] { return node.children ?? []; }
export function cragBoulders(node: MoneyCragNode): MoneyCragNode[] { return node.boulders ?? []; }
export function cragProblems(node: MoneyCragNode): MoneyCragNode[] { return node.problems ?? []; }

export const W = 1000;
export const H = 680;

export function geometryPoints(geojson: MoneyGeoJSON): MoneyPosition[] {
  const geometry: MoneyGeometry | undefined = geojson.type === 'Feature' ? geojson.geometry : geojson.type === 'FeatureCollection' ? geojson.features[0]?.geometry : geojson;
  if (!geometry) return [];
  if (geometry.type === 'Point') return [geometry.coordinates];
  if (geometry.type === 'LineString') return geometry.coordinates;
  if (geometry.type === 'Polygon') return geometry.coordinates[0] ?? [];
  return [];
}

export function polygonGeoJSON(points: MoneyPosition[]): MoneyGeoJSON {
  const closed = points.length && (points[0][0] !== points[points.length - 1][0] || points[0][1] !== points[points.length - 1][1]) ? [...points, points[0]] : points;
  return { type: 'Polygon', coordinates: [closed] };
}

export function lineGeoJSON(points: MoneyPosition[]): MoneyGeoJSON { return { type: 'LineString', coordinates: points }; }
export function pointGeoJSON(point: MoneyPosition): MoneyGeoJSON { return { type: 'Point', coordinates: point }; }

export function centroid(points: MoneyPosition[]): MoneyPosition {
  if (!points.length) return [500, 340];
  return points.reduce<MoneyPosition>((a, p) => [a[0] + p[0] / points.length, a[1] + p[1] / points.length], [0, 0]);
}

export function bbox(node: MoneyCragNode): [number, number, number, number] {
  const pts = geometryPoints(node.feature.geojson);
  if (!pts.length) return [440, 280, 560, 400];
  return pts.reduce<[number, number, number, number]>((a, p) => [Math.min(a[0], p[0]), Math.min(a[1], p[1]), Math.max(a[2], p[0]), Math.max(a[3], p[1])], [1e9, 1e9, -1e9, -1e9]);
}

export function poly(points: MoneyPosition[]) { return points.map(p => p.join(',')).join(' '); }

export function smoothOpen(pts: MoneyPosition[]) {
  if (pts.length < 2) return '';
  let d = `M ${pts[0][0]} ${pts[0][1]}`;
  for (let i = 0; i < pts.length - 1; i += 1) {
    const p0 = pts[i - 1] || pts[i], p1 = pts[i], p2 = pts[i + 1], p3 = pts[i + 2] || p2;
    d += ` C ${p1[0] + (p2[0] - p0[0]) / 6} ${p1[1] + (p2[1] - p0[1]) / 6}, ${p2[0] - (p3[0] - p1[0]) / 6} ${p2[1] - (p3[1] - p1[1]) / 6}, ${p2[0]} ${p2[1]}`;
  }
  return d;
}

export function flattenAreas(root: MoneyCragNode | null): MoneyCragNode[] {
  const out: MoneyCragNode[] = [];
  const walk = (node: MoneyCragNode) => { out.push(node); cragChildren(node).forEach(walk); };
  if (root) walk(root);
  return out;
}

export function flattenBoulders(root: MoneyCragNode | null): MoneyCragNode[] {
  const out: MoneyCragNode[] = [];
  const walk = (node: MoneyCragNode) => { out.push(...cragBoulders(node)); cragChildren(node).forEach(walk); };
  if (root) walk(root);
  return out;
}

export function flattenProblems(root: MoneyCragNode | null): Array<MoneyCragNode & { boulder: MoneyCragNode; area: MoneyCragNode }> {
  const out: Array<MoneyCragNode & { boulder: MoneyCragNode; area: MoneyCragNode }> = [];
  const walk = (area: MoneyCragNode) => {
    cragBoulders(area).forEach(boulder => cragProblems(boulder).forEach(problem => out.push({ ...problem, boulder, area })));
    cragChildren(area).forEach(walk);
  };
  if (root) walk(root);
  return out;
}

export function findNode(root: MoneyCragNode | null, id?: string | null): MoneyCragNode | null {
  if (!root || !id) return null;
  if (root.feature.id === id) return root;
  for (const boulder of cragBoulders(root)) {
    if (boulder.feature.id === id) return boulder;
    const problem = cragProblems(boulder).find(p => p.feature.id === id);
    if (problem) return problem;
  }
  for (const child of cragChildren(root)) {
    const found = findNode(child, id);
    if (found) return found;
  }
  return null;
}

export function parentArea(root: MoneyCragNode | null, id?: string | null): MoneyCragNode | null {
  if (!root || !id) return null;
  for (const boulder of cragBoulders(root)) if (boulder.feature.id === id) return root;
  for (const child of cragChildren(root)) {
    if (child.feature.id === id) return root;
    const found = parentArea(child, id);
    if (found) return found;
  }
  return null;
}

export function pathTo(root: MoneyCragNode | null, id?: string | null): MoneyCragNode[] {
  if (!root || !id) return [];
  if (root.feature.id === id) return [root];
  for (const child of cragChildren(root)) {
    const p = pathTo(child, id);
    if (p.length) return [root, ...p];
  }
  for (const boulder of cragBoulders(root)) if (boulder.feature.id === id) return [root, boulder];
  return [];
}

export function stats(node: MoneyCragNode) {
  const dev: Record<MoneyDevStatus, number> = { scouted: 0, 'needs-work': 0, cleaning: 0, established: 0 };
  let boulders = 0, problems = 0, sent = 0, projects = 0;
  const walk = (area: MoneyCragNode) => {
    cragBoulders(area).forEach(b => {
      boulders += 1;
      const status = b.feature.status as MoneyDevStatus;
      if (status in dev) dev[status] += 1;
      cragProblems(b).forEach(p => { problems += 1; if (p.feature.status === 'sent') sent += 1; if (p.feature.status === 'project') projects += 1; });
    });
    cragChildren(area).forEach(walk);
  };
  walk(node);
  return { boulders, problems, sent, projects, subareas: cragChildren(node).length, dev };
}

export function problemMeta(feature: MoneyFeature) {
  return {
    grade: String(feature.properties.grade ?? 'V?'),
    stars: Number(feature.properties.stars ?? 1),
    fa: typeof feature.properties.fa === 'string' ? feature.properties.fa : null,
    types: Array.isArray(feature.properties.types) ? feature.properties.types.map(String) : [],
    status: feature.status as MoneyProblemStatus,
  };
}

export function notesFor(notes: MoneyNote[], type: string, id: string) {
  return notes.filter(n => n.target_type === type && n.target_ref === id || n.feature_id === id);
}
