import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, expect, it, vi } from 'vitest';
import MoneyCreekApp from '../MoneyCreekApp';

vi.mock('../../../contexts/AuthContext', () => ({
  AuthProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  useAuth: () => ({ user: { id: 'u', email: 'dev@example.com', display_name: 'Dev', role: 'developer' }, isAuthenticated: true, isBootstrapping: false, canWrite: true, logout: vi.fn() }),
}));

vi.mock('../../../services/money', () => ({
  moneyApi: {
    getProject: vi.fn().mockResolvedValue({ project: { id: 'project-1', slug: 'money-creek', name: 'Money Creek', center_lat: 0, center_lon: 0, default_zoom: 14, created_at: '', updated_at: '' }, user: {}, permissions: { can_write: true } }),
    getCragSnapshot: vi.fn().mockResolvedValue({ project: { id: 'project-1', slug: 'money-creek', name: 'Money Creek', center_lat: 0, center_lon: 0, default_zoom: 14, created_at: '', updated_at: '' }, trails: null, notes: null, uploads: null, root: { feature: { id: 'area-1', project_id: 'project-1', feature_type: 'area', title: 'Money Creek', description: 'Reference crag', status: 'active', geojson: { type: 'Polygon', coordinates: [[[0,0],[100,0],[100,100],[0,0]]] }, style: {}, properties: { kind: 'Crag', aspect: 'Skykomish' }, sort_order: 0, created_by: 'u', updated_by: 'u', created_at: '', updated_at: '' }, children: null, boulders: null, problems: null } }),
  },
}));

describe('MoneyCreekApp reference shell', () => {
  it('renders the reference-style workspace navigation', async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(<QueryClientProvider client={client}><MoneyCreekApp /></QueryClientProvider>);
    expect(await screen.findAllByText('Money Creek')).toHaveLength(3);
    expect(screen.getByText('Workspace')).toBeTruthy();
    expect(screen.getAllByText('Boulders').length).toBeGreaterThan(0);
  });
});
