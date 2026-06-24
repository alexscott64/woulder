import { useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Loader2 } from 'lucide-react';
import { AuthProvider, useAuth } from '../../contexts/AuthContext';
import { moneyApi } from '../../services/money';
import { MoneyFeature, MoneyFeatureFilters, MoneyFeatureType, MoneyPosition } from '../../types/money';
import { buildGeoJSON, featureTypeLabel, minimumPointCount } from './geometry';
import { FieldLens, FieldPageKey, FieldPages } from './FieldPages';
import { LoginScreen } from './LoginScreen';

const PROJECT_SLUG = 'money-creek';
const EMPTY_FEATURES: MoneyFeature[] = [];

function featurePassesFilters(feature: MoneyFeature, filters: MoneyFeatureFilters) {
  if (filters.type && filters.type !== 'all' && feature.feature_type !== filters.type) return false;
  if (filters.status && filters.status !== 'all' && feature.status !== filters.status) return false;
  if (!filters.status || filters.status === 'all') {
    if (feature.status === 'archived') return false;
  }
  const search = filters.search?.toLowerCase().trim();
  if (search && !`${feature.title} ${feature.description ?? ''} ${feature.properties?.poi_label ?? ''}`.toLowerCase().includes(search)) return false;
  return true;
}

function pageForType(type: MoneyFeatureType): FieldPageKey {
  if (type === 'trail') return 'approach';
  if (type === 'topo' || type === 'drawing') return 'topos';
  return 'scratch';
}

