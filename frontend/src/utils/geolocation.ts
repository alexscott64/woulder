/**
 * Geolocation utilities for automatic area selection
 */

interface GeolocationResult {
  latitude: number;
  longitude: number;
}

// Area center coordinates (approximate center of climbing locations in each area)
const AREA_COORDINATES: Record<number, { lat: number; lon: number; name: string }> = {
  1: { lat: 47.75, lon: -121.5, name: 'Pacific Northwest' },   // Central Cascades
  2: { lat: 34.1, lon: -117.0, name: 'Southern California' },  // Between Joshua Tree and LA
};

/**
 * Get user's approximate location from IP address using ip-api.com (free, no key required)
 */
export async function getUserLocation(): Promise<GeolocationResult | null> {
  try {
    const response = await fetch('http://ip-api.com/json/?fields=lat,lon');
    if (!response.ok) {
      console.warn('Failed to fetch user location');
      return null;
    }
    const data = await response.json();
    return {
      latitude: data.lat,
      longitude: data.lon,
    };
  } catch (error) {
    console.error('Error fetching user location:', error);
    return null;
  }
}

/**
 * Calculate distance between two coordinates using Haversine formula (in miles)
 */
function calculateDistance(lat1: number, lon1: number, lat2: number, lon2: number): number {
  const R = 3959; // Earth's radius in miles
  const dLat = toRadians(lat2 - lat1);
  const dLon = toRadians(lon2 - lon1);

  const a =
    Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos(toRadians(lat1)) * Math.cos(toRadians(lat2)) *
    Math.sin(dLon / 2) * Math.sin(dLon / 2);

  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  return R * c;
}

function toRadians(degrees: number): number {
  return degrees * (Math.PI / 180);
}

/**
 * Find the closest area based on user's location
 */
export function findClosestArea(userLat: number, userLon: number): number | null {
  let closestAreaId: number | null = null;
  let minDistance = Infinity;

  for (const [areaIdStr, coords] of Object.entries(AREA_COORDINATES)) {
    const distance = calculateDistance(userLat, userLon, coords.lat, coords.lon);
    if (distance < minDistance) {
      minDistance = distance;
      closestAreaId = parseInt(areaIdStr, 10);
    }
  }

  return closestAreaId;
}

/**
 * Get the closest area to the user's IP location
 */
export async function getClosestAreaFromIP(): Promise<number | null> {
  const location = await getUserLocation();
  if (!location) {
    return null;
  }

  const closestArea = findClosestArea(location.latitude, location.longitude);
  console.log(`User location: ${location.latitude}, ${location.longitude} -> Closest area: ${closestArea}`);
  return closestArea;
}
