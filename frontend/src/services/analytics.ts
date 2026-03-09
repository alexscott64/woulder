/**
 * Lightweight analytics tracking client for Woulder.
 * Tracks page views, user interactions, and feature usage
 * for the /iglooghost analytics dashboard.
 *
 * Privacy-friendly: No cookies, fingerprint uses only screen+timezone+language.
 * Events are batched and sent every 5 seconds to minimize HTTP requests.
 */

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';
const ANALYTICS_URL = `${API_BASE}/analytics`;
const BATCH_INTERVAL = 5000; // 5 seconds
const HEARTBEAT_INTERVAL = 30000; // 30 seconds

// --- Types ---

interface TrackEvent {
  event_type: string;
  event_name: string;
  page_path?: string;
  metadata?: Record<string, unknown>;
}

// --- State ---

let sessionId: string | null = null;
let visitorId: string | null = null;
let eventBuffer: TrackEvent[] = [];
let batchTimer: ReturnType<typeof setInterval> | null = null;
let heartbeatTimer: ReturnType<typeof setInterval> | null = null;
let initialized = false;

// --- Utilities ---

/** Generate a UUID v4 */
function generateUUID(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

/** Generate a privacy-friendly visitor fingerprint */
function generateVisitorId(): string {
  const components = [
    screen.width,
    screen.height,
    Intl.DateTimeFormat().resolvedOptions().timeZone,
    navigator.language,
    screen.colorDepth,
    new Date().getTimezoneOffset(),
  ].join('|');

  // Simple hash
  let hash = 0;
  for (let i = 0; i < components.length; i++) {
    const char = components.charCodeAt(i);
    hash = ((hash << 5) - hash + char) | 0;
  }
  return Math.abs(hash).toString(36);
}

/** Detect device type from screen width */
function getDeviceType(): string {
  const width = screen.width;
  if (width < 768) return 'mobile';
  if (width < 1024) return 'tablet';
  return 'desktop';
}

/** Detect browser from user agent */
function getBrowser(): string {
  const ua = navigator.userAgent;
  if (ua.includes('Firefox')) return 'Firefox';
  if (ua.includes('Edg')) return 'Edge';
  if (ua.includes('Chrome')) return 'Chrome';
  if (ua.includes('Safari')) return 'Safari';
  if (ua.includes('Opera') || ua.includes('OPR')) return 'Opera';
  return 'Other';
}

/** Detect OS from user agent */
function getOS(): string {
  const ua = navigator.userAgent;
  if (ua.includes('Windows')) return 'Windows';
  if (ua.includes('Mac OS')) return 'macOS';
  if (ua.includes('Linux')) return 'Linux';
  if (ua.includes('Android')) return 'Android';
  if (ua.includes('iPhone') || ua.includes('iPad')) return 'iOS';
  return 'Other';
}

/** Safe fetch that silently fails (analytics should never break the app) */
async function safeFetch(url: string, options: RequestInit): Promise<void> {
  try {
    await fetch(url, options);
  } catch {
    // Silently ignore analytics failures
  }
}

// --- Core Functions ---

/** Initialize analytics tracking. Call once on app mount. */
export function initAnalytics(): void {
  if (initialized) return;
  initialized = true;

  // Get or create session ID (persists for tab lifetime)
  sessionId = sessionStorage.getItem('woulder_session_id');
  if (!sessionId) {
    sessionId = generateUUID();
    sessionStorage.setItem('woulder_session_id', sessionId);
  }

  // Generate visitor fingerprint
  visitorId = generateVisitorId();

  // Create session on backend
  const referrer = document.referrer || null;
  safeFetch(`${ANALYTICS_URL}/session`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      session_id: sessionId,
      visitor_id: visitorId,
      referrer,
      device_type: getDeviceType(),
      browser: getBrowser(),
      os: getOS(),
      screen_width: screen.width,
      screen_height: screen.height,
      user_agent: navigator.userAgent,
    }),
  });

  // Start batch event sender
  batchTimer = setInterval(flushEvents, BATCH_INTERVAL);

  // Start heartbeat
  heartbeatTimer = setInterval(sendHeartbeat, HEARTBEAT_INTERVAL);

  // Flush on page unload
  window.addEventListener('beforeunload', flushEventsSync);

  // Flush on visibility change (tab hidden)
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'hidden') {
      flushEventsSync();
    }
  });
}

