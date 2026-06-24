import { useEffect, useMemo, useRef, useState } from 'react';
import DeckGL from '@deck.gl/react';
import { Map } from 'react-map-gl/maplibre';
import { PathLayer, PolygonLayer, ScatterplotLayer, TextLayer } from '@deck.gl/layers';
import { Layers, LocateFixed, Plus, RotateCcw, X } from 'lucide-react';
import { MoneyCragNode, MoneyFeature, MoneyPosition } from '../../../types/money';
import { bbox, bboxCenter, centroid, flattenAreas, flattenBoulders, geometryPoints, MONEY_CREEK_CENTER, polygonGeoJSON } from './model';
import { DEV, T } from './theme';
import 'maplibre-gl/dist/maplibre-gl.css';
import type { StyleSpecification } from 'maplibre-gl';

type LayerState = { base: string; contours: boolean; trails: boolean; areas: Record<string, boolean>; dev: Record<string, boolean> };
type ViewState = { longitude: number; latitude: number; zoom: number; pitch: number; bearing: number };
type Rgba = [number, number, number, number];
type AreaDatum = { node: MoneyCragNode; polygon: MoneyPosition[]; center: MoneyPosition; focused: boolean };
type BoulderDatum = { node: MoneyCragNode; polygon: MoneyPosition[]; center: MoneyPosition; color: Rgba; line: Rgba; rank: number; focused: boolean };
type TrailDatum = { node: MoneyCragNode; path: MoneyPosition[] };
type TerrainBand = { polygon: MoneyPosition[]; fill: Rgba; line: Rgba };
type TerrainPath = { path: MoneyPosition[]; color: Rgba; width: number; major?: boolean };
type TerrainLabel = { position: MoneyPosition; text: string; size: number; color: Rgba; background: Rgba };

interface Props {
  root: MoneyCragNode;
  area: MoneyCragNode;
  trails: MoneyCragNode[];
  selectedBoulderId: string | null;
  selectedTrailId: string | null;
  mode: 'view' | 'create-area' | 'create-boulder';
  layers: LayerState;
  mobile: boolean;
  onEnter: (id: string) => void;
  onSelectBoulder: (id: string | null) => void;
  onSelectTrail: (id: string | null) => void;
  onCreateDone: (points: MoneyPosition[]) => void;
  onCreateCancel: () => void;
  setLayers: (layers: LayerState) => void;
}

const REAL_TOPO_STYLE = 'https://basemaps.cartocdn.com/gl/voyager-gl-style/style.json';
const REAL_LIGHT_STYLE = 'https://basemaps.cartocdn.com/gl/positron-gl-style/style.json';
const SATELLITE_STYLE: StyleSpecification = {
  version: 8,
  sources: { esri: { type: 'raster', tiles: ['https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}'], tileSize: 256, attribution: 'Tiles © Esri' } },
  layers: [{ id: 'esri', type: 'raster', source: 'esri' }],
};
const STYLIZED_STYLE: StyleSpecification = {
  version: 8,
  sources: {},
  layers: [{ id: 'money-reference-stylized-bg', type: 'background', paint: { 'background-color': T.map.bg2 } }],
};

const MIN_ZOOM = 12.2;
const MAX_ZOOM = 22;
const DETAIL_ZOOM = 16.15;
const FINE_DETAIL_ZOOM = 17.25;

