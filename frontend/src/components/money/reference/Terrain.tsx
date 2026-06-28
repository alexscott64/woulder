import { T, } from './theme';

const W = 1000, H = 680;
function mulberry32(a: number) { return function () { a |= 0; a = (a + 0x6d2b79f5) | 0; let t = Math.imul(a ^ (a >>> 15), 1 | a); t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t; return ((t ^ (t >>> 14)) >>> 0) / 4294967296; }; }
function smoothClosed(points: number[][]) { const p = points, n = p.length; let d = `M ${p[0][0].toFixed(1)} ${p[0][1].toFixed(1)}`; for (let i = 0; i < n; i += 1) { const p0 = p[(i - 1 + n) % n], p1 = p[i], p2 = p[(i + 1) % n], p3 = p[(i + 2) % n]; d += ` C ${(p1[0] + (p2[0] - p0[0]) / 6).toFixed(1)} ${(p1[1] + (p2[1] - p0[1]) / 6).toFixed(1)}, ${(p2[0] - (p3[0] - p1[0]) / 6).toFixed(1)} ${(p2[1] - (p3[1] - p1[1]) / 6).toFixed(1)}, ${p2[0].toFixed(1)} ${p2[1].toFixed(1)}`; } return d + ' Z'; }
function hill(cx: number, cy: number, baseR: number, irregularity: number, count: number, seed: number, scales: number[]) { const rng = mulberry32(seed); const factors = Array.from({ length: count }, () => 1 + (rng() - 0.5) * irregularity); return scales.map(s => smoothClosed(factors.map((f, i) => { const ang = (i / count) * Math.PI * 2; const r = baseR * s * f; return [cx + Math.cos(ang) * r, cy + Math.sin(ang) * r]; }))); }
const SC = [1, 0.84, 0.68, 0.53, 0.39, 0.26];
const HILLS = [hill(360, 272, 172, 0.34, 13, 7, SC), hill(720, 204, 134, 0.4, 12, 21, SC), hill(560, 470, 106, 0.46, 11, 88, [1, 0.78, 0.55, 0.32])];
const CREEK = 'M -30 612 C 150 580, 250 486, 372 512 C 486 536, 556 470, 656 432 C 772 388, 880 430, 1040 372';
function lerp(a: number, b: number, t: number) { return a + (b - a) * t; }
function hex(c: string) { return [parseInt(c.slice(1, 3), 16), parseInt(c.slice(3, 5), 16), parseInt(c.slice(5, 7), 16)]; }
function mix(c1: string, c2: string, t: number) { const a = hex(c1), b = hex(c2); return `rgb(${Math.round(lerp(a[0], b[0], t))},${Math.round(lerp(a[1], b[1], t))},${Math.round(lerp(a[2], b[2], t))})`; }
const SLOPE = ['#3E7A4E', '#6E9C4E', '#C9B84A', '#D08A3C', '#C8572F'];
function slopeColor(t: number) { const x = t * (SLOPE.length - 1); const i = Math.min(SLOPE.length - 2, Math.floor(x)); return mix(SLOPE[i], SLOPE[i + 1], x - i); }

export function Terrain({ base, contours }: { base: string; contours: boolean }) {
  const p = T.map;
  return <svg viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="none" width={W} height={H} style={{ display: 'block' }}>
    <defs><linearGradient id="terr-bg-react" x1="0" y1="0" x2="0.3" y2="1"><stop offset="0" stopColor={base === 'satellite' ? '#222A18' : p.bg1} /><stop offset="1" stopColor={base === 'satellite' ? '#1A2012' : p.bg2} /></linearGradient></defs>
    <rect x="0" y="0" width={W} height={H} fill="url(#terr-bg-react)" />
    <path d={`${CREEK} L ${W + 40} ${H} L -30 ${H} Z`} fill={p.basin} opacity={base === 'satellite' ? 0.7 : 0.5} />
    {base === 'satellite' && HILLS.map((rings, hi) => rings.map((d, ri) => { const t = ri / (rings.length - 1); const col = t < 0.45 ? mix(p.forestOuter, p.forestMid, t / 0.45) : t < 0.78 ? mix(p.forestMid, p.talus, (t - 0.45) / 0.33) : mix(p.talus, p.rock, (t - 0.78) / 0.22); return <path key={`${hi}-${ri}`} d={d} fill={col} stroke="rgba(0,0,0,0.15)" strokeWidth="0.6" />; }))}
    {base === 'slope' && HILLS.map((rings, hi) => rings.map((d, ri) => <path key={`${hi}-${ri}`} d={d} fill={slopeColor(ri / (rings.length - 1))} opacity="0.82" />))}
    {base === 'topo' && <g stroke={p.grid} strokeWidth="0.7" opacity="0.5">{[0, 1, 2, 3, 4, 5].map(i => <line key={`v${i}`} x1={i * W / 5} y1="0" x2={i * W / 5} y2={H} />)}{[0, 1, 2, 3, 4].map(i => <line key={`h${i}`} x1="0" y1={i * H / 4} x2={W} y2={i * H / 4} />)}</g>}
    {contours && (base === 'stylized' || base === 'topo') && HILLS.map((rings, hi) => rings.map((d, ri) => <path key={`c${hi}-${ri}`} d={d} fill="none" stroke={p.contour} strokeWidth={base === 'topo' ? (ri % 2 === 0 ? 1.5 : 0.8) : (ri === 0 ? 1.6 : 1.1)} opacity={base === 'topo' ? (ri % 2 === 0 ? 0.85 : 0.5) : 0.32 + ri * 0.11} />))}
    {contours && base === 'stylized' && HILLS.map((rings, hi) => <path key={`idx${hi}`} d={rings[0]} fill="none" stroke={p.ridge} strokeWidth="2" opacity="0.5" />)}
    {contours && (base === 'satellite' || base === 'slope') && HILLS.map((rings, hi) => rings.map((d, ri) => <path key={`o${hi}-${ri}`} d={d} fill="none" stroke="rgba(20,16,10,0.35)" strokeWidth="0.7" opacity="0.7" />))}
  </svg>;
}
