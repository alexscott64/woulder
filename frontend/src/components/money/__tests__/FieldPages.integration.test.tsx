import { ReactNode, useMemo, useState } from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider, UseQueryResult } from '@tanstack/react-query';
import { FieldPages, FieldPageKey } from '../FieldPages';
import { MoneyFeature, MoneyFeatureDetail, MoneyFeatureFilters, MoneyNote, MoneyProject } from '../../../types/money';
import { moneyApi } from '../../../services/money';

vi.mock('../../../services/money', () => ({
  moneyApi: {
    createNote: vi.fn(),
    uploadImage: vi.fn(),
    archiveFeature: vi.fn(),
    updateFeature: vi.fn(),
    deleteNote: vi.fn(),
    deleteUpload: vi.fn(),
    getUploadDownloadURL: vi.fn(),
    getUploadBlobUrl: vi.fn(),
  },
}));

vi.mock('../../../contexts/AuthContext', () => ({
  useAuth: () => ({
    user: { id: 'user-1', email: 'dev@example.com', display_name: 'Dev User', role: 'developer' },
    canWrite: true,
  }),
}));

vi.mock('../MoneyMap', () => ({
  MoneyMap: ({ features, selectedFeatureId, onSelectFeature }: { features: MoneyFeature[]; selectedFeatureId: string | null; onSelectFeature: (feature: MoneyFeature) => void }) => (
    <div data-testid="money-map-mock" data-feature-ids={features.map(feature => feature.id).join(',')} data-selected-feature-id={selectedFeatureId ?? ''}>
      {features.map(feature => <button key={feature.id} type="button" onClick={() => onSelectFeature(feature)}>Map select {feature.title}</button>)}
    </div>
  ),
}));

const project: MoneyProject = {
  id: 'project-1',
  slug: 'money-creek',
  name: 'Money Creek',
  center_lat: 47.7,
  center_lon: -121.46,
  default_zoom: 14,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
};

const pinFeature: MoneyFeature = {
  id: 'pin-1',
  project_id: project.id,
  feature_type: 'poi',
  title: 'Warmup Pin',
  description: 'Landing dries first',
  status: 'active',
  geojson: { type: 'Point', coordinates: [-121.45, 47.71] },
  style: {},
  properties: { poi_category: 'general', poi_label: 'General note' },
  created_by: 'user-1',
  updated_by: 'user-1',
  created_at: '2026-01-02T00:00:00Z',
  updated_at: '2026-01-02T00:00:00Z',
};

function queryResult(data: MoneyFeatureDetail | undefined, isLoading = false): UseQueryResult<MoneyFeatureDetail> {
  return {
    data,
    isLoading,
    isFetching: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
  } as unknown as UseQueryResult<MoneyFeatureDetail>;
}

function renderWithClient(children: ReactNode) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(<QueryClientProvider client={client}>{children}</QueryClientProvider>);
}

function Harness({ initialSelected = pinFeature, initialPage = 'scratch' as FieldPageKey }: { initialSelected?: MoneyFeature | null; initialPage?: FieldPageKey }) {
  const [selectedFeature, setSelectedFeature] = useState<MoneyFeature | null>(initialSelected);
  const [page, setPage] = useState<FieldPageKey>(initialPage);
  const [filters, setFilters] = useState<MoneyFeatureFilters>({ type: 'all', status: 'all', search: '' });
  const [notes, setNotes] = useState<MoneyNote[]>([]);
  const detail = useMemo<MoneyFeatureDetail | undefined>(() => selectedFeature ? { feature: selectedFeature, notes, uploads: [] } : undefined, [notes, selectedFeature]);

  return (
    <FieldPages
      user={{ id: 'user-1', email: 'dev@example.com', display_name: 'Dev User', role: 'developer' }}
      project={project}
      features={[pinFeature]}
      visibleMapFeatures={[pinFeature]}
      selectedFeature={selectedFeature}
      selectedFeatureId={selectedFeature?.id ?? null}
      detailQuery={queryResult(detail)}
      projectId={project.id}
      canWrite
      isFetching={false}
      isOnline
      loadError={false}
      savingFeature={false}
      filters={filters}
      noteCounts={notes.length ? { [pinFeature.id]: notes.length } : {}}
      primaryUploads={{}}
      lens="all"
      page={page}
      drawingType={null}
      draftPoints={[]}
      onLensChange={vi.fn()}
      onPageChange={setPage}
      onFiltersChange={setFilters}
      onSelectFeature={feature => { setSelectedFeature(feature); setPage('scratch'); }}
      onClearSelection={() => setSelectedFeature(null)}
      onRefresh={vi.fn()}
      onLogout={vi.fn()}
      onChanged={() => setNotes([{ id: 'note-1', project_id: project.id, feature_id: pinFeature.id, body: 'Saved beta note', visibility: 'team', created_by: 'user-1', updated_by: 'user-1', created_at: '2026-01-03T00:00:00Z', updated_at: '2026-01-03T00:00:00Z' }])}
      onFeatureSaved={setSelectedFeature}
      onFeatureArchived={() => setSelectedFeature(null)}
      onStartDrawing={vi.fn()}
      onAddDraftPoint={vi.fn()}
      onUndoDraftPoint={vi.fn()}
      onCancelDrawing={vi.fn()}
      onFinishDrawing={vi.fn()}
    />
  );
}