export function CragMap({ root, area, trails, selectedBoulderId, selectedTrailId, mode, layers, mobile, onEnter, onSelectBoulder, onSelectTrail, onCreateDone, onCreateCancel, setLayers }: Props) {
  const wrapRef = useRef<HTMLDivElement | null>(null);
  const [viewState, setViewState] = useState<ViewState>(() => ({ longitude: MONEY_CREEK_CENTER[0], latitude: MONEY_CREEK_CENTER[1], zoom: 14.8, pitch: 0, bearing: 0 }));
  const [draft, setDraft] = useState<MoneyPosition[]>([]);
  const [layersOpen, setLayersOpen] = useState(!mobile);
  const creating = mode !== 'view';

  useEffect(() => { setViewState(viewForBBox(bbox(area))); }, [area.feature.id, area]);
  useEffect(() => { if (creating) setDraft([]); }, [creating]);

  const areaBox = useMemo(() => bbox(area), [area]);
  const terrain = useMemo(() => buildStylizedTerrain(areaBox, viewState.zoom, trailDataSeed(trails)), [areaBox, trails, viewState.zoom]);

  const focusedAreaIds = useMemo(() => new Set(flattenAreas(area).map(node => node.feature.id)), [area]);
  const focusedBoulderIds = useMemo(() => new Set(flattenBoulders(area).map(node => node.feature.id)), [area]);

  const areaData = useMemo<AreaDatum[]>(() => flattenAreas(root)
    .filter(node => node.feature.id !== root.feature.id && layers.areas[node.feature.id] !== false)
    .map(node => ({ node, polygon: geometryPoints(node.feature.geojson), center: featureCenter(node.feature), focused: focusedAreaIds.has(node.feature.id) }))
    .filter(d => usablePolygon(d.polygon)), [root, layers.areas, focusedAreaIds]);

  const boulderData = useMemo<BoulderDatum[]>(() => flattenBoulders(root).flatMap((node, index) => {
    const dev = String(node.feature.status);
    if (layers.dev[dev] === false) return [];
    const meta = DEV.meta[(dev in DEV.meta ? dev : 'scouted') as keyof typeof DEV.meta];
    const center = featureCenter(node.feature);
    if (!isLonLatPosition(center)) return [];
    return [{ node, polygon: boulderFootprint(geometryPoints(node.feature.geojson), center, node.feature.id), center, color: hexToRgba(meta.bg, 86), line: hexToRgba(meta.c, 235), rank: stableHash(node.feature.id || String(index)) % 5, focused: focusedBoulderIds.has(node.feature.id) }];
  }), [root, layers.dev, focusedBoulderIds]);

  const trailData = useMemo<TrailDatum[]>(() => layers.trails ? trails.map(node => ({ node, path: geometryPoints(node.feature.geojson) })).filter(d => usablePath(d.path)) : [], [layers.trails, trails]);
  const boulderLabelData = useMemo(() => boulderData.filter(d => d.node.feature.id === selectedBoulderId || d.focused && (viewState.zoom >= FINE_DETAIL_ZOOM || viewState.zoom >= DETAIL_ZOOM && d.rank === 0)), [boulderData, selectedBoulderId, viewState.zoom]);

  const deckLayers = useMemo(() => {
    const draftGeo = draft.length > 0 ? polygonGeoJSON(draft) : null;
    const stylized = layers.base === 'stylized' ? [
      new PolygonLayer<TerrainBand>({
        id: 'money-reference-stylized-terrain-wash', data: terrain.bands, pickable: false, stroked: false, filled: true,
        getPolygon: d => d.polygon, getFillColor: d => d.fill,
      }),
      layers.contours && new PathLayer<TerrainPath>({
        id: 'money-reference-stylized-contours', data: terrain.contours, pickable: false, getPath: d => d.path, getColor: d => d.color, getWidth: d => d.major ? d.width * 1.2 : d.width, widthMinPixels: 0.45, widthMaxPixels: viewState.zoom >= DETAIL_ZOOM ? 1.75 : 1.2, jointRounded: true, capRounded: true,
      }),
      new PathLayer<TerrainPath>({
        id: 'money-reference-stylized-texture', data: terrain.texture, pickable: false, getPath: d => d.path, getColor: d => d.color, getWidth: d => d.width, widthMinPixels: 0.35, widthMaxPixels: 0.9, jointRounded: true, capRounded: true,
      }),
      new PathLayer<TerrainPath>({
        id: 'money-reference-stylized-ravines-halo', data: terrain.ravines, pickable: false, getPath: d => d.path, getColor: [28, 20, 16, 58], getWidth: d => d.width + 2, widthMinPixels: 1.6, widthMaxPixels: 4.5, jointRounded: true, capRounded: true,
      }),
      new PathLayer<TerrainPath>({
        id: 'money-reference-stylized-ravines', data: terrain.ravines, pickable: false, getPath: d => d.path, getColor: d => d.color, getWidth: d => d.width, widthMinPixels: 1.5, widthMaxPixels: 5, jointRounded: true, capRounded: true,
      }),
      new TextLayer<TerrainLabel>({
        id: 'money-reference-stylized-terrain-labels', data: terrain.labels, pickable: false, getPosition: d => d.position, getText: d => d.text, getSize: d => d.size, getColor: d => d.color, getBackgroundColor: d => d.background, background: true, backgroundPadding: [7, 4], fontFamily: T.mono, fontWeight: 700,
      }),
    ].filter(Boolean) : [];

    return [
      ...stylized,
      new PolygonLayer<AreaDatum>({
        id: 'money-reference-areas', data: areaData, pickable: !creating, stroked: true, filled: true, lineWidthMinPixels: 0.65, lineWidthMaxPixels: 2.4,
        getPolygon: d => d.polygon, getFillColor: d => areaFill(d, layers.base), getLineColor: d => d.focused ? [174, 185, 116, 205] : [174, 185, 116, 82], getLineWidth: d => d.focused ? 1.35 : 0.65,
        onClick: info => { if (info.object) onEnter(info.object.node.feature.id); return true; },
      }),
      new ScatterplotLayer<BoulderDatum>({
        id: 'money-reference-boulder-markers', data: boulderData, pickable: !creating, stroked: true, filled: true, radiusMinPixels: 4, radiusMaxPixels: 16, lineWidthMinPixels: 1, lineWidthMaxPixels: 3,
        getPosition: d => d.center, getRadius: d => d.node.feature.id === selectedBoulderId ? 3.5 : d.focused ? 2.25 : 1.55, getFillColor: d => d.node.feature.id === selectedBoulderId ? withAlpha(d.line, 245) : d.focused ? withAlpha(d.line, 210) : withAlpha(d.line, 118), getLineColor: d => d.node.feature.id === selectedBoulderId ? [255, 245, 220, 255] : d.focused ? [28, 24, 18, 230] : [28, 24, 18, 120], getLineWidth: d => d.node.feature.id === selectedBoulderId ? 1.5 : 0.85,
        onClick: info => { if (info.object) onSelectBoulder(info.object.node.feature.id); return true; },
      }),
      new PathLayer<TrailDatum>({ id: 'money-reference-trails-halo', data: trailData, getPath: d => d.path, getColor: [17, 24, 20, layers.base === 'stylized' ? 92 : 130], getWidth: d => d.node.feature.id === selectedTrailId ? 4.2 : 3, widthMinPixels: 2, widthMaxPixels: 5.5, jointRounded: true, capRounded: true }),
      new PathLayer<TrailDatum>({
        id: 'money-reference-trails', data: trailData, pickable: !creating, getPath: d => d.path, getColor: d => d.node.feature.id === selectedTrailId ? [174, 185, 116, 255] : [185, 128, 80, 225], getWidth: d => d.node.feature.id === selectedTrailId ? 2.6 : 1.45, widthMinPixels: 1.2, widthMaxPixels: 3.8, jointRounded: true, capRounded: true,
        onClick: info => { if (info.object) onSelectTrail(info.object.node.feature.id); return true; },
      }),
      viewState.zoom >= DETAIL_ZOOM && new PathLayer<TrailDatum>({
        id: 'money-reference-trails-fine-detail', data: trailData, pickable: false, getPath: d => densifyPath(d.path, 12), getColor: [238, 225, 211, 95], getWidth: 0.45, widthMinPixels: 0.6, widthMaxPixels: 1.5, jointRounded: true, capRounded: true,
      }),
      new TextLayer<AreaDatum>({ id: 'money-reference-area-labels', data: areaData.filter(d => d.focused || viewState.zoom < DETAIL_ZOOM), getPosition: d => d.center, getText: d => d.node.feature.title, getSize: d => d.focused && viewState.zoom >= DETAIL_ZOOM ? 12 : 10.5, getColor: d => d.focused ? [238, 225, 211, 255] : [238, 225, 211, 155], getBackgroundColor: d => d.focused ? [42, 36, 29, layers.base === 'stylized' ? 196 : 220] : [42, 36, 29, 128], background: true, backgroundPadding: [7, 4], fontWeight: 700, pickable: false }),
      new TextLayer<BoulderDatum>({ id: 'money-reference-boulder-labels', data: boulderLabelData, getPosition: d => d.center, getText: d => d.node.feature.title, getSize: d => d.node.feature.id === selectedBoulderId ? 11 : viewState.zoom >= FINE_DETAIL_ZOOM ? 9.5 : 8.5, getColor: d => d.line, getBackgroundColor: [42, 36, 29, 205], background: true, backgroundPadding: [5, 2], fontWeight: 700, pickable: false, getPixelOffset: [0, -12] }),
      draftGeo && new PolygonLayer({ id: 'money-reference-draft', data: [{ polygon: geometryPoints(draftGeo) }], getPolygon: (d: { polygon: MoneyPosition[] }) => d.polygon, getFillColor: [174, 185, 116, 70], getLineColor: [174, 185, 116, 255], getLineWidth: 1.8, lineWidthMinPixels: 1.2, lineWidthMaxPixels: 4 }),
      draft.length > 0 && new ScatterplotLayer({ id: 'money-reference-draft-points', data: draft.map(position => ({ position })), getPosition: (d: { position: MoneyPosition }) => d.position, getFillColor: [255, 255, 255, 245], getLineColor: [174, 185, 116, 255], getRadius: 1.2, radiusMinPixels: 4, radiusMaxPixels: 8, stroked: true }),
    ].filter(Boolean);
  }, [areaData, boulderData, boulderLabelData, creating, draft, layers.base, layers.contours, onEnter, onSelectBoulder, onSelectTrail, selectedBoulderId, selectedTrailId, terrain, trailData, viewState.zoom]);

  const finish = () => { if (draft.length >= 3) onCreateDone(draft); };
  const focus = () => setViewState(viewForBBox(bbox(area)));
  const zoomBy = (delta: number) => setViewState(current => ({ ...current, zoom: clamp(current.zoom + delta, MIN_ZOOM, MAX_ZOOM) }));

  return <div ref={wrapRef} style={{ position: 'absolute', inset: 0, overflow: 'hidden', background: T.map.bg2 }}>
    <DeckGL controller={{ dragPan: !creating, scrollZoom: true, doubleClickZoom: !creating, touchZoom: true, keyboard: true }} layers={deckLayers} viewState={viewState} onViewStateChange={(event: { viewState: unknown }) => {
      const next = event.viewState as Partial<ViewState>;
      setViewState(current => ({ longitude: next.longitude ?? current.longitude, latitude: next.latitude ?? current.latitude, zoom: clamp(next.zoom ?? current.zoom, MIN_ZOOM, MAX_ZOOM), pitch: next.pitch ?? current.pitch, bearing: next.bearing ?? current.bearing }));
    }} onClick={(info: { coordinate?: number[]; picked?: boolean }) => {
      if (creating && info.coordinate && info.coordinate.length >= 2) setDraft(d => [...d, [info.coordinate![0], info.coordinate![1]]]);
      if (!creating && !info.picked) { onSelectBoulder(null); onSelectTrail(null); }
    }} getCursor={() => creating ? 'crosshair' : 'grab'}>
      <Map mapStyle={mapStyle(layers.base) as never} reuseMaps minZoom={MIN_ZOOM} maxZoom={MAX_ZOOM} />
    </DeckGL>
    {layers.base === 'stylized' && <div className="pointer-events-none absolute inset-0" style={{ background: 'radial-gradient(circle at 46% 36%, rgba(174,185,116,0.13), transparent 32%), radial-gradient(circle at 74% 65%, rgba(134,160,182,0.08), transparent 28%), linear-gradient(180deg, rgba(23,17,15,0.06), rgba(23,17,15,0.30))', mixBlendMode: 'soft-light' }} />}
    {layers.base === 'slope' && <div className="pointer-events-none absolute inset-0" style={{ background: 'linear-gradient(135deg, rgba(62,122,78,0.23), rgba(201,184,74,0.18) 45%, rgba(200,87,47,0.20))' }} />}

    {!creating && <div style={{ position: 'absolute', left: 14, top: 14, display: 'flex', flexDirection: 'column', gap: 8, zIndex: 10 }}>
      <button onClick={focus} style={mapButton}><LocateFixed size={16} />Focus</button>
      <div style={{ display: 'flex', background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 10, overflow: 'hidden', boxShadow: T.shadow }}>
        <button onClick={() => zoomBy(1)} style={smallButton}>+</button><button onClick={() => zoomBy(-1)} style={smallButton}>−</button>
      </div>
    </div>}
    {!creating && (layersOpen ? <LayersPanel layers={layers} setLayers={setLayers} onClose={() => setLayersOpen(false)} root={root} /> : <button onClick={() => setLayersOpen(true)} style={{ position: 'absolute', top: 14, right: 14, zIndex: 10, display: 'flex', gap: 7, background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 10, padding: '9px 13px', color: T.ink, fontWeight: 700, boxShadow: T.shadow }}><Layers size={16} />Layers</button>)}
    {creating && <div style={{ position: 'absolute', bottom: mobile ? 22 : 20, left: '50%', transform: 'translateX(-50%)', display: 'flex', gap: 8, background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 14, padding: 8, boxShadow: T.shadow, zIndex: 10 }}><button onClick={onCreateCancel} style={ctrl(false)}><X size={16} />Cancel</button><button onClick={() => setDraft(d => d.slice(0, -1))} style={ctrl(false)}><RotateCcw size={16} />Undo</button><button disabled={draft.length < 3} onClick={finish} style={ctrl(true, draft.length < 3)}><Plus size={16} />Done</button></div>}
    {creating && <div style={{ position: 'absolute', left: 14, bottom: mobile ? 92 : 84, background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 10, padding: '8px 11px', color: T.mut, fontSize: 12, zIndex: 10 }}>Tap the real map to add lon/lat vertices · {draft.length} point{draft.length === 1 ? '' : 's'}</div>}
  </div>;
}

