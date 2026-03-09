/**
 * Grade parsing, ordering, and classification utilities for climbing route grades.
 * Mirrors the backend grades package (backend/internal/grades/grades.go).
 *
 * Supported grade families:
 *   - V-scale (bouldering): V0 through V17
 *   - YDS (sport/trad): 5.4 through 5.15d
 *   - Ice (WI): WI1 through WI7
 *   - Alpine Ice (AI): AI1 through AI6
 *   - Mixed (M): M1 through M13
 */

// Grade family identifiers
export const FAMILY_V = 'v' as const;
export const FAMILY_YDS = 'yds' as const;
export const FAMILY_WI = 'wi' as const;
export const FAMILY_AI = 'ai' as const;
export const FAMILY_MIXED = 'mixed' as const;
export const FAMILY_AID = 'aid' as const;

export type GradeFamily = typeof FAMILY_V | typeof FAMILY_YDS | typeof FAMILY_WI | typeof FAMILY_AI | typeof FAMILY_MIXED | typeof FAMILY_AID;

// Ordered grade lists per family
export const V_SCALE_GRADES = [
  'V0', 'V1', 'V2', 'V3', 'V4', 'V5', 'V6', 'V7', 'V8', 'V9',
  'V10', 'V11', 'V12', 'V13', 'V14', 'V15', 'V16', 'V17',
] as const;

export const YDS_GRADES = [
  '5.4', '5.5', '5.6', '5.7', '5.8', '5.9',
  '5.10a', '5.10b', '5.10c', '5.10d',
  '5.11a', '5.11b', '5.11c', '5.11d',
  '5.12a', '5.12b', '5.12c', '5.12d',
  '5.13a', '5.13b', '5.13c', '5.13d',
  '5.14a', '5.14b', '5.14c', '5.14d',
  '5.15a', '5.15b', '5.15c', '5.15d',
] as const;

export const WI_GRADES = [
  'WI1', 'WI2', 'WI3', 'WI4', 'WI5', 'WI6', 'WI7',
] as const;

export const AI_GRADES = [
  'AI1', 'AI2', 'AI3', 'AI4', 'AI5', 'AI6',
] as const;

export const MIXED_GRADES = [
  'M1', 'M2', 'M3', 'M4', 'M5', 'M6', 'M7', 'M8', 'M9',
  'M10', 'M11', 'M12', 'M13',
] as const;

// Order offsets per family — must match backend/internal/grades/grades.go
const OFFSET_V = 0;
const OFFSET_YDS = 100;
const OFFSET_WI = 200;
const OFFSET_MIXED = 300;
const OFFSET_AI = 400;

// Precomputed lookup map: uppercase grade → numeric order
const gradeToOrder: Record<string, number> = {};

V_SCALE_GRADES.forEach((g, i) => { gradeToOrder[g.toUpperCase()] = OFFSET_V + i; });
YDS_GRADES.forEach((g, i) => { gradeToOrder[g.toUpperCase()] = OFFSET_YDS + i; });
WI_GRADES.forEach((g, i) => { gradeToOrder[g.toUpperCase()] = OFFSET_WI + i; });
AI_GRADES.forEach((g, i) => { gradeToOrder[g.toUpperCase()] = OFFSET_AI + i; });
MIXED_GRADES.forEach((g, i) => { gradeToOrder[g.toUpperCase()] = OFFSET_MIXED + i; });

/**
 * Classify a grade string into its grade family.
 * Returns undefined for unrecognized grades.
 */
export function gradeFamily(grade: string): GradeFamily | undefined {
  const g = grade.trim().toUpperCase();
  if (!g) return undefined;

  if (g.startsWith('V') && g.length > 1 && g[1] >= '0' && g[1] <= '9') return FAMILY_V;
  if (g.startsWith('WI')) return FAMILY_WI;
  if (g.startsWith('AI') && g.length > 2 && g[2] >= '0' && g[2] <= '9') return FAMILY_AI;
  if (g.startsWith('M') && g.length > 1 && g[1] >= '0' && g[1] <= '9') return FAMILY_MIXED;
  if (g.startsWith('A') || g.startsWith('C')) return FAMILY_AID;
  if (g.startsWith('5.')) return FAMILY_YDS;
  if (g.startsWith('5') && g.length > 1 && g[1] >= '0' && g[1] <= '9') return FAMILY_YDS;

  return undefined;
}

/**
 * Normalize a grade for lookup (strip +/-, handle slash grades, add 'a' to bare YDS).
 */
function normalizeGrade(g: string): string {
  // Strip +/- suffix
  g = g.replace(/[+-]+$/, '');

  // Handle slash grades: "5.10A/B" → "5.10A"
  const slashIdx = g.indexOf('/');
  if (slashIdx > 0) {
    g = g.substring(0, slashIdx);
  }

  // Handle bare YDS like "5.10" → "5.10A"
  const family = gradeFamily(g);
  if (family === FAMILY_YDS && g.length >= 4) {
    const last = g[g.length - 1];
    if (last >= '0' && last <= '9') {
      const numPart = g.substring(2); // Strip "5."
      if (numPart.length >= 2) {
        g = g + 'A';
      }
    }
  }

  return g;
}

/**
 * Convert a grade string to its numeric sort order.
 * Returns -1 if unrecognized.
 */
