import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import MoneyCreekApp, { resolveSubmittedBlocks } from '../MoneyCreekApp';
import type { MoneyCragNode, MoneyFeature, MoneyUpload } from '../../../types/money';

vi.mock('../../../contexts/AuthContext', () => ({
  AuthProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  useAuth: () => ({ user: { id: 'u', email: 'dev@example.com', display_name: 'Dev', role: 'developer' }, isAuthenticated: true, isBootstrapping: false, canWrite: true, logout: vi.fn() }),
}));

vi.mock('../reference/CragMap', () => ({
  CragMap: ({ area, selectedBoulderId, goToPoint, onSelectBoulder, onAddBoulderAt }: { area: MoneyCragNode; selectedBoulderId: string | null; goToPoint?: { position: [number, number]; nonce: number } | null; onSelectBoulder: (id: string | null) => void; onAddBoulderAt: (position: [number, number], parentId: string | null) => void }) => <div data-testid="crag-map" data-area-id={area.feature.id} data-selected-boulder-id={selectedBoulderId ?? ''} data-go-to-point={goToPoint ? goToPoint.position.join(',') : ''}><button onClick={() => onSelectBoulder('nested-boulder')}>Map select nested boulder</button><button onClick={() => onAddBoulderAt([-121.4703, 47.6997], 'nested-area')}>Map add boulder here</button>{goToPoint && <button onClick={() => onAddBoulderAt(goToPoint.position, area.feature.id)}>Search add boulder action</button>}</div>,
}));

const moneyMocks = vi.hoisted(() => ({
  uploadImage: vi.fn(),
  createProjectNote: vi.fn(),
  getCragSnapshot: vi.fn(),
  updateFeature: vi.fn(),
  updateNote: vi.fn(),
  deleteUpload: vi.fn(),
  createBoulder: vi.fn(),
}));

const baseFeature = {
  project_id: 'project-1',
  style: {},
  properties: {},
  sort_order: 0,
  created_by: 'u',
  updated_by: 'u',
  created_at: '',
  updated_at: '',
};

function feature(overrides: Partial<MoneyFeature> & Pick<MoneyFeature, 'id' | 'feature_type' | 'title' | 'status' | 'geojson'>): MoneyFeature {
  return { ...baseFeature, ...overrides } as MoneyFeature;
}

function makeRoot(nestedBoulderTitle = 'tiny boulder'): MoneyCragNode {
  const nestedProblem: MoneyCragNode = { feature: feature({ id: 'nested-problem', parent_feature_id: 'nested-boulder', feature_type: 'problem', title: 'Tiny Arete', status: 'project', geojson: { type: 'Point', coordinates: [5, 5] }, properties: {} }), children: null, boulders: null, problems: null };
  const rootBoulder: MoneyCragNode = { feature: feature({ id: 'root-boulder', feature_type: 'boulder', title: 'roadside boulder', status: 'scouted', geojson: { type: 'Polygon', coordinates: [[[20, 0], [30, 0], [30, 10], [20, 0]]] } }), children: null, boulders: null, problems: null };
  const nestedBoulder: MoneyCragNode = { feature: feature({ id: 'nested-boulder', parent_feature_id: 'nested-area', feature_type: 'boulder', title: nestedBoulderTitle, status: 'scouted', geojson: { type: 'Polygon', coordinates: [[[0, 0], [10, 0], [10, 10], [0, 0]]] }, properties: { landing: 'flat' } }), children: null, boulders: null, problems: [nestedProblem] };
  const nestedArea: MoneyCragNode = { feature: feature({ id: 'nested-area', feature_type: 'area', title: 'Sub Zone', description: 'Nested area', status: 'active', geojson: { type: 'Polygon', coordinates: [[[0, 0], [50, 0], [50, 50], [0, 0]]] }, properties: { kind: 'Boulders', aspect: 'forest' } }), children: null, boulders: [nestedBoulder], problems: null };
  return { feature: feature({ id: 'area-1', feature_type: 'area', title: 'Money Creek', description: 'Reference crag', status: 'active', geojson: { type: 'Polygon', coordinates: [[[0, 0], [100, 0], [100, 100], [0, 0]]] }, properties: { kind: 'Crag', aspect: 'Skykomish' } }), children: [nestedArea], boulders: [rootBoulder], problems: null };
}

