import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useState } from 'react';
import { ContentView, ReferenceFilters, defaultReferenceFilters } from './ContentViews';
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

const problem: MoneyCragNode = {
  feature: { ...featureBase, id: 'problem-1', feature_type: 'problem', title: 'Tiny Arete', status: 'project', properties: { grade: 'V2', stars: 2, types: ['arete'] } } as MoneyFeature,
  children: null,
  boulders: null,
  problems: null,
};

const boulder: MoneyCragNode = {
  feature: { ...featureBase, id: 'boulder-1', feature_type: 'boulder', title: 'tiny boulder', status: 'scouted' } as MoneyFeature,
  children: null,
  boulders: null,
  problems: [problem],
};

const childProblem: MoneyCragNode = {
  feature: { ...featureBase, id: 'problem-2', feature_type: 'problem', title: 'Moss Traverse', status: 'project', properties: { grade: 'V1', stars: 1, types: ['traverse'] } } as MoneyFeature,
  children: null,
  boulders: null,
  problems: null,
};

const noPhotoBoulder: MoneyCragNode = {
  feature: { ...featureBase, id: 'boulder-2', feature_type: 'boulder', title: 'blank boulder', status: 'established' } as MoneyFeature,
  children: null,
  boulders: null,
  problems: [childProblem],
};

const childArea: MoneyCragNode = {
  feature: { ...featureBase, id: 'area-2', feature_type: 'area', title: 'Upper Forest', description: 'Second sector', status: 'active', properties: { kind: 'Area' } } as MoneyFeature,
  children: null,
  boulders: [noPhotoBoulder],
  problems: null,
};

const area: MoneyCragNode = {
  feature: { ...featureBase, id: 'area-1', feature_type: 'area', title: 'Money Creek', description: 'Reference crag', status: 'active', properties: { kind: 'Crag' } } as MoneyFeature,
  children: [childArea],
  boulders: [boulder],
  problems: null,
};

const trail: MoneyCragNode = {
  feature: { ...featureBase, id: 'trail-1', feature_type: 'trail', title: 'Creek connector', status: 'active', geojson: { type: 'LineString', coordinates: [[-121.5, 47.7], [-121.51, 47.71]] }, properties: { trail_category: 'connector' } } as MoneyFeature,
  children: null,
  boulders: null,
  problems: null,
};

