import { MoneyPosition } from '../../../types/money';

export interface ParsedCoordinates {
  latitude: number;
  longitude: number;
  position: MoneyPosition;
}

export function parseCoordinateSearch(input: string): ParsedCoordinates | null {
  const text = input.trim();
  if (!text) return null;

  const hemisphere = parseHemispherePair(text);
  if (hemisphere) return hemisphere;

  const numbers = text.match(/[+-]?\d+(?:\.\d+)?/g)?.map(Number) ?? [];
  if (numbers.length !== 2 || numbers.some(value => !Number.isFinite(value))) return null;

  const [first, second] = numbers;
  if (isLatitude(first) && isLongitude(second)) return toParsed(first, second);
  if (!isLatitude(first) && isLongitude(first) && isLatitude(second)) return toParsed(second, first);
  return null;
}

export function validLatitude(value: number): boolean {
  return Number.isFinite(value) && value >= -90 && value <= 90;
}

export function validLongitude(value: number): boolean {
  return Number.isFinite(value) && value >= -180 && value <= 180;
}

function parseHemispherePair(input: string): ParsedCoordinates | null {
  const matches = [...input.matchAll(/([+-]?\d+(?:\.\d+)?)\s*°?\s*([NSEW])/gi)];
  if (matches.length !== 2) return null;

  let latitude: number | null = null;
  let longitude: number | null = null;
  for (const match of matches) {
    const raw = Number(match[1]);
    const hemi = match[2].toUpperCase();
    if (!Number.isFinite(raw)) return null;
    const absolute = Math.abs(raw);
    if (hemi === 'N' || hemi === 'S') latitude = hemi === 'S' ? -absolute : absolute;
    if (hemi === 'E' || hemi === 'W') longitude = hemi === 'W' ? -absolute : absolute;
  }
  return latitude == null || longitude == null ? null : toParsed(latitude, longitude);
}

function toParsed(latitude: number, longitude: number): ParsedCoordinates | null {
  if (!validLatitude(latitude) || !validLongitude(longitude)) return null;
  return { latitude, longitude, position: [longitude, latitude] };
}

function isLatitude(value: number): boolean {
  return value >= -90 && value <= 90;
}

function isLongitude(value: number): boolean {
  return value >= -180 && value <= 180;
}
