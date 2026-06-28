import { useEffect, useState } from 'react';
import { Loader2, Trash2, X } from 'lucide-react';
import { moneyApi } from '../../../services/money';
import { MoneyUpload } from '../../../types/money';
import { T } from './theme';

export interface PhotoLightboxItem {
  upload: MoneyUpload;
  title?: string;
  contextLabel?: string;
  noteBody?: string;
}

export function PhotoLightbox({ item, canDelete = false, onDelete, onClose }: { item: PhotoLightboxItem; canDelete?: boolean; onDelete?: (upload: MoneyUpload) => void; onClose: () => void }) {
  const { src, loading } = useUploadImageUrl(item.upload.id);
  const title = item.title || item.upload.original_filename;
  const meta = photoMeta(item.upload);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [onClose]);

  return <div role="presentation" onClick={onClose} style={{ position: 'fixed', inset: 0, zIndex: 90, background: 'rgba(8,5,4,0.82)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 18 }}>
    <div role="dialog" aria-modal="true" aria-label={title} onClick={event => event.stopPropagation()} style={{ width: 'min(1120px, 100%)', maxHeight: 'min(88vh, 900px)', display: 'grid', gridTemplateRows: 'auto minmax(0,1fr)', background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 16, boxShadow: T.shadow, overflow: 'hidden' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 14px', borderBottom: `1px solid ${T.line}` }}>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ fontSize: 15, fontWeight: 800, color: T.ink, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{title}</div>
          {item.contextLabel && <div style={{ fontFamily: T.mono, fontSize: 11, color: T.mut, marginTop: 3 }}>{item.contextLabel}</div>}
        </div>
        {canDelete && <button type="button" aria-label="Delete photo" onClick={() => { if (window.confirm('Delete this photo?')) onDelete?.(item.upload); }} style={{ border: `1px solid #8F4E45`, background: 'rgba(143,78,69,0.14)', color: '#E6A299', borderRadius: 9, height: 38, padding: '0 12px', display: 'flex', alignItems: 'center', gap: 7, cursor: 'pointer', fontWeight: 800 }}><Trash2 size={16} />Delete</button>}
        <button type="button" aria-label="Close photo" onClick={onClose} style={{ border: `1px solid ${T.line2}`, background: T.inset, color: T.ink, borderRadius: 9, width: 38, height: 38, display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}><X size={18} /></button>
      </div>
      <div style={{ minHeight: 0, display: 'grid', gridTemplateColumns: 'minmax(0,1fr) minmax(240px,320px)' }}>
        <div style={{ minHeight: 0, background: T.inset, display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 14 }}>
          {loading || !src ? <Loader2 size={28} color={T.accent} /> : <img src={src} alt={title} style={{ maxWidth: '100%', maxHeight: 'calc(88vh - 92px)', objectFit: 'contain', borderRadius: 10 }} />}
        </div>
        <aside style={{ borderLeft: `1px solid ${T.line}`, padding: 16, overflowY: 'auto' }}>
          <InfoLabel>Photo</InfoLabel>
          <div style={{ fontSize: 14, fontWeight: 800, color: T.ink, overflowWrap: 'anywhere', marginBottom: 12 }}>{title}</div>
          {item.contextLabel && <><InfoLabel>From</InfoLabel><div style={{ fontSize: 13, color: '#D9CBBD', lineHeight: 1.45, marginBottom: 12 }}>{item.contextLabel}</div></>}
          {item.noteBody && <><InfoLabel>Associated note</InfoLabel><p style={{ margin: '0 0 12px', fontSize: 13, lineHeight: 1.55, color: '#D9CBBD', whiteSpace: 'pre-wrap' }}>{item.noteBody}</p></>}
          {meta.length > 0 && <><InfoLabel>Metadata</InfoLabel><div style={{ display: 'grid', gap: 6 }}>{meta.map(row => <div key={row} style={{ fontFamily: T.mono, fontSize: 11, color: T.mut }}>{row}</div>)}</div></>}
        </aside>
      </div>
    </div>
  </div>;
}

export function UploadPhotoButton({ upload, ratio = '1 / 1', title, onOpen, canDelete = false, onDelete }: { upload: MoneyUpload; ratio?: string; title?: string; onOpen: () => void; canDelete?: boolean; onDelete?: (upload: MoneyUpload) => void }) {
  const { src, loading } = useUploadImageUrl(upload.id);
  const label = title || upload.original_filename;
  const baseStyle: React.CSSProperties = { width: '100%', aspectRatio: ratio, borderRadius: 8, border: `1px solid ${T.line}`, overflow: 'hidden' };

  if (loading || !src) return <div style={{ ...baseStyle, background: T.map.slot, display: 'flex', alignItems: 'center', justifyContent: 'center' }}><Loader2 size={16} color={T.accent} /></div>;

  return <div style={{ position: 'relative' }}>
    <button type="button" aria-label={`Open photo ${label}`} onClick={onOpen} style={{ ...baseStyle, display: 'block', padding: 0, background: 'transparent', cursor: 'zoom-in' }}>
      <img src={src} alt={label} style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }} />
    </button>
    {canDelete && <button type="button" aria-label={`Delete photo ${label}`} onClick={event => { event.stopPropagation(); if (window.confirm('Delete this photo?')) onDelete?.(upload); }} style={{ position: 'absolute', top: 6, right: 6, width: 28, height: 28, borderRadius: 8, border: `1px solid #8F4E45`, background: 'rgba(20,10,8,0.82)', color: '#E6A299', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}><Trash2 size={14} /></button>}
  </div>;
}

export function useUploadImageUrl(uploadId: string) {
  const [src, setSrc] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let objectUrl: string | null = null;
    let cancelled = false;
    setSrc(null);
    if (!uploadId) {
      setLoading(false);
      return () => {
        cancelled = true;
      };
    }
    setLoading(true);
    moneyApi.getUploadBlobUrl(uploadId).then(url => {
      if (cancelled) {
        if (url.startsWith('blob:')) URL.revokeObjectURL(url);
        return;
      }
      objectUrl = url.startsWith('blob:') ? url : null;
      setSrc(url);
    }).catch(() => {
      if (!cancelled) setSrc(null);
    }).finally(() => {
      if (!cancelled) setLoading(false);
    });
    return () => {
      cancelled = true;
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [uploadId]);

  return { src, loading };
}

function InfoLabel({ children }: { children: React.ReactNode }) {
  return <div style={{ fontFamily: T.mono, fontSize: 10.5, letterSpacing: 0.8, color: T.mut, textTransform: 'uppercase', margin: '0 0 6px' }}>{children}</div>;
}

function photoMeta(upload: MoneyUpload): string[] {
  const rows: string[] = [];
  if (upload.content_type) rows.push(upload.content_type);
  if (upload.width && upload.height) rows.push(`${upload.width} × ${upload.height}`);
  if (upload.created_at) rows.push(`Uploaded ${new Date(upload.created_at).toLocaleString()}`);
  return rows;
}