const childTrail: MoneyCragNode = {
  feature: { ...featureBase, id: 'trail-2', feature_type: 'trail', title: 'Upper spur', status: 'active', geojson: { type: 'LineString', coordinates: [[-121.5, 47.7], [-121.51, 47.71]] }, properties: { trail_category: 'trail_to_area', trail_destination_feature_id: 'area-2' } } as MoneyFeature,
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

const problemUpload: MoneyUpload = {
  ...upload,
  id: 'upload-2',
  feature_id: undefined,
  note_id: 'note-2',
  original_filename: 'tiny-arete.jpg',
  created_at: '2026-01-03T00:00:00Z',
};

const childUpload: MoneyUpload = {
  ...upload,
  id: 'upload-3',
  feature_id: 'boulder-2',
  original_filename: 'blank-boulder.jpg',
  created_at: '2026-01-04T00:00:00Z',
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

const problemNote: MoneyNote = {
  ...note,
  id: 'note-2',
  target_type: 'feature',
  target_ref: 'problem-1',
  body: 'Tiny arete photo note',
  blocks: [{ kind: 'photo', upload_id: 'upload-2', name: 'tiny-arete.jpg' }],
  created_at: '2026-01-03T00:00:00Z',
  updated_at: '2026-01-03T00:00:00Z',
};

function renderContent(view: string, options: { notes?: MoneyNote[]; uploads?: MoneyUpload[]; trails?: MoneyCragNode[]; initialFilters?: Partial<ReferenceFilters>; mobile?: boolean; currentAreaId?: string | null } = {}) {
  function Harness() {
    const [filters, setFilters] = useState<ReferenceFilters>({ ...defaultReferenceFilters, ...options.initialFilters });
    return <ContentView view={view} root={area} trails={options.trails ?? []} notes={options.notes ?? []} uploads={options.uploads ?? []} trash={[]} canWrite mobile={options.mobile ?? false} filters={filters} setFilters={setFilters} currentAreaId={options.currentAreaId} openBoulder={vi.fn()} selectTrail={vi.fn()} onOpenComposer={vi.fn()} onEditNote={vi.fn()} onDeleteNote={vi.fn()} onDeleteUpload={vi.fn()} onRestore={vi.fn()} onCreateTrail={vi.fn()} onUpdateTrail={vi.fn()} />;
  }
  return render(<Harness />);
}

describe('ContentView photos', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('opens Photos view uploads with associated note and area/boulder context', async () => {
    renderContent('photos', { notes: [note], uploads: [upload] });

    fireEvent.click(await screen.findByRole('button', { name: 'Open photo tiny-boulder.jpg' }));

    expect(await screen.findByRole('dialog', { name: 'tiny-boulder.jpg' })).toBeTruthy();
    expect(screen.getByText('Tiny boulder note')).toBeTruthy();
    expect(screen.getAllByText('Money Creek / tiny boulder').length).toBeGreaterThan(0);
  });
});

describe('ContentView contextual filters and thumbnails', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('filters boulders by contextual area and photo controls', async () => {
    renderContent('boulders', { notes: [note], uploads: [upload], mobile: true });

    expect(screen.getByText('tiny boulder')).toBeTruthy();
    expect(screen.getByText('blank boulder')).toBeTruthy();
    expect(await screen.findByAltText('tiny boulder')).toBeTruthy();

    fireEvent.change(screen.getByLabelText('Boulders filters area'), { target: { value: 'area-2' } });
    expect(screen.queryByText('tiny boulder')).toBeNull();
    expect(screen.getByText('blank boulder')).toBeTruthy();

    fireEvent.change(screen.getByLabelText('Boulder photo filter'), { target: { value: 'with' } });
    expect(screen.queryByText('blank boulder')).toBeNull();
    expect(screen.getByText('No boulders match these filters.')).toBeTruthy();
  });

  it('defaults Boulders to the current area and exposes dependent sub-area choices', () => {
    renderContent('boulders', { currentAreaId: 'area-2', mobile: true });

    expect(screen.queryByText('tiny boulder')).toBeNull();
    expect(screen.getByText('blank boulder')).toBeTruthy();
    expect(screen.queryByLabelText('Boulders filters sub-area')).toBeNull();
  });

  it('narrows boulders, problems, photos, notes, and trails to the selected sub-area subtree', async () => {
    const childNote: MoneyNote = { ...note, id: 'note-4', target_ref: 'boulder-2', body: 'Upper Forest boulder note', blocks: [], created_at: '2026-01-04T00:00:00Z' };

    renderContent('boulders', { notes: [note, childNote], uploads: [upload, childUpload], trails: [trail, childTrail], mobile: true });
    expect(screen.getByText('tiny boulder')).toBeTruthy();
    expect(screen.getByText('blank boulder')).toBeTruthy();
    fireEvent.change(screen.getByLabelText('Boulders filters sub-area'), { target: { value: 'area-2' } });
    expect(screen.queryByText('tiny boulder')).toBeNull();
    expect(screen.getByText('blank boulder')).toBeTruthy();

    cleanup();
    renderContent('problems', { initialFilters: { subAreaId: 'area-2' }, mobile: true });
    expect(screen.queryByText('Tiny Arete')).toBeNull();
    expect(screen.getByText('Moss Traverse')).toBeTruthy();

    cleanup();
    renderContent('photos', { notes: [note, childNote], uploads: [upload, childUpload], initialFilters: { subAreaId: 'area-2' }, mobile: true });
    expect(screen.queryByRole('button', { name: 'Open photo tiny-boulder.jpg' })).toBeNull();
    expect(await screen.findByRole('button', { name: 'Open photo blank-boulder.jpg' })).toBeTruthy();

    cleanup();
    renderContent('notes', { notes: [note, childNote], uploads: [upload, childUpload], initialFilters: { subAreaId: 'area-2' }, mobile: true });
    expect(screen.queryByText('Tiny boulder note')).toBeNull();
    expect(screen.getByText('Upper Forest boulder note')).toBeTruthy();

    cleanup();
    renderContent('trails', { trails: [trail, childTrail], initialFilters: { subAreaId: 'area-2' }, mobile: true });
    expect(screen.queryByText('Creek connector')).toBeNull();
    expect(screen.getByText('Upper spur')).toBeTruthy();
  });

  it('renders problem thumbnails through associated note uploads', async () => {
    renderContent('problems', { notes: [problemNote], uploads: [problemUpload] });

    expect(screen.getByText('Tiny Arete')).toBeTruthy();
    expect(await screen.findByAltText('Tiny Arete')).toBeTruthy();
  });

  it('filters trails by category while retaining area scope controls', () => {
    const approachTrail: MoneyCragNode = { ...trail, feature: { ...trail.feature, id: 'trail-2', title: 'Main approach', properties: { trail_category: 'approach' } } as MoneyFeature };
    renderContent('trails', { trails: [trail, approachTrail], mobile: true });

    expect(screen.getByText('Creek connector')).toBeTruthy();
    expect(screen.getByText('Main approach')).toBeTruthy();
    expect(screen.getByTestId('reference-filter-rail')).toBeTruthy();

    fireEvent.change(screen.getByLabelText('Trail category filter'), { target: { value: 'approach' } });
    expect(screen.queryByText('Creek connector')).toBeNull();
    expect(screen.getByText('Main approach')).toBeTruthy();
  });

  it('filters notes by photo presence', () => {
    const textNote: MoneyNote = { ...note, id: 'note-3', target_ref: 'area-2', body: 'No photos here', blocks: [], created_at: '2026-01-04T00:00:00Z' };
    renderContent('notes', { notes: [note, textNote], uploads: [upload], mobile: true });

    expect(screen.getByText('Tiny boulder note')).toBeTruthy();
    expect(screen.getByText('No photos here')).toBeTruthy();

    fireEvent.change(screen.getByLabelText('Note photo filter'), { target: { value: 'with' } });
    expect(screen.getByText('Tiny boulder note')).toBeTruthy();
    expect(screen.queryByText('No photos here')).toBeNull();
  });
});

