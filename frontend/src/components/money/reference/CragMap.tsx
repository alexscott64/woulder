import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import DeckGL from '@deck.gl/react';
import { Map } from 'react-map-gl/maplibre';
import { PathLayer, PolygonLayer, ScatterplotLayer, TextLayer } from '@deck.gl/layers';
import { PathStyleExtension } from '@deck.gl/extensions';
import { Layers, LocateFixed, Plus, RotateCcw, Save, Trash2, X } from 'lucide-react';
import { MoneyCragNode, MoneyFeature, MoneyPosition } from '../../../types/money';
import { bbox, bboxCenter, closedPolygonPoints, centroid, deletePolygonVertex, flattenAreas, flattenBoulders, geometryPoints, insertPolygonVertexAfter, isValidAreaEditRing, MONEY_CREEK_CENTER, openPolygonPoints, polygonGeoJSON, replacePolygonVertex } from './model';
import { loadMoneyCreekLineMap, type LineMapLabel, type LineMapPath, type MoneyCreekLineMapData } from './lineMap';
import { DEV, T } from './theme';
import 'maplibre-gl/dist/maplibre-gl.css';
import type { StyleSpecification } from 'maplibre-gl';

type LayerState = { base: string; roads: boolean; water: boolean; contours: boolean; trails: boolean; areas: Record<string, boolean>; dev: Record<string, boolean> };
type ViewState = { longitude: number; latitude: number; zoom: number; pitch: number; bearing: number };
type Rgba = [number, number, number, number];
type AreaDatum = { node: MoneyCragNode; polygon: MoneyPosition[]; center: MoneyPosition };
type BoulderDatum = { node: MoneyCragNode; polygon: MoneyPosition[]; center: MoneyPosition; color: Rgba; line: Rgba; rank: number };
type TrailDatum = { node: MoneyCragNode; path: MoneyPosition[] };
type TrailFineDatum = TrailDatum & { finePath: MoneyPosition[] };
type LineMapLoadState =
  | { status: 'idle' | 'loading'; data: null; error?: undefined }
  | { status: 'ready'; data: MoneyCreekLineMapData; error?: undefined }
  | { status: 'error'; data: null; error: unknown };
type IdleWindow = Window & typeof globalThis & {
  requestIdleCallback?: (callback: () => void, options?: { timeout?: number }) => number;
  cancelIdleCallback?: (handle: number) => void;
};

interface Props {
  root: MoneyCragNode;
  area: MoneyCragNode;
  trails: MoneyCragNode[];
  selectedBoulderId: string | null;
  selectedTrailId: string | null;
  mode: 'view' | 'create-area' | 'create-boulder' | 'edit-area';
  layers: LayerState;
  mobile: boolean;
  onEnter: (id: string) => void;
  onSelectBoulder: (id: string | null) => void;
  onSelectTrail: (id: string | null) => void;
  onCreateDone: (points: MoneyPosition[]) => void;
  onCreateCancel: () => void;
  onEditSave: (points: MoneyPosition[]) => void;
  onEditCancel: () => void;
  setLayers: (layers: LayerState) => void;
}

const REAL_TOPO_STYLE = 'https://basemaps.cartocdn.com/gl/voyager-gl-style/style.json';
const REAL_LIGHT_STYLE = 'https://basemaps.cartocdn.com/gl/positron-gl-style/style.json';
const SATELLITE_STYLE: StyleSpecification = {
  version: 8,
  sources: { esri: { type: 'raster', tiles: ['https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}'], tileSize: 256, attribution: 'Tiles © Esri' } },
  layers: [{ id: 'esri', type: 'raster', source: 'esri' }],
};
const STYLIZED_STYLE: StyleSpecification = { version: 8, sources: {}, layers: [] };

const MIN_ZOOM = 12.2;
const MAX_ZOOM = 22;
const DETAIL_ZOOM = 16.15;
const FINE_DETAIL_ZOOM = 17.25;
const DASHED_PATH_STYLE = new PathStyleExtension({ dash: true });

