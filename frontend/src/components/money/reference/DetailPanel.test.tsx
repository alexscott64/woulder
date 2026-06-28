import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { DetailPanel } from './DetailPanel';
import { MoneyCragNode, MoneyFeature, MoneyGeoJSON, MoneyNote, MoneyTrailCategory, MoneyUpload } from '../../../types/money';

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
  feature: { ...featureBase, id: 'area-1', feature_type: 'area', title: 'Money Creek', description: 'Reference crag', status: 'active', properties: { kind: 'Crag', aspect: 'forest' } } as MoneyFeature,
  children: null,
  boulders: [boulder],
  problems: null,
};

const trail: MoneyCragNode = {
  feature: { ...featureBase, id: 'trail-1', feature_type: 'trail', title: 'old trail', status: 'active', geojson: { type: 'LineString', coordinates: [[0, 0], [1, 1]] }, properties: { trail_category: 'connector' } } as MoneyFeature,
  children: null,
  boulders: null,
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

function renderDetailPanel(overrides: Partial<React.ComponentProps<typeof DetailPanel>> = {}) {
  return render(<DetailPanel root={area} area={area} selectedBoulder={boulder} selectedTrail={null} notes={[note]} uploads={[upload]} tab="overview" mobile={false} expanded canWrite canEditArea={false} setExpanded={vi.fn()} setTab={vi.fn()} onEnter={vi.fn()} onSelectBoulder={vi.fn()} onNewArea={vi.fn()} onNewBoulder={vi.fn()} onEditArea={vi.fn()} onDeleteArea={vi.fn()} onSetDev={vi.fn()} onRenameBoulder={vi.fn()} onAddProblem={vi.fn()} onOpenComposer={vi.fn()} onEditNote={vi.fn()} onDeleteNote={vi.fn()} onDeleteUpload={vi.fn()} onUpdateTrail={vi.fn()} onDeleteTrail={vi.fn()} {...overrides} />);
}

describe('DetailPanel uploaded photos', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders boulder uploads instead of default photo placeholders', async () => {
    renderDetailPanel();

    const images = await screen.findAllByAltText('tiny-boulder.jpg');
    await waitFor(() => expect(images[0].getAttribute('src')).toBe('https://example.invalid/photo.jpg'));
    expect(screen.queryByText('topo photo')).toBeNull();
  });

  it('opens a lightbox when a detail photo is clicked', async () => {
    renderDetailPanel();

    const photoButtons = await screen.findAllByRole('button', { name: 'Open photo tiny-boulder.jpg' });
    fireEvent.click(photoButtons[0]);

    expect(await screen.findByRole('dialog', { name: 'tiny-boulder.jpg' })).toBeTruthy();
    expect(screen.getAllByText('Money Creek / tiny boulder').length).toBeGreaterThan(0);
  });
});

describe('DetailPanel boulder management', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('saves boulder name edits from the compact header controls', () => {
    const onRenameBoulder = vi.fn();
    renderDetailPanel({ onRenameBoulder });

    expect(screen.queryByLabelText('Boulder name')).toBeNull();
    expect(screen.queryByText('Save boulder name')).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: 'Edit boulder name' }));
    fireEvent.change(screen.getByLabelText('Boulder name'), { target: { value: 'Tiny Roof' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save boulder name' }));

    expect(onRenameBoulder).toHaveBeenCalledWith(boulder, 'Tiny Roof');
  });

  it('discards boulder name edits without calling rename', () => {
    const onRenameBoulder = vi.fn();
    renderDetailPanel({ onRenameBoulder });

    fireEvent.click(screen.getByRole('button', { name: 'Edit boulder name' }));
    fireEvent.change(screen.getByLabelText('Boulder name'), { target: { value: 'Tiny Roof' } });
    fireEvent.click(screen.getByRole('button', { name: 'Discard boulder name changes' }));

    expect(onRenameBoulder).not.toHaveBeenCalled();
    expect(screen.queryByLabelText('Boulder name')).toBeNull();
    expect(screen.getAllByText('tiny boulder').length).toBeGreaterThan(0);
  });

  it('supports keyboard save and cancel in mobile boulder headers', () => {
    const onRenameBoulder = vi.fn();
    renderDetailPanel({ mobile: true, onRenameBoulder });

    fireEvent.click(screen.getByRole('button', { name: 'Edit boulder name' }));
    const input = screen.getByLabelText('Boulder name');
    fireEvent.change(input, { target: { value: 'Tiny Roof' } });
    fireEvent.keyDown(input, { key: 'Escape' });
    expect(onRenameBoulder).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole('button', { name: 'Edit boulder name' }));
    const nextInput = screen.getByLabelText('Boulder name');
    fireEvent.change(nextInput, { target: { value: 'Tiny Roof' } });
    fireEvent.keyDown(nextInput, { key: 'Enter' });

    expect(onRenameBoulder).toHaveBeenCalledWith(boulder, 'Tiny Roof');
  });
});
 
describe('DetailPanel trail management', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('saves trail label and destination metadata', () => {
    const onUpdateTrail = vi.fn();
    renderDetailPanel({ selectedBoulder: null, selectedTrail: trail, onUpdateTrail });

    fireEvent.change(screen.getByLabelText('Trail label'), { target: { value: 'Trail to Main Wall' } });
    fireEvent.change(screen.getByLabelText('Trail category'), { target: { value: 'trail_to_area' satisfies MoneyTrailCategory } });
    fireEvent.change(screen.getByLabelText('Destination area'), { target: { value: 'area-1' } });
    fireEvent.click(screen.getByText('Save trail'));

    expect(onUpdateTrail).toHaveBeenCalledWith(trail, { title: 'Trail to Main Wall', category: 'trail_to_area', destinationFeatureId: 'area-1', destinationLabel: null });
  });

  it('confirms trail deletion before calling delete handler', () => {
    const onDeleteTrail = vi.fn();
    Object.defineProperty(window, 'confirm', { value: vi.fn().mockReturnValue(true), configurable: true });
    renderDetailPanel({ selectedBoulder: null, selectedTrail: trail, onDeleteTrail });

    fireEvent.click(screen.getByText('Delete trail'));

    expect(window.confirm).toHaveBeenCalledWith('Delete trail old trail?');
    expect(onDeleteTrail).toHaveBeenCalledWith(trail);
  });
});