function featureCenter(feature: MoneyFeature): MoneyPosition {
  const center = feature.properties?.center;
  if (Array.isArray(center) && typeof center[0] === 'number' && typeof center[1] === 'number') return [center[0], center[1]];
  return centroid(geometryPoints(feature.geojson));
}

function viewForBBox(box: [number, number, number, number]): ViewState {
  const center = bboxCenter(box);
  const span = Math.max(box[2] - box[0], box[3] - box[1]);
  const zoom = clamp(Math.log2(360 / Math.max(span, 0.00025)) - 0.65, 13.35, 19.2);
  return { longitude: center[0], latitude: center[1], zoom, pitch: 0, bearing: 0 };
}

function mapStyle(base: string): string | StyleSpecification {
  if (base === 'satellite') return SATELLITE_STYLE;
  if (base === 'topo') return REAL_TOPO_STYLE;
  if (base === 'slope') return REAL_LIGHT_STYLE;
  return STYLIZED_STYLE;
}

function buildStylizedTerrain(box: [number, number, number, number], zoom: number, trailSeed: number) {
  const p = T.map;
  const [minLon, minLat, maxLon, maxLat] = padBox(box, 0.18);
  const cx = (minLon + maxLon) / 2;
  const cy = (minLat + maxLat) / 2;
  const sx = maxLon - minLon;
  const sy = maxLat - minLat;
  const fine = zoom >= DETAIL_ZOOM;
  const micro = zoom >= FINE_DETAIL_ZOOM;
  const ringScales = fine ? [1, 0.78, 0.58, 0.4, 0.25] : [1, 0.7, 0.44];
  const hills = [
    { c: [cx - sx * 0.18, cy + sy * 0.12] as MoneyPosition, rx: sx * 0.30, ry: sy * 0.38, seed: 31 },
    { c: [cx + sx * 0.26, cy + sy * 0.22] as MoneyPosition, rx: sx * 0.22, ry: sy * 0.31, seed: 73 },
    { c: [cx + sx * 0.05, cy - sy * 0.26] as MoneyPosition, rx: sx * 0.25, ry: sy * 0.23, seed: 117 },
  ];

  const bands: TerrainBand[] = [
    { polygon: [[minLon, minLat], [maxLon, minLat], [maxLon, maxLat], [minLon, maxLat], [minLon, minLat]], fill: hexToRgba(p.basin, 112), line: [0, 0, 0, 0] },
  ];
  const contours: TerrainPath[] = [];
  hills.forEach((hill, hillIndex) => {
    ringScales.forEach((scale, ringIndex) => {
      const ring = terrainRing(hill.c, hill.rx * scale, hill.ry * scale, hill.seed + ringIndex * 13, fine ? 72 : 48);
      if (ringIndex === 0) bands.push({ polygon: ring, fill: terrainFill(hillIndex), line: [0, 0, 0, 0] });
      contours.push({ path: ring, color: ringIndex % 2 === 0 ? hexToRgba(p.ridge, hillIndex === 0 ? 92 : 72) : hexToRgba(p.contour, fine ? 74 : 54), width: ringIndex % 2 === 0 ? 0.95 : 0.65, major: ringIndex % 2 === 0 });
    });
  });

  if (micro) {
    hills.forEach((hill, hillIndex) => [0.88, 0.33].forEach((scale, i) => contours.push({ path: terrainRing(hill.c, hill.rx * scale, hill.ry * scale, hill.seed + trailSeed + i * 19, 84), color: hexToRgba(p.slotStripe, 42), width: 0.35, major: hillIndex === 0 && i === 0 })));
  }

  const creek = creekPath(minLon, minLat, maxLon, maxLat, trailSeed);
  const ravines: TerrainPath[] = [
    { path: creek, color: hexToRgba(p.creek, 205), width: 2.15, major: true },
    { path: ravinePath(hills[0].c, creek[Math.floor(creek.length * 0.42)], 9), color: hexToRgba(p.slot, 116), width: 1.05 },
    { path: ravinePath(hills[1].c, creek[Math.floor(creek.length * 0.62)], 17), color: hexToRgba(p.slot, 104), width: 0.95 },
  ];

  const texture = textureStrokes(minLon, minLat, maxLon, maxLat, fine, micro);
  const labels: TerrainLabel[] = [
    { position: creek[Math.floor(creek.length * 0.55)], text: 'MONEY CREEK', size: 10, color: hexToRgba(p.creek, 235), background: [24, 17, 14, 188] },
    { position: [hills[0].c[0], hills[0].c[1] + sy * 0.15], text: 'WEST RIDGE', size: 9, color: hexToRgba(p.ridge, 230), background: [24, 17, 14, 170] },
    ...(fine ? [{ position: [hills[2].c[0] + sx * 0.04, hills[2].c[1] - sy * 0.13] as MoneyPosition, text: 'TALUS BENCH', size: 8.5, color: hexToRgba(p.talus, 230), background: [24, 17, 14, 158] as Rgba }] : []),
  ];

  return { bands, contours, ravines, texture, labels };
}