/** Track an analytics event. Events are buffered and sent in batches. */
export function trackEvent(
  eventType: string,
  eventName: string,
  metadata?: Record<string, unknown>
): void {
  if (!initialized || !sessionId) return;

  eventBuffer.push({
    event_type: eventType,
    event_name: eventName,
    page_path: window.location.pathname + window.location.hash,
    metadata: metadata || {},
  });
}

/** Convenience: Track a page view */
export function trackPageView(pageName: string, metadata?: Record<string, unknown>): void {
  trackEvent('page_view', pageName, metadata);
}

/** Convenience: Track a location view */
export function trackLocationView(
  locationId: number,
  locationName: string,
  areaId?: number
): void {
  trackEvent('location_view', 'location_expand', {
    location_id: locationId,
    location_name: locationName,
    ...(areaId !== undefined && { area_id: areaId }),
  });
}

/** Convenience: Track an area view */
export function trackAreaView(
  areaId: string | number,
  areaName: string,
  eventName: string = 'area_select',
  locationId?: number
): void {
  trackEvent('area_view', eventName, {
    area_id: String(areaId),
    area_name: areaName,
    ...(locationId !== undefined && { location_id: locationId }),
  });
}

/** Convenience: Track a route/boulder view */
export function trackRouteView(
  routeId: string | number,
  routeName: string,
  routeType?: string
): void {
  trackEvent('route_view', 'route_detail', {
    route_id: String(routeId),
    route_name: routeName,
    ...(routeType && { route_type: routeType }),
  });
}

/** Convenience: Track a modal open */
export function trackModalOpen(modalName: string, metadata?: Record<string, unknown>): void {
  trackEvent('modal_open', modalName, metadata);
}

/** Convenience: Track a heat map interaction */
export function trackHeatMapAction(action: string, metadata?: Record<string, unknown>): void {
  trackEvent('heatmap', action, metadata);
}

/** Convenience: Track a search */
export function trackSearch(query: string, resultsCount: number): void {
  trackEvent('search', 'route_search', { query, results_count: resultsCount });
}

/** Convenience: Track a settings change */
export function trackSettingsChange(settingName: string, value: unknown): void {
  trackEvent('settings', settingName, { value });
}

// --- Internal ---

/** Flush buffered events to the backend (async) */
async function flushEvents(): Promise<void> {
  if (eventBuffer.length === 0 || !sessionId) return;

  const events = [...eventBuffer];
  eventBuffer = [];

  await safeFetch(`${ANALYTICS_URL}/events`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      session_id: sessionId,
      events,
    }),
  });
}

/** Flush events synchronously (for beforeunload). Uses sendBeacon for reliability. */
function flushEventsSync(): void {
  if (eventBuffer.length === 0 || !sessionId) return;

  const events = [...eventBuffer];
  eventBuffer = [];

  const payload = JSON.stringify({
    session_id: sessionId,
    events,
  });

  // sendBeacon is more reliable than fetch during page unload
  if (navigator.sendBeacon) {
    navigator.sendBeacon(
      `${ANALYTICS_URL}/events`,
      new Blob([payload], { type: 'application/json' })
    );
  }
}

/** Send heartbeat to keep session alive */
async function sendHeartbeat(): Promise<void> {
  if (!sessionId) return;

  await safeFetch(`${ANALYTICS_URL}/heartbeat`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ session_id: sessionId }),
  });
}

/** Clean up timers (for testing or unmounting) */
export function destroyAnalytics(): void {
  if (batchTimer) clearInterval(batchTimer);
  if (heartbeatTimer) clearInterval(heartbeatTimer);
  window.removeEventListener('beforeunload', flushEventsSync);
  flushEventsSync();
  initialized = false;
  sessionId = null;
  visitorId = null;
  eventBuffer = [];
}
