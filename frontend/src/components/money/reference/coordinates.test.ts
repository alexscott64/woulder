import { describe, expect, it } from 'vitest';
import { parseCoordinateSearch } from './coordinates';

describe('parseCoordinateSearch', () => {
  it('parses comma-separated latitude and longitude', () => {
    expect(parseCoordinateSearch('47.6997, -121.4703')).toEqual({ latitude: 47.6997, longitude: -121.4703, position: [-121.4703, 47.6997] });
  });

  it('parses whitespace-separated latitude and longitude', () => {
    expect(parseCoordinateSearch('47.6997 -121.4703')).toEqual({ latitude: 47.6997, longitude: -121.4703, position: [-121.4703, 47.6997] });
  });

  it('parses unambiguous longitude-latitude input', () => {
    expect(parseCoordinateSearch('-121.4703, 47.6997')).toEqual({ latitude: 47.6997, longitude: -121.4703, position: [-121.4703, 47.6997] });
  });

  it('parses hemisphere coordinates', () => {
    expect(parseCoordinateSearch('47.6997N 121.4703W')).toEqual({ latitude: 47.6997, longitude: -121.4703, position: [-121.4703, 47.6997] });
  });

  it('rejects invalid ranges and non-coordinate search text', () => {
    expect(parseCoordinateSearch('91, -121')).toBeNull();
    expect(parseCoordinateSearch('Money Creek boulder')).toBeNull();
  });
});