describe('FieldPages Money integration flows', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(moneyApi.createNote).mockResolvedValue({ id: 'note-1', project_id: project.id, feature_id: pinFeature.id, body: 'Saved beta note', visibility: 'team', created_by: 'user-1', updated_by: 'user-1', created_at: '2026-01-03T00:00:00Z', updated_at: '2026-01-03T00:00:00Z' });
    vi.mocked(moneyApi.archiveFeature).mockResolvedValue(undefined);
  });

  it('saves a selected feature quick note through the API and shows persisted selected-feature note state', async () => {
    renderWithClient(<Harness />);

    fireEvent.click(screen.getByRole('button', { name: 'Note' }));
    fireEvent.change(screen.getByPlaceholderText('Beta, access note, cleanup task, photo reminder...'), { target: { value: 'Saved beta note' } });
    fireEvent.click(screen.getAllByRole('button', { name: 'Save note' }).at(-1)!);

    await waitFor(() => expect(moneyApi.createNote).toHaveBeenCalledWith(pinFeature.id, { body: 'Saved beta note', visibility: 'team' }));
    expect(await screen.findAllByText('Saved beta note')).toHaveLength(2);
    expect(screen.queryByText('Scratch note')).toBeNull();
  });

  it('keeps photo capture contextual and prompts for a selected map item without switching to Inbox', () => {
    renderWithClient(<Harness initialSelected={null} />);

    fireEvent.click(screen.getByRole('button', { name: 'Photo' }));

    expect(screen.getByText('Select or create a map item first')).toBeTruthy();
    expect(screen.getAllByText(/Scratch/).length).toBeGreaterThan(0);
    expect(screen.queryByText(/Inbox/)).toBeNull();
    expect(moneyApi.uploadImage).not.toHaveBeenCalled();
  });

  it('selects an existing list/map pin and passes visible selected feature data to the map', async () => {
    renderWithClient(<Harness initialSelected={null} />);

    let map = screen.getAllByTestId('money-map-mock')[0];
    expect(map.getAttribute('data-feature-ids')).toContain(pinFeature.id);
    expect(map.getAttribute('data-selected-feature-id')).toBe('');

    fireEvent.click(screen.getAllByRole('button', { name: /Warmup Pin/ })[0]);

    await waitFor(() => {
      map = screen.getAllByTestId('money-map-mock')[0];
      expect(map.getAttribute('data-selected-feature-id')).toBe(pinFeature.id);
    });

    fireEvent.click(screen.getAllByText('Map select Warmup Pin')[0]);
    await waitFor(() => expect(screen.getAllByTestId('money-map-mock')[0].getAttribute('data-selected-feature-id')).toBe(pinFeature.id));
  });

  it('shows an archive action for a selected user-created pin and calls the API', async () => {
    renderWithClient(<Harness />);

    const archiveButton = screen.getByRole('button', { name: 'Archive pin' });
    expect(archiveButton).toBeTruthy();
    fireEvent.click(archiveButton);

    await waitFor(() => expect(moneyApi.archiveFeature).toHaveBeenCalledWith(pinFeature.id));
  });

  it('renders map controls above selected detail overlays at component state level', () => {
    renderWithClient(<Harness />);

    const map = screen.getAllByTestId('money-map-mock')[0];
    expect(map.getAttribute('data-selected-feature-id')).toBe(pinFeature.id);
    expect(screen.getByRole('button', { name: 'Archive pin' })).toBeTruthy();
  });
});