// Procedural terrain is deterministic in lon/lat because no DEM/topo dataset is shipped locally.
// The generated rings, ravines, texture, and labels share the MapLibre camera/projection with the real basemap and feature overlays.
function terrainRing(center: MoneyPosition, rx: number, ry: number, seed: number, count: number): MoneyPosition[] {
  const points = Array.from({ length: count }, (_, i) => {
    const a = i / count * Math.PI * 2;
    const wobble = 1 + Math.sin(a * 3 + seed) * 0.055 + Math.cos(a * 5 + seed * 0.37) * 0.035 + Math.sin(a * 9 + seed * 0.19) * 0.018;
    return [center[0] + Math.cos(a) * rx * wobble, center[1] + Math.sin(a) * ry * wobble] as MoneyPosition;
  });
  return [...points, points[0]];
}

function terrainFill(index: number): Rgba {
  const fills: Rgba[] = [hexToRgba(T.map.forestOuter, 34), hexToRgba(T.map.forestMid, 30), hexToRgba(T.map.talus, 26)];
  return fills[Math.min(index, fills.length - 1)];
}

function creekPath(minLon: number, minLat: number, maxLon: number, maxLat: number, seed: number): MoneyPosition[] {
  const sx = maxLon - minLon;
  const sy = maxLat - minLat;
  return Array.from({ length: 54 }, (_, i) => {
    const t = i / 53;
    const lon = minLon + sx * t;
    const lat = minLat + sy * (0.72 - t * 0.28 + Math.sin(t * Math.PI * 2.6 + seed * 0.01) * 0.035 + Math.sin(t * Math.PI * 7.1) * 0.012);
    return [lon, lat] as MoneyPosition;
  });
}

