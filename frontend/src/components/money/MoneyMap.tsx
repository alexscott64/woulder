import { useEffect, useMemo, useState } from 'react';
import DeckGL from '@deck.gl/react';
import { Map } from 'react-map-gl/maplibre';
import { PathLayer, PolygonLayer, ScatterplotLayer } from '@deck.gl/layers';
import { MapPin } from 'lucide-react';
import { MoneyFeature, MoneyFeatureType, MoneyPosition, MoneyProject } from '../../types/money';
import { buildGeoJSON, featureTypeLabel, geometryFromGeoJSON, getFeatureCoordinates } from './geometry';
import 'maplibre-gl/dist/maplibre-gl.css';

interface MoneyMapProps {
  project: MoneyProject;
  features: MoneyFeature[];
  selectedFeatureId: string | null;
  drawingType: MoneyFeatureType | null;
  draftPoints: MoneyPosition[];
  onSelectFeature: (feature: MoneyFeature) => void;
  onAddDraftPoint: (point: MoneyPosition) => void;
}

const VIEW_STORAGE_KEY = 'money-creek-map-view';

const featureColor: Record<MoneyFeatureType, [number, number, number, number]> = {
  trail: [16, 185, 129, 220],
  topo: [56, 189, 248, 150],
  poi: [251, 191, 36, 230],
  drawing: [244, 114, 182, 220],
};

function loadView(project: MoneyProject) {
  try {
    const stored = localStorage.getItem(VIEW_STORAGE_KEY);
    if (stored) return JSON.parse(stored);
  } catch {
    // ignore corrupt viewport state
  }
  return {
    longitude: project.center_lon,
    latitude: project.center_lat,
    zoom: project.default_zoom,
    pitch: 0,
    bearing: 0,
  };
}

function pointData(features: MoneyFeature[]) {
  return features
    .filter(feature => geometryFromGeoJSON(feature.geojson)?.type === 'Point')
    .map(feature => ({ feature, position: getFeatureCoordinates(feature)[0] }));
}

function pathData(features: MoneyFeature[]) {
  return features
    .filter(feature => geometryFromGeoJSON(feature.geojson)?.type === 'LineString')
    .map(feature => ({ feature, path: getFeatureCoordinates(feature) }));
}

function polygonData(features: MoneyFeature[]) {
  return features
    .map(feature => {
      const geometry = geometryFromGeoJSON(feature.geojson);
      if (geometry?.type !== 'Polygon') return null;
      return { feature, polygon: geometry.coordinates[0] ?? [] };
    })
    .filter((item): item is { feature: MoneyFeature; polygon: MoneyPosition[] } => Boolean(item));
}