vi.mock('../../../services/money', () => ({
  moneyApi: {
    getProject: vi.fn().mockResolvedValue({ project: { id: 'project-1', slug: 'money-creek', name: 'Money Creek', center_lat: 0, center_lon: 0, default_zoom: 14, created_at: '', updated_at: '' }, user: {}, permissions: { can_write: true } }),
    getCragSnapshot: moneyMocks.getCragSnapshot,
    listTrash: vi.fn().mockResolvedValue({ items: [] }),
    archiveFeature: vi.fn().mockResolvedValue(undefined),
    restoreFeature: vi.fn().mockResolvedValue(undefined),
    moveFeatureParent: vi.fn().mockResolvedValue({}),
    uploadImage: moneyMocks.uploadImage,
    createProjectNote: moneyMocks.createProjectNote,
    updateFeature: moneyMocks.updateFeature,
    updateNote: moneyMocks.updateNote,
    deleteUpload: moneyMocks.deleteUpload,
    createBoulder: moneyMocks.createBoulder,
    getUploadBlobUrl: vi.fn().mockResolvedValue('https://example.invalid/existing.jpg'),
  },
}));

function moneyUpload(id: string, filename: string, featureId = 'nested-boulder'): MoneyUpload {
  return { id, project_id: 'project-1', feature_id: featureId, original_filename: filename, content_type: 'image/jpeg', byte_size: 1, checksum_sha256: `sha-${id}`, block_kind: 'photo', metadata: {}, asset_kind: 'original', storage_backend: 'r2', visibility: 'private', sync_status: 'available', uploaded_by: 'u', created_at: '2026-01-01T00:00:00Z' };
}

function renderApp() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(<QueryClientProvider client={client}><MoneyCreekApp /></QueryClientProvider>);
}

