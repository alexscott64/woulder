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

const secondUpload: MoneyUpload = {
  ...upload,
  id: 'upload-2',
  original_filename: 'second.jpeg',
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
    Object.defineProperty(URL, 'createObjectURL', { value: vi.fn((file: File) => `blob:${file.name}`), configurable: true });
    Object.defineProperty(URL, 'revokeObjectURL', { value: vi.fn(), configurable: true });
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

  it('appends multiple selected photos and files to the pending attachment list', () => {
    const onSubmit = vi.fn();
    render(<NoteComposer root={area} area={area} boulder={boulder} mobile={false} onClose={vi.fn()} onSubmit={onSubmit} />);

    const photoInput = screen.getByLabelText('Photos') as HTMLInputElement;
    const fileInput = screen.getByLabelText('Files') as HTMLInputElement;
    const firstPhoto = new File(['one'], 'one.jpg', { type: 'image/jpeg' });
    const secondPhoto = new File(['two'], 'two.jpg', { type: 'image/jpeg' });
    const topoFile = new File(['topo'], 'topo.pdf', { type: 'application/pdf' });

    expect(photoInput.multiple).toBe(true);
    expect(fileInput.multiple).toBe(true);
    fireEvent.change(photoInput, { target: { files: [firstPhoto, secondPhoto] } });
    fireEvent.change(fileInput, { target: { files: [topoFile] } });
    fireEvent.click(screen.getByText('Post note'));

    expect(screen.getByAltText('one.jpg').getAttribute('src')).toBe('blob:one.jpg');
    expect(screen.getByAltText('two.jpg').getAttribute('src')).toBe('blob:two.jpg');
    expect(screen.getByText('topo.pdf')).toBeTruthy();
    expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({
      files: [
        expect.objectContaining({ file: firstPhoto, kind: 'photo', blockKey: expect.any(String) }),
        expect.objectContaining({ file: secondPhoto, kind: 'photo', blockKey: expect.any(String) }),
        expect.objectContaining({ file: topoFile, kind: 'file', blockKey: expect.any(String) }),
      ],
      blocks: [
        expect.objectContaining({ kind: 'photo', name: 'one.jpg', url: 'blob:one.jpg', metadata: { local_block_key: expect.any(String) } }),
        expect.objectContaining({ kind: 'photo', name: 'two.jpg', url: 'blob:two.jpg', metadata: { local_block_key: expect.any(String) } }),
        expect.objectContaining({ kind: 'file', name: 'topo.pdf', metadata: { local_block_key: expect.any(String) } }),
      ],
    }));
  });

  it('edits multiple existing uploads and new files without duplicates', async () => {
    const onSubmit = vi.fn();
    const multiNote: MoneyNote = { ...note, blocks: [{ kind: 'photo', upload_id: 'upload-1', name: 'tiny.jpeg' }, { kind: 'photo', upload_id: 'upload-2', name: 'second.jpeg' }] };
    render(<NoteComposer root={area} area={area} boulder={boulder} initialNote={multiNote} uploads={[upload, secondUpload]} mobile={false} onClose={vi.fn()} onSubmit={onSubmit} />);

    await screen.findByAltText('tiny.jpeg');
    await screen.findByAltText('second.jpeg');
    fireEvent.click(screen.getByLabelText('Remove attachment tiny.jpeg'));
    const newPhoto = new File(['new'], 'new.jpg', { type: 'image/jpeg' });
    fireEvent.change(screen.getByLabelText('Photos'), { target: { files: [newPhoto] } });
    fireEvent.click(screen.getByText('Save note'));

    expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({
      files: [expect.objectContaining({ file: newPhoto, kind: 'photo', blockKey: expect.any(String) })],
      blocks: [
        { kind: 'photo', upload_id: 'upload-2', name: 'second.jpeg' },
        expect.objectContaining({ kind: 'photo', name: 'new.jpg', url: 'blob:new.jpg', metadata: { local_block_key: expect.any(String) } }),
      ],
    }));
  });

  it('opens a sketch drawing pad and saves vector sketch metadata', () => {
    const onSubmit = vi.fn();
    render(<NoteComposer root={area} area={area} boulder={boulder} mobile={false} onClose={vi.fn()} onSubmit={onSubmit} />);

    fireEvent.click(screen.getByRole('button', { name: 'Sketch' }));
    const pad = screen.getByTestId('sketch-drawing-pad');
    Object.defineProperty(pad, 'getBoundingClientRect', { value: () => ({ left: 0, top: 0, width: 200, height: 100 }), configurable: true });
    fireEvent.pointerDown(pad, { clientX: 20, clientY: 10, pointerId: 1 });
    fireEvent.pointerMove(pad, { clientX: 100, clientY: 50, pointerId: 1 });
    fireEvent.pointerUp(pad, { clientX: 100, clientY: 50, pointerId: 1 });
    fireEvent.click(screen.getByText('Save sketch'));
    fireEvent.click(screen.getByText('Post note'));

    expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({
      files: [],
      blocks: [expect.objectContaining({ kind: 'sketch', metadata: expect.objectContaining({ vector_schema: 'money-sketch-v1', sketchpad: expect.objectContaining({ strokes: expect.any(Array) }) }) })],
    }));
  });
});