export function CragMap({ root, area, trails, selectedBoulderId, selectedTrailId, mode, layers, mobile, onEnter, onSelectBoulder, onSelectTrail, onCreateDone, onCreateCancel, onEditSave, onEditCancel, setLayers }: Props) {
  const wrapRef = useRef<HTMLDivElement | null>(null);
  const [viewState, setViewState] = useState<ViewState>(() => ({ longitude: MONEY_CREEK_CENTER[0], latitude: MONEY_CREEK_CENTER[1], zoom: 14.8, pitch: 0, bearing: 0 }));
  const [draft, setDraft] = useState<MoneyPosition[]>([]);
  const [draftHistory, setDraftHistory] = useState<MoneyPosition[][]>([]);
  const [selectedDraftVertex, setSelectedDraftVertex] = useState<number | null>(null);
  const [editDraft, setEditDraft] = useState<MoneyPosition[]>([]);
  const [editHistory, setEditHistory] = useState<MoneyPosition[][]>([]);
  const [selectedEditVertex, setSelectedEditVertex] = useState<number | null>(null);
  const [vertexMessage, setVertexMessage] = useState('');
  const [layersOpen, setLayersOpen] = useState(!mobile);
  const [lineMapState, setLineMapState] = useState<LineMapLoadState>({ status: 'idle', data: null });
  const lineMapStartedRef = useRef(false);
  const creating = mode === 'create-area' || mode === 'create-boulder';
  const editing = mode === 'edit-area';
  const lastFocusBBoxKeyRef = useRef<string | null>(null);

  const focusBBox = useMemo(() => bbox(area), [area]);
  const focusBBoxKey = `${focusBBox[0]},${focusBBox[1]},${focusBBox[2]},${focusBBox[3]}`;

  useEffect(() => {
    if (editing || lastFocusBBoxKeyRef.current === focusBBoxKey) return;
    lastFocusBBoxKeyRef.current = focusBBoxKey;
    setViewState(viewForBBox(focusBBox));
  }, [focusBBox, focusBBoxKey, editing]);
  useEffect(() => { if (creating) { setDraft([]); setDraftHistory([]); setSelectedDraftVertex(null); setVertexMessage('Tap the map to add vertices.'); } }, [creating]);
  useEffect(() => { if (editing) { setEditDraft(openPolygonPoints(geometryPoints(area.feature.geojson))); setEditHistory([]); setSelectedEditVertex(null); setVertexMessage('Select a white vertex to delete it, or tap green midpoint dots to add vertices.'); } }, [editing, area.feature.geojson]);
  useEffect(() => {
    if (layers.base !== 'stylized' || lineMapStartedRef.current) return;
    lineMapStartedRef.current = true;
    let loadStarted = false;
    let timeoutId: number | undefined;
    let idleId: number | undefined;
    const idleWindow = window as IdleWindow;
    const startLoad = () => {
      loadStarted = true;
      setLineMapState({ status: 'loading', data: null });
      loadMoneyCreekLineMap()
        .then(data => setLineMapState({ status: 'ready', data }))
        .catch(error => setLineMapState({ status: 'error', data: null, error }));
    };

    if (idleWindow.requestIdleCallback) idleId = idleWindow.requestIdleCallback(startLoad, { timeout: 1200 });
    else timeoutId = window.setTimeout(startLoad, 0);

    return () => {
      if (loadStarted) return;
      lineMapStartedRef.current = false;
      if (idleId !== undefined && idleWindow.cancelIdleCallback) idleWindow.cancelIdleCallback(idleId);
      if (timeoutId !== undefined) window.clearTimeout(timeoutId);
    };
  }, [layers.base]);

  const focusedAreaIds = useMemo(() => new Set(flattenAreas(area).map(node => node.feature.id)), [area]);
  const focusedBoulderIds = useMemo(() => new Set(flattenBoulders(area).map(node => node.feature.id)), [area]);
  const isFocusedArea = useCallback((id: string) => focusedAreaIds.has(id), [focusedAreaIds]);
  const isFocusedBoulder = useCallback((id: string) => focusedBoulderIds.has(id), [focusedBoulderIds]);

  const allAreaData = useMemo<AreaDatum[]>(() => flattenAreas(root)
    .filter(node => node.feature.id !== root.feature.id)
    .map(node => ({ node, polygon: geometryPoints(node.feature.geojson), center: featureCenter(node.feature) }))
    .filter(d => usablePolygon(d.polygon)), [root]);
  const areaData = useMemo(() => allAreaData.filter(d => layers.areas[d.node.feature.id] !== false), [allAreaData, layers.areas]);

  const allBoulderData = useMemo<BoulderDatum[]>(() => flattenBoulders(root).flatMap((node, index) => {
    const dev = String(node.feature.status);
    const meta = DEV.meta[(dev in DEV.meta ? dev : 'scouted') as keyof typeof DEV.meta];
    const center = featureCenter(node.feature);
    if (!isLonLatPosition(center)) return [];
    return [{ node, polygon: boulderFootprint(geometryPoints(node.feature.geojson), center, node.feature.id), center, color: hexToRgba(meta.bg, 86), line: hexToRgba(meta.c, 235), rank: stableHash(node.feature.id || String(index)) % 5 }];
  }), [root]);
  const boulderData = useMemo(() => allBoulderData.filter(d => layers.dev[String(d.node.feature.status)] !== false), [allBoulderData, layers.dev]);

  const trailData = useMemo<TrailDatum[]>(() => layers.trails ? trails.map(node => ({ node, path: geometryPoints(node.feature.geojson) })).filter(d => usablePath(d.path)) : [], [layers.trails, trails]);
  const trailFineData = useMemo<TrailFineDatum[]>(() => trailData.map(d => ({ ...d, finePath: densifyPath(d.path, 12) })), [trailData]);
  const areaLabelData = useMemo(() => areaData.filter(d => isFocusedArea(d.node.feature.id) || viewState.zoom < DETAIL_ZOOM), [areaData, isFocusedArea, viewState.zoom]);
  const boulderLabelData = useMemo(() => boulderData.filter(d => d.node.feature.id === selectedBoulderId || isFocusedBoulder(d.node.feature.id) && (viewState.zoom >= FINE_DETAIL_ZOOM || viewState.zoom >= DETAIL_ZOOM && d.rank === 0)), [boulderData, isFocusedBoulder, selectedBoulderId, viewState.zoom]);
  const lineMapData = lineMapState.status === 'ready' ? lineMapState.data : null;
  const lineLabelData = useMemo(() => (lineMapData?.labels ?? []).filter(label => lineLabelVisible(label, viewState.zoom, layers)), [layers, lineMapData, viewState.zoom]);
  const roadData = useMemo(() => lineMapData?.byCategory.road.filter(path => path.visible !== false) ?? [], [lineMapData]);
  const waterData = useMemo(() => lineMapData ? [...lineMapData.byCategory.creek, ...lineMapData.byCategory.reservoir] : [], [lineMapData]);
  const contourData = useMemo(() => visibleContours(lineMapData?.byCategory.contour ?? [], viewState.zoom), [lineMapData, viewState.zoom]);
  const indexContourData = useMemo(() => visibleContours(lineMapData?.byCategory['index-contour'] ?? [], viewState.zoom), [lineMapData, viewState.zoom]);

  const addDraftVertex = useCallback((position: MoneyPosition) => {
    setDraft(points => { setDraftHistory(history => [...history, points].slice(-50)); return [...points, position]; });
    setSelectedDraftVertex(null);
    setVertexMessage('Vertex added.');
  }, []);
  const insertDraftVertex = useCallback((index: number, position: MoneyPosition) => {
    setDraft(points => { setDraftHistory(history => [...history, points].slice(-50)); return insertPolygonVertexAfter(points, index, position); });
    setSelectedDraftVertex(index + 1);
    setVertexMessage('Vertex inserted at midpoint.');
  }, []);
  const moveDraftVertex = useCallback((index: number, position: MoneyPosition) => {
    setDraft(points => replacePolygonVertex(points, index, position));
    setSelectedDraftVertex(index);
  }, []);
  const deleteDraftVertex = useCallback(() => {
    if (selectedDraftVertex == null) { setVertexMessage('Select a draft vertex first.'); return; }
    setDraft(points => {
      if (selectedDraftVertex < 0 || selectedDraftVertex >= points.length) return points;
      setDraftHistory(history => [...history, points].slice(-50));
      return deletePolygonVertex(points, selectedDraftVertex, 0);
    });
    setSelectedDraftVertex(null);
    setVertexMessage('Draft vertex deleted.');
  }, [selectedDraftVertex]);
  const undoDraft = useCallback(() => {
    setDraftHistory(history => {
      const previous = history[history.length - 1];
      if (!previous) return history;
      setDraft(previous);
      setSelectedDraftVertex(null);
      setVertexMessage('Undid last draft edit.');
      return history.slice(0, -1);
    });
  }, []);
  const clearDraft = useCallback(() => {
    setDraft(points => { if (points.length) setDraftHistory(history => [...history, points].slice(-50)); return []; });
    setSelectedDraftVertex(null);
    setVertexMessage('Draft cleared.');
  }, []);
  const insertEditVertex = useCallback((index: number, position: MoneyPosition) => {
    setEditDraft(points => { setEditHistory(history => [...history, points].slice(-50)); return insertPolygonVertexAfter(points, index, position); });
    setSelectedEditVertex(index + 1);
    setVertexMessage('Vertex inserted at midpoint.');
  }, []);
  const moveEditVertex = useCallback((index: number, position: MoneyPosition) => {
    setEditDraft(points => replacePolygonVertex(points, index, position));
    setSelectedEditVertex(index);
  }, []);
  const deleteEditVertex = useCallback(() => {
    if (selectedEditVertex == null) { setVertexMessage('Select a vertex first.'); return; }
    if (editDraft.length <= 3) { setVertexMessage('Areas need at least 3 vertices.'); return; }
    setEditDraft(points => {
      if (selectedEditVertex < 0 || selectedEditVertex >= points.length || points.length <= 3) return points;
      setEditHistory(history => [...history, points].slice(-50));
      return deletePolygonVertex(points, selectedEditVertex);
    });
    setSelectedEditVertex(null);
    setVertexMessage('Vertex deleted.');
  }, [editDraft.length, selectedEditVertex]);
  const undoEdit = useCallback(() => {
    setEditHistory(history => {
      const previous = history[history.length - 1];
      if (!previous) return history;
      setEditDraft(previous);
      setSelectedEditVertex(null);
      setVertexMessage('Undid last reshape edit.');
      return history.slice(0, -1);
    });
  }, []);

  const deckLayers = useMemo(() => {
    const draftGeo = draft.length > 0 ? polygonGeoJSON(draft) : null;
    const draftPolygon = draft.length >= 3 ? closedPolygonPoints(draft) : draft;
    const draftHandles = creating ? draft.map((position, index) => ({ position, index })) : [];
    const draftEdgeHandles = creating && draft.length >= 2 ? draft.map((position, index) => {
      const lastOpenEdge = index === draft.length - 1 && draft.length < 3;
      if (lastOpenEdge) return null;
      return { position: midpoint(position, draft[(index + 1) % draft.length]), index };
    }).filter((handle): handle is { position: MoneyPosition; index: number } => Boolean(handle)) : [];
    const editPolygon = editing ? closedPolygonPoints(editDraft) : [];
    const editHandles = editing ? editDraft.map((position, index) => ({ position, index })) : [];
    const editEdgeHandles = editing ? editDraft.map((position, index) => ({ position: midpoint(position, editDraft[(index + 1) % editDraft.length]), index })) : [];
    const stylized = layers.base === 'stylized' ? [
      layers.contours && new PathLayer<LineMapPath>({
        id: 'money-reference-line-map-contours', data: contourData, pickable: false, getPath: contourPath, getColor: d => contourColor(d, viewState.zoom), getWidth: contourWidth, widthMinPixels: 0.28, widthMaxPixels: 0.9, jointRounded: true, capRounded: true,
      }),
      layers.contours && new PathLayer<LineMapPath>({
        id: 'money-reference-line-map-index-contours', data: indexContourData, pickable: false, getPath: contourPath, getColor: d => indexContourColor(d, viewState.zoom), getWidth: d => contourWidth(d) + 0.2, widthMinPixels: 0.48, widthMaxPixels: 1.25, jointRounded: true, capRounded: true,
      }),
      layers.roads && new PathLayer<LineMapPath>({
        id: 'money-reference-line-map-roads-halo', data: roadData, pickable: false, getPath: d => d.path, getColor: [31, 24, 18, 70], getWidth: d => roadWidth(d) + 2.1, widthMinPixels: 1.8, widthMaxPixels: 6.2, jointRounded: true, capRounded: true,
      }),
      layers.roads && new PathLayer<LineMapPath>({
        id: 'money-reference-line-map-roads', data: roadData, pickable: false, getPath: d => d.path, getColor: roadColor, getWidth: roadWidth, widthMinPixels: 0.95, widthMaxPixels: 3.35, jointRounded: true, capRounded: true, ...ROAD_DASH_PROPS,
      }),
      layers.water && new PathLayer<LineMapPath>({
        id: 'money-reference-line-map-water-halo', data: waterData, pickable: false, getPath: d => d.path, getColor: [13, 26, 29, 98], getWidth: d => waterWidth(d) + 2.15, widthMinPixels: 1.7, widthMaxPixels: 7, jointRounded: true, capRounded: true,
      }),
      layers.water && new PathLayer<LineMapPath>({
        id: 'money-reference-line-map-water', data: waterData, pickable: false, getPath: d => d.path, getColor: waterColor, getWidth: waterWidth, widthMinPixels: 0.95, widthMaxPixels: 4.2, jointRounded: true, capRounded: true,
      }),
      new TextLayer<LineMapLabel>({
        id: 'money-reference-line-map-labels', data: lineLabelData, pickable: false, getPosition: d => d.position, getText: d => d.name, getSize: d => lineLabelSize(d, viewState.zoom), getColor: lineLabelColor, getBackgroundColor: lineLabelBackground, background: true, backgroundPadding: [5, 2], fontWeight: 700, getPixelOffset: lineLabelOffset, getTextAnchor: 'middle', getAlignmentBaseline: 'center', billboard: true,
      }),
    ].filter(Boolean) : [];

    return [
      ...stylized,
      new PolygonLayer<AreaDatum>({
        id: 'money-reference-areas', data: areaData, pickable: !creating && !editing, stroked: true, filled: true, lineWidthMinPixels: 0.65, lineWidthMaxPixels: 2.4,
        getPolygon: d => editing && d.node.feature.id === area.feature.id ? editPolygon : d.polygon, getFillColor: d => editing && d.node.feature.id === area.feature.id ? [174, 185, 116, 82] : areaFill(d, layers.base, isFocusedArea(d.node.feature.id)), getLineColor: d => editing && d.node.feature.id === area.feature.id ? [255, 245, 220, 255] : isFocusedArea(d.node.feature.id) ? [174, 185, 116, 205] : [174, 185, 116, 82], getLineWidth: d => editing && d.node.feature.id === area.feature.id ? 2 : isFocusedArea(d.node.feature.id) ? 1.35 : 0.65,
        onClick: info => { if (info.object) onEnter(info.object.node.feature.id); return true; },
      }),
      new ScatterplotLayer<BoulderDatum>({
        id: 'money-reference-boulder-markers', data: boulderData, pickable: !creating && !editing, stroked: true, filled: true, radiusMinPixels: 4, radiusMaxPixels: 16, lineWidthMinPixels: 1, lineWidthMaxPixels: 3,
        getPosition: d => d.center, getRadius: d => d.node.feature.id === selectedBoulderId ? 3.5 : isFocusedBoulder(d.node.feature.id) ? 2.25 : 1.55, getFillColor: d => d.node.feature.id === selectedBoulderId ? withAlpha(d.line, 245) : isFocusedBoulder(d.node.feature.id) ? withAlpha(d.line, 210) : withAlpha(d.line, 118), getLineColor: d => d.node.feature.id === selectedBoulderId ? [255, 245, 220, 255] : isFocusedBoulder(d.node.feature.id) ? [28, 24, 18, 230] : [28, 24, 18, 120], getLineWidth: d => d.node.feature.id === selectedBoulderId ? 1.5 : 0.85,
        onClick: info => { if (info.object) onSelectBoulder(info.object.node.feature.id); return true; },
      }),
      new PathLayer<TrailDatum>({ id: 'money-reference-trails-halo', data: trailData, getPath: d => d.path, getColor: [17, 24, 20, layers.base === 'stylized' ? 92 : 130], getWidth: d => d.node.feature.id === selectedTrailId ? 4.2 : 3, widthMinPixels: 2, widthMaxPixels: 5.5, jointRounded: true, capRounded: true }),
      new PathLayer<TrailDatum>({
        id: 'money-reference-trails', data: trailData, pickable: !creating && !editing, getPath: d => d.path, getColor: d => d.node.feature.id === selectedTrailId ? [174, 185, 116, 255] : [185, 128, 80, 225], getWidth: d => d.node.feature.id === selectedTrailId ? 2.6 : 1.45, widthMinPixels: 1.2, widthMaxPixels: 3.8, jointRounded: true, capRounded: true,
        onClick: info => { if (info.object) onSelectTrail(info.object.node.feature.id); return true; },
      }),
      viewState.zoom >= DETAIL_ZOOM && new PathLayer<TrailFineDatum>({
        id: 'money-reference-trails-fine-detail', data: trailFineData, pickable: false, getPath: d => d.finePath, getColor: [238, 225, 211, 95], getWidth: 0.45, widthMinPixels: 0.6, widthMaxPixels: 1.5, jointRounded: true, capRounded: true,
      }),
      new TextLayer<AreaDatum>({ id: 'money-reference-area-labels', data: areaLabelData, getPosition: d => d.center, getText: d => d.node.feature.title, getSize: d => isFocusedArea(d.node.feature.id) && viewState.zoom >= DETAIL_ZOOM ? 12 : 10.5, getColor: d => isFocusedArea(d.node.feature.id) ? [238, 225, 211, 255] : [238, 225, 211, 155], getBackgroundColor: d => isFocusedArea(d.node.feature.id) ? [42, 36, 29, layers.base === 'stylized' ? 196 : 220] : [42, 36, 29, 128], background: true, backgroundPadding: [7, 4], fontWeight: 700, pickable: false }),
      new TextLayer<BoulderDatum>({ id: 'money-reference-boulder-labels', data: boulderLabelData, getPosition: d => d.center, getText: d => d.node.feature.title, getSize: d => d.node.feature.id === selectedBoulderId ? 11 : viewState.zoom >= FINE_DETAIL_ZOOM ? 9.5 : 8.5, getColor: d => d.line, getBackgroundColor: [42, 36, 29, 205], background: true, backgroundPadding: [5, 2], fontWeight: 700, pickable: false, getPixelOffset: [0, -12] }),
      draftGeo && new PolygonLayer({ id: 'money-reference-draft', data: [{ polygon: geometryPoints(draftGeo) }], getPolygon: (d: { polygon: MoneyPosition[] }) => d.polygon, getFillColor: [174, 185, 116, 70], getLineColor: [174, 185, 116, 255], getLineWidth: 1.8, lineWidthMinPixels: 1.2, lineWidthMaxPixels: 4 }),
      creating && draft.length >= 2 && new PathLayer({ id: 'money-reference-draft-outline', data: [{ path: draftPolygon }], getPath: (d: { path: MoneyPosition[] }) => d.path, getColor: [174, 185, 116, 255], getWidth: 2, widthMinPixels: 2, widthMaxPixels: 5, jointRounded: true, capRounded: true }),
      creating && draftEdgeHandles.length > 0 && new ScatterplotLayer({ id: 'money-reference-draft-edge-handles', data: draftEdgeHandles, pickable: true, getPosition: (d: { position: MoneyPosition }) => d.position, getFillColor: [174, 185, 116, 150], getLineColor: [255, 245, 220, 235], getRadius: 1, radiusMinPixels: 5, radiusMaxPixels: 9, stroked: true, onClick: info => { if (info.object) insertDraftVertex(info.object.index, info.object.position); return true; } }),
      draft.length > 0 && new ScatterplotLayer({ id: 'money-reference-draft-points', data: draftHandles, pickable: creating, getPosition: (d: { position: MoneyPosition }) => d.position, getFillColor: (d: { index: number }) => d.index === selectedDraftVertex ? [255, 210, 115, 255] : [255, 255, 255, 245], getLineColor: (d: { index: number }) => d.index === selectedDraftVertex ? [42, 36, 29, 255] : [174, 185, 116, 255], getRadius: (d: { index: number }) => d.index === selectedDraftVertex ? 1.65 : 1.25, radiusMinPixels: 6, radiusMaxPixels: 14, lineWidthMinPixels: 1.4, stroked: true, onClick: info => { if (info.object) { setSelectedDraftVertex(info.object.index); setVertexMessage(`Draft vertex ${info.object.index + 1} selected.`); } return true; }, onDrag: info => { const coordinate = info.coordinate; if (info.object && coordinate && coordinate.length >= 2) moveDraftVertex(info.object.index, [coordinate[0], coordinate[1]]); return true; } }),
      editing && new PathLayer({ id: 'money-reference-edit-outline', data: [{ path: editPolygon }], getPath: (d: { path: MoneyPosition[] }) => d.path, getColor: [255, 245, 220, 255], getWidth: 2.2, widthMinPixels: 2, widthMaxPixels: 5, jointRounded: true, capRounded: true }),
      editing && new ScatterplotLayer({ id: 'money-reference-edit-edge-handles', data: editEdgeHandles, pickable: true, getPosition: (d: { position: MoneyPosition }) => d.position, getFillColor: [174, 185, 116, 150], getLineColor: [255, 245, 220, 235], getRadius: 1, radiusMinPixels: 5, radiusMaxPixels: 9, stroked: true, onClick: info => { if (info.object) insertEditVertex(info.object.index, info.object.position); return true; } }),
      editing && new ScatterplotLayer({ id: 'money-reference-edit-handles', data: editHandles, pickable: true, getPosition: (d: { position: MoneyPosition }) => d.position, getFillColor: (d: { index: number }) => d.index === selectedEditVertex ? [255, 210, 115, 255] : [255, 245, 220, 255], getLineColor: [42, 36, 29, 255], getRadius: (d: { index: number }) => d.index === selectedEditVertex ? 1.7 : 1.35, radiusMinPixels: 7, radiusMaxPixels: 14, lineWidthMinPixels: 1.5, stroked: true, onClick: info => { if (info.object) { setSelectedEditVertex(info.object.index); setVertexMessage(`Vertex ${info.object.index + 1} selected.`); } return true; }, onDrag: info => { const coordinate = info.coordinate; if (info.object && coordinate && coordinate.length >= 2) moveEditVertex(info.object.index, [coordinate[0], coordinate[1]]); return true; } }),
    ].filter(Boolean);
  }, [area.feature.id, areaData, areaLabelData, boulderData, boulderLabelData, contourData, creating, draft, editDraft, editing, indexContourData, insertDraftVertex, insertEditVertex, isFocusedArea, isFocusedBoulder, layers.base, layers.contours, layers.roads, layers.water, lineLabelData, moveDraftVertex, moveEditVertex, onEnter, onSelectBoulder, onSelectTrail, roadData, selectedBoulderId, selectedDraftVertex, selectedEditVertex, selectedTrailId, trailData, trailFineData, viewState.zoom, waterData]);

  const finish = () => { if (draft.length >= 3) onCreateDone(draft); };
  const saveEdit = () => { if (isValidAreaEditRing(editDraft)) onEditSave(editDraft); };
  const focus = () => setViewState(viewForBBox(focusBBox));
  const zoomBy = (delta: number) => setViewState(current => ({ ...current, zoom: clamp(current.zoom + delta, MIN_ZOOM, MAX_ZOOM) }));

  return <div ref={wrapRef} style={{ position: 'absolute', inset: 0, overflow: 'hidden', background: layers.base === 'stylized' ? 'transparent' : T.map.bg2 }}>
    <DeckGL controller={{ dragPan: !creating && !editing, scrollZoom: true, doubleClickZoom: !creating && !editing, touchZoom: true, keyboard: true }} layers={deckLayers} viewState={viewState} onViewStateChange={(event: { viewState: unknown }) => {
      const next = event.viewState as Partial<ViewState>;
      setViewState(current => ({ longitude: next.longitude ?? current.longitude, latitude: next.latitude ?? current.latitude, zoom: clamp(next.zoom ?? current.zoom, MIN_ZOOM, MAX_ZOOM), pitch: next.pitch ?? current.pitch, bearing: next.bearing ?? current.bearing }));
    }} onClick={(info: { coordinate?: number[]; picked?: boolean }) => {
      if (creating && info.coordinate && info.coordinate.length >= 2 && !info.picked) addDraftVertex([info.coordinate[0], info.coordinate[1]]);
      if (editing && info.coordinate && info.coordinate.length >= 2 && !info.picked) insertEditVertex(nearestEdgeIndex(editDraft, [info.coordinate[0], info.coordinate[1]]), [info.coordinate[0], info.coordinate[1]]);
      if (!creating && !editing && !info.picked) { onSelectBoulder(null); onSelectTrail(null); }
    }} getCursor={() => creating ? 'crosshair' : editing ? 'pointer' : 'grab'}>
      <Map mapStyle={mapStyle(layers.base) as never} reuseMaps minZoom={MIN_ZOOM} maxZoom={MAX_ZOOM} onLoad={() => window.requestAnimationFrame(() => window.dispatchEvent(new Event('resize')))} />
    </DeckGL>
    {layers.base === 'slope' && <div className="pointer-events-none absolute inset-0" style={{ background: 'linear-gradient(135deg, rgba(62,122,78,0.23), rgba(201,184,74,0.18) 45%, rgba(200,87,47,0.20))' }} />}

    {layers.base === 'stylized' && lineMapState.status !== 'ready' && <div style={{ position: 'absolute', left: 14, bottom: 14, zIndex: 10, background: 'rgba(42, 36, 29, 0.82)', border: `1px solid ${T.line2}`, borderRadius: 10, padding: '7px 10px', color: lineMapState.status === 'error' ? '#f0b4a8' : T.mut, fontSize: 11.5, boxShadow: T.shadow, pointerEvents: 'none' }}>{lineMapState.status === 'error' ? 'Static line map unavailable' : 'Loading static line map…'}</div>}
    {!creating && !editing && <div style={{ position: 'absolute', left: 14, top: 14, display: 'flex', flexDirection: 'column', gap: 8, zIndex: 10 }}>
      <button onClick={focus} style={mapButton}><LocateFixed size={16} />Focus</button>
      <div style={{ display: 'flex', background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 10, overflow: 'hidden', boxShadow: T.shadow }}>
        <button onClick={() => zoomBy(1)} style={smallButton}>+</button><button onClick={() => zoomBy(-1)} style={smallButton}>−</button>
      </div>
    </div>}
    {!creating && !editing && (layersOpen ? <LayersPanel layers={layers} setLayers={setLayers} onClose={() => setLayersOpen(false)} root={root} /> : <button onClick={() => setLayersOpen(true)} style={{ position: 'absolute', top: 14, right: 14, zIndex: 10, display: 'flex', gap: 7, background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 10, padding: '9px 13px', color: T.ink, fontWeight: 700, boxShadow: T.shadow }}><Layers size={16} />Layers</button>)}
    {creating && <div style={{ position: 'absolute', bottom: mobile ? 22 : 20, left: '50%', transform: 'translateX(-50%)', display: 'flex', gap: 8, flexWrap: 'wrap', justifyContent: 'center', background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 14, padding: 8, boxShadow: T.shadow, zIndex: 10, maxWidth: mobile ? '92vw' : 720 }}><button onClick={onCreateCancel} style={ctrl(false)}><X size={16} />Cancel</button><button onClick={undoDraft} disabled={!draftHistory.length} style={ctrl(false, !draftHistory.length)}><RotateCcw size={16} />Undo last</button><button onClick={clearDraft} disabled={!draft.length} style={ctrl(false, !draft.length)}><X size={16} />Clear</button><button onClick={deleteDraftVertex} disabled={selectedDraftVertex == null} style={ctrl(false, selectedDraftVertex == null)}><Trash2 size={16} />Delete vertex</button><button disabled={draft.length < 3} onClick={finish} style={ctrl(true, draft.length < 3)}><Plus size={16} />Done</button></div>}
    {creating && <div style={{ position: 'absolute', left: 14, bottom: mobile ? 126 : 84, background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 10, padding: '8px 11px', color: T.mut, fontSize: 12, zIndex: 10, maxWidth: mobile ? '88vw' : 480 }}>Tap map to add vertices · drag/select white vertices · tap green midpoints to insert · {draft.length} point{draft.length === 1 ? '' : 's'} · {draft.length < 3 ? `${3 - draft.length} more needed` : 'ready'}{vertexMessage ? ` · ${vertexMessage}` : ''}</div>}
    {editing && <div style={{ position: 'absolute', bottom: mobile ? 22 : 20, left: '50%', transform: 'translateX(-50%)', display: 'flex', gap: 8, flexWrap: 'wrap', justifyContent: 'center', background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 14, padding: 8, boxShadow: T.shadow, zIndex: 10, maxWidth: mobile ? '92vw' : 680 }}><button onClick={onEditCancel} style={ctrl(false)}><X size={16} />Cancel</button><button onClick={undoEdit} disabled={!editHistory.length} style={ctrl(false, !editHistory.length)}><RotateCcw size={16} />Undo last</button><button onClick={deleteEditVertex} disabled={selectedEditVertex == null || editDraft.length <= 3} style={ctrl(false, selectedEditVertex == null || editDraft.length <= 3)}><Trash2 size={16} />Delete vertex</button><button disabled={!isValidAreaEditRing(editDraft)} onClick={saveEdit} style={ctrl(true, !isValidAreaEditRing(editDraft))}><Save size={16} />Save</button></div>}
    {editing && <div style={{ position: 'absolute', left: 14, bottom: mobile ? 126 : 84, background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 10, padding: '8px 11px', color: T.mut, fontSize: 12, zIndex: 10, maxWidth: mobile ? '88vw' : 500 }}>Drag/select white vertices · tap green midpoint handles to add · tap map to add at nearest edge · {editDraft.length} vertices{editDraft.length <= 3 ? ' · minimum reached' : ''}{vertexMessage ? ` · ${vertexMessage}` : ''}</div>}
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

function smoothPath(path: MoneyPosition[]): MoneyPosition[] {
  if (path.length < 4) return densifyPath(path, 6);
  const out: MoneyPosition[] = [path[0]];
  for (let i = 0; i < path.length - 1; i += 1) {
    const p0 = path[Math.max(0, i - 1)], p1 = path[i], p2 = path[i + 1], p3 = path[Math.min(path.length - 1, i + 2)];
    for (let step = 1; step <= 5; step += 1) {
      const t = step / 5;
      out.push([catmullRom(p0[0], p1[0], p2[0], p3[0], t), catmullRom(p0[1], p1[1], p2[1], p3[1], t)]);
    }
  }
  return out;
}

function closePolygon(points: MoneyPosition[]): MoneyPosition[] {
  if (!points.length) return points;
  const first = points[0], last = points[points.length - 1];
  return first[0] === last[0] && first[1] === last[1] ? points : [...points, first];
}

function midpoint(a: MoneyPosition, b: MoneyPosition): MoneyPosition {
  return [(a[0] + b[0]) / 2, (a[1] + b[1]) / 2];
}

function nearestEdgeIndex(points: MoneyPosition[], point: MoneyPosition): number {
  if (points.length < 2) return Math.max(0, points.length - 1);
  let best = 0;
  let bestDistance = Number.POSITIVE_INFINITY;
  points.forEach((a, index) => {
    const b = points[(index + 1) % points.length];
    const distance = distanceToSegment(point, a, b);
    if (distance < bestDistance) {
      best = index;
      bestDistance = distance;
    }
  });
  return best;
}

function distanceToSegment(point: MoneyPosition, a: MoneyPosition, b: MoneyPosition): number {
  const dx = b[0] - a[0];
  const dy = b[1] - a[1];
  const lengthSq = dx * dx + dy * dy;
  if (!lengthSq) return Math.hypot(point[0] - a[0], point[1] - a[1]);
  const t = clamp(((point[0] - a[0]) * dx + (point[1] - a[1]) * dy) / lengthSq, 0, 1);
  return Math.hypot(point[0] - (a[0] + dx * t), point[1] - (a[1] + dy * t));
}

function lonLatToMeters(origin: MoneyPosition, point: MoneyPosition): [number, number] {
  const cosLat = Math.max(0.1, Math.cos(origin[1] * Math.PI / 180));
  return [(point[0] - origin[0]) * 111320 * cosLat, (point[1] - origin[1]) * 111320];
}

function offsetMeters(origin: MoneyPosition, east: number, north: number): MoneyPosition {
  const cosLat = Math.max(0.1, Math.cos(origin[1] * Math.PI / 180));
  return [origin[0] + east / (111320 * cosLat), origin[1] + north / 111320];
}

const smoothedContourPathCache = new WeakMap<LineMapPath, MoneyPosition[]>();
function contourPath(d: LineMapPath): MoneyPosition[] {
  if (d.path.length > 24) return d.path;
  const cached = smoothedContourPathCache.get(d);
  if (cached) return cached;
  const smoothed = smoothPath(d.path);
  smoothedContourPathCache.set(d, smoothed);
  return smoothed;
}
function visibleContours(paths: LineMapPath[], zoom: number): LineMapPath[] {
  if (zoom >= DETAIL_ZOOM) return paths;
  return paths.filter((path, index) => path.category === 'index-contour' || (path.elevationM ?? index) % 30 === 0);
}
function contourColor(d: LineMapPath, zoom: number): Rgba {
  const high = (d.elevationM ?? 0) >= 1200;
  return high ? [214, 187, 154, zoom >= DETAIL_ZOOM ? 38 : 27] : [152, 122, 91, zoom >= DETAIL_ZOOM ? 32 : 22];
}
function indexContourColor(d: LineMapPath, zoom: number): Rgba {
  const high = (d.elevationM ?? 0) >= 1200;
  return high ? [235, 217, 196, zoom >= DETAIL_ZOOM ? 76 : 55] : [194, 158, 119, zoom >= DETAIL_ZOOM ? 68 : 48];
}
function contourWidth(d: LineMapPath): number { return d.elevationM != null && d.elevationM >= 1200 ? 0.68 : 0.5; }
function roadColor(d: LineMapPath): Rgba { return d.importance === 'minor' ? [171, 139, 102, 150] : [205, 174, 132, 210]; }
function roadWidth(d: LineMapPath): number { return d.importance === 'minor' ? 1.15 : d.surface === 'paved' ? 2.2 : 1.75; }
function roadDash(d: LineMapPath): [number, number] { return d.importance === 'minor' || d.sourceKind === 'trail' ? [5, 4] : [0, 0]; }
const ROAD_DASH_PROPS = { extensions: [DASHED_PATH_STYLE], getDashArray: roadDash, dashJustified: true };
function waterColor(d: LineMapPath): Rgba { return d.category === 'reservoir' ? [139, 168, 187, 205] : [107, 157, 153, d.seasonal ? 150 : 220]; }
function waterWidth(d: LineMapPath): number { return d.category === 'reservoir' ? 2.45 : d.importance === 'major' ? 2.05 : 1.35; }
function lineLabelVisible(label: LineMapLabel, zoom: number, layers: LayerState): boolean {
  if (zoom < label.minZoom) return false;
  if (label.priority === 'low' && zoom < 14.25) return false;
  if (label.kind === 'road') return layers.roads;
  if (label.kind === 'creek' || label.kind === 'reservoir') return layers.water;
  return zoom >= DETAIL_ZOOM;
}
function lineLabelSize(d: LineMapLabel, zoom: number): number { return d.priority === 'high' ? 12.5 : d.kind === 'context' ? zoom >= DETAIL_ZOOM ? 11 : 10 : d.kind === 'road' ? 10 : 11.25; }
function lineLabelColor(d: LineMapLabel): Rgba { return d.kind === 'road' ? [218, 195, 166, 218] : d.kind === 'context' ? [238, 225, 211, 205] : d.kind === 'reservoir' ? [178, 207, 224, 230] : [157, 207, 199, 230]; }
function lineLabelBackground(d: LineMapLabel): Rgba { return d.kind === 'context' ? [42, 36, 29, 118] : d.priority === 'high' ? [27, 33, 31, 118] : [36, 30, 24, 86]; }
function lineLabelOffset(d: LineMapLabel): [number, number] { return d.kind === 'road' ? [0, -10] : d.kind === 'creek' ? [0, 12] : [0, 0]; }
function catmullRom(p0: number, p1: number, p2: number, p3: number, t: number): number { const t2 = t * t, t3 = t2 * t; return 0.5 * ((2 * p1) + (-p0 + p2) * t + (2 * p0 - 5 * p1 + 4 * p2 - p3) * t2 + (-p0 + 3 * p1 - 3 * p2 + p3) * t3); }
function areaFill(_d: AreaDatum, base: string, focused: boolean): Rgba { return focused ? base === 'stylized' ? [174, 185, 116, 30] : [174, 185, 116, 44] : base === 'stylized' ? [174, 185, 116, 8] : [174, 185, 116, 14]; }
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
function LayersPanel({ layers, setLayers, onClose, root }: { layers: LayerState; setLayers: (l: LayerState) => void; onClose: () => void; root: MoneyCragNode }) { const bases = ['stylized', 'topo', 'satellite', 'slope']; const areaRows = useMemo(() => flattenAreas(root).filter(a => a.feature.id !== root.feature.id), [root]); return <div style={{ position: 'absolute', top: 14, right: 14, zIndex: 10, width: 224, background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 12, boxShadow: T.shadow, overflow: 'hidden' }}><div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '11px 13px', borderBottom: `1px solid ${T.line}`, color: T.ink }}><Layers size={16} /><b style={{ fontSize: 13.5 }}>Layers</b><button onClick={onClose} style={{ marginLeft: 'auto', border: 'none', background: 'transparent', color: T.mut, cursor: 'pointer' }}>×</button></div><div style={{ padding: '11px 13px', maxHeight: 380, overflowY: 'auto' }}><Label>Base map</Label>{bases.map(base => <Row key={base} onClick={() => setLayers({ ...layers, base })} on={layers.base === base} label={base === 'satellite' ? 'Satellite' : base} />)}<Label>Areas</Label>{areaRows.map(c => <Row key={c.feature.id} onClick={() => setLayers({ ...layers, areas: { ...layers.areas, [c.feature.id]: layers.areas[c.feature.id] === false } })} on={layers.areas[c.feature.id] !== false} label={c.feature.title} />)}<Label>Development</Label>{DEV.order.map(k => <Row key={k} onClick={() => setLayers({ ...layers, dev: { ...layers.dev, [k]: layers.dev[k] === false } })} on={layers.dev[k] !== false} label={DEV.meta[k].short} color={DEV.meta[k].c} />)}<div style={{ borderTop: `1px solid ${T.line}`, margin: '10px 0 8px' }} /><Label>Line map</Label>{(['roads', 'water', 'contours', 'trails'] as const).map(k => <Row key={k} onClick={() => setLayers({ ...layers, [k]: !layers[k] })} on={layers[k]} label={k === 'contours' ? 'contours' : k} />)}</div></div>; }
function Label({ children }: { children: React.ReactNode }) { return <div style={{ fontFamily: T.mono, fontSize: 10, letterSpacing: 0.6, color: T.faint, textTransform: 'uppercase', margin: '14px 0 7px' }}>{children}</div>; }
function Row({ on, label, onClick, color }: { on: boolean; label: string; onClick: () => void; color?: string }) { return <div onClick={onClick} style={{ display: 'flex', alignItems: 'center', gap: 9, padding: '5px 0', cursor: 'pointer' }}><span style={{ width: 9, height: 9, borderRadius: color ? '50%' : 2, background: color ?? T.accent, transform: color ? 'none' : 'rotate(45deg)' }} /><span style={{ flex: 1, fontSize: 12.5, color: on ? T.ink : T.mut, textTransform: label.length < 9 ? 'capitalize' : undefined }}>{label}</span><span style={{ width: 28, height: 16, borderRadius: 9, background: on ? (color ?? T.accent) : T.line2, position: 'relative' }}><span style={{ position: 'absolute', top: 2, left: on ? 14 : 2, width: 12, height: 12, borderRadius: '50%', background: '#fff' }} /></span></div>; }