describe('ContentView trails', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('parses an uploaded GPX trail and submits a trail feature payload', async () => {
    const onCreateTrail = vi.fn();
    const file = new File(['<gpx><trk><name>Main approach</name><trkseg><trkpt lat="47.70" lon="-121.50"/><trkpt lat="47.71" lon="-121.51"/></trkseg></trk></gpx>'], 'main-approach.gpx', { type: 'application/gpx+xml' });
    render(<ContentView view="trails" root={area} trails={[]} notes={[]} uploads={[]} trash={[]} canWrite mobile={false} openBoulder={vi.fn()} selectTrail={vi.fn()} onOpenComposer={vi.fn()} onEditNote={vi.fn()} onDeleteNote={vi.fn()} onDeleteUpload={vi.fn()} onRestore={vi.fn()} onCreateTrail={onCreateTrail} onUpdateTrail={vi.fn()} />);

    fireEvent.change(document.querySelector('input[type="file"]') as HTMLInputElement, { target: { files: [file] } });

    await waitFor(() => expect(onCreateTrail).toHaveBeenCalledTimes(1));
    expect(onCreateTrail).toHaveBeenCalledWith(expect.objectContaining({
      title: 'Main approach',
      filename: 'main-approach.gpx',
      sourceFormat: 'gpx',
      pointCount: 2,
      geojson: { type: 'LineString', coordinates: [[-121.5, 47.7], [-121.51, 47.71]] },
    }));
  });

  it('edits trail info from the Trails view', () => {
    const onUpdateTrail = vi.fn();
    render(<ContentView view="trails" root={area} trails={[trail]} notes={[]} uploads={[]} trash={[]} canWrite mobile={false} openBoulder={vi.fn()} selectTrail={vi.fn()} onOpenComposer={vi.fn()} onEditNote={vi.fn()} onDeleteNote={vi.fn()} onDeleteUpload={vi.fn()} onRestore={vi.fn()} onCreateTrail={vi.fn()} onUpdateTrail={onUpdateTrail} />);

    fireEvent.click(screen.getByRole('button', { name: 'Edit trail Creek connector' }));
    fireEvent.change(screen.getByLabelText('Trail label'), { target: { value: 'Trail to Main Wall' } });
    fireEvent.change(screen.getByLabelText('Trail category'), { target: { value: 'trail_to_area' } });
    fireEvent.change(screen.getByLabelText('Destination area'), { target: { value: 'area-1' } });
    fireEvent.click(screen.getByText('Save trail info'));

    expect(onUpdateTrail).toHaveBeenCalledWith(trail, { title: 'Trail to Main Wall', category: 'trail_to_area', destinationFeatureId: 'area-1', destinationLabel: null });
  });
});
