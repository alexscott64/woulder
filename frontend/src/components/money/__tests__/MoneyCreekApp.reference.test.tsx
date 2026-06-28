import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import MoneyCreekApp from '../MoneyCreekApp';
import type { MoneyCragNode, MoneyFeature } from '../../../types/money';

vi.mock('../../../contexts/AuthContext', () => ({
  AuthProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  useAuth: () => ({ user: { id: 'u', email: 'dev@example.com', display_name: 'Dev', role: 'developer' }, isAuthenticated: true, isBootstrapping: false, canWrite: true, logout: vi.fn() }),
}));

vi.mock('../reference/CragMap', () => ({
  CragMap: ({ onSelectBoulder }: { onSelectBoulder: (id: string | null) => void }) => <div data-testid="crag-map"><button onClick={() => onSelectBoulder('nested-boulder')}>Map select nested boulder</button></div>,
}));

const moneyMocks = vi.hoisted(() => ({
  uploadImage: vi.fn(),
  createProjectNote: vi.fn(),
  getCragSnapshot: vi.fn(),
  updateFeature: vi.fn(),
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

function makeRoot(): MoneyCragNode {
  const rootBoulder: MoneyCragNode = { feature: feature({ id: 'root-boulder', feature_type: 'boulder', title: 'roadside boulder', status: 'scouted', geojson: { type: 'Polygon', coordinates: [[[20, 0], [30, 0], [30, 10], [20, 0]]] } }), children: null, boulders: null, problems: null };
  const nestedBoulder: MoneyCragNode = { feature: feature({ id: 'nested-boulder', feature_type: 'boulder', title: 'tiny boulder', status: 'scouted', geojson: { type: 'Polygon', coordinates: [[[0, 0], [10, 0], [10, 10], [0, 0]]] } }), children: null, boulders: null, problems: null };
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
  },
}));

function renderApp() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(<QueryClientProvider client={client}><MoneyCreekApp /></QueryClientProvider>);
}

describe('MoneyCreekApp reference shell', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    moneyMocks.getCragSnapshot.mockResolvedValue({ project: { id: 'project-1', slug: 'money-creek', name: 'Money Creek', center_lat: 0, center_lon: 0, default_zoom: 14, created_at: '', updated_at: '' }, trails: null, notes: null, uploads: null, root: makeRoot() });
    moneyMocks.uploadImage.mockResolvedValue({ id: 'upload-1', original_filename: 'tiny-boulder.jpg' });
    moneyMocks.createProjectNote.mockResolvedValue({ id: 'note-1' });
    moneyMocks.updateFeature.mockImplementation(async (_id, payload) => ({ id: _id, ...payload }));
  });

  it('renders the reference-style workspace navigation', async () => {
    renderApp();
    expect(await screen.findAllByText('Money Creek')).toHaveLength(3);
    expect(screen.getByText('Workspace')).toBeTruthy();
    expect(screen.getAllByText('Boulders').length).toBeGreaterThan(0);
  });

  it('links photo note uploads to the selected boulder feature', async () => {
    renderApp();

    fireEvent.click(await screen.findByText('Map select nested boulder'));
    fireEvent.click(screen.getByText('Add'));
    fireEvent.click(screen.getByText('Add a photo'));

    const input = document.querySelector('input[type="file"]') as HTMLInputElement;
    const file = new File(['photo'], 'tiny-boulder.jpg', { type: 'image/jpeg' });
    Object.defineProperty(input, 'files', { value: [file], configurable: true });
    fireEvent.change(input);
    fireEvent.click(screen.getByText('Post note'));

    await waitFor(() => expect(moneyMocks.uploadImage).toHaveBeenCalledWith('project-1', file, { featureId: 'nested-boulder', blockKind: 'photo' }));
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
});
