import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { NoteComposer } from './NoteComposer';
import { MoneyCragNode, MoneyFeature, MoneyGeoJSON, MoneyNote, MoneyUpload } from '../../../types/money';

vi.mock('../../../services/money', () => ({
  moneyApi: {
    getUploadBlobUrl: vi.fn().mockResolvedValue('https://example.invalid/tiny.jpeg'),
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
  feature: { ...featureBase, id: 'area-1', feature_type: 'area', title: 'Money Creek', description: 'Reference crag', status: 'active', properties: { kind: 'Crag', aspect: 'forest' } } as MoneyFeature,
  children: null,
  boulders: [boulder],
  problems: null,
};

const upload: MoneyUpload = {
  id: 'upload-1',
  project_id: 'project-1',
  feature_id: 'boulder-1',
  note_id: 'note-1',
  original_filename: 'tiny.jpeg',
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
  blocks: [{ kind: 'photo', upload_id: 'upload-1', name: 'tiny.jpeg' }],
  created_by: 'user-1',
  updated_by: 'user-1',
  created_at: '2026-01-02T00:00:00Z',
  updated_at: '2026-01-02T00:00:00Z',
};

describe('NoteComposer edit media', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders an existing uploaded photo once with its authenticated preview', async () => {
    render(<NoteComposer root={area} area={area} boulder={boulder} initialNote={note} uploads={[upload]} mobile={false} onClose={vi.fn()} onSubmit={vi.fn()} />);

    const image = await screen.findByAltText('tiny.jpeg');
    await waitFor(() => expect(image.getAttribute('src')).toBe('https://example.invalid/tiny.jpeg'));
    expect(screen.getAllByAltText('tiny.jpeg')).toHaveLength(1);
    expect(screen.queryByText('tiny.jpeg')).toBeNull();
  });

  it('saves existing upload blocks without duplicating them as new files', async () => {
    const onSubmit = vi.fn();
    render(<NoteComposer root={area} area={area} boulder={boulder} initialNote={note} uploads={[upload]} mobile={false} onClose={vi.fn()} onSubmit={onSubmit} />);

    fireEvent.click(screen.getByText('Save note'));

    expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({
      files: [],
      blocks: [{ kind: 'photo', upload_id: 'upload-1', name: 'tiny.jpeg' }],
    }));
  });

  it('does not render a file input for Sketch', () => {
    render(<NoteComposer root={area} area={area} boulder={boulder} mobile={false} onClose={vi.fn()} onSubmit={vi.fn()} />);

    fireEvent.click(screen.getByRole('button', { name: 'Sketch' }));

    expect(screen.getAllByLabelText(/Photo|File/)).toHaveLength(2);
    expect(screen.getByRole('status').textContent).toContain('Sketch is not available');
  });
});