function ravinePath(from: MoneyPosition, to: MoneyPosition, seed: number): MoneyPosition[] {
  return Array.from({ length: 22 }, (_, i) => {
    const t = i / 21;
    const bend = Math.sin(t * Math.PI) * 0.00016;
    return [lerp(from[0], to[0], t) + bend * Math.cos(seed), lerp(from[1], to[1], t) + bend * Math.sin(seed)] as MoneyPosition;
  });
}

function textureStrokes(minLon: number, minLat: number, maxLon: number, maxLat: number, fine: boolean, micro: boolean): TerrainPath[] {
  const cols = micro ? 12 : fine ? 10 : 7;
  const rows = micro ? 8 : fine ? 7 : 5;
  const sx = maxLon - minLon;
  const sy = maxLat - minLat;
  const strokes: TerrainPath[] = [];
  for (let y = 0; y < rows; y += 1) for (let x = 0; x < cols; x += 1) {
    const h = stableHash(`${x}:${y}:money-terrain`);
    if (h % (fine ? 2 : 3) === 0) continue;
    const lon = minLon + sx * ((x + 0.35 + (h % 17) / 50) / cols);
    const lat = minLat + sy * ((y + 0.35 + (h % 23) / 60) / rows);
    const len = sx * (micro ? 0.008 : 0.011) * (0.65 + (h % 9) / 12);
    const angle = -0.65 + (h % 100) / 180;
    strokes.push({ path: [[lon - Math.cos(angle) * len, lat - Math.sin(angle) * len * 0.55], [lon + Math.cos(angle) * len, lat + Math.sin(angle) * len * 0.55]], color: hexToRgba(T.map.slotStripe, micro ? 34 : 26), width: 0.35 });
  }
  return strokes;
}

