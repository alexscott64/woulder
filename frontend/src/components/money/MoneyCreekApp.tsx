import { useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { AlertCircle, Loader2, LogOut, Map, Mountain, RefreshCw, Users } from 'lucide-react';
import { AuthProvider, useAuth } from '../../contexts/AuthContext';
import { moneyApi } from '../../services/money';
import { MoneyFeature, MoneyFeatureFilters, MoneyFeatureType, MoneyPosition } from '../../types/money';
import { buildGeoJSON, featureTypeLabel, minimumPointCount } from './geometry';
import { DrawingToolbar } from './DrawingToolbar';
import { FeatureEditor } from './FeatureEditor';
import { FeatureList } from './FeatureList';
import { ImageUploader } from './ImageUploader';
import { LoginScreen } from './LoginScreen';
import { MobileBottomSheet } from './MobileBottomSheet';
import { MoneyMap } from './MoneyMap';
import { NotesPanel } from './NotesPanel';
import { OfflineBanner } from './OfflineBanner';

const PROJECT_SLUG = 'money-creek';
const EMPTY_FEATURES: MoneyFeature[] = [];

function MoneyCreekWorkspace() {
  const { user, isAuthenticated, isBootstrapping, canWrite, logout } = useAuth();
  const queryClient = useQueryClient();
  const [filters, setFilters] = useState<MoneyFeatureFilters>({ type: 'all', status: 'all', search: '' });
  const [selectedFeatureId, setSelectedFeatureId] = useState<string | null>(null);
  const [sheetCollapsed, setSheetCollapsed] = useState(false);
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
    mutationFn: async ({ type, points, title }: { type: MoneyFeatureType; points: MoneyPosition[]; title: string }) => {
      if (!projectId) throw new Error('Project not loaded');
      return moneyApi.createFeature(projectId, {
        feature_type: type,
        title,
        description: null,
        status: 'draft',
        geojson: buildGeoJSON(type, points),
        style: {},
        properties: { source: 'frontend-drawing' },
      });
    },
    onSuccess: feature => {
      void queryClient.invalidateQueries({ queryKey: ['money-snapshot', projectId] });
      setSelectedFeatureId(feature.id);
      setDrawingType(null);
      setDraftPoints([]);
      setSheetCollapsed(false);
    },
  });

  const features = snapshotQuery.data?.features ?? EMPTY_FEATURES;
  const filteredMapFeatures = useMemo(() => features.filter(feature => feature.status !== 'archived' || filters.status === 'archived'), [features, filters.status]);

  if (isBootstrapping) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-950 text-white">
        <Loader2 className="mr-3 h-6 w-6 animate-spin text-emerald-200" />
        Restoring toolkit session
      </div>
    );
  }

  if (!isAuthenticated) {
    return <LoginScreen />;
  }

  const project = projectQuery.data?.project ?? snapshotQuery.data?.project;
  const selectedFeature = detailQuery.data?.feature ?? features.find(feature => feature.id === selectedFeatureId) ?? null;
  const sheetTitle = selectedFeature ? selectedFeature.title : 'Money Creek Features';
  const sheetSubtitle = selectedFeature ? `${featureTypeLabel(selectedFeature.feature_type)} · ${selectedFeature.status}` : `${features.length} mapped items`;

  const handleSelectFeature = (feature: MoneyFeature) => {
    setSelectedFeatureId(feature.id);
    setSheetCollapsed(false);
  };

  const startDrawing = (type: MoneyFeatureType) => {
    setDrawingType(type);
    setDraftPoints([]);
    setSheetCollapsed(true);
  };

  const finishDrawing = () => {
    if (!drawingType || draftPoints.length < minimumPointCount(drawingType)) return;
    const title = window.prompt(`Name this ${featureTypeLabel(drawingType).toLowerCase()}`);
    if (!title?.trim()) return;
    createFeature.mutate({ type: drawingType, points: draftPoints, title: title.trim() });
  };

  const refreshData = () => {
    void projectQuery.refetch();
    void snapshotQuery.refetch();
    if (selectedFeatureId) void detailQuery.refetch();
  };

  return (
    <main className="fixed inset-0 overflow-hidden bg-slate-950 text-white">
      {project ? (
        <MoneyMap
          project={project}
          features={filteredMapFeatures}
          selectedFeatureId={selectedFeatureId}
          drawingType={drawingType}
          draftPoints={draftPoints}
          onSelectFeature={handleSelectFeature}
          onAddDraftPoint={point => setDraftPoints(points => drawingType === 'poi' ? [point] : [...points, point])}
        />
      ) : (
        <div className="flex h-full items-center justify-center">
          <Loader2 className="mr-3 h-6 w-6 animate-spin text-emerald-200" />
          Loading Money Creek project
        </div>
      )}

      <header className="absolute left-3 right-3 top-3 z-20 rounded-3xl border border-white/20 bg-slate-950/88 px-3 py-3 shadow-2xl backdrop-blur-xl md:left-4 md:right-4">
        <div className="flex items-center justify-between gap-3">
          <div className="flex min-w-0 items-center gap-3">
            <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-emerald-300 text-slate-950">
              <Mountain className="h-6 w-6" />
            </div>
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <h1 className="truncate text-base font-black sm:text-xl">Money Creek</h1>
                <span className="hidden rounded-full bg-white/10 px-2 py-0.5 text-[0.65rem] font-bold uppercase tracking-widest text-emerald-100 sm:inline">Toolkit</span>
              </div>
              <p className="truncate text-xs text-slate-400"><Users className="mr-1 inline h-3.5 w-3.5" />{user?.display_name} · {user?.role}</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button onClick={refreshData} className="rounded-2xl bg-white/10 p-2.5 text-slate-200 hover:bg-white/15" title="Refresh">
              <RefreshCw className={`h-4 w-4 ${snapshotQuery.isFetching ? 'animate-spin' : ''}`} />
            </button>
            <a href="/" className="hidden rounded-2xl bg-white/10 p-2.5 text-slate-200 hover:bg-white/15 sm:block" title="Dashboard">
              <Map className="h-4 w-4" />
            </a>
            <button onClick={() => void logout()} className="rounded-2xl bg-white/10 p-2.5 text-slate-200 hover:bg-white/15" title="Sign out">
              <LogOut className="h-4 w-4" />
            </button>
          </div>
        </div>
      </header>

      <DrawingToolbar
        drawingType={drawingType}
        draftPoints={draftPoints}
        canWrite={canWrite}
        onStart={startDrawing}
        onUndo={() => setDraftPoints(points => points.slice(0, -1))}
        onCancel={() => { setDrawingType(null); setDraftPoints([]); }}
        onFinish={finishDrawing}
      />

      {!isOnline && <OfflineBanner />}

      {(projectQuery.error || snapshotQuery.error) && (
        <div className="absolute left-3 right-3 top-20 z-30 rounded-2xl border border-red-300/30 bg-red-950/90 px-4 py-3 text-sm text-red-100 shadow-xl backdrop-blur md:left-4 md:right-auto md:w-96">
          <AlertCircle className="mr-2 inline h-4 w-4" />
          Failed to load toolkit data.
        </div>
      )}

      {createFeature.isPending && (
        <div className="absolute left-1/2 top-24 z-30 -translate-x-1/2 rounded-full bg-slate-950/90 px-4 py-2 text-sm shadow-xl backdrop-blur">
          <Loader2 className="mr-2 inline h-4 w-4 animate-spin text-emerald-200" />Saving drawing
        </div>
      )}

      <MobileBottomSheet title={sheetTitle} subtitle={sheetSubtitle} collapsed={sheetCollapsed} onToggle={() => setSheetCollapsed(value => !value)}>
        {selectedFeature && projectId ? (
          <div className="space-y-5">
            <button onClick={() => setSelectedFeatureId(null)} className="rounded-full bg-white/10 px-3 py-1.5 text-xs font-bold text-slate-200 hover:bg-white/15">← Feature list</button>
            <FeatureEditor
              key={selectedFeature.id}
              feature={selectedFeature}
              canWrite={canWrite}
              onSaved={feature => {
                queryClient.setQueryData(['money-feature', feature.id], { ...(detailQuery.data ?? { notes: [], uploads: [] }), feature });
                void queryClient.invalidateQueries({ queryKey: ['money-snapshot', projectId] });
              }}
              onArchived={() => {
                setSelectedFeatureId(null);
                void queryClient.invalidateQueries({ queryKey: ['money-snapshot', projectId] });
              }}
            />
            {detailQuery.isLoading ? (
              <div className="rounded-3xl border border-white/10 bg-white/8 p-5 text-center text-sm text-slate-300"><Loader2 className="mr-2 inline h-4 w-4 animate-spin" />Loading detail</div>
            ) : (
              <>
                <NotesPanel featureId={selectedFeature.id} notes={detailQuery.data?.notes ?? []} canWrite={canWrite} onChanged={() => { void detailQuery.refetch(); void queryClient.invalidateQueries({ queryKey: ['money-snapshot', projectId] }); }} />
                <ImageUploader projectId={projectId} featureId={selectedFeature.id} uploads={detailQuery.data?.uploads ?? []} canWrite={canWrite} onUploaded={() => { void detailQuery.refetch(); void queryClient.invalidateQueries({ queryKey: ['money-snapshot', projectId] }); }} onDeleted={() => { void detailQuery.refetch(); void queryClient.invalidateQueries({ queryKey: ['money-snapshot', projectId] }); }} />
              </>
            )}
          </div>
        ) : (
          <FeatureList
            features={features}
            selectedFeatureId={selectedFeatureId}
            filters={filters}
            noteCounts={snapshotQuery.data?.note_counts ?? {}}
            primaryUploads={snapshotQuery.data?.primary_uploads ?? {}}
            onSelect={handleSelectFeature}
            onFiltersChange={setFilters}
          />
        )}
      </MobileBottomSheet>
    </main>
  );
}

export default function MoneyCreekApp() {
  return (
    <AuthProvider>
      <MoneyCreekWorkspace />
    </AuthProvider>
  );
}
