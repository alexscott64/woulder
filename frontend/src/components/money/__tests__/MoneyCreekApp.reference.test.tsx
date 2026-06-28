import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import MoneyCreekApp from '../MoneyCreekApp';

vi.mock('../../../contexts/AuthContext', () => ({
  AuthProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  useAuth: () => ({ user: { id: 'u', email: 'dev@example.com', display_name: 'Dev', role: 'developer' }, isAuthenticated: true, isBootstrapping: false, canWrite: true, logout: vi.fn() }),
}));

const moneyMocks = vi.hoisted(() => ({
  uploadImage: vi.fn(),
  createProjectNote: vi.fn(),
}));

vi.mock('../../../services/money', () => ({
  moneyApi: {
    getProject: vi.fn().mockResolvedValue({ project: { id: 'project-1', slug: 'money-creek', name: 'Money Creek', center_lat: 0, center_lon: 0, default_zoom: 14, created_at: '', updated_at: '' }, user: {}, permissions: { can_write: true } }),
    getCragSnapshot: vi.fn().mockResolvedValue({ project: { id: 'project-1', slug: 'money-creek', name: 'Money Creek', center_lat: 0, center_lon: 0, default_zoom: 14, created_at: '', updated_at: '' }, trails: null, notes: null, uploads: null, root: { feature: { id: 'area-1', project_id: 'project-1', feature_type: 'area', title: 'Money Creek', description: 'Reference crag', status: 'active', geojson: { type: 'Polygon', coordinates: [[[0,0],[100,0],[100,100],[0,0]]] }, style: {}, properties: { kind: 'Crag', aspect: 'Skykomish' }, sort_order: 0, created_by: 'u', updated_by: 'u', created_at: '', updated_at: '' }, children: null, boulders: [{ feature: { id: 'boulder-1', project_id: 'project-1', feature_type: 'boulder', title: 'tiny boulder', description: null, status: 'scouted', geojson: { type: 'Polygon', coordinates: [[[0,0],[10,0],[10,10],[0,0]]] }, style: {}, properties: {}, sort_order: 0, created_by: 'u', updated_by: 'u', created_at: '', updated_at: '' }, children: null, boulders: null, problems: null }], problems: null } }),
    listTrash: vi.fn().mockResolvedValue({ items: [] }),
    archiveFeature: vi.fn().mockResolvedValue(undefined),
    restoreFeature: vi.fn().mockResolvedValue(undefined),
    moveFeatureParent: vi.fn().mockResolvedValue({}),
    uploadImage: moneyMocks.uploadImage,
    createProjectNote: moneyMocks.createProjectNote,
  },
}));

describe('MoneyCreekApp reference shell', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    moneyMocks.uploadImage.mockResolvedValue({ id: 'upload-1', original_filename: 'tiny-boulder.jpg' });
    moneyMocks.createProjectNote.mockResolvedValue({ id: 'note-1' });
  });

  it('renders the reference-style workspace navigation', async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(<QueryClientProvider client={client}><MoneyCreekApp /></QueryClientProvider>);
    expect(await screen.findAllByText('Money Creek')).toHaveLength(3);
    expect(screen.getByText('Workspace')).toBeTruthy();
    expect(screen.getAllByText('Boulders').length).toBeGreaterThan(0);
  });
  it('links photo note uploads to the selected boulder feature', async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
    render(<QueryClientProvider client={client}><MoneyCreekApp /></QueryClientProvider>);

    fireEvent.click(await screen.findByText('tiny boulder'));
    fireEvent.click(screen.getByText('Add'));
    fireEvent.click(screen.getByText('Add a photo'));

    const input = document.querySelector('input[type="file"]') as HTMLInputElement;
    const file = new File(['photo'], 'tiny-boulder.jpg', { type: 'image/jpeg' });
    Object.defineProperty(input, 'files', { value: [file], configurable: true });
    fireEvent.change(input);
    fireEvent.click(screen.getByText('Post note'));

    await waitFor(() => expect(moneyMocks.uploadImage).toHaveBeenCalledWith('project-1', file, { featureId: 'boulder-1', blockKind: 'photo' }));
  });
});
