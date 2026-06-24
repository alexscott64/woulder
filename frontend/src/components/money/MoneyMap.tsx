import { useEffect, useMemo, useState } from 'react';
import DeckGL from '@deck.gl/react';
import { Map } from 'react-map-gl/maplibre';
import { PathLayer, PolygonLayer, ScatterplotLayer } from '@deck.gl/layers';
import { LocateFixed, MapPin, Minus, Plus, RotateCcw } from 'lucide-react';
import { getPOIIconOption } from './POIIcons';
import { MoneyFeature, MoneyFeatureType, MoneyPosition, MoneyProject } from '../../types/money';
import { buildGeoJSON, featureTypeLabel, geometryFromGeoJSON, getFeatureCoordinates } from './geometry';
import 'maplibre-gl/dist/maplibre-gl.css';

interface MoneyMapProps {
  project: MoneyProject;
  features: MoneyFeature[];
  selectedFeatureId: string | null;
  drawingType: MoneyFeatureType | null;
  draftPoints: MoneyPosition[];
  focusMode: boolean;
  onSelectFeature: (feature: MoneyFeature) => void;
  onAddDraftPoint: (point: MoneyPosition) => void;
}

type MoneyViewState = {
  longitude: number;
  latitude: number;
  zoom: number;
  pitch: number;
  bearing: number;
};

const MONEY_CREEK_CENTER: MoneyPosition = [-121.46678408962231, 47.70007999015064];
const MONEY_CREEK_ZOOM = 14;
const VIEW_STORAGE_KEY = 'money-creek-map-view-v2';

const featureColor: Record<MoneyFeatureType, [number, number, number, number]> = {
  trail: [38, 112, 86, 230],
  topo: [38, 88, 87, 125],
  poi: [180, 129, 42, 230],
  drawing: [82, 82, 74, 210],
};

function isFiniteNumber(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value);
}

function defaultView(): MoneyViewState {
  return { longitude: MONEY_CREEK_CENTER[0], latitude: MONEY_CREEK_CENTER[1], zoom: MONEY_CREEK_ZOOM, pitch: 0, bearing: 0 };
}

function isValidView(value: unknown): value is MoneyViewState {
  if (!value || typeof value !== 'object') return false;
  const candidate = value as Partial<MoneyViewState>;
  return isFiniteNumber(candidate.longitude)
    && isFiniteNumber(candidate.latitude)
    && isFiniteNumber(candidate.zoom)
    && isFiniteNumber(candidate.pitch)
    && isFiniteNumber(candidate.bearing)
    && candidate.longitude >= -180
    && candidate.longitude <= 180
    && candidate.latitude >= -85
    && candidate.latitude <= 85
    && candidate.zoom >= 0
    && candidate.zoom <= 22;
}

function loadView() {
  try {
    const stored = localStorage.getItem(VIEW_STORAGE_KEY);
    if (!stored) return defaultView();
    const parsed: unknown = JSON.parse(stored);
    if (isValidView(parsed)) return parsed;
  } catch {
    // ignore corrupt viewport state
  }
  return defaultView();
}

function pointData(features: MoneyFeature[]) {
  return features.filter(feature => geometryFromGeoJSON(feature.geojson)?.type === 'Point').map(feature => ({ feature, position: getFeatureCoordinates(feature)[0] }));
}

