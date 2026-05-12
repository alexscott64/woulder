import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, act } from '@testing-library/react';
import { ConditionsModal } from '../ConditionsModal';
import type { RockTemperatureStatus, WeatherCondition } from '../../types/weather';

/**
 * Tests for the `initialFocus` deep-link prop on ConditionsModal.
 *
 * The prop is what wires the WeatherCard's clickable "Rock Temp" tile to the
 * "Surface temperature & friction" section inside the modal. The behavior we
 * lock in here:
 *   1. When `initialFocus='rock-surface-temperature'` and a rockTempStatus is
 *      provided, the modal opens on the "Rock" tab (not the default "Today").
 *   2. The rock-surface-temperature section is rendered and gets
 *      scrollIntoView() called on it after mount.
 *   3. When no `initialFocus` is provided, the modal opens on the default
 *      "Today" tab and does not call scrollIntoView.
 */

const todayCondition: WeatherCondition = {
  level: 'good',
  reasons: ['Dry conditions', 'Mild temperatures'],
};

const rockTempStatus: RockTemperatureStatus = {
  estimated_surface_temp_f: 58,
  air_temp_f: 55,
  temp_differential_f: 3,
  condition: 'good',
  friction_quality: 'good',
  message: 'Surface temperature is in the ideal range for friction.',
  confidence_score: 0.9,
  rock_type: 'Granite',
};

describe('ConditionsModal - initialFocus', () => {
  let scrollIntoViewSpy: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    scrollIntoViewSpy = vi.fn();
    // happy-dom doesn't implement scrollIntoView; stub it on the prototype so
    // every Element instance inherits the spy.
    Element.prototype.scrollIntoView = scrollIntoViewSpy as unknown as typeof Element.prototype.scrollIntoView;
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('opens on the Rock tab and scrolls the rock-surface-temperature section into view when initialFocus is set', async () => {
    render(
      <ConditionsModal
        locationName="Test Crag"
        rockTempStatus={rockTempStatus}
        todayCondition={todayCondition}
        initialFocus="rock-surface-temperature"
        onClose={() => {}}
      />,
    );

    // The Rock tab content (the "Surface temperature & friction" heading) is
    // only rendered when activeTab === 'rock', so finding it confirms the
    // modal switched tabs based on initialFocus.
    expect(
      screen.getByRole('heading', { name: /surface temperature & friction/i }),
    ).toBeTruthy();

    // The anchor div carries the stable id callers deep-link to.
    const section = document.getElementById('rock-surface-temperature');
    expect(section).not.toBeNull();

    // scrollIntoView fires inside a requestAnimationFrame, so wait for it.
    await waitFor(() => {
      expect(scrollIntoViewSpy).toHaveBeenCalled();
    });

    // Verify it was the rock-surface-temperature element that got scrolled.
    const calledOnEl = scrollIntoViewSpy.mock.instances[0] as HTMLElement;
    expect(calledOnEl).toBe(section);
  });

  it('defaults to the Today tab and does not scroll when initialFocus is omitted', async () => {
    render(
      <ConditionsModal
        locationName="Test Crag"
        rockTempStatus={rockTempStatus}
        todayCondition={todayCondition}
        onClose={() => {}}
      />,
    );

    // Today tab content is visible (the "Contributing Factors:" heading lives
    // inside the today tab body).
    expect(
      screen.getByRole('heading', { name: /contributing factors/i }),
    ).toBeTruthy();

    // Rock-tab heading should NOT be in the DOM since we're on the Today tab.
    expect(
      screen.queryByRole('heading', { name: /surface temperature & friction/i }),
    ).toBeNull();

    // Flush a frame to make sure no deferred scrollIntoView is queued.
    await act(async () => {
      await new Promise((resolve) => requestAnimationFrame(() => resolve(null)));
    });
    expect(scrollIntoViewSpy).not.toHaveBeenCalled();
  });
});
