export const T = {
  app: '#17110F', surf: '#221A17', surf2: '#2B211D', inset: '#1B1411', raise: '#322620',
  ink: '#EEE1D3', mut: '#A48F80', faint: '#766150', line: '#3A2C24', line2: '#473630',
  accent: '#AEB974', onAccent: '#1C160E', accentSoft: 'rgba(174,185,116,0.16)', accentDim: 'rgba(174,185,116,0.5)',
  blue: '#86A0B6', blueSoft: 'rgba(134,160,182,0.16)', av: '#B0937D', warn: '#D38A52', danger: '#CD6A5A',
  shadow: '0 12px 38px rgba(0,0,0,0.5)', shadowSm: '0 2px 10px rgba(0,0,0,0.42)', mapLabelBg: 'rgba(24,17,14,0.86)',
  font: 'Figtree, system-ui, sans-serif', mono: 'Space Mono, ui-monospace, monospace',
  map: {
    bg1: '#2E251E', bg2: '#241B16', basin: '#33271F', contour: '#7C6149', ridge: '#9E7C58', grid: '#4A382B',
    creek: '#5C8C92', trail: '#B98050', road: '#9A8266', forestOuter: '#2C3622', forestMid: '#3A4A2C', forestIn: '#4C5A38',
    talus: '#6B6052', rock: '#827564', slot: '#2A201B', slotStripe: '#3B2D24', slotText: '#998468',
  },
} as const;

export const DEV = {
  order: ['scouted', 'needs-work', 'cleaning', 'established'] as const,
  meta: {
    scouted: { label: 'Scouted', short: 'Scouted', c: '#A48F80', bg: 'rgba(164,143,128,0.18)', desc: 'Found and marked — not worked yet' },
    'needs-work': { label: 'Needs cleaning', short: 'Needs work', c: '#D38A52', bg: 'rgba(211,138,82,0.18)', desc: 'Flagged for cleanup — not started' },
    cleaning: { label: 'Cleaning', short: 'Cleaning', c: '#86A0B6', bg: 'rgba(134,160,182,0.18)', desc: 'Actively being developed' },
    established: { label: 'Established', short: 'Established', c: '#AEB974', bg: 'rgba(174,185,116,0.18)', desc: 'Clean and ready to climb' },
  },
} as const;

export const PTYPES = ['Slab', 'Vertical', 'Overhang', 'Roof', 'Arête', 'Dihedral', 'Compression', 'Crimpy', 'Slopers', 'Pinchy', 'Pockets', 'Jugs', 'Dynamic', 'Technical', 'Powerful', 'Mantel', 'Highball', 'Traverse', 'Lowball', 'Stemming'];
