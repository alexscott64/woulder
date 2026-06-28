import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { MoneyCragNode, MoneyFeature, MoneyGeoJSON, MoneyUpload } from '../../../types/money';
import { TopoOverlayEditor, TopoOverlaySvg } from './TopoOverlayEditor';

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

describe('TopoOverlayEditor', () => {
  it('renders a large-photo drawing stage with non-overlay responsive toolbar controls', () => {
    const onSave = vi.fn();
    const { container } = render(<TopoOverlayEditor root={root} boulder={boulder} upload={upload} canWrite imageSrc="https://example.invalid/tiny.jpg" onSave={onSave} />);

    expect(screen.getByRole('region', { name: 'Photo topo editor' })).toBeTruthy();
    expect(screen.getByText('Draw topo')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Line' }).getAttribute('aria-pressed')).toBe('true');
    expect(screen.getByRole('button', { name: 'Start' })).toBeTruthy();
    expect((screen.getByRole('button', { name: 'Save' }) as HTMLButtonElement).disabled).toBe(true);
    const toolbar = container.querySelector('.money-topo-toolbar') as HTMLElement;
    expect(toolbar).toBeTruthy();
    expect(toolbar.style.position).toBe('');
    const stage = container.querySelector('.money-topo-stage') as HTMLElement;
    const img = container.querySelector('.money-topo-stage img') as HTMLImageElement;
    expect(img).toBeTruthy();
    expect(stage.style.display).toBe('flex');
    expect(stage.style.minHeight).toBe('0');
    expect(stage.style.minWidth).toBe('0');
    expect(stage.style.overflow).toBe('auto');
    expect(stage.style.alignItems).toBe('center');
    expect(img.style.objectFit).toBe('contain');
    expect(img.style.maxWidth).toBe('100%');
    expect(img.style.maxHeight).toBe('100%');
    expect(img.style.height).toBe('auto');
    expect(img.style.maxHeight).not.toContain('100vh');
    expect((screen.getByRole('region', { name: 'Photo topo editor' }) as HTMLElement).style.overflow).toBe('hidden');
  });

  it('places typed start markers directly on the large photo canvas and saves normalized points', () => {
    const onSave = vi.fn();
    const { container } = render(<TopoOverlayEditor root={root} boulder={boulder} upload={upload} canWrite imageSrc="https://example.invalid/tiny.jpg" onSave={onSave} />);

    expect(screen.getByText(/Draw on the enlarged photo/)).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Start' }));
    const canvas = screen.getByTestId('topo-photo-canvas');
    Object.defineProperty(canvas, 'getBoundingClientRect', { value: () => ({ left: 10, top: 20, width: 200, height: 100 }), configurable: true });

    fireEvent.pointerDown(canvas, { clientX: 50, clientY: 90, pointerId: 1 });
    expect(screen.getAllByTestId('topo-start-marker')).toHaveLength(1);
    expect(container.querySelectorAll('[data-testid="topo-start-marker"] line')).toHaveLength(2);
    expect(container.querySelectorAll('[data-testid="topo-start-marker"] text')).toHaveLength(0);
    fireEvent.click(screen.getByRole('button', { name: 'Left' }));
    fireEvent.pointerDown(canvas, { clientX: 150, clientY: 70, pointerId: 2 });
    expect(screen.getAllByTestId('topo-start-marker')).toHaveLength(2);
    fireEvent.click(screen.getByRole('button', { name: 'Right' }));
    fireEvent.pointerDown(canvas, { clientX: 110, clientY: 60, pointerId: 3 });
    expect(screen.getAllByTestId('topo-start-marker')).toHaveLength(3);
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    expect(onSave).toHaveBeenCalledWith(problem, expect.objectContaining({
      upload_id: 'upload-1',
      photo_id: 'upload-1',
      problem_id: 'problem-1',
      starts: [
        expect.objectContaining({ point: [0.2, 0.7], label: 'X', type: 'generic' }),
        expect.objectContaining({ point: [0.7, 0.5], label: 'L', type: 'left' }),
        expect.objectContaining({ point: [0.5, 0.4], label: 'R', type: 'right' }),
      ],
    }));
  });

  it('uses a single-row photo stage when controls are portaled outside the image area', () => {
    const onSave = vi.fn();
    const portalTarget = document.createElement('div');
    document.body.appendChild(portalTarget);

    const { container, unmount } = render(<TopoOverlayEditor root={root} boulder={boulder} upload={upload} canWrite imageSrc="https://example.invalid/tiny.jpg" controlsPortalTarget={portalTarget} onSave={onSave} />);

    const editor = screen.getByRole('region', { name: 'Photo topo editor' }) as HTMLElement;
    expect(editor.style.gridTemplateRows.replace(/\s/g, '')).toBe('minmax(0,1fr)');
    expect(container.querySelector('.money-topo-toolbar')).toBeNull();
    expect(portalTarget.querySelector('.money-topo-toolbar')).toBeTruthy();

    unmount();
    portalTarget.remove();
  });

  it('renders one saved generic X marker without a duplicate text X', () => {
    const { container } = render(<TopoOverlaySvg overlays={[{ problem, overlay: { id: 'topo-1', upload_id: 'upload-1', problem_id: 'problem-1', color: '#F97316', width: 5, order: 0, paths: [], starts: [{ id: 'start-x', point: [0.1, 0.8], label: 'X', type: 'generic' }] } }]} />);

    expect(screen.getAllByTestId('topo-start-marker')).toHaveLength(1);
    expect(container.querySelectorAll('[data-testid="topo-start-marker"] line')).toHaveLength(2);
    expect(container.querySelectorAll('[data-testid="topo-start-marker"] text')).toHaveLength(0);
  });

  it('renders saved typed start markers in the photo overlay viewer', () => {
    const { container } = render(<TopoOverlaySvg overlays={[{ problem, overlay: { id: 'topo-1', upload_id: 'upload-1', problem_id: 'problem-1', color: '#F97316', width: 5, order: 0, paths: [], starts: [{ id: 'start-l', point: [0.2, 0.7], label: 'Start L' }, { id: 'start-r', point: [0.7, 0.5], label: 'R', type: 'right' }] } }]} />);

    expect(screen.getAllByTestId('topo-start-marker')).toHaveLength(2);
    expect(container.querySelectorAll('[data-testid="topo-start-marker"] line')).toHaveLength(4);
    expect(screen.getByText('L')).toBeTruthy();
    expect(screen.getByText('R')).toBeTruthy();
  });
});
