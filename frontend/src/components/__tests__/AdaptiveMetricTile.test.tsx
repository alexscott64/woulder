import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AdaptiveMetricTile, type AdaptiveTileVariant } from '../AdaptiveMetricTile';

/**
 * Tests for AdaptiveMetricTile.
 *
 * The component is the extracted 3rd-column tile from WeatherCard. The
 * critical contract under test is:
 *   - `hot` and `rock` variants render as a clickable `<button>` and invoke
 *     `onClick` (this drives the deep-link into ConditionsModal).
 *   - `snow`, `wet`, `unknown` variants render as a passive `<div>` and do
 *     not respond to clicks.
 *   - aria-label / title pass through correctly.
 *   - The provided icon + label + value + sub are rendered.
 */

const passiveVariants: AdaptiveTileVariant[] = ['snow', 'wet', 'unknown'];

describe('AdaptiveMetricTile', () => {
  it('renders as a <button> for kind="hot" and invokes onClick', () => {
    const onClick = vi.fn();
    render(
      <AdaptiveMetricTile
        kind="hot"
        icon={<svg data-testid="icon" />}
        label="Rock Temp"
        value="92°F"
        onClick={onClick}
        ariaLabel="Open rock temperature and friction details"
      />,
    );

    const btn = screen.getByRole('button', {
      name: /open rock temperature and friction details/i,
    });
    fireEvent.click(btn);
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it('renders as a <button> for kind="rock" and invokes onClick', () => {
    const onClick = vi.fn();
    render(
      <AdaptiveMetricTile
        kind="rock"
        icon={<svg data-testid="icon" />}
        label="Rock Temp"
        value="58°F"
        onClick={onClick}
        ariaLabel="Open rock temperature and friction details"
      />,
    );

    const btn = screen.getByRole('button');
    fireEvent.click(btn);
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it.each(passiveVariants)(
    'renders as a non-button <div> for kind="%s" and does not invoke onClick on click',
    (kind) => {
      const onClick = vi.fn();
      const { container } = render(
        <AdaptiveMetricTile
          kind={kind}
          icon={<svg data-testid="icon" />}
          label="On Ground"
          value='0"'
          onClick={onClick}
        />,
      );

      // No button role present.
      expect(screen.queryByRole('button')).toBeNull();

      // Click the wrapper anyway and confirm onClick is not wired up.
      const wrapper = container.firstChild as HTMLElement;
      expect(wrapper.tagName).toBe('DIV');
      fireEvent.click(wrapper);
      expect(onClick).not.toHaveBeenCalled();
    },
  );

  it('passes aria-label and title through to the button', () => {
    render(
      <AdaptiveMetricTile
        kind="rock"
        icon={<svg data-testid="icon" />}
        label="Rock Temp"
        value="58°F"
        onClick={() => {}}
        ariaLabel="Open rock temperature and friction details"
        title="View rock temperature & friction details"
      />,
    );

    const btn = screen.getByRole('button', {
      name: /open rock temperature and friction details/i,
    });
    expect(btn.getAttribute('title')).toBe(
      'View rock temperature & friction details',
    );
    expect(btn.getAttribute('aria-label')).toBe(
      'Open rock temperature and friction details',
    );
  });

  it('renders the provided icon, label, value, and sub content', () => {
    render(
      <AdaptiveMetricTile
        kind="snow"
        icon={<svg data-testid="snow-icon" />}
        label="On Ground"
        value='1.2"'
        sub="Heavy"
      />,
    );

    expect(screen.getByTestId('snow-icon')).toBeTruthy();
    expect(screen.getByText('On Ground')).toBeTruthy();
    expect(screen.getByText('1.2"')).toBeTruthy();
    expect(screen.getByText('Heavy')).toBeTruthy();
  });
});
