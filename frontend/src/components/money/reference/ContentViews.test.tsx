import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ContentView } from './ContentViews';
import { MoneyCragNode, MoneyFeature, MoneyGeoJSON, MoneyNote, MoneyUpload } from '../../../types/money';

vi.mock('../../../services/money', () => ({
  moneyApi: {
    getUploadBlobUrl: vi.fn().mockResolvedValue('https://example.invalid/photo.jpg'),
  },
}));

const featureBase = {
  project_id: 'project-1',
  geojson: { type: 'Polygon', coordinates: [[[0, 0], [1, 0], [1, 1], [0, 0]]] } as MoneyGeoJSON,
  style: {},
  properties: {},
  sort_order: 0,
  created_by: 'user-1',
  updated_by: 'user-1',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
};

const boulder: MoneyCragNode = {
  feature: { ...featureBase, id: 'boulder-1', feature_type: 'boulder', title: 'tiny boulder', status: 'scouted' } as MoneyFeature,
  children: null,
  boulders: null,
  problems: null,
};

const area: MoneyCragNode = {
  feature: { ...featureBase, id: 'area-1', feature_type: 'area', title: 'Money Creek', description: 'Reference crag', status: 'active', properties: { kind: 'Crag' } } as MoneyFeature,
  children: null,
  boulders: [boulder],
  problems: null,
};

const upload: MoneyUpload = {
  id: 'upload-1',
  project_id: 'project-1',
  feature_id: 'boulder-1',
  original_filename: 'tiny-boulder.jpg',
  content_type: 'image/jpeg',
  byte_size: 1234,
  checksum_sha256: 'sha',
  block_kind: 'photo',
  metadata: {},
  asset_kind: 'original',
  storage_backend: 'r2',
  visibility: 'private',
  sync_status: 'available',
  uploaded_by: 'user-1',
  created_at: '2026-01-02T00:00:00Z',
};

const note: MoneyNote = {
  id: 'note-1',
  project_id: 'project-1',
  target_type: 'boulder',
  target_ref: 'boulder-1',
  body: 'Tiny boulder note',
  visibility: 'team',
  tags: [],
  blocks: [{ kind: 'photo', upload_id: 'upload-1', name: 'tiny-boulder.jpg' }],
  created_by: 'user-1',
  updated_by: 'user-1',
  created_at: '2026-01-02T00:00:00Z',
  updated_at: '2026-01-02T00:00:00Z',
};

describe('ContentView photos', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('opens Photos view uploads with associated note and area/boulder context', async () => {
    render(<ContentView view="photos" root={area} trails={[]} notes={[note]} uploads={[upload]} trash={[]} canWrite mobile={false} openBoulder={vi.fn()} selectTrail={vi.fn()} onOpenComposer={vi.fn()} onEditNote={vi.fn()} onDeleteNote={vi.fn()} onDeleteUpload={vi.fn()} onRestore={vi.fn()} />);

    fireEvent.click(await screen.findByRole('button', { name: 'Open photo tiny-boulder.jpg' }));

    expect(await screen.findByRole('dialog', { name: 'tiny-boulder.jpg' })).toBeTruthy();
    expect(screen.getByText('Tiny boulder note')).toBeTruthy();
    expect(screen.getAllByText('Money Creek / tiny boulder').length).toBeGreaterThan(0);
  });
});