function pathData(features: MoneyFeature[]) {
  return features.filter(feature => geometryFromGeoJSON(feature.geojson)?.type === 'LineString').map(feature => ({ feature, path: getFeatureCoordinates(feature) }));
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

export function MoneyMap({ project, features, selectedFeatureId, drawingType, draftPoints, focusMode, onSelectFeature, onAddDraftPoint }: MoneyMapProps) {
  const [viewState, setViewState] = useState<MoneyViewState>(loadView);
  const [hoveredFeature, setHoveredFeature] = useState<MoneyFeature | null>(null);
  const selectedFeature = features.find(feature => feature.id === selectedFeatureId) ?? null;

  useEffect(() => {
    localStorage.setItem(VIEW_STORAGE_KEY, JSON.stringify(viewState));
  }, [viewState]);

  useEffect(() => {
    if (!selectedFeature) return;
    const selectedPosition = getFeatureCoordinates(selectedFeature)[0];
    if (!selectedPosition) return;
    const focusTimer = window.setTimeout(() => {
      setViewState(current => ({
        ...current,
        longitude: selectedPosition[0],
        latitude: selectedPosition[1],
        zoom: Math.max(current.zoom, 16),
      }));
    }, 0);
    return () => window.clearTimeout(focusTimer);
  }, [selectedFeature]);

  const focusMoneyCreek = () => setViewState(defaultView());
  const zoomBy = (delta: number) => setViewState(current => ({ ...current, zoom: Math.max(0, Math.min(22, current.zoom + delta)) }));
  const resetTilt = () => setViewState(current => ({ ...current, pitch: 0, bearing: 0 }));

  const layers = useMemo(() => {
    const selectedLine = (feature: MoneyFeature) => feature.id === selectedFeatureId ? [255, 255, 255, 255] as [number, number, number, number] : featureColor[feature.feature_type];
    const selectedWidth = (feature: MoneyFeature) => feature.id === selectedFeatureId ? 8 : 4;
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
        getFillColor: (d: { feature: MoneyFeature }) => d.feature.id === selectedFeatureId ? [38, 88, 87, 150] : featureColor[d.feature.feature_type],
        getLineColor: (d: { feature: MoneyFeature }) => selectedLine(d.feature),
        getLineWidth: (d: { feature: MoneyFeature }) => selectedWidth(d.feature),
        onClick: (info: { object?: { feature: MoneyFeature } }) => {
          if (info.object) onSelectFeature(info.object.feature);
          return true;
        },
        onHover: (info: { object?: { feature: MoneyFeature } }) => setHoveredFeature(info.object?.feature ?? null),
      }),
      new PathLayer({
        id: 'money-paths-shadow',
        data: pathData(features),
        rounded: true,
        widthMinPixels: 5,
        getPath: (d: { path: MoneyPosition[] }) => d.path,
        getColor: [15, 23, 20, 105],
        getWidth: (d: { feature: MoneyFeature }) => selectedWidth(d.feature) + 3,
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
        id: 'money-points-halo',
        data: pointData(features),
        stroked: false,
        filled: true,
        radiusMinPixels: 14,
        radiusMaxPixels: 26,
        getPosition: (d: { position?: MoneyPosition }) => d.position ?? [project.center_lon, project.center_lat],
        getFillColor: [255, 255, 255, 110],
        getRadius: 18,
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
        getLineColor: (d: { feature: MoneyFeature }) => d.feature.id === selectedFeatureId ? [15, 23, 42, 255] : [255, 255, 255, 230],
        getRadius: 12,
        onClick: (info: { object?: { feature: MoneyFeature } }) => {
          if (info.object) onSelectFeature(info.object.feature);
          return true;
        },
        onHover: (info: { object?: { feature: MoneyFeature } }) => setHoveredFeature(info.object?.feature ?? null),
      }),
      draftGeometry?.type === 'LineString' && new PathLayer({ id: 'money-draft-path', data: [{ path: draftGeometry.coordinates }], getPath: (d: { path: MoneyPosition[] }) => d.path, getColor: [255, 255, 255, 245], getWidth: 6, widthMinPixels: 3 }),
      draftGeometry?.type === 'Polygon' && new PolygonLayer({ id: 'money-draft-polygon', data: [{ polygon: draftGeometry.coordinates[0] ?? [] }], getPolygon: (d: { polygon: MoneyPosition[] }) => d.polygon, getFillColor: [255, 255, 255, 70], getLineColor: [255, 255, 255, 245], getLineWidth: 5, lineWidthMinPixels: 2 }),
      draftGeometry?.type === 'Point' && new ScatterplotLayer({ id: 'money-draft-point', data: [{ position: draftGeometry.coordinates }], getPosition: (d: { position: MoneyPosition }) => d.position, getFillColor: [255, 255, 255, 245], getLineColor: [38, 112, 86, 255], getRadius: 15, radiusMinPixels: 11, radiusMaxPixels: 20, stroked: true }),
    ].filter(Boolean);
  }, [draftPoints, drawingType, features, onSelectFeature, project.center_lat, project.center_lon, selectedFeatureId]);

  return (
    <div className="absolute inset-0 bg-[#171b18]">
      <DeckGL controller={{ dragPan: true, scrollZoom: true, doubleClickZoom: true, touchZoom: true, keyboard: true }} layers={layers} viewState={viewState} onViewStateChange={(event: { viewState: unknown }) => {
        const next = event.viewState as Partial<MoneyViewState>;
        setViewState(current => ({
          longitude: next.longitude ?? current.longitude,
          latitude: next.latitude ?? current.latitude,
          zoom: next.zoom ?? current.zoom,
          pitch: next.pitch ?? current.pitch,
          bearing: next.bearing ?? current.bearing,
        }));
      }} onClick={(info: { coordinate?: number[] }) => {
        if (drawingType && info.coordinate && info.coordinate.length >= 2) onAddDraftPoint([info.coordinate[0], info.coordinate[1]]);
      }}>
        <Map mapStyle={focusMode ? 'https://basemaps.cartocdn.com/gl/positron-gl-style/style.json' : 'https://basemaps.cartocdn.com/gl/voyager-gl-style/style.json'} reuseMaps />
      </DeckGL>

      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_50%_50%,transparent_0,transparent_58%,rgba(15,23,20,0.18)_100%)]" />

      <div data-testid="money-map-controls" className="absolute left-3 top-3 z-40 flex flex-col gap-2">
        <button type="button" onClick={focusMoneyCreek} className="inline-flex items-center gap-2 rounded-full border border-stone-200 bg-white/95 px-3 py-2 text-xs font-black text-teal-950 shadow-lg backdrop-blur hover:bg-white" title="Focus Money Creek">
          <LocateFixed className="h-4 w-4" /> Money Creek
        </button>
        <div className="flex overflow-hidden rounded-full border border-stone-200 bg-white/95 text-teal-950 shadow-lg backdrop-blur">
          <button type="button" onClick={() => zoomBy(1)} className="border-r border-stone-200 px-3 py-2 hover:bg-stone-100" title="Zoom in"><Plus className="h-4 w-4" /></button>
          <button type="button" onClick={() => zoomBy(-1)} className="border-r border-stone-200 px-3 py-2 hover:bg-stone-100" title="Zoom out"><Minus className="h-4 w-4" /></button>
          <button type="button" onClick={resetTilt} className="px-3 py-2 hover:bg-stone-100" title="Reset tilt"><RotateCcw className="h-4 w-4" /></button>
        </div>
      </div>

      <div className="pointer-events-none absolute bottom-3 right-3 z-30 hidden max-w-64 rounded-2xl border border-stone-200 bg-white/95 px-4 py-3 text-xs font-semibold text-slate-700 shadow-lg backdrop-blur md:block">
        <p className="font-black uppercase tracking-[0.16em] text-teal-900">Map tools</p>
        <p className="mt-1">Drag to pan east/west or north/south. Scroll or pinch to zoom. Use drawing controls for trails, topos, sketches, and pins.</p>
      </div>

      {selectedFeature && (
        <div className="pointer-events-none absolute right-3 top-3 z-30 hidden max-w-72 rounded-2xl border border-stone-200 bg-white/95 px-4 py-3 text-slate-950 shadow-lg backdrop-blur lg:block">
          <p className="text-xs font-black uppercase tracking-[0.18em] text-teal-900">Selected {featureTypeLabel(selectedFeature.feature_type)}</p>
          <p className="mt-1 font-black leading-tight">{selectedFeature.title}</p>
          <p className="mt-1 text-xs font-semibold text-slate-500">{selectedFeature.status}</p>
        </div>
      )}

      {hoveredFeature && !selectedFeature && (
        <div className="pointer-events-none absolute right-3 top-3 z-30 hidden rounded-2xl border border-stone-200 bg-white/95 px-4 py-3 text-slate-950 shadow-lg backdrop-blur md:block">
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-teal-900">{featureTypeLabel(hoveredFeature.feature_type)}</p>
          <p className="font-semibold">{hoveredFeature.title}</p>
          {hoveredFeature.feature_type === 'poi' && <p className="mt-1 text-xs text-slate-500">{getPOIIconOption(hoveredFeature.properties?.poi_category).label}</p>}
        </div>
      )}

      {drawingType && (
        <div className="absolute bottom-24 left-3 right-3 z-30 rounded-2xl border border-stone-200 bg-white/95 p-3 text-sm text-slate-800 shadow-lg backdrop-blur md:left-4 md:right-auto md:w-96">
          <div className="flex items-start gap-2">
            <MapPin className="mt-0.5 h-4 w-4 text-teal-900" />
            <p>Tap the map to draw a {featureTypeLabel(drawingType).toLowerCase()}. Save from the drawing controls after adding enough points.</p>
          </div>
        </div>
      )}
    </div>
  );
}