describe('MoneyCreekApp reference shell', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(URL, 'createObjectURL', { value: vi.fn((file: File) => `blob:${file.name}`), configurable: true });
    Object.defineProperty(URL, 'revokeObjectURL', { value: vi.fn(), configurable: true });
    let root = makeRoot();
    moneyMocks.getCragSnapshot.mockImplementation(async () => ({ project: { id: 'project-1', slug: 'money-creek', name: 'Money Creek', center_lat: 0, center_lon: 0, default_zoom: 14, created_at: '', updated_at: '' }, trails: null, notes: null, uploads: null, root }));
    moneyMocks.uploadImage.mockResolvedValue({ id: 'upload-1', original_filename: 'tiny-boulder.jpg' });
    moneyMocks.createProjectNote.mockResolvedValue({ id: 'note-1' });
    moneyMocks.updateNote.mockResolvedValue({ id: 'note-existing' });
    moneyMocks.deleteUpload.mockResolvedValue(undefined);
    moneyMocks.updateFeature.mockImplementation(async (_id, payload) => { root = makeRoot(payload.title); return { id: _id, ...payload }; });
    moneyMocks.createBoulder.mockResolvedValue({ id: 'new-boulder', parent_feature_id: 'nested-area' });
  });

  it('renders the reference-style workspace navigation', async () => {
    renderApp();
    expect(await screen.findAllByText('Money Creek')).toHaveLength(3);
    expect(screen.getByText('Workspace')).toBeTruthy();
    expect(screen.getAllByText('Boulders').length).toBeGreaterThan(0);
  });

  it('resolves multiple local attachment blocks to uploaded blocks in selection order', () => {
    const uploaded = new Map([
      ['local-1', { kind: 'photo' as const, upload_id: 'upload-1', name: 'first.jpg' }],
      ['local-2', { kind: 'photo' as const, upload_id: 'upload-2', name: 'second.jpg' }],
      ['local-3', { kind: 'file' as const, upload_id: 'upload-3', name: 'topo.pdf' }],
    ]);

    expect(resolveSubmittedBlocks([
      { kind: 'photo', name: 'first.jpg', url: 'blob:first.jpg', metadata: { local_block_key: 'local-1' } },
      { kind: 'photo', name: 'second.jpg', url: 'blob:second.jpg', metadata: { local_block_key: 'local-2' } },
      { kind: 'file', name: 'topo.pdf', metadata: { local_block_key: 'local-3' } },
    ], uploaded)).toEqual([
      { kind: 'photo', upload_id: 'upload-1', name: 'first.jpg' },
      { kind: 'photo', upload_id: 'upload-2', name: 'second.jpg' },
      { kind: 'file', upload_id: 'upload-3', name: 'topo.pdf' },
    ]);
  });

  it('resolves edited note blocks by preserving existing attachments and appending new uploads once', () => {
    const uploaded = new Map([
      ['local-new', { kind: 'photo' as const, upload_id: 'new-upload', name: 'new.jpg' }],
    ]);

    expect(resolveSubmittedBlocks([
      { kind: 'photo', upload_id: 'existing-1', name: 'existing.jpg' },
      { kind: 'photo', name: 'new.jpg', url: 'blob:new.jpg', metadata: { local_block_key: 'local-new' } },
    ], uploaded)).toEqual([
      { kind: 'photo', upload_id: 'existing-1', name: 'existing.jpg' },
      { kind: 'photo', upload_id: 'new-upload', name: 'new.jpg' },
    ]);
  });

  it('syncs the active area to a boulder selected from the map', async () => {
    renderApp();

    fireEvent.click(await screen.findByText('Map select nested boulder'));
    await screen.findAllByText('Sub Zone');
    fireEvent.click(screen.getAllByText('Boulders')[0]);

    expect((await screen.findAllByText('tiny boulder')).length).toBeGreaterThan(0);
    expect(screen.queryByText('roadside boulder')).toBeNull();
    const areaFilter = screen.getByLabelText('Boulders filters area') as HTMLSelectElement;
    expect(within(areaFilter).getByText('Current area')).toBeTruthy();
  });

  it('renames the selected boulder from the icon header editor', async () => {
    renderApp();

    fireEvent.click(await screen.findByText('Map select nested boulder'));
    fireEvent.click(screen.getByRole('button', { name: 'Edit boulder name' }));
    fireEvent.change(screen.getByLabelText('Boulder name'), { target: { value: 'Tiny Roof' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save boulder name' }));

    await waitFor(() => expect(moneyMocks.updateFeature).toHaveBeenCalledWith('nested-boulder', expect.objectContaining({ feature_type: 'boulder', title: 'Tiny Roof', parent_feature_id: 'nested-area', status: 'scouted', geojson: { type: 'Polygon', coordinates: [[[0, 0], [10, 0], [10, 10], [0, 0]]] }, properties: { landing: 'flat' } })));
    await waitFor(() => expect(screen.getByTestId('crag-map').dataset.selectedBoulderId).toBe('nested-boulder'));
    expect(screen.getByTestId('crag-map').dataset.areaId).toBe('nested-area');
    expect((await screen.findAllByText('Tiny Roof')).length).toBeGreaterThan(0);
  });

  it('navigates when breadcrumb area items are clicked', async () => {
    renderApp();

    fireEvent.click(await screen.findByText('Map select nested boulder'));
    await waitFor(() => expect(screen.getByTestId('crag-map').dataset.areaId).toBe('nested-area'));
    fireEvent.click(screen.getByRole('button', { name: 'Go to Money Creek' }));

    await waitFor(() => expect(screen.getByTestId('crag-map').dataset.areaId).toBe('area-1'));
    expect(screen.getByTestId('crag-map').dataset.selectedBoulderId).toBe('');
  });

  it('creates a point boulder from map right-click add flow', async () => {
    renderApp();

    fireEvent.click(await screen.findByText('Map add boulder here'));
    fireEvent.change(screen.getByLabelText('Boulder name'), { target: { value: 'Coordinate Bloc' } });
    fireEvent.click(screen.getByText('Create boulder'));

    await waitFor(() => expect(moneyMocks.createBoulder).toHaveBeenCalledWith('project-1', expect.objectContaining({
      parent_feature_id: 'nested-area',
      title: 'Coordinate Bloc',
      dev_status: 'scouted',
      geojson: { type: 'Point', coordinates: [-121.4703, 47.6997] },
    })));
  });

  it('clears the coordinate add-boulder action after creating from search', async () => {
    renderApp();

    fireEvent.change(await screen.findByLabelText('Map search'), { target: { value: '47.6997, -121.4703' } });
    fireEvent.submit(screen.getByLabelText('Map search'));
    await waitFor(() => expect(screen.getByTestId('crag-map').dataset.goToPoint).toBe('-121.4703,47.6997'));
    fireEvent.click(screen.getByText('Search add boulder action'));
    fireEvent.change(screen.getByLabelText('Boulder name'), { target: { value: 'Search Bloc' } });
    fireEvent.click(screen.getByText('Create boulder'));

    await waitFor(() => expect(moneyMocks.createBoulder).toHaveBeenCalledWith('project-1', expect.objectContaining({
      parent_feature_id: 'area-1',
      title: 'Search Bloc',
      geojson: { type: 'Point', coordinates: [-121.4703, 47.6997] },
    })));
    await waitFor(() => expect(screen.getByTestId('crag-map').dataset.goToPoint).toBe(''));
    expect(screen.queryByText('Search add boulder action')).toBeNull();
  });

  it('creates a point boulder from Boulders menu with typed lat/long', async () => {
    renderApp();

    fireEvent.click((await screen.findAllByText('Boulders'))[0]);
    fireEvent.click(await screen.findByRole('button', { name: 'Add boulder' }));
    fireEvent.change(screen.getByLabelText('Boulder name'), { target: { value: 'Menu Bloc' } });
    fireEvent.change(screen.getByLabelText('Boulder latitude'), { target: { value: '47.7' } });
    fireEvent.change(screen.getByLabelText('Boulder longitude'), { target: { value: '-121.47' } });
    fireEvent.click(screen.getByText('Create boulder'));

    await waitFor(() => expect(moneyMocks.createBoulder).toHaveBeenCalledWith('project-1', expect.objectContaining({
      parent_feature_id: 'area-1',
      title: 'Menu Bloc',
      geojson: { type: 'Point', coordinates: [-121.47, 47.7] },
    })));
  });

  it('deletes removed existing note attachments after saving an edited note', async () => {
    let root = makeRoot();
    moneyMocks.getCragSnapshot.mockImplementation(async () => ({
      project: { id: 'project-1', slug: 'money-creek', name: 'Money Creek', center_lat: 0, center_lon: 0, default_zoom: 14, created_at: '', updated_at: '' },
      trails: null,
      root,
      notes: [{ id: 'note-existing', project_id: 'project-1', target_type: 'boulder', target_ref: 'nested-boulder', body: 'photo note', visibility: 'team', tags: [], blocks: [{ kind: 'photo', upload_id: 'upload-existing', name: 'existing.jpg' }], created_by: 'u', updated_by: 'u', created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' }],
      uploads: [{ id: 'upload-existing', project_id: 'project-1', note_id: 'note-existing', original_filename: 'existing.jpg', content_type: 'image/jpeg', byte_size: 1, checksum_sha256: 'sha', block_kind: 'photo', metadata: {}, asset_kind: 'original', storage_backend: 'r2', visibility: 'private', sync_status: 'available', uploaded_by: 'u', created_at: '2026-01-01T00:00:00Z' }],
    }));
    renderApp();

    fireEvent.click(await screen.findByText('Map select nested boulder'));
    fireEvent.click(await screen.findByRole('button', { name: 'Edit note' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Remove attachment existing.jpg' }));
    fireEvent.click(screen.getByText('Save note'));

    await waitFor(() => expect(moneyMocks.updateNote).toHaveBeenCalledWith('note-existing', expect.objectContaining({ blocks: [] })));
    await waitFor(() => expect(moneyMocks.deleteUpload).toHaveBeenCalledWith('upload-existing'));
  });

  it('opens selected boulder photos view-only before entering topo edit mode', async () => {
    moneyMocks.getCragSnapshot.mockResolvedValue({
      project: { id: 'project-1', slug: 'money-creek', name: 'Money Creek', center_lat: 0, center_lon: 0, default_zoom: 14, created_at: '', updated_at: '' },
      trails: null,
      root: makeRoot(),
      notes: [],
      uploads: [moneyUpload('upload-existing', 'existing.jpg')],
    });
    renderApp();

    fireEvent.click(await screen.findByText('Map select nested boulder'));
    fireEvent.click(await screen.findByRole('button', { name: 'Open photo existing.jpg' }));

    expect(await screen.findByRole('dialog', { name: 'existing.jpg' })).toBeTruthy();
    expect(screen.queryByRole('region', { name: 'Photo topo editor' })).toBeNull();
    expect(screen.queryByText('Draw topo')).toBeNull();
    fireEvent.click(within(screen.getByRole('dialog', { name: 'existing.jpg' })).getByRole('button', { name: 'Draw topo lines' }));
    expect(await screen.findByRole('region', { name: 'Photo topo editor' })).toBeTruthy();
    expect(screen.getByText('Draw topo')).toBeTruthy();
  });

  it('shows a photo picker for external draw topo action before opening editor', async () => {
    moneyMocks.getCragSnapshot.mockResolvedValue({
      project: { id: 'project-1', slug: 'money-creek', name: 'Money Creek', center_lat: 0, center_lon: 0, default_zoom: 14, created_at: '', updated_at: '' },
      trails: null,
      root: makeRoot(),
      notes: [],
      uploads: [moneyUpload('upload-one', 'one.jpg'), moneyUpload('upload-two', 'two.jpg')],
    });
    renderApp();

    fireEvent.click(await screen.findByText('Map select nested boulder'));
    fireEvent.click(await screen.findByRole('button', { name: 'Draw topo lines' }));

    expect(await screen.findByRole('dialog', { name: 'Select topo photo' })).toBeTruthy();
    expect(screen.queryByRole('region', { name: 'Photo topo editor' })).toBeNull();
    fireEvent.click(screen.getByText('two.jpg'));

    expect(await screen.findByRole('dialog', { name: 'two.jpg' })).toBeTruthy();
    expect(await screen.findByRole('region', { name: 'Photo topo editor' })).toBeTruthy();
    expect(screen.getByText('Draw topo')).toBeTruthy();
  });
});