export function MoneyMap({ project, features, selectedFeatureId, drawingType, draftPoints, onSelectFeature, onAddDraftPoint }: MoneyMapProps) {
  const [viewState, setViewState] = useState(loadView(project));
  const [hoveredFeature, setHoveredFeature] = useState<MoneyFeature | null>(null);

  useEffect(() => {
    localStorage.setItem(VIEW_STORAGE_KEY, JSON.stringify(viewState));
  }, [viewState]);

  const layers = useMemo(() => {
    const selectedLine = (feature: MoneyFeature) => feature.id === selectedFeatureId ? [255, 255, 255, 255] as [number, number, number, number] : featureColor[feature.feature_type];
    const selectedWidth = (feature: MoneyFeature) => feature.id === selectedFeatureId ? 7 : 4;

    const draftGeoJSON = drawingType && draftPoints.length > 0 ? buildGeoJSON(drawingType, draftPoints) : null;
    const draftGeometry = draftGeoJSON ? geometryFromGeoJSON(draftGeoJSON) : null;

    return [
      new PolygonLayer({
        id: 'money-polygons',
        data: polygonData(features),
        pickable: true,
        stroked: true,
        filled: true,
        lineWidthMinPixels: 2,
        getPolygon: (d: { polygon: MoneyPosition[] }) => d.polygon,
        getFillColor: (d: { feature: MoneyFeature }) => d.feature.id === selectedFeatureId ? [14, 165, 233, 110] : featureColor[d.feature.feature_type],
        getLineColor: (d: { feature: MoneyFeature }) => selectedLine(d.feature),
        getLineWidth: (d: { feature: MoneyFeature }) => selectedWidth(d.feature),
        onClick: (info: { object?: { feature: MoneyFeature } }) => {
          if (info.object) onSelectFeature(info.object.feature);
          return true;
        },
        onHover: (info: { object?: { feature: MoneyFeature } }) => setHoveredFeature(info.object?.feature ?? null),
      }),
      new PathLayer({
        id: 'money-paths',
        data: pathData(features),
        pickable: true,
        rounded: true,
        widthMinPixels: 3,
        getPath: (d: { path: MoneyPosition[] }) => d.path,
        getColor: (d: { feature: MoneyFeature }) => selectedLine(d.feature),
        getWidth: (d: { feature: MoneyFeature }) => selectedWidth(d.feature),
        onClick: (info: { object?: { feature: MoneyFeature } }) => {
          if (info.object) onSelectFeature(info.object.feature);
          return true;
        },
        onHover: (info: { object?: { feature: MoneyFeature } }) => setHoveredFeature(info.object?.feature ?? null),
      }),
      new ScatterplotLayer({
        id: 'money-points',
        data: pointData(features),
        pickable: true,
        stroked: true,
        filled: true,
        radiusMinPixels: 8,
        radiusMaxPixels: 18,
        lineWidthMinPixels: 2,
        getPosition: (d: { position?: MoneyPosition }) => d.position ?? [project.center_lon, project.center_lat],
        getFillColor: (d: { feature: MoneyFeature }) => d.feature.id === selectedFeatureId ? [255, 255, 255, 255] : featureColor[d.feature.feature_type],
        getLineColor: (d: { feature: MoneyFeature }) => featureColor[d.feature.feature_type],
        getRadius: 12,
        onClick: (info: { object?: { feature: MoneyFeature } }) => {
          if (info.object) onSelectFeature(info.object.feature);
          return true;
        },
        onHover: (info: { object?: { feature: MoneyFeature } }) => setHoveredFeature(info.object?.feature ?? null),
      }),
      draftGeometry?.type === 'LineString' && new PathLayer({
        id: 'money-draft-path',
        data: [{ path: draftGeometry.coordinates }],
        getPath: (d: { path: MoneyPosition[] }) => d.path,
        getColor: [255, 255, 255, 240],
        getWidth: 5,
        widthMinPixels: 3,
      }),
      draftGeometry?.type === 'Polygon' && new PolygonLayer({
        id: 'money-draft-polygon',
        data: [{ polygon: draftGeometry.coordinates[0] ?? [] }],
        getPolygon: (d: { polygon: MoneyPosition[] }) => d.polygon,
        getFillColor: [255, 255, 255, 60],
        getLineColor: [255, 255, 255, 240],
        getLineWidth: 4,
        lineWidthMinPixels: 2,
      }),
      draftGeometry?.type === 'Point' && new ScatterplotLayer({
        id: 'money-draft-point',
        data: [{ position: draftGeometry.coordinates }],
        getPosition: (d: { position: MoneyPosition }) => d.position,
        getFillColor: [255, 255, 255, 240],
        getLineColor: [16, 185, 129, 255],
        getRadius: 14,
        radiusMinPixels: 10,
        radiusMaxPixels: 18,
        stroked: true,
      }),
    ].filter(Boolean);
  }, [draftPoints, drawingType, features, onAddDraftPoint, onSelectFeature, project.center_lat, project.center_lon, selectedFeatureId]);

  return (
    <div className="absolute inset-0 bg-slate-900">
      <DeckGL
        controller
        layers={layers}
        viewState={viewState}
        onViewStateChange={({ viewState: nextViewState }) => setViewState(nextViewState)}
        onClick={(info: { coordinate?: number[] }) => {
          if (drawingType && info.coordinate && info.coordinate.length >= 2) {
            onAddDraftPoint([info.coordinate[0], info.coordinate[1]]);
          }
        }}
      >
        <Map mapStyle="https://basemaps.cartocdn.com/gl/positron-gl-style/style.json" reuseMaps />
      </DeckGL>

      {hoveredFeature && (
        <div className="pointer-events-none absolute left-4 top-4 z-10 hidden rounded-2xl bg-slate-950/90 px-4 py-3 text-white shadow-xl backdrop-blur md:block">
          <p className="text-xs font-bold uppercase tracking-widest text-emerald-200">{featureTypeLabel(hoveredFeature.feature_type)}</p>
          <p className="font-bold">{hoveredFeature.title}</p>
        </div>
      )}

      {drawingType && (
        <div className="absolute bottom-28 left-3 right-3 z-10 rounded-2xl border border-white/20 bg-slate-950/90 p-3 text-sm text-white shadow-xl backdrop-blur md:bottom-6 md:left-4 md:right-auto md:w-80">
          <div className="flex items-start gap-2">
            <MapPin className="mt-0.5 h-4 w-4 text-emerald-200" />
            <p>Tap the map to draw a {featureTypeLabel(drawingType).toLowerCase()}. Save from the drawing toolbar when enough points are added.</p>
          </div>
        </div>
      )}
    </div>
  );
}
