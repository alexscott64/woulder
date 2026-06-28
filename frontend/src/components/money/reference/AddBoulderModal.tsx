import { useEffect, useMemo, useState } from 'react';
import { MoneyCragNode, MoneyDevStatus, MoneyPosition } from '../../../types/money';
import { flattenAreas } from './model';
import { DEV, T } from './theme';
import { validLatitude, validLongitude } from './coordinates';

export interface AddBoulderPayload {
  name: string;
  latitude: number;
  longitude: number;
  position: MoneyPosition;
  parentId: string;
  devStatus: MoneyDevStatus;
}

interface Props {
  root: MoneyCragNode | null;
  initialPosition?: MoneyPosition | null;
  initialParentId?: string | null;
  fallbackParentId?: string | null;
  saving?: boolean;
  onClose: () => void;
  onSave: (payload: AddBoulderPayload) => void;
}

export function AddBoulderModal({ root, initialPosition, initialParentId, fallbackParentId, saving = false, onClose, onSave }: Props) {
  const areas = useMemo(() => flattenAreas(root), [root]);
  const defaultParent = parentExists(areas, initialParentId) ? initialParentId! : parentExists(areas, fallbackParentId) ? fallbackParentId! : areas[0]?.feature.id ?? '';
  const [name, setName] = useState('');
  const [latitude, setLatitude] = useState(initialPosition ? formatCoord(initialPosition[1]) : '');
  const [longitude, setLongitude] = useState(initialPosition ? formatCoord(initialPosition[0]) : '');
  const [parentId, setParentId] = useState(defaultParent);
  const [devStatus, setDevStatus] = useState<MoneyDevStatus>('scouted');

  useEffect(() => {
    setLatitude(initialPosition ? formatCoord(initialPosition[1]) : '');
    setLongitude(initialPosition ? formatCoord(initialPosition[0]) : '');
  }, [initialPosition]);

  useEffect(() => {
    setParentId(defaultParent);
  }, [defaultParent]);

  const lat = Number(latitude);
  const lon = Number(longitude);
  const invalidLat = latitude.trim() !== '' && !validLatitude(lat);
  const invalidLon = longitude.trim() !== '' && !validLongitude(lon);
  const canSave = Boolean(name.trim() && parentId && validLatitude(lat) && validLongitude(lon) && !saving);
  const submit = () => {
    if (!canSave) return;
    onSave({ name: name.trim(), latitude: lat, longitude: lon, position: [lon, lat], parentId, devStatus });
  };

  return <div role="dialog" aria-modal="true" aria-label="Add boulder" onClick={onClose} style={{ position: 'fixed', inset: 0, zIndex: 70, background: 'rgba(8,5,4,0.62)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 16 }}>
    <div onClick={event => event.stopPropagation()} style={{ width: 460, maxWidth: '100%', background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 16, boxShadow: T.shadow, padding: 22 }}>
      <div style={{ fontSize: 18, fontWeight: 800, marginBottom: 4 }}>Add boulder</div>
      <div style={{ color: T.mut, fontSize: 13, marginBottom: 16 }}>Create a point boulder and place it under an area or sub-area.</div>
      <div style={{ display: 'grid', gap: 12 }}>
        <label style={label}>Name<input aria-label="Boulder name" autoFocus value={name} onChange={event => setName(event.target.value)} onKeyDown={event => { if (event.key === 'Enter') submit(); }} style={field} /></label>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
          <label style={label}>Latitude<input aria-label="Boulder latitude" inputMode="decimal" value={latitude} onChange={event => setLatitude(event.target.value)} style={{ ...field, borderColor: invalidLat ? '#B65B4D' : T.line2 }} /></label>
          <label style={label}>Longitude<input aria-label="Boulder longitude" inputMode="decimal" value={longitude} onChange={event => setLongitude(event.target.value)} style={{ ...field, borderColor: invalidLon ? '#B65B4D' : T.line2 }} /></label>
        </div>
        {(invalidLat || invalidLon) && <div role="alert" style={{ color: '#E6A299', fontSize: 12 }}>Latitude must be between -90 and 90; longitude must be between -180 and 180.</div>}
        <label style={label}>Parent area<select aria-label="Boulder parent area" value={parentId} onChange={event => setParentId(event.target.value)} style={field}>{areas.map(area => <option key={area.feature.id} value={area.feature.id}>{areaPath(root, area.feature.id)}</option>)}</select></label>
        <label style={label}>Development status<select aria-label="Boulder development status" value={devStatus} onChange={event => setDevStatus(event.target.value as MoneyDevStatus)} style={field}>{DEV.order.map(status => <option key={status} value={status}>{DEV.meta[status].label}</option>)}</select></label>
      </div>
      <div style={{ display: 'flex', gap: 8, marginTop: 18 }}>
        <button type="button" disabled={!canSave} onClick={submit} style={{ flex: 1, border: 'none', borderRadius: 10, padding: 13, background: canSave ? T.accent : T.line2, color: canSave ? T.onAccent : T.faint, fontWeight: 800, cursor: canSave ? 'pointer' : 'default' }}>{saving ? 'Creating…' : 'Create boulder'}</button>
        <button type="button" onClick={onClose} style={{ border: `1px solid ${T.line2}`, borderRadius: 10, padding: '13px 18px', background: 'transparent', color: T.ink, fontWeight: 700, cursor: 'pointer' }}>Cancel</button>
      </div>
    </div>
  </div>;
}

function parentExists(areas: MoneyCragNode[], id?: string | null): boolean {
  return Boolean(id && areas.some(area => area.feature.id === id));
}

function areaPath(root: MoneyCragNode | null, id: string): string {
  const parts: string[] = [];
  const walk = (node: MoneyCragNode, trail: string[]): boolean => {
    const next = [...trail, node.feature.title];
    if (node.feature.id === id) {
      parts.push(...next);
      return true;
    }
    return (node.children ?? []).some(child => walk(child, next));
  };
  if (root) walk(root, []);
  return parts.join(' / ') || id;
}

function formatCoord(value: number): string {
  return Number.isFinite(value) ? String(Number(value.toFixed(7))) : '';
}

const label: React.CSSProperties = { display: 'grid', gap: 5, color: T.mut, fontFamily: T.mono, fontSize: 10.5, textTransform: 'uppercase', letterSpacing: 0.6 };
const field: React.CSSProperties = { width: '100%', background: T.surf2, border: `1px solid ${T.line2}`, borderRadius: 9, padding: '10px 11px', color: T.ink, fontFamily: T.font, fontSize: 13, outline: 'none', minHeight: 40, textTransform: 'none', letterSpacing: 0 };