function boulderFootprint(points: MoneyPosition[], center: MoneyPosition, id: string): MoneyPosition[] {
  const seed = stableHash(id);
  const maxMeters = 1.15 + seed % 100 / 100 * 1.45;
  if (points.length >= 3) {
    const shrunk = points.map(point => capOffset(center, point, maxMeters));
    return closePolygon(shrunk);
  }
  const count = 9;
  const major = maxMeters;
  const minor = Math.max(0.55, maxMeters * (0.48 + seed % 17 / 70));
  const angle = seed % 360 * Math.PI / 180;
  const generated = Array.from({ length: count }, (_, i) => {
    const a = i / count * Math.PI * 2;
    const r = 0.88 + Math.sin(a * 3 + seed) * 0.08;
    const east = (Math.cos(a) * major * Math.cos(angle) - Math.sin(a) * minor * Math.sin(angle)) * r;
    const north = (Math.cos(a) * major * Math.sin(angle) + Math.sin(a) * minor * Math.cos(angle)) * r;
    return offsetMeters(center, east, north);
  });
  return closePolygon(generated);
}

function capOffset(center: MoneyPosition, point: MoneyPosition, maxMeters: number): MoneyPosition {
  const meters = lonLatToMeters(center, point);
  const distance = Math.hypot(meters[0], meters[1]);
  if (distance < 0.05) return offsetMeters(center, maxMeters * 0.45, 0);
  const scale = Math.min(1, maxMeters / distance);
  return offsetMeters(center, meters[0] * scale, meters[1] * scale);
}

