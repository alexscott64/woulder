import { MoneyFeature, MoneyFeatureType, MoneyGeoJSON, MoneyGeometry, MoneyPosition } from '../../types/money';

export function geometryFromGeoJSON(geojson: MoneyGeoJSON): MoneyGeometry | null {
  if (geojson.type === 'Feature') {
    return geojson.geometry;
  }
  if (geojson.type === 'FeatureCollection') {
    return geojson.features[0]?.geometry ?? null;
  }
  return geojson;
}

export function getFeatureCoordinates(feature: MoneyFeature): MoneyPosition[] {
  const geometry = geometryFromGeoJSON(feature.geojson);
  if (!geometry) return [];
  if (geometry.type === 'Point') return [geometry.coordinates];
  if (geometry.type === 'LineString') return geometry.coordinates;
  return geometry.coordinates[0] ?? [];
}

export function buildGeoJSON(type: MoneyFeatureType, points: MoneyPosition[]): MoneyGeoJSON {
  if (type === 'poi') {
    return { type: 'Point', coordinates: points[0] ?? [-121.55, 47.72] };
  }
  if (type === 'topo') {
    const ring = [...points];
    const first = ring[0];
    const last = ring[ring.length - 1];
    if (first && last && (first[0] !== last[0] || first[1] !== last[1])) {
      ring.push(first);
    }
    return { type: 'Polygon', coordinates: [ring] };
  }
  return { type: 'LineString', coordinates: points };
}

export function minimumPointCount(type: MoneyFeatureType): number {
  if (type === 'poi') return 1;
  if (type === 'topo') return 3;
  return 2;
}

export function featureTypeLabel(type: MoneyFeatureType): string {
  const labels: Record<MoneyFeatureType, string> = {
    trail: 'Trail',
    topo: 'Topo',
    poi: 'Map pin',
    drawing: 'Sketch',
  };
  return labels[type];
}

export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
