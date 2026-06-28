import { MoneyGeoJSON, MoneyPosition } from '../../../types/money';

export interface ParsedTrailImport {
  title: string;
  geojson: MoneyGeoJSON;
  pointCount: number;
  sourceFormat: 'gpx' | 'geojson';
}

export async function parseTrailFile(file: File): Promise<ParsedTrailImport> {
  const text = await file.text();
  const lowerName = file.name.toLowerCase();
  if (lowerName.endsWith('.gpx') || text.trimStart().startsWith('<')) {
    return parseGPXTrail(text, file.name);
  }
  if (lowerName.endsWith('.geojson') || lowerName.endsWith('.json') || text.trimStart().startsWith('{')) {
    return parseGeoJSONTrail(text, file.name);
  }
  throw new Error('Unsupported trail file. Upload GPX or GeoJSON.');
}

export function parseGPXTrail(text: string, filename = 'trail.gpx'): ParsedTrailImport {
  const doc = new DOMParser().parseFromString(text, 'application/xml');
  if (doc.querySelector('parsererror')) throw new Error('Invalid GPX file.');
  const name = textFor(doc.querySelector('trk > name, rte > name, metadata > name')) || titleFromFilename(filename);
  const trkSegments = Array.from(doc.querySelectorAll('trkseg'))
    .map(segment => Array.from(segment.querySelectorAll('trkpt')).map(pointFromGPXElement).filter(isPosition))
    .filter(points => points.length >= 2);
  const routePoints = Array.from(doc.querySelectorAll('rtept')).map(pointFromGPXElement).filter(isPosition);
  const coordinates = trkSegments.length > 0 ? trkSegments.flat() : routePoints;
  if (coordinates.length < 2) throw new Error('GPX trail needs at least two track or route points.');
  return { title: name, geojson: { type: 'LineString', coordinates }, pointCount: coordinates.length, sourceFormat: 'gpx' };
}

export function parseGeoJSONTrail(text: string, filename = 'trail.geojson'): ParsedTrailImport {
  let parsed: unknown;
  try {
    parsed = JSON.parse(text);
  } catch {
    throw new Error('Invalid GeoJSON file.');
  }
  const result = lineFromGeoJSON(parsed);
  if (!result) throw new Error('GeoJSON trail must contain a LineString or MultiLineString with at least two points.');
  const title = result.title || titleFromFilename(filename);
  return { title, geojson: { type: 'LineString', coordinates: result.coordinates }, pointCount: result.coordinates.length, sourceFormat: 'geojson' };
}

function lineFromGeoJSON(value: unknown): { title?: string; coordinates: MoneyPosition[] } | null {
  if (!isRecord(value)) return null;
  if (value.type === 'FeatureCollection' && Array.isArray(value.features)) {
    for (const feature of value.features) {
      const line = lineFromGeoJSON(feature);
      if (line) return line;
    }
    return null;
  }
  if (value.type === 'Feature') {
    const line = lineFromGeoJSON(value.geometry);
    if (!line) return null;
    const props = isRecord(value.properties) ? value.properties : {};
    const title = typeof props.name === 'string' ? props.name : typeof props.title === 'string' ? props.title : undefined;
    return { ...line, title };
  }
  if (value.type === 'LineString' && Array.isArray(value.coordinates)) {
    const coordinates = value.coordinates.map(positionFromGeoJSON).filter(isPosition);
    return coordinates.length >= 2 ? { coordinates } : null;
  }
  if (value.type === 'MultiLineString' && Array.isArray(value.coordinates)) {
    const coordinates = value.coordinates.flatMap(line => Array.isArray(line) ? line.map(positionFromGeoJSON).filter(isPosition) : []);
    return coordinates.length >= 2 ? { coordinates } : null;
  }
  return null;
}

function pointFromGPXElement(el: Element): MoneyPosition | null {
  const lat = Number(el.getAttribute('lat'));
  const lon = Number(el.getAttribute('lon'));
  if (!Number.isFinite(lat) || !Number.isFinite(lon) || lat < -90 || lat > 90 || lon < -180 || lon > 180) return null;
  return [lon, lat];
}

function positionFromGeoJSON(value: unknown): MoneyPosition | null {
  if (!Array.isArray(value) || value.length < 2) return null;
  const lon = Number(value[0]);
  const lat = Number(value[1]);
  if (!Number.isFinite(lat) || !Number.isFinite(lon) || lat < -90 || lat > 90 || lon < -180 || lon > 180) return null;
  return [lon, lat];
}

function isPosition(value: MoneyPosition | null): value is MoneyPosition {
  return Boolean(value);
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value && typeof value === 'object');
}

function textFor(el: Element | null): string {
  return el?.textContent?.trim() ?? '';
}

function titleFromFilename(filename: string): string {
  return filename.replace(/\.[^.]+$/, '').replace(/[-_]+/g, ' ').trim() || 'Uploaded trail';
}