function densifyPath(path: MoneyPosition[], minSegments: number): MoneyPosition[] {
  if (path.length < 2) return path;
  const out: MoneyPosition[] = [];
  for (let i = 0; i < path.length - 1; i += 1) {
    const a = path[i], b = path[i + 1];
    const steps = i % 2 === 0 ? minSegments : Math.max(3, Math.floor(minSegments / 2));
    for (let s = 0; s < steps; s += 1) out.push([lerp(a[0], b[0], s / steps), lerp(a[1], b[1], s / steps)]);
  }
  out.push(path[path.length - 1]);
  return out;
}

function padBox(box: [number, number, number, number], ratio: number): [number, number, number, number] {
  const dx = Math.max(0.001, (box[2] - box[0]) * ratio);
  const dy = Math.max(0.001, (box[3] - box[1]) * ratio);
  return [box[0] - dx, box[1] - dy, box[2] + dx, box[3] + dy];
}

function trailDataSeed(trails: MoneyCragNode[]): number {
  return trails.reduce((sum, trail) => sum + stableHash(trail.feature.id), 0);
}

function closePolygon(points: MoneyPosition[]): MoneyPosition[] {
  if (!points.length) return points;
  const first = points[0], last = points[points.length - 1];
  return first[0] === last[0] && first[1] === last[1] ? points : [...points, first];
}

function lonLatToMeters(origin: MoneyPosition, point: MoneyPosition): [number, number] {
  const cosLat = Math.max(0.1, Math.cos(origin[1] * Math.PI / 180));
  return [(point[0] - origin[0]) * 111320 * cosLat, (point[1] - origin[1]) * 111320];
}

function offsetMeters(origin: MoneyPosition, east: number, north: number): MoneyPosition {
  const cosLat = Math.max(0.1, Math.cos(origin[1] * Math.PI / 180));
  return [origin[0] + east / (111320 * cosLat), origin[1] + north / 111320];
}

function areaFill(d: AreaDatum, base: string): Rgba { return d.focused ? base === 'stylized' ? [174, 185, 116, 30] : [174, 185, 116, 44] : base === 'stylized' ? [174, 185, 116, 8] : [174, 185, 116, 14]; }
function isLonLatPosition(p: MoneyPosition): boolean { return Number.isFinite(p[0]) && Number.isFinite(p[1]) && Math.abs(p[0]) <= 180 && Math.abs(p[1]) <= 90; }
function usablePath(path: MoneyPosition[]): boolean { return path.length >= 2 && path.every(isLonLatPosition); }
function usablePolygon(points: MoneyPosition[]): boolean { return points.length >= 3 && points.every(isLonLatPosition); }
function withAlpha(color: Rgba, alpha: number): Rgba { return [color[0], color[1], color[2], alpha]; }
function clamp(value: number, min: number, max: number) { return Math.max(min, Math.min(max, value)); }
function lerp(a: number, b: number, t: number) { return a + (b - a) * t; }

function stableHash(input: string): number {
  let hash = 2166136261;
  for (let i = 0; i < input.length; i += 1) hash = Math.imul(hash ^ input.charCodeAt(i), 16777619);
  return hash >>> 0;
}