export function gradeToOrderNum(grade: string): number {
  const g = grade.trim().toUpperCase();
  if (!g) return -1;

  // Direct lookup
  if (g in gradeToOrder) return gradeToOrder[g];

  // Try normalized
  const normalized = normalizeGrade(g);
  if (normalized in gradeToOrder) return gradeToOrder[normalized];

  return -1;
}

/**
 * Convert a numeric order back to a grade string.
 * Returns undefined if invalid.
 */
export function orderToGrade(order: number): string | undefined {
  if (order >= OFFSET_AI && order < OFFSET_AI + AI_GRADES.length) {
    return AI_GRADES[order - OFFSET_AI];
  }
  if (order >= OFFSET_MIXED && order < OFFSET_MIXED + MIXED_GRADES.length) {
    return MIXED_GRADES[order - OFFSET_MIXED];
  }
  if (order >= OFFSET_WI && order < OFFSET_WI + WI_GRADES.length) {
    return WI_GRADES[order - OFFSET_WI];
  }
  if (order >= OFFSET_YDS && order < OFFSET_YDS + YDS_GRADES.length) {
    return YDS_GRADES[order - OFFSET_YDS];
  }
  if (order >= OFFSET_V && order < OFFSET_V + V_SCALE_GRADES.length) {
    return V_SCALE_GRADES[order - OFFSET_V];
  }
  return undefined;
}

/** Grade scale definition for a route type group */
export interface GradeScale {
  /** Unique key for the scale */
  key: string;
  /** Display label */
  label: string;
  /** Emoji for the scale */
  emoji: string;
  /** Ordered list of grade labels */
  grades: readonly string[];
  /** Numeric order values corresponding to each grade */
  orders: number[];
}

/** Get the grade scales that apply to the given selected route types */
export function getGradeScalesForTypes(selectedTypes: string[]): GradeScale[] {
  const scales: GradeScale[] = [];

  if (selectedTypes.includes('Boulder')) {
    scales.push({
      key: 'boulder',
      label: 'Boulder',
      emoji: '🪨',
      grades: V_SCALE_GRADES,
      orders: V_SCALE_GRADES.map((_, i) => OFFSET_V + i),
    });
  }

  if (selectedTypes.includes('Sport') || selectedTypes.includes('Trad')) {
    scales.push({
      key: 'rope',
      label: selectedTypes.includes('Sport') && selectedTypes.includes('Trad')
        ? 'Sport / Trad'
        : selectedTypes.includes('Sport') ? 'Sport' : 'Trad',
      emoji: '🧗',
      grades: YDS_GRADES,
      orders: YDS_GRADES.map((_, i) => OFFSET_YDS + i),
    });
  }

  if (selectedTypes.includes('Ice')) {
    // Ice scale: WI + AI grades
    const iceGrades = [...WI_GRADES, ...AI_GRADES] as const;
    scales.push({
      key: 'ice',
      label: 'Ice',
      emoji: '🧊',
      grades: iceGrades,
      orders: [
        ...WI_GRADES.map((_, i) => OFFSET_WI + i),
        ...AI_GRADES.map((_, i) => OFFSET_AI + i),
      ],
    });

    // Mixed scale: M grades (separate from ice)
    scales.push({
      key: 'mixed',
      label: 'Mixed',
      emoji: '🔀',
      grades: MIXED_GRADES,
      orders: MIXED_GRADES.map((_, i) => OFFSET_MIXED + i),
    });
  }

  return scales;
}

/** Grade range selection: indices into the scale's grades array */
export interface GradeRangeSelection {
  [scaleKey: string]: [number, number]; // [minIndex, maxIndex]
}

/**
 * Convert the current grade selections + route types into API params.
 * Returns comma-separated integer grade_order values that form an allowlist.
 *
 * When ANY scale is narrowed, we send orders for ALL scales:
 * - Narrowed scales contribute only their selected range
 * - Full-range scales contribute their entire range (so those routes pass through)
 *
 * This ensures that narrowing Boulder to V16-V17 doesn't exclude Ice/Trad routes.
 */
export function gradeRangeToApiParams(
  selections: GradeRangeSelection,
  selectedTypes: string[],
): { gradeOrders?: string } {
  const scales = getGradeScalesForTypes(selectedTypes);
  if (scales.length === 0) return {};

  // First pass: check if any scale is narrowed
  let hasNarrowed = false;
  for (const scale of scales) {
    const sel = selections[scale.key];
    const minIdx = sel ? sel[0] : 0;
    const maxIdx = sel ? sel[1] : scale.grades.length - 1;
    if (minIdx !== 0 || maxIdx !== scale.grades.length - 1) {
      hasNarrowed = true;
      break;
    }
  }

  if (!hasNarrowed) return {};

  // Second pass: collect orders from all scales
  // Narrowed scales → only selected range; full-range scales → all orders
  const allOrders: number[] = [];
  for (const scale of scales) {
    const sel = selections[scale.key];
    const minIdx = sel ? sel[0] : 0;
    const maxIdx = sel ? sel[1] : scale.grades.length - 1;

    for (let i = minIdx; i <= maxIdx; i++) {
      allOrders.push(scale.orders[i]);
    }
  }

  return {
    gradeOrders: allOrders.join(','),
  };
}
