import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MoneyMap } from '../MoneyMap';
import { MoneyFeature, MoneyProject } from '../../../types/money';

type DeckProps = {
  viewState: { longitude: number; latitude: number; zoom: number; pitch: number; bearing: number };
  layers: Array<{ id: string; props: Record<string, unknown> }>;
  children: React.ReactNode;
};

let lastDeckProps: DeckProps | null = null;

vi.mock('@deck.gl/react', () => ({
  default: (props: DeckProps) => {
    lastDeckProps = props;
    return <div data-testid="deck-gl" data-longitude={props.viewState.longitude} data-latitude={props.viewState.latitude} data-zoom={props.viewState.zoom}>{props.children}</div>;
  },
}));

vi.mock('@deck.gl/layers', () => {
  class MockLayer {
    id: string;
    props: Record<string, unknown>;

    constructor(props: Record<string, unknown>) {
      this.id = String(props.id);
      this.props = props;
    }
  }

  return {
    PathLayer: MockLayer,
    PolygonLayer: MockLayer,
    ScatterplotLayer: MockLayer,
  };
});

vi.mock('react-map-gl/maplibre', () => ({
  Map: () => <div data-testid="maplibre-map" />,
}));

vi.mock('maplibre-gl/dist/maplibre-gl.css', () => ({}));

const project: MoneyProject = {
  id: 'project-1',
  slug: 'money-creek',
  name: 'Money Creek',
  center_lat: 47.7,
  center_lon: -121.46,
  default_zoom: 14,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
};

const selectedCoordinates: [number, number] = [-121.444, 47.722];

const selectedPin: MoneyFeature = {
  id: 'selected-pin',
  project_id: project.id,
  feature_type: 'poi',
  title: 'Selected Pin',
  status: 'active',
  geojson: { type: 'Point', coordinates: selectedCoordinates },
  style: {},
  properties: {},
  created_by: 'user-1',
  updated_by: 'user-1',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
};

describe('MoneyMap integration state', () => {
  beforeEach(() => {
    lastDeckProps = null;
    localStorage.clear();
  });

  it('focuses the viewport on the selected pin and exposes high-priority map controls', async () => {
    render(
      <MoneyMap
        project={project}
        features={[selectedPin]}
        selectedFeatureId={selectedPin.id}
        drawingType={null}
        draftPoints={[]}
        focusMode
        onSelectFeature={vi.fn()}
        onAddDraftPoint={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(lastDeckProps?.viewState.longitude).toBe(selectedCoordinates[0]);
      expect(lastDeckProps?.viewState.latitude).toBe(selectedCoordinates[1]);
      expect(lastDeckProps?.viewState.zoom).toBeGreaterThanOrEqual(16);
    });

    const controls = screen.getByTestId('money-map-controls');
    expect(controls.className).toContain('z-40');
    expect(screen.getByText('Selected Pin')).toBeTruthy();
  });

  it('keeps the selected feature in visible point layer data', () => {
    render(
      <MoneyMap
        project={project}
        features={[selectedPin]}
        selectedFeatureId={selectedPin.id}
        drawingType={null}
        draftPoints={[]}
        focusMode
        onSelectFeature={vi.fn()}
        onAddDraftPoint={vi.fn()}
      />
    );

    const pointsLayer = lastDeckProps?.layers.find(layer => layer.id === 'money-points');
    expect(pointsLayer).toBeTruthy();
    expect(pointsLayer?.props.data).toMatchObject([{ feature: { id: selectedPin.id }, position: selectedCoordinates }]);
  });
});