function hexToRgba(hexOrRgba: string, fallbackAlpha: number): Rgba {
  const rgba = hexOrRgba.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)(?:,\s*([\d.]+))?\)/i);
  if (rgba) return [Number(rgba[1]), Number(rgba[2]), Number(rgba[3]), rgba[4] ? Math.round(Number(rgba[4]) * 255) : fallbackAlpha];
  const hex = hexOrRgba.replace('#', '');
  if (hex.length >= 6) return [parseInt(hex.slice(0, 2), 16), parseInt(hex.slice(2, 4), 16), parseInt(hex.slice(4, 6), 16), fallbackAlpha];
  return [174, 185, 116, fallbackAlpha];
}

const mapButton: React.CSSProperties = { display: 'flex', alignItems: 'center', gap: 7, border: `1px solid ${T.line2}`, borderRadius: 10, padding: '9px 12px', background: T.surf, color: T.ink, fontFamily: T.font, fontWeight: 700, cursor: 'pointer', boxShadow: T.shadow };
const smallButton: React.CSSProperties = { width: 40, height: 38, border: 'none', borderRight: `1px solid ${T.line}`, background: 'transparent', color: T.ink, cursor: 'pointer', fontSize: 18 };
function ctrl(accent: boolean, disabled = false): React.CSSProperties { return { display: 'flex', alignItems: 'center', gap: 7, border: accent ? 'none' : `1px solid ${T.line2}`, borderRadius: 9, padding: '11px 18px', background: disabled ? T.line : accent ? T.accent : 'transparent', color: disabled ? T.faint : accent ? T.onAccent : T.ink, fontFamily: T.font, fontWeight: 700, cursor: disabled ? 'default' : 'pointer' }; }
function LayersPanel({ layers, setLayers, onClose, root }: { layers: LayerState; setLayers: (l: LayerState) => void; onClose: () => void; root: MoneyCragNode }) { const bases = ['stylized', 'topo', 'satellite', 'slope']; const areaRows = flattenAreas(root).filter(a => a.feature.id !== root.feature.id); return <div style={{ position: 'absolute', top: 14, right: 14, zIndex: 10, width: 224, background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 12, boxShadow: T.shadow, overflow: 'hidden' }}><div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '11px 13px', borderBottom: `1px solid ${T.line}`, color: T.ink }}><Layers size={16} /><b style={{ fontSize: 13.5 }}>Layers</b><button onClick={onClose} style={{ marginLeft: 'auto', border: 'none', background: 'transparent', color: T.mut, cursor: 'pointer' }}>×</button></div><div style={{ padding: '11px 13px', maxHeight: 380, overflowY: 'auto' }}><Label>Base map</Label>{bases.map(base => <Row key={base} onClick={() => setLayers({ ...layers, base })} on={layers.base === base} label={base} />)}<Label>Areas</Label>{areaRows.map(c => <Row key={c.feature.id} onClick={() => setLayers({ ...layers, areas: { ...layers.areas, [c.feature.id]: layers.areas[c.feature.id] === false } })} on={layers.areas[c.feature.id] !== false} label={c.feature.title} />)}<Label>Development</Label>{DEV.order.map(k => <Row key={k} onClick={() => setLayers({ ...layers, dev: { ...layers.dev, [k]: layers.dev[k] === false } })} on={layers.dev[k] !== false} label={DEV.meta[k].short} color={DEV.meta[k].c} />)}<div style={{ borderTop: `1px solid ${T.line}`, margin: '10px 0 8px' }} />{(['contours', 'trails'] as const).map(k => <Row key={k} onClick={() => setLayers({ ...layers, [k]: !layers[k] })} on={layers[k]} label={k === 'contours' ? 'terrain tint' : k} />)}</div></div>; }
function Label({ children }: { children: React.ReactNode }) { return <div style={{ fontFamily: T.mono, fontSize: 10, letterSpacing: 0.6, color: T.faint, textTransform: 'uppercase', margin: '14px 0 7px' }}>{children}</div>; }
function Row({ on, label, onClick, color }: { on: boolean; label: string; onClick: () => void; color?: string }) { return <div onClick={onClick} style={{ display: 'flex', alignItems: 'center', gap: 9, padding: '5px 0', cursor: 'pointer' }}><span style={{ width: 9, height: 9, borderRadius: color ? '50%' : 2, background: color ?? T.accent, transform: color ? 'none' : 'rotate(45deg)' }} /><span style={{ flex: 1, fontSize: 12.5, color: on ? T.ink : T.mut, textTransform: label.length < 9 ? 'capitalize' : undefined }}>{label}</span><span style={{ width: 28, height: 16, borderRadius: 9, background: on ? (color ?? T.accent) : T.line2, position: 'relative' }}><span style={{ position: 'absolute', top: 2, left: on ? 14 : 2, width: 12, height: 12, borderRadius: '50%', background: '#fff' }} /></span></div>; }