function MoneyCreekWorkspace() {
  const { user, isAuthenticated, isBootstrapping, canWrite, logout } = useAuth();
  const queryClient = useQueryClient();
  const [activePage, setActivePage] = useState<FieldPageKey>('scratch');
  const [lens, setLens] = useState<FieldLens>('all');
  const [filters, setFilters] = useState<MoneyFeatureFilters>({ type: 'all', status: 'all', search: '' });
  const [selectedFeatureId, setSelectedFeatureId] = useState<string | null>(null);
  const [drawingType, setDrawingType] = useState<MoneyFeatureType | null>(null);
  const [draftPoints, setDraftPoints] = useState<MoneyPosition[]>([]);
  const [isOnline, setIsOnline] = useState(navigator.onLine);

  useEffect(() => {
    const online = () => setIsOnline(true);
    const offline = () => setIsOnline(false);
    window.addEventListener('online', online);
    window.addEventListener('offline', offline);
    return () => {
      window.removeEventListener('online', online);
      window.removeEventListener('offline', offline);
    };
  }, []);

  const projectQuery = useQuery({
    queryKey: ['money-project', PROJECT_SLUG],
    queryFn: () => moneyApi.getProject(PROJECT_SLUG),
    enabled: isAuthenticated,
    staleTime: 10 * 60 * 1000,
  });

  const projectId = projectQuery.data?.project.id;

  const snapshotQuery = useQuery({
    queryKey: ['money-snapshot', projectId],
    queryFn: () => moneyApi.getSnapshot(projectId!),
    enabled: Boolean(projectId),
    staleTime: 60 * 1000,
  });

  const detailQuery = useQuery({
    queryKey: ['money-feature', selectedFeatureId],
    queryFn: () => moneyApi.getFeature(selectedFeatureId!),
    enabled: Boolean(selectedFeatureId),
    staleTime: 30 * 1000,
  });

  const createFeature = useMutation({
    mutationFn: async ({ type, points, title, properties = {} }: { type: MoneyFeatureType; points: MoneyPosition[]; title: string; properties?: Record<string, unknown> }) => {
      if (!projectId) throw new Error('Project not loaded');
      return moneyApi.createFeature(projectId, {
        feature_type: type,
        title,
        description: null,
        status: 'draft',
        geojson: buildGeoJSON(type, points),
        style: {},
        properties: { source: type === 'drawing' ? 'crag-notebook-sketch' : 'crag-notebook-map', ...properties },
      });
    },
    onSuccess: feature => {
      void queryClient.invalidateQueries({ queryKey: ['money-snapshot', projectId] });
      setSelectedFeatureId(feature.id);
      setDrawingType(null);
      setDraftPoints([]);
      setActivePage(pageForType(feature.feature_type));
    },
  });

  const features = snapshotQuery.data?.features ?? EMPTY_FEATURES;
  const project = projectQuery.data?.project ?? snapshotQuery.data?.project;
  const selectedFeature = detailQuery.data?.feature ?? features.find(feature => feature.id === selectedFeatureId) ?? null;
  const noteCounts = snapshotQuery.data?.note_counts ?? {};
  const primaryUploads = snapshotQuery.data?.primary_uploads ?? {};
  const visibleMapFeatures = useMemo(() => features.filter(feature => featurePassesFilters(feature, filters)), [features, filters]);

  if (isBootstrapping) {
    return <div className="flex min-h-screen items-center justify-center bg-[#23201A] text-[#FBF7EA]"><Loader2 className="mr-3 h-6 w-6 animate-spin text-[#C6922E]" />Restoring Money Creek session</div>;
  }

  if (!isAuthenticated) return <LoginScreen />;

  const handleSelectFeature = (feature: MoneyFeature) => {
    setSelectedFeatureId(feature.id);
    setActivePage(pageForType(feature.feature_type));
  };

  const startDrawing = (type: MoneyFeatureType) => {
    setActivePage(pageForType(type));
    setDrawingType(type);
    setDraftPoints([]);
  };

  const finishDrawing = () => {
    if (!drawingType || draftPoints.length < minimumPointCount(drawingType)) return;
    const title = window.prompt(`Name this ${featureTypeLabel(drawingType).toLowerCase()} note`);
    if (!title?.trim()) return;
    const properties = drawingType === 'poi' ? { poi_category: 'general', poi_label: 'General note' } : undefined;
    createFeature.mutate({ type: drawingType, points: draftPoints, title: title.trim(), properties });
  };

  const refreshData = () => {
    void projectQuery.refetch();
    void snapshotQuery.refetch();
    if (selectedFeatureId) void detailQuery.refetch();
  };

  const onFeatureChanged = () => {
    if (selectedFeatureId) void detailQuery.refetch();
    void queryClient.invalidateQueries({ queryKey: ['money-snapshot', projectId] });
  };

  const onFeatureSaved = (feature: MoneyFeature) => {
    queryClient.setQueryData(['money-feature', feature.id], { notes: detailQuery.data?.notes ?? [], uploads: detailQuery.data?.uploads ?? [], feature });
    void queryClient.invalidateQueries({ queryKey: ['money-snapshot', projectId] });
  };

  const onFeatureArchived = () => {
    setSelectedFeatureId(null);
    void queryClient.invalidateQueries({ queryKey: ['money-snapshot', projectId] });
  };

  return (
    <FieldPages
      user={user}
      project={project}
      features={features}
      visibleMapFeatures={visibleMapFeatures}
      selectedFeature={selectedFeature}
      selectedFeatureId={selectedFeatureId}
      detailQuery={detailQuery}
      projectId={projectId}
      canWrite={canWrite}
      isFetching={snapshotQuery.isFetching || projectQuery.isFetching}
      isOnline={isOnline}
      loadError={Boolean(projectQuery.error || snapshotQuery.error)}
      savingFeature={createFeature.isPending}
      filters={filters}
      noteCounts={noteCounts}
      primaryUploads={primaryUploads}
      lens={lens}
      page={activePage}
      drawingType={drawingType}
      draftPoints={draftPoints}
      onLensChange={setLens}
      onPageChange={setActivePage}
      onFiltersChange={setFilters}
      onSelectFeature={handleSelectFeature}
      onClearSelection={() => setSelectedFeatureId(null)}
      onRefresh={refreshData}
      onLogout={() => void logout()}
      onChanged={onFeatureChanged}
      onFeatureSaved={onFeatureSaved}
      onFeatureArchived={onFeatureArchived}
      onStartDrawing={startDrawing}
      onAddDraftPoint={point => setDraftPoints(points => drawingType === 'poi' ? [point] : [...points, point])}
      onUndoDraftPoint={() => setDraftPoints(points => points.slice(0, -1))}
      onCancelDrawing={() => { setDrawingType(null); setDraftPoints([]); }}
      onFinishDrawing={finishDrawing}
    />
  );
}

export default function MoneyCreekApp() {
  return (
    <AuthProvider>
      <MoneyCreekWorkspace />
    </AuthProvider>
  );
}
