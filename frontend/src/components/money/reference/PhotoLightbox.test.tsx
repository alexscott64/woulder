import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { PhotoLightbox } from './PhotoLightbox';
import { MoneyCragNode, MoneyFeature, MoneyGeoJSON, MoneyUpload } from '../../../types/money';

vi.mock('../../../services/money', () => ({
  moneyApi: {
    getUploadBlobUrl: vi.fn().mockResolvedValue('https://example.invalid/photo.jpg'),
  },
}));

const upload: MoneyUpload = {
  id: 'upload-1',
  project_id: 'project-1',
  feature_id: 'boulder-1',
  original_filename: 'tiny-boulder.jpg',
  title: 'Tiny Boulder Topo',
  comments: 'First comment',
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

const featureBase = {
  project_id: 'project-1',
  geojson: { type: 'Point', coordinates: [-121.47, 47.7] } as MoneyGeoJSON,
  style: {},
  properties: {},
  sort_order: 0,
  created_by: 'user-1',
  updated_by: 'user-1',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
};

const problem: MoneyCragNode = {
  feature: { ...featureBase, id: 'problem-1', parent_feature_id: 'boulder-1', feature_type: 'problem', title: 'Sit Start', status: 'project', properties: {} } as MoneyFeature,
  children: null,
  boulders: null,
  problems: null,
};

const boulder: MoneyCragNode = {
  feature: { ...featureBase, id: 'boulder-1', feature_type: 'boulder', title: 'Tiny Boulder', status: 'scouted' } as MoneyFeature,
  children: null,
  boulders: null,
  problems: [problem],
};

const root: MoneyCragNode = {
  feature: { ...featureBase, id: 'area-1', feature_type: 'area', title: 'Money Creek', status: 'active' } as MoneyFeature,
  children: null,
  boulders: [boulder],
  problems: null,
};

describe('PhotoLightbox metadata layout and editing', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders photo details in a separate side panel, not over the photo stage', async () => {
    render(<PhotoLightbox item={{ upload, contextLabel: 'Money Creek / tiny boulder' }} canWrite onClose={vi.fn()} />);

    expect(await screen.findByRole('dialog', { name: 'Tiny Boulder Topo' })).toBeTruthy();
    const stage = document.querySelector('.money-photo-lightbox-stage');
    const panel = document.querySelector('.money-photo-details-panel');
    expect(stage).toBeTruthy();
    expect(panel).toBeTruthy();
    expect(stage?.contains(panel)).toBe(false);
    expect(screen.getByText('First comment')).toBeTruthy();
  });

  it('renders topo controls in the side details panel instead of the photo stage', async () => {
    render(<PhotoLightbox item={{ upload, startInTopoEdit: true }} root={root} boulder={boulder} canWrite onSaveTopo={vi.fn()} onClose={vi.fn()} />);

    expect(await screen.findByRole('region', { name: 'Photo topo editor' })).toBeTruthy();
    const stage = document.querySelector('.money-photo-lightbox-stage') as HTMLElement;
    const slot = document.querySelector('.money-topo-controls-slot');
    const toolbar = document.querySelector('.money-topo-toolbar');
    expect(stage).toBeTruthy();
    expect(slot).toBeTruthy();
    expect(toolbar).toBeTruthy();
    expect(stage.contains(toolbar)).toBe(false);
    expect(slot?.contains(toolbar)).toBe(true);
    expect(stage.style.minHeight).toBe('0');
    expect(stage.style.minWidth).toBe('0');
    expect(stage.style.overflow).toBe('auto');
  });

  it('keeps the topo edit stage inside a definite scrollable lightbox body', async () => {
    render(<PhotoLightbox item={{ upload, startInTopoEdit: true }} root={root} boulder={boulder} canWrite onSaveTopo={vi.fn()} onClose={vi.fn()} />);

    const dialog = await screen.findByRole('dialog', { name: 'Tiny Boulder Topo' }) as HTMLElement;
    const content = document.querySelector('.money-photo-lightbox-content') as HTMLElement;
    const stage = document.querySelector('.money-photo-lightbox-stage') as HTMLElement;
    const panel = document.querySelector('.money-photo-details-panel') as HTMLElement;

    expect(dialog.style.height).toBe('88vh');
    expect(dialog.style.maxHeight).toBe('900px');
    expect(dialog.getAttribute('style')).toContain('grid-template-rows: auto minmax(0,1fr)');
    expect(content.style.minHeight).toBe('0');
    expect(content.style.overflow).toBe('hidden');
    expect(content.style.gridTemplateColumns.replace(/\s/g, '')).toBe('minmax(0,1fr)minmax(280px,340px)');
    expect(stage.style.overflow).toBe('auto');
    expect(panel.style.minHeight).toBe('0');
    expect(panel.style.overflowY).toBe('auto');
  });

  it('saves edited photo title and comments', async () => {
    const onUpdateMetadata = vi.fn().mockResolvedValue({ ...upload, title: 'New Topo', comments: 'Better notes' });
    render(<PhotoLightbox item={{ upload }} canWrite onUpdateMetadata={onUpdateMetadata} onClose={vi.fn()} />);

    fireEvent.click(screen.getByRole('button', { name: 'Edit photo details' }));
    fireEvent.change(screen.getByLabelText('Photo title'), { target: { value: 'New Topo' } });
    fireEvent.change(screen.getByLabelText('Photo comments'), { target: { value: 'Better notes' } });
    fireEvent.click(screen.getByRole('button', { name: /Save/ }));

    await waitFor(() => expect(onUpdateMetadata).toHaveBeenCalledWith(upload, { title: 'New Topo', comments: 'Better notes' }));
    expect(await screen.findByText('Better notes')).toBeTruthy();
  });
});
